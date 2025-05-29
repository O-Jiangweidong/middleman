package models

import "time"

type UserGroup struct {
	ID          string   `json:"id" gorm:"type:char(36);primaryKey;not null"`
	Name        string   `json:"name" gorm:"type:varchar(128);not null"`
	Comment     string   `json:"comment" gorm:"type:text"`
	CreatedBy   string   `json:"created_by" gorm:"type:varchar(128);default:null"`
	OrgID       string   `json:"org_id" gorm:"type:varchar(36);not null;index"`
	UpdatedBy   string   `json:"updated_by" gorm:"type:varchar(128);default:null"`
	DateCreated *UTCTime `json:"date_created" gorm:"type:timestamp with time zone;default:null"`
	DateUpdated *UTCTime `json:"date_updated" gorm:"type:timestamp with time zone;not null"`

	Users []User `gorm:"many2many:users_user_groups;joinForeignKey:UserGroupID;joinReferences:UserID;constraint:OnDelete:CASCADE"`
}

type User struct {
	ID                      string   `json:"id" gorm:"type:uuid;primaryKey;not null"`
	Username                string   `json:"username" gorm:"type:varchar(128);not null;unique"`
	Name                    string   `json:"name" gorm:"type:varchar(128);not null"`
	Email                   string   `json:"email" gorm:"type:varchar(128);not null;unique"`
	IsActive                bool     `json:"is_active" gorm:"type:boolean;not null"`
	Comment                 string   `json:"comment,omitempty" gorm:"type:text"`
	IsFirstLogin            bool     `json:"is_first_login" gorm:"type:boolean;not null"`
	CreatedBy               string   `json:"created_by" gorm:"type:varchar(256);not null"`
	Source                  string   `json:"source" gorm:"type:varchar(30);not null"`
	NeedUpdatePassword      bool     `json:"need_update_password" gorm:"type:boolean;not null"`
	UpdatedBy               string   `json:"updated_by" gorm:"type:varchar(256);not null"`
	MFALevel                int16    `json:"mfa_level" gorm:"type:smallint;not null"`
	Phone                   string   `json:"phone" gorm:"type:text"`
	Wechat                  string   `json:"wechat" gorm:"type:varchar(256)"`
	LastLogin               *UTCTime `json:"last_login,omitempty" gorm:"type:timestamp with time zone;default null"`
	DateJoined              *UTCTime `json:"date_joined" gorm:"type:timestamp with time zone;default null"`
	DateExpired             *UTCTime `json:"date_expired,omitempty" gorm:"type:timestamp with time zone;not null"`
	DatePasswordLastUpdated *UTCTime `json:"date_password_last_updated,omitempty" gorm:"type:timestamp with time zone;default null"`
	DateUpdated             *UTCTime `json:"date_updated" gorm:"type:timestamp with time zone;default null"`

	Roles      []RbacRole  `json:"roles" gorm:"many2many:rbac_role_bindings;joinForeignKey:UserID;joinReferences:RoleID;constraint:OnDelete:CASCADE;SaveReference:false"`
	UserGroups []UserGroup `json:"groups" gorm:"many2many:users_user_groups;joinForeignKey:UserID;joinReferences:UserGroupID;constraint:OnDelete:CASCADE;SaveReference:false"`

	PasswordStrategy string `json:"password_strategy" gorm:"-"`
	Password         string `json:"password" gorm:"-"`
}

func (u *User) IsValid() bool {
	if !u.IsActive {
		return false
	}
	return u.DateExpired.After(time.Now())
}

func (u *User) ToJMSUser() JMSUser {
	user := JMSUser{User: *u}
	var orgRoles, systemRoles []string
	for _, role := range u.Roles {
		if role.Scope == "system" {
			systemRoles = append(systemRoles, role.ID)
		} else {
			orgRoles = append(orgRoles, role.ID)
		}
	}
	user.Roles = nil
	user.OrgRoles = orgRoles
	user.SystemRoles = systemRoles

	var userGroups []string
	for _, group := range u.UserGroups {
		userGroups = append(userGroups, group.ID)
	}
	user.Groups = userGroups
	return user
}

type JMSUser struct {
	User

	Groups      []string `json:"groups,omitempty"`
	OrgRoles    []string `json:"org_roles"`
	SystemRoles []string `json:"system_roles"`
}
