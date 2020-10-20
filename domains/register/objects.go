package register

import (
	"time"

	"github.com/drd-engineering/TwinCape/db"
)

// UserRegistrationData json data definition for User registration
type UserRegistrationData struct {
	Name         string `json:"name"`
	Gender       string `json:"gender"`
	Email        string `json:"email"`
	KtpNumber    int64  `json:"ktpNumber"`
	Address      string `json:"address"`
	PhoneNumber  string `json:"phoneNumber"`
	DateOfBirth  string `json:"dateofBirth"`
	Cityzenship  string `json:"cityzenship"`
	PlaceOfBirth string `json:"placeofBirth"`
}

// ResponseSaveUser json data definition for response from server to http request
type ResponseSaveUser struct {
	ID           string    `json:"id"`
	Name         string    `json:"name"`
	Gender       string    `json:"gender"`
	Email        string    `json:"email"`
	KtpNumber    int64     `json:"ktpNumber"`
	Address      string    `json:"address"`
	PhoneNumber  string    `json:"phoneNumber"`
	Password     string    `json:"password"`
	DateOfBirth  time.Time `json:"dateofBirth"`
	Cityzenship  string    `json:"cityzenship"`
	PlaceOfBirth string    `json:"placeofBirth"`
}

// CreateResponse from database
func (t ResponseSaveUser) CreateResponse(user db.User) ResponseSaveUser {
	t.ID = user.ID
	t.Name = user.Name
	t.Gender = user.Gender
	t.Email = user.Email
	t.KtpNumber = user.KtpNumber
	t.Address = user.Address
	t.PhoneNumber = user.PhoneNumber
	t.DateOfBirth = user.DateOfBirth
	t.Cityzenship = user.Cityzenship
	t.PlaceOfBirth = user.PlaceOfBirth
	return t
}
