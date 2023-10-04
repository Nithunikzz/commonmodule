package domain

import (
	"github.com/google/uuid"
	"gorm.io/gorm"
)

//	type Organisation struct {
//		ID          string `gorm:"type:char(36);primary_key" json:"id"`
//		Name        string `gorm:"size:255;not null;unique" json:"name"`
//		DisplayName string `gorm:"size:255;not null" json:"displayname"`
//		Description string `gorm:"size:255" json:"description"`
//		OryId       string `gorm:"size:255" json:"oryid"`
//	}
type Organisation struct {
	ID          string `gorm:"type:char(36);primary_key" json:"id"`
	Name        string `gorm:"size:255;not null;unique" json:"name"`
	DisplayName string `gorm:"size:255;not null" json:"displayname"`
	Description string `gorm:"size:255" json:"description"`
}

func (o *Organisation) BeforeCreate(tx *gorm.DB) (err error) {
	o.ID = uuid.New().String()
	return
}

type OrganisationIDResponse struct {
	ID string `json:"id"`
}

//	func NewOrganisation(name, description, oryid string) *Organisation {
//		return &Organisation{
//			Name:        name,
//			Description: description,
//			OryId:       oryid,
//		}
//	}
func NewOrganisation(name, description string) *Organisation {
	return &Organisation{
		Name:        name,
		Description: description,
	}
}
