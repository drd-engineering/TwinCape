package register

import (
	"errors"
	"math/rand"
	"net/http"
	"time"

	"github.com/drd-engineering/TwinCape/db"
	"github.com/drd-engineering/TwinCape/environments"
	"github.com/gin-gonic/gin"
	"github.com/jinzhu/gorm"

	"golang.org/x/crypto/bcrypt"
)

// SaveUser service handler for user registration
func SaveUser(c *gin.Context) {
	var input UserRegistrationData
	c.ShouldBindJSON(&input)

	var dbInstance = db.GetDb()

	strChan := make(chan string)
	errChan := make(chan error)

	go createPassword(8, strChan)

	validationMessage, isValid := isDataRegistrationValid(input)
	if !isValid {
		c.Abort()
		c.JSON(http.StatusBadRequest, gin.H{"message": validationMessage})
		return
	}
	message, isExist := isUserExist(input, dbInstance)
	if isExist {
		c.Abort()
		c.JSON(http.StatusBadRequest, gin.H{"message": message})
		return
	}

	go createUniqueID(dbInstance, strChan, errChan)

	var userBirthDate time.Time
	var err error
	if len(input.DateOfBirth) != 0 {
		userBirthDate, err = time.Parse("2006-01-02", input.DateOfBirth)
	}
	if err != nil {
		c.Abort()
		c.JSON(http.StatusBadRequest,
			gin.H{"message": "Date of birth format: (YYYY-MM-DD"})
		return
	}
	passwordUser := <-strChan
	go secureUserPassword(passwordUser, strChan, errChan)

	storedID := <-strChan
	err = <-errChan
	if err != nil {
		c.Abort()
		c.JSON(http.StatusInternalServerError,
			gin.H{"message": err.Error()})
		return
	}

	storedPassword := <-strChan
	err = <-errChan
	if err != nil {
		c.Abort()
		c.JSON(http.StatusInternalServerError,
			gin.H{"message": "Failed process when hashing user password"})
		return
	}

	var userDb = db.User{
		ID:           storedID,
		Name:         input.Name,
		Gender:       input.Gender,
		Email:        input.Email,
		KtpNumber:    input.KtpNumber,
		Address:      input.Address,
		PhoneNumber:  input.PhoneNumber,
		Password:     storedPassword,
		DateOfBirth:  userBirthDate,
		Cityzenship:  input.Cityzenship,
		PlaceOfBirth: input.PlaceOfBirth,
	}
	dbInstance.Create(&userDb)
	responseSaveUser := ResponseSaveUser{}
	responseSaveUser = responseSaveUser.CreateResponse(userDb)
	responseSaveUser.Password = passwordUser
	c.JSON(http.StatusOK, gin.H{"user": responseSaveUser, "message": "User saved"})
}

func isUserExist(user UserRegistrationData, dbInstance *gorm.DB) (string, bool) {
	var existingUserCount int
	dbInstance.Model(&db.User{}).Where("ktp_number = ?", user.KtpNumber).Count(&existingUserCount)
	if existingUserCount > 0 {
		return "User with same KTP number already exists", true
	}
	dbInstance.Model(&db.User{}).Where("email = ?", user.Email).Count(&existingUserCount)
	if existingUserCount > 0 {
		return "User with same email already exists", true
	}
	dbInstance.Model(&db.User{}).Where("phone_number = ?", user.PhoneNumber).Count(&existingUserCount)
	if existingUserCount > 0 {
		return "User with same phone number already exists", true
	}
	return "", false
}

func isDataRegistrationValid(user UserRegistrationData) (string, bool) {
	if user.KtpNumber < 1 {
		return "User KTP number must not be empty", false
	}
	if len(user.Email) == 0 {
		return "User email must not be empty", false
	}
	if len(user.PhoneNumber) == 0 {
		return "User phone number must not be empty", false
	}
	return "", true
}

func createPassword(passwordLength int, c chan string) {
	environments.LoadEnvironmentVariableFile()
	passwordBaseString := environments.Get("PASSWORD_BASE_STRING")
	byteResult := make([]byte, passwordLength)
	for i := range byteResult {
		byteResult[i] = passwordBaseString[rand.Int63()%int64(len(passwordBaseString))]
	}
	c <- string(byteResult)
}

func secureUserPassword(password string, c chan string, r chan error) {
	// Use GenerateFromPassword to hash & salt password
	hashValue, err := bcrypt.GenerateFromPassword([]byte(password), 6)
	if err != nil {
		c <- ""
		r <- err
		return
	}
	c <- string(hashValue)
	r <- nil
}

func createUniqueID(dbInstance *gorm.DB, c chan string, r chan error) {
	environments.LoadEnvironmentVariableFile()
	idBaseString := environments.Get("ID_BASE_STRING")
	// max 3 times trial
	for i := 0; i < 3; i++ {
		byteResult := make([]byte, 6)
		for i := range byteResult {
			byteResult[i] = idBaseString[rand.Int63()%int64(len(idBaseString))]
		}
		result := "DRD-" + string(byteResult)
		var existingUserCount int
		dbInstance.Model(&db.User{}).Where("id = ?", result).Count(&existingUserCount)
		if existingUserCount < 1 {
			c <- result
			r <- nil
			return
		}
	}
	// if fail all trial
	c <- ""
	r <- errors.New("UniqueID: failed to generate unique ID")
}
