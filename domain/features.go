package domain

import (
	"github.com/google/uuid"
	"gorm.io/gorm"
)

type Feature struct {
	ID        string  `gorm:"type:char(36);primary_key" json:"id"`
	Name      string  `gorm:"size:255"`
	ServiceID string  `gorm:"type:char(36)"`
	Service   Service `gorm:"foreignKey:ServiceID"`
}

func (a *Feature) BeforeCreate(tx *gorm.DB) (err error) {
	a.ID = uuid.New().String()
	return
}
