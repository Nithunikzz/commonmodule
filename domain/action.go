package domain

import (
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

type Action struct {
	ID          string    `gorm:"type:char(36);primary_key" json:"id"`
	Name        string    `gorm:"size:255;not null" json:"name"`
	DisplayName string    `gorm:"size:255;not null" json:"displayname"`
	Description string    `gorm:"size:255;not null" json:"description"`
	ServiceID   string    `gorm:"type:char(36)"`
	Service     Service   `gorm:"foreignKey:ServiceID"`
	Created_At  time.Time `gorm:"autoCreateTime;"`
	Updated_At  time.Time `gorm:"autoUpdateTime;"`
}

func (a *Action) BeforeCreate(tx *gorm.DB) (err error) {
	a.ID = uuid.New().String()
	return
}

type ActionIDResponse struct {
	ID string `json:"id"`
}

func NewAction(name, displayname, ServiceID, description string) *Action {
	return &Action{
		Name:        name,
		DisplayName: displayname,
		ServiceID:   ServiceID,
		Description: description,
	}
}
