package authenticator

import (
	"fmt"
	"net/http"
	"time"

	"github.com/dgrijalva/jwt-go"
	"github.com/drd-engineering/TwinCape/db"
	"github.com/drd-engineering/TwinCape/environments"
	"github.com/gin-gonic/gin"

	"golang.org/x/crypto/bcrypt"
)

// Login service handler for client to login
func Login(c *gin.Context) {
	var input UserLogin
	c.ShouldBindJSON(&input)

	var dbInstance = db.GetDb()
	var userInDb db.User
	if len(input.ID) > 0 {
		dbInstance.Where("id = ?", input.ID).First(&userInDb)
	} else if len(input.Email) > 0 {
		dbInstance.Where("email = ?", input.Email).First(&userInDb)
	} else {
		c.Abort()
		c.JSON(http.StatusUnauthorized,
			gin.H{"message": "Please provide valid login details"})
		return
	}
	if len(userInDb.ID) == 0 {
		c.Abort()
		c.JSON(http.StatusUnauthorized,
			gin.H{"message": "Please provide valid login details"})
		return
	}

	if result := comparePasswords(userInDb.Password, input.Password); !result {
		c.Abort()
		c.JSON(http.StatusUnauthorized,
			gin.H{"message": "Please provide valid login details"})
		return
	}
	token, err := createToken(userInDb.ID)
	if err != nil {
		c.Abort()
		c.JSON(http.StatusInternalServerError,
			gin.H{"message": "Error when creating token"})
		return
	}

	c.JSON(http.StatusOK, token)
}

func createToken(userID string) (TokenDetails, error) {
	tokenDetails := TokenDetails{}

	accessTokenClaims := jwt.StandardClaims{
		Audience:  userID,
		ExpiresAt: time.Now().Add(time.Minute * 30).Unix(),
		IssuedAt:  time.Now().Unix(),
		Issuer:    "SSO_TWINCAPE",
		Subject:   "SSO_ACCESS",
	}
	accessToken := jwt.NewWithClaims(jwt.SigningMethodHS256, accessTokenClaims)
	signAccessToken, err := accessToken.SignedString([]byte(environments.Get("ACCESS_SECRET_KEY")))
	if err != nil {
		return TokenDetails{}, err
	}
	tokenDetails.AccessToken = signAccessToken

	refreshTokenClaims := jwt.StandardClaims{
		Audience:  userID,
		ExpiresAt: time.Now().Add(time.Hour * 24 * 7).Unix(),
		IssuedAt:  time.Now().Unix(),
		Issuer:    "SSO_TWINCAPE",
		Subject:   "SSO_REFRESH",
	}
	refreshToken := jwt.NewWithClaims(jwt.SigningMethodHS256, refreshTokenClaims)
	signRefreshToken, err := refreshToken.SignedString([]byte(environments.Get("REFRESH_SECRET_KEY")))
	if err != nil {
		return TokenDetails{}, err
	}
	tokenDetails.RefreshToken = signRefreshToken

	return tokenDetails, nil
}

func comparePasswords(hashedPassword string, plainPassword string) bool {
	// Since we'll be getting the hashed password from the DB it
	// will be a string so we'll need to convert it to a byte slice
	byteHashedPassword := []byte(hashedPassword)
	bytePlainPassword := []byte(plainPassword)
	err := bcrypt.CompareHashAndPassword(byteHashedPassword, bytePlainPassword)
	if err != nil {
		return false
	}
	return true
}

// CheckToken service handler to check if the token given is valid
func CheckToken(c *gin.Context) {
	c.JSON(http.StatusOK, gin.H{"userID": c.MustGet("userID"), "message": "You are authorized"})
}

// GetLoginDetails service handler to give client user logged in details
func GetLoginDetails(c *gin.Context) {
	userID, ok := c.MustGet("userID").(string)
	if !ok {
		c.Abort()
		c.JSON(http.StatusInternalServerError,
			gin.H{"message": "Failed to get user id from token payload"})
		return
	}
	userDb := db.User{}
	dbInstance := db.GetDb()

	dbInstance.Where(&db.User{ID: userID}).First(&userDb)

	if len(userDb.ID) == 0 {
		c.Abort()
		c.JSON(http.StatusUnauthorized,
			gin.H{"message": "Invalid user logged in"})
		return
	}

	response := ResponseLoginDetails{}
	response = response.CreateResponse(userDb)

	c.JSON(http.StatusOK,
		gin.H{"user": response, "message": "You are authorized"})
}

// RefreshToken service handler to create new token for user logged using refresh token
func RefreshToken(c *gin.Context) {
	var input RequestRefreshToken
	c.ShouldBindJSON(&input)

	if len(input.RefreshToken) == 0 {
		c.Abort()
		c.JSON(http.StatusBadRequest,
			gin.H{"message": "provide refresh token in body"})
		return
	}

	claims := jwt.StandardClaims{}
	_, err := jwt.ParseWithClaims(input.RefreshToken, &claims, func(token *jwt.Token) (interface{}, error) {
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, fmt.Errorf("unexpected signing method: %v", token.Header["alg"])
		}
		return []byte(environments.Get("REFRESH_SECRET_KEY")), nil
	})
	if err != nil {
		c.Abort()
		c.JSON(http.StatusBadRequest,
			gin.H{"message": err.Error()})
		return
	}
	if claims.Issuer != "SSO_TWINCAPE" {
		c.Abort()
		c.JSON(http.StatusBadRequest,
			gin.H{"message": "Invalid refresh token"})
		return
	}
	userID := claims.Audience
	newToken, err := createToken(userID)
	if err != nil {
		c.Abort()
		c.JSON(http.StatusInternalServerError,
			gin.H{"message": "Error when creating token"})
		return
	}

	c.JSON(http.StatusOK, newToken)
}
