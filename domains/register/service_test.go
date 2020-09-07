package register_test

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"os"
	"path"
	"runtime"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"

	"github.com/drd-engineering/TwinCape/db"
	"github.com/drd-engineering/TwinCape/domains/register"
	"github.com/drd-engineering/TwinCape/environments"
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
	dir := path.Join(path.Dir(filename), "../..")
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
func TestSaveUser(t *testing.T) {
	tests := []struct {
		name  string
		input []byte
		code  int
		body  []string
	}{
		{
			name: "OK",
			input: []byte(`{"name":"test", "email":"test@test.com","ktpNumber":1111,
							"address":"jalan test","phoneNumber":"+6200000000000",
							"dateofBirth":"2000-12-22","cityzenship":"WNI","placeofBirth":"Jakarta"}`),
			code: 200,
			body: []string{"message", "user"},
		},
		{
			name: "FailNoKTPNumber",
			input: []byte(`{"name":"test", "email":"test@test.com",
							"address":"jalan test","phoneNumber":"+6200000000000",
							"dateofBirth":"2000-12-22","cityzenship":"WNI","placeofBirth":"Jakarta"}`),
			code: 400,
			body: []string{"message"},
		},
		{
			name: "FailNoPhoneNumber",
			input: []byte(`{"name":"test", "email":"test@test.com",
							"ktpNumber":1111,"address":"jalan test",
							"dateofBirth":"2000-12-22","cityzenship":"WNI","placeofBirth":"Jakarta"}`),
			code: 400,
			body: []string{"message"},
		},
		{
			name: "FailNoEmail",
			input: []byte(`{"name":"test", "ktpNumber":1111,
							"address":"jalan test","phoneNumber":"+6200000000000",
							"dateofBirth":"2000-12-22","cityzenship":"WNI","placeofBirth":"Jakarta"}`),
			code: 400,
			body: []string{"message"},
		},
		{
			name: "OKDataIncomplete",
			input: []byte(`{"name":"test", "email":"test2@test.com","ktpNumber":1112,
							"address":"jalan test","phoneNumber":"+6200000000001"}`),
			code: 200,
			body: []string{"message", "user"},
		},
		{
			name: "FailSameEmail",
			input: []byte(`{"name":"test", "email":"test@test.com","ktpNumber":1113,
							"address":"jalan test","phoneNumber":"+6200000000002"}`),
			code: 400,
			body: []string{"message"},
		},
		{
			name: "FailSameKTPNumber",
			input: []byte(`{"name":"test", "email":"test3@test.com","ktpNumber":1111,
							"address":"jalan test","phoneNumber":"+6200000000002"}`),
			code: 400,
			body: []string{"message"},
		},
		{
			name: "FailSamePhoneNumber",
			input: []byte(`{"name":"test", "email":"test3@test.com","ktpNumber":1113,
							"address":"jalan test","phoneNumber":"+6200000000000"}`),
			code: 400,
			body: []string{"message"},
		},
		{
			name: "FailedBecauseFalseBirthDateFormat",
			input: []byte(`{"name":"test", "email":"est@test.com","ktpNumber":11911,
							"address":"jalan test","phoneNumber":"+6200000090000",
							"dateofBirth":"2000-30-12","cityzenship":"WNI","placeofBirth":"Jakarta"}`),
			code: 400,
			body: []string{"message"},
		},
	}
	r := gin.Default()
	r.POST("/t/saveUser", register.SaveUser)

	set := setupTestCase(t)
	defer set(t)
	for _, tc := range tests {
		w := httptest.NewRecorder()
		req, _ := http.NewRequest("POST", "/t/saveUser", bytes.NewBuffer(tc.input))
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
