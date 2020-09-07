package authenticator_test

import (
	"bytes"
	"encoding/json"
	"golang.org/x/crypto/bcrypt"
	"net/http"
	"net/http/httptest"
	"os"
	"path"
	"runtime"
	"testing"
	"time"

	"github.com/dgrijalva/jwt-go"
	"github.com/drd-engineering/TwinCape/db"
	"github.com/drd-engineering/TwinCape/domains/authenticator"
	"github.com/drd-engineering/TwinCape/environments"
	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
)

func dbTestConfig() *db.Config {
	return &db.Config{
		Host:     environments.Get("HOST_DB"),
		Username: environments.Get("USERNAME_DB"),
		DBName:   environments.Get("DB_NAME"),
		Password: environments.Get("PASSWORD_DB"),
	}
}
func getUserLoginTest() db.User {
	return db.User{
		ID:          "testid",
		Name:        "test",
		Email:       "test@test.com",
		KtpNumber:   10201021020102,
		PhoneNumber: "+6200000000000",
	}
}
func secureUserPassword(password string) string {
	// Use GenerateFromPassword to hash & salt password
	hashValue, _ := bcrypt.GenerateFromPassword([]byte(password), 6)
	return string(hashValue)
}
func setupTestCase(t *testing.T) func(t *testing.T) {
	environments.Set("RELEASE_TYPE", "localhost")
	_, filename, _, _ := runtime.Caller(0)
	dir := path.Join(path.Dir(filename), "../..")
	err := os.Chdir(dir)
	if err != nil {
		panic(err)
	}
	environments.LoadEnvironmentVariableFile()
	db.InitPostgre(dbTestConfig())
	dbInstance := db.GetDb()
	testUser := getUserLoginTest()
	testUser.Password = secureUserPassword("testing")
	dbInstance.Create(&testUser)
	return func(t *testing.T) {
		os.Clearenv()
		dbInstance := db.GetDb()
		dbInstance.DropTable(db.User{}, db.APILog{})
	}
}

func TestLogin(t *testing.T) {
	tests := []struct {
		name  string
		input []byte
		code  int
		body  []string
	}{
		{
			name:  "SuccessLoginUseId",
			input: []byte(`{"id":"testid", "password":"testing"}`),
			code:  200,
			body:  []string{"accessToken", "refreshToken"},
		},
		{
			name:  "SuccessLoginEmail",
			input: []byte(`{"email":"test@test.com", "password":"testing"}`),
			code:  200,
			body:  []string{"accessToken", "refreshToken"},
		},
		{
			name:  "FailedLoginDetailsNotMatch",
			input: []byte(`{"id":"testid", "password":"tesing"}`),
			code:  401,
			body:  []string{"message"},
		},
		{
			name:  "FailedLoginNoUser",
			input: []byte(`{"id":"testid2", "password":"testing"}`),
			code:  401,
			body:  []string{"message"},
		},
		{
			name:  "FailedLoginNoJSONRequestBody",
			input: []byte(`{"password":"testing"}`),
			code:  401,
			body:  []string{"message"},
		},
	}
	r := gin.Default()
	r.POST("/t/login", authenticator.Login)

	set := setupTestCase(t)
	defer set(t)
	for _, tc := range tests {
		w := httptest.NewRecorder()
		req, _ := http.NewRequest("POST", "/t/login", bytes.NewBuffer(tc.input))
		r.ServeHTTP(w, req)

		assert.Equal(t, tc.code, w.Code, "test "+tc.name+" case")
		var got gin.H
		err := json.Unmarshal(w.Body.Bytes(), &got)
		if err != nil {
			t.Fatal(err)
		}
		for _, keyBody := range tc.body {
			val, ok := got[keyBody]
			assert.True(t, ok, "The return body should contain "+keyBody+" as json data in test "+tc.name+" case")
			assert.NotEmpty(t, val, "the value of json data should not be empty in test "+tc.name+" case")
		}
	}
}

func mockMiddleware(expectedUser interface{}) gin.HandlerFunc {
	return func(c *gin.Context) {
		c.Set("userID", expectedUser)
		c.Next()
	}
}
func TestCheckToken(t *testing.T) {
	tests := []struct {
		name  string
		input []byte
		code  int
		body  []string
	}{
		{
			name:  "Success",
			input: []byte(`{}`),
			code:  200,
			body:  []string{"userID", "message"},
		},
	}
	r := gin.Default()
	r.Use(mockMiddleware("testid"))
	r.POST("/t/check-token", authenticator.CheckToken)

	set := setupTestCase(t)
	defer set(t)
	for _, tc := range tests {
		w := httptest.NewRecorder()
		req, _ := http.NewRequest("POST", "/t/check-token", bytes.NewBuffer(tc.input))
		r.ServeHTTP(w, req)

		assert.Equal(t, tc.code, w.Code, "test "+tc.name+" case")
		var got gin.H
		err := json.Unmarshal(w.Body.Bytes(), &got)
		if err != nil {
			t.Fatal(err)
		}
		for _, keyBody := range tc.body {
			val, ok := got[keyBody]
			assert.True(t, ok, "The return body should contain "+keyBody+" as json data in test "+tc.name+" case")
			assert.NotEmpty(t, val, "the value of json data should not be empty in test "+tc.name+" case")
		}
	}
}

func TestGetLoginDetails(t *testing.T) {
	tests := []struct {
		name  string
		input interface{}
		code  int
		body  []string
	}{
		{
			name:  "Success",
			input: "testid",
			code:  200,
			body:  []string{"user", "message"},
		},
		{
			name:  "Failed",
			input: "test",
			code:  401,
			body:  []string{"message"},
		},
		{
			name:  "Success but error middleware",
			input: 12,
			code:  500,
			body:  []string{"message"},
		},
	}
	set := setupTestCase(t)
	defer set(t)
	for _, tc := range tests {
		r := gin.Default()
		r.Use(mockMiddleware(tc.input))
		r.POST("/t/get-login-details", authenticator.GetLoginDetails)
		w := httptest.NewRecorder()
		req, _ := http.NewRequest("POST", "/t/get-login-details", bytes.NewBuffer([]byte{}))
		r.ServeHTTP(w, req)

		assert.Equal(t, tc.code, w.Code, "test "+tc.name+" case")
		var got gin.H
		err := json.Unmarshal(w.Body.Bytes(), &got)
		if err != nil {
			t.Fatal(err)
		}
		for _, keyBody := range tc.body {
			val, ok := got[keyBody]
			assert.True(t, ok, "The return body should contain "+keyBody+" as json data in test "+tc.name+" case")
			assert.NotEmpty(t, val, "the value of json data should not be empty in test "+tc.name+" case")
		}
	}
}

type jwtTestDetails struct {
	create     bool
	signMethod jwt.SigningMethod
	signString string
	issuer     string
	userID     string
	expiredAt  int64
}

func mockJWTCreation(jwtDetails *jwtTestDetails) authenticator.TokenDetails {
	if !jwtDetails.create {
		return authenticator.TokenDetails{}
	}
	tokenDetails := authenticator.TokenDetails{}
	refreshTokenClaims := jwt.StandardClaims{
		Audience:  jwtDetails.userID,
		ExpiresAt: jwtDetails.expiredAt,
		Issuer:    jwtDetails.issuer,
		IssuedAt:  time.Now().Unix(),
	}
	refreshToken := jwt.NewWithClaims(jwtDetails.signMethod, refreshTokenClaims)
	signRefreshToken, err := refreshToken.SignedString([]byte(environments.Get(jwtDetails.signString)))
	if err != nil {
		return authenticator.TokenDetails{}
	}
	tokenDetails.RefreshToken = signRefreshToken

	return tokenDetails
}

func TestRefreshToken(t *testing.T) {
	tests := []struct {
		name  string
		input jwtTestDetails
		code  int
		body  []string
	}{
		{
			name: "Success",
			input: jwtTestDetails{create: true, userID: "testid", expiredAt: time.Now().Add(time.Minute * 3).Unix(),
				signMethod: jwt.SigningMethodHS256, signString: "REFRESH_SECRET_KEY", issuer: "SSO_TWINCAPE"},
			code: 200,
			body: []string{"accessToken", "refreshToken"},
		},
		{
			name: "FailedEXPDateShowTokenExpired",
			input: jwtTestDetails{create: true, userID: "testid", expiredAt: time.Now().Add(time.Minute * -3).Unix(),
				signMethod: jwt.SigningMethodHS256, signString: "REFRESH_SECRET_KEY", issuer: "SSO_TWINCAPE"},
			code: 400,
			body: []string{"message"},
		},
		{
			name: "FailedNoRefreshTokenInJSONBody",
			input: jwtTestDetails{create: false, userID: "testid", expiredAt: time.Now().Unix(),
				signMethod: jwt.SigningMethodHS256, signString: "REFRESH_SECRET_KEY", issuer: "SSO_TWINCAPE"},
			code: 400,
			body: []string{"message"},
		},
		{
			name: "FailedDifferentSignMethod",
			input: jwtTestDetails{create: true, userID: "testid", expiredAt: time.Now().Unix(),
				signMethod: jwt.SigningMethodRS256, signString: "REFRESH_SECRET_KEY", issuer: "SSO_TWINCAPE"},
			code: 400,
			body: []string{"message"},
		},
		{
			name: "FailedDifferentSignString",
			input: jwtTestDetails{create: true, userID: "testid", expiredAt: time.Now().Unix(),
				signMethod: jwt.SigningMethodHS256, signString: "ACCESS_SECRET_KEY", issuer: "SSO_TWINCAPE"},
			code: 400,
			body: []string{"message"},
		},
		{
			name: "FailedIssuerJWTtokenIsNotSSO",
			input: jwtTestDetails{create: true, userID: "testid", expiredAt: time.Now().Add(time.Minute * 3).Unix(),
				signMethod: jwt.SigningMethodHS256, signString: "REFRESH_SECRET_KEY", issuer: "DRD"},
			code: 400,
			body: []string{"message"},
		},
	}
	set := setupTestCase(t)
	defer set(t)
	for _, tc := range tests {
		r := gin.Default()
		r.Use(mockMiddleware(tc.input.userID))
		r.POST("/t/refresh-token", authenticator.RefreshToken)
		w := httptest.NewRecorder()
		mockToken := mockJWTCreation(&tc.input)
		jsonData := []byte(`{"refreshToken":"` + mockToken.RefreshToken + `"}`)
		req, _ := http.NewRequest("POST", "/t/refresh-token", bytes.NewBuffer(jsonData))
		r.ServeHTTP(w, req)

		assert.Equal(t, tc.code, w.Code, "test "+tc.name+" case")
		var got gin.H
		err := json.Unmarshal(w.Body.Bytes(), &got)
		if err != nil {
			t.Fatal(err)
		}
		for _, keyBody := range tc.body {
			val, ok := got[keyBody]
			assert.True(t, ok, "The return body should contain "+keyBody+" as json data in test "+tc.name+" case")
			assert.NotEmpty(t, val, "the value of json data should not be empty in test "+tc.name+" case")
		}
	}
}
