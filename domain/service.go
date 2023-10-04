package domain

import (
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

type Service struct {
	ID          string    `gorm:"type:char(36);primary_key" json:"id"`
	Name        string    `gorm:"size:255;not null" json:"name"`
	Description string    `gorm:"size:255" json:"description"`
	Version     int       `gorm:"default:1" json:"version"`
	Created_At  time.Time `gorm:"autoCreateTime;"`
	Updated_At  time.Time `gorm:"autoUpdateTime;"`
}

func (u *Service) BeforeCreate(tx *gorm.DB) (err error) {
	u.ID = uuid.New().String()
	return
}

type ServiceIDResponse struct {
	//Id uuid.UUID `json:"id"`
	Id string `json:"id"`
}

type RegisterServiceInput struct {
	ClientID           string
	ClientSecret       string
	ServiceName        string
	ServiceDescription string
}

type RegisterServiceOutput struct {
	Id string
}

func NewRegisterService(cliId, cliSecret, regSerName, regSerDes string) *RegisterServiceOutput {
	return &RegisterServiceOutput{
		Id: cliId,
	}
}

func NewMyServiceInfo(serviceName string, serviceDescription string) *Service {
	return &Service{
		ID:          uuid.New().String(),
		Name:        serviceName,
		Description: serviceDescription,
	}
}
