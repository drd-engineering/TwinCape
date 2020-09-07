package routes_test

import (
	"bytes"
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
	"github.com/drd-engineering/TwinCape/routes"
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
func setupTestCase(t *testing.T) func(t *testing.T) {
	environments.Set("RELEASE_TYPE", "localhost")
	_, filename, _, _ := runtime.Caller(0)
	dir := path.Join(path.Dir(filename), "..")
	err := os.Chdir(dir)
	if err != nil {
		panic(err)
	}
	environments.LoadEnvironmentVariableFile()
	db.InitPostgre(dbTestConfig())
	return func(t *testing.T) {
		os.Clearenv()
		dbInstance := db.GetDb()
		dbInstance.DropTable(db.User{}, db.APILog{})
	}
}
func mockHandler(c *gin.Context) {
	c.JSON(200, gin.H{})
}

// authorization testing
func TestDRDApplicationIdentification(t *testing.T) {
	tests := []struct {
		name      string
		useHeader bool
		input     string
		code      int
		body      []string
	}{
		{
			name:      "OK",
			useHeader: true,
			input:     "DRD_IDENTIFICATION",
			code:      200,
		},
		{
			name:      "FailedNoHeaderFound",
			useHeader: false,
			input:     "",
			code:      401,
		},
		{
			name:      "FailedHeaderIsWrong",
			useHeader: true,
			input:     "USERNAME_DB",
			code:      401,
		},
	}
	set := setupTestCase(t)
	defer set(t)
	r := gin.Default()
	test := r.Group("/t")
	{
		test.Use(routes.DRDApplicationIdentification())
		test.POST("/testanycall", mockHandler)
	}
	for _, tc := range tests {
		w := httptest.NewRecorder()
		req, _ := http.NewRequest("POST", "/t/testanycall", bytes.NewBuffer([]byte{}))
		if tc.useHeader {
			req.Header.Set("Drd-Identification", environments.Get(tc.input))
		}
		r.ServeHTTP(w, req)

		assert.Equal(t, tc.code, w.Code, "test "+tc.name+" case")
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
	accessTokenClaims := jwt.StandardClaims{
		Audience:  jwtDetails.userID,
		ExpiresAt: jwtDetails.expiredAt,
		Issuer:    jwtDetails.issuer,
		IssuedAt:  time.Now().Unix(),
	}
	accessToken := jwt.NewWithClaims(jwtDetails.signMethod, accessTokenClaims)
	signAccessToken, _ := accessToken.SignedString([]byte(environments.Get(jwtDetails.signString)))
	tokenDetails.AccessToken = signAccessToken

	return tokenDetails
}

func TestAuthorizationBearer(t *testing.T) {
	tests := []struct {
		name  string
		input jwtTestDetails
		code  int
	}{
		{
			name: "Success",
			input: jwtTestDetails{create: true, userID: "testid", expiredAt: time.Now().Add(time.Minute * 3).Unix(),
				signMethod: jwt.SigningMethodHS256, signString: "ACCESS_SECRET_KEY", issuer: "SSO_TWINCAPE"},
			code: 200,
		},
		{
			name: "FailedNoAuthSendInHeader",
			input: jwtTestDetails{create: false, userID: "testid", expiredAt: time.Now().Add(time.Minute * 3).Unix(),
				signMethod: jwt.SigningMethodHS256, signString: "ACCESS_SECRET_KEY", issuer: "SSO_TWINCAPE"},
			code: 401,
		},
		{
			name: "FailedExpired",
			input: jwtTestDetails{create: true, userID: "testid", expiredAt: time.Now().Add(time.Minute * -3).Unix(),
				signMethod: jwt.SigningMethodHS256, signString: "ACCESS_SECRET_KEY", issuer: "SSO_TWINCAPE"},
			code: 401,
		},
		{
			name: "FailedNotIssuedbyDRD",
			input: jwtTestDetails{create: true, userID: "testid", expiredAt: time.Now().Add(time.Minute * 3).Unix(),
				signMethod: jwt.SigningMethodHS256, signString: "ACCESS_SECRET_KEY", issuer: "UNKNOWN"},
			code: 401,
		},
	}
	set := setupTestCase(t)
	defer set(t)
	r := gin.Default()
	test := r.Group("/t")
	{
		test.Use(routes.AuthorizationBearer())
		test.POST("/testanycall", mockHandler)
	}
	for _, tc := range tests {
		w := httptest.NewRecorder()
		req, _ := http.NewRequest("POST", "/t/testanycall", bytes.NewBuffer([]byte{}))
		mockToken := mockJWTCreation(&tc.input)
		authString := "Bearer"
		if tc.input.create {
			authString += " " + mockToken.AccessToken
		}
		req.Header.Set("Authorization", authString)
		r.ServeHTTP(w, req)

		assert.Equal(t, tc.code, w.Code, "test "+tc.name+" case")
	}
}

// main routes testing
func TestGetInstance(t *testing.T) {
	routeInstance := routes.GetInstance()
	assert.IsType(t, &gin.Engine{}, routeInstance, "Should return the gin engine")
}

func TestMiddlewareWorksFine(t *testing.T) {
	set := setupTestCase(t)
	defer set(t)
	routeInstance := routes.GetInstance()
	routeInstance.POST("/t/testrun", mockHandler)
	w := httptest.NewRecorder()
	req, _ := http.NewRequest("POST", "/t/testrun", bytes.NewBuffer([]byte{}))
	routeInstance.ServeHTTP(w, req)

	assert.Equal(t, 200, w.Code, "should response with code 200")
}
