package domain

import (
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

// 	ID            string         `gorm:"type:char(36);primary_key" json:"id"`

type User struct {
	ID string `gorm:"type:char(36);primary_key" json:"id"`
	//Username string `gorm:"size:255;not null" json:"username"`
	Firstname  string    `gorm:"size:255;not null" json:"firstname"`
	Lastname   string    `gorm:"size:255;not null" json:"lastname"`
	Email      string    `gorm:"size:255;not null;unique" json:"email"`
	OryID      string    `json:"ory_id"`
	Created_At time.Time `gorm:"autoCreateTime;"`
	Updated_At time.Time `gorm:"autoUpdateTime;"`
}

func (u *User) BeforeCreate(tx *gorm.DB) (err error) {
	u.ID = uuid.New().String()
	return
}

type UserIDResponse struct {
	UserID string `json:"id"`
}

func NewUser(firstname string, lastname string, email string, OryID string) *User {
	return &User{
		//Username: username,
		Firstname: firstname,
		Lastname:  lastname,
		Email:     email,
		OryID:     OryID,
	}
}
