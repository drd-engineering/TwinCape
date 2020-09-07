package db

import (
	"fmt"
	"sync"

	"github.com/jinzhu/gorm"

	_ "github.com/jinzhu/gorm/dialects/postgres"
)

var db *gorm.DB //database
var once sync.Once

// Config is db Config for db initiation
type Config struct {
	Host     string
	Username string
	DBName   string
	Password string
}

// InitPostgre cocnnection start
func InitPostgre(config *Config) error {
	dbURI := fmt.Sprintf("host=%s user=%s dbname=%s sslmode=disable password=%s", config.Host, config.Username, config.DBName, config.Password) //Build connection string
	conn, err := gorm.Open("postgres", dbURI)
	if err != nil {
		return err
	}
	db = conn
	db.Debug().AutoMigrate(&User{}, &APILog{})
	return nil
}

// GetDb function for getting the singleton Db
func GetDb() *gorm.DB {
	return db
}
