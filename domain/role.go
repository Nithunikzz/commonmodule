package domain

import (
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

type Role struct {
	ID             string       `gorm:"type:char(36);primary_key" json:"id"`
	Name           string       `gorm:"size:255;not null" json:"name"`
	DisplayName    string       `gorm:"size:255;not null" json:"displayname"`
	Description    string       `gorm:"size:255" json:"description"` // new field
	Owner          string       `json:"owner,omitempty"`
	ServiceID      string       `gorm:"type:char(36)"`
	Service        Service      `gorm:"foreignKey:ServiceID"`
	OrganisationID string       `gorm:"type:char(36)"`
	Organisation   Organisation `gorm:"foreignKey:OrganisationID"`
	Created_At     time.Time    `gorm:"autoCreateTime;"`
	Updated_At     time.Time    `gorm:"autoUpdateTime;"`
	Actions        []Action     `gorm:"many2many:role_actions"`
}

func (r *Role) BeforeCreate(tx *gorm.DB) (err error) {
	r.ID = uuid.New().String()
	return
}

type Resource struct {
	ID        uuid.UUID `gorm:"type:uuid;primary_key" json:"id"`
	Name      string    `gorm:"size:255;not null" json:"name"`
	ServiceID uuid.UUID `gorm:"type:uuid;not null" json:"-"`
	Roles     []Role    `gorm:"many2many:resource_roles;" json:"roles"`
}
type RoleIDResponse struct {
	ID string `json:"id"`
}

func NewRole(name, displayname, serviceid, Owner, Orgid, description string) *Role {
	return &Role{
		Name:           name,
		DisplayName:    displayname,
		ServiceID:      serviceid,
		Owner:          Owner,
		OrganisationID: Orgid,
		Description:    description, // set the value of the new field
	}
}

func NewResource(name string) *Resource {
	return &Resource{
		ID:   uuid.New(),
		Name: name,
	}
}
func NewService(name, description string) *Service {
	return &Service{
		Name:        name,
		Description: description,
	}
}
func NewResponse(id string) *RoleIDResponse {
	return &RoleIDResponse{
		ID: id,
	}
}
