package db_test

import (
	"os"
	"path"
	"runtime"
	"testing"

	"github.com/jinzhu/gorm"
	"github.com/stretchr/testify/assert"

	"github.com/drd-engineering/TwinCape/db"
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
	dir := path.Join(path.Dir(filename), "..")
	err := os.Chdir(dir)
	if err != nil {
		panic(err)
	}
	environments.LoadEnvironmentVariableFile()
	return func(t *testing.T) {
		os.Clearenv()
	}
}
func TestInitPostgreSuccess(t *testing.T) {
	set := setupTestCase(t)
	defer set(t)
	err := db.InitPostgre(dbTestConfig())
	assert.Nil(t, err, "Should return nil because there is no error")
}

func TestGetDbSuccess(t *testing.T) {
	set := setupTestCase(t)
	defer set(t)
	db.InitPostgre(dbTestConfig())
	assert.IsType(t, &gorm.DB{}, db.GetDb(),
		"Should return a Gorm DB pointers object")
}

func TestInitPostgreFailed(t *testing.T) {
	err := db.InitPostgre(dbTestConfig())
	assert.Error(t, err, "Should return an error")
}
