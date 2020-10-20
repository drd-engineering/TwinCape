package authenticator

import (
	"time"

	"github.com/drd-engineering/TwinCape/db"
)

// UserLogin is user data requested to login to system
type UserLogin struct {
	ID       string `json:"id"`
	Email    string `json:"email"`
	Password string `json:"password"`
}

// TokenDetails is response containing access token and refresh token
type TokenDetails struct {
	AccessToken  string `json:"accessToken"`
	RefreshToken string `json:"refreshToken"`
}

// RequestRefreshToken is json request body for refreshing token
type RequestRefreshToken struct {
	RefreshToken string `json:"refreshToken"`
}

// ResponseLoginDetails is data containing user logged in details
type ResponseLoginDetails struct {
	ID           string    `json:"id"`
	Name         string    `json:"name"`
	Gender       string    `json:"gender"`
	Email        string    `json:"email"`
	KtpNumber    int64     `json:"ktpNumber"`
	Address      string    `json:"address"`
	PhoneNumber  string    `json:"phoneNumber"`
	DateOfBirth  time.Time `json:"dateofBirth"`
	Cityzenship  string    `json:"cityzenship"`
	PlaceOfBirth string    `json:"placeofBirth"`
}

// CreateResponse from database
func (t ResponseLoginDetails) CreateResponse(user db.User) ResponseLoginDetails {
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
