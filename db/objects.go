package db

import (
	"time"
)

// User is db definition of a user in SSO System
type User struct {
	ID           string
	CreatedAt    time.Time
	UpdatedAt    time.Time
	DeletedAt    *time.Time
	Name         string
	Gender       string
	Email        string
	KtpNumber    int64 `gorm:"unique;not null"`
	Address      string
	PhoneNumber  string
	Password     string
	DateOfBirth  time.Time
	Cityzenship  string
	PlaceOfBirth string
}

// APILog is db definition of a Log of API service consume
type APILog struct {
	ID             int `gorm:"primary_key"`
	Timestamp      time.Time
	TTL            string
	ResponseStatus int
	Path           string
	Method         string
	ClientIP       string
	ClientTools    string
	Protocol       string
}
