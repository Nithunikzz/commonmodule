package domain

import (
	"github.com/google/uuid"
	"gorm.io/gorm"
)

type OrganisationUserRole struct {
	ID             string `gorm:"type:char(36);primary_key" json:"id"`
	OrganisationID string `gorm:"type:char(36)"`
	UserID         string `gorm:"type:char(36)"`
	RoleID         string `gorm:"type:char(36)"`
	Organisation   Organisation
	User           User
	Role           Role
}
type AssociationResponse struct {
	ID string `json:"id"`
}

func (a *OrganisationUserRole) BeforeCreate(tx *gorm.DB) (err error) {
	a.ID = uuid.New().String()
	return
}

func NewOrganisationUserRole(orgid, userid, roleid string) *OrganisationUserRole {
	return &OrganisationUserRole{
		OrganisationID: orgid,
		UserID:         userid,
		RoleID:         roleid,
	}
}

type OrganisationFeature struct {
	ID             string `gorm:"type:char(36);primary_key" json:"id"`
	OrganisationID string `gorm:"type:char(36)"`
	FeatureID      string `gorm:"type:char(36)"`
	Organisation   Organisation
	Feature        Feature
}

func (a *OrganisationFeature) BeforeCreate(tx *gorm.DB) (err error) {
	a.ID = uuid.New().String()
	return
}

func NewOrganisationFeature(orgid, featureid string) *OrganisationFeature {
	return &OrganisationFeature{
		OrganisationID: orgid,
		FeatureID:      featureid,
	}
}

type RoleAction struct {
	ID       string `gorm:"type:char(36);primary_key" json:"id"`
	RoleID   string `gorm:"type:char(36)"`
	ActionID string `gorm:"type:char(36)"`
	Role     Role
	Action   Action
}

func (a *RoleAction) BeforeCreate(tx *gorm.DB) (err error) {
	a.ID = uuid.New().String()
	return
}
func NewRoleAction(roleid, actionid string) *RoleAction {
	return &RoleAction{
		RoleID:   roleid,
		ActionID: actionid,
	}
}
