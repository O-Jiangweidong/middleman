package models

import "time"

type UserGroup struct {
	ID          string   `json:"id" gorm:"type:char(36);primaryKey;not null"`
	Name        string   `json:"name,omitempty" gorm:"type:varchar(128);not null"`
	Comment     string   `json:"comment,omitempty" gorm:"type:text"`
	CreatedBy   string   `json:"created_by,omitempty" gorm:"type:varchar(128);default:null"`
	OrgID       string   `json:"org_id,omitempty" gorm:"type:varchar(36);not null;index"`
	UpdatedBy   string   `json:"updated_by,omitempty" gorm:"type:varchar(128);default:null"`
	DateCreated *UTCTime `json:"date_created,omitempty" gorm:"type:timestamp with time zone;default:null"`
	DateUpdated *UTCTime `json:"date_updated,omitempty" gorm:"type:timestamp with time zone;not null"`

	Users       []User            `json:"-" gorm:"many2many:users_user_groups;joinForeignKey:UserGroupID;joinReferences:UserID;constraint:OnDelete:CASCADE"`
	Permissions []AssetPermission `json:"-" gorm:"many2many:perms_assetpermission_user_groups;joinForeignKey:usergroup_id;joinReferences:assetpermission_id;constraint:OnDelete:CASCADE"`
}

type User struct {
	ID                      string   `json:"id" gorm:"type:uuid;primaryKey;not null"`
	Username                string   `json:"username,omitempty" gorm:"type:varchar(128);not null;unique"`
	Name                    string   `json:"name,omitempty" gorm:"type:varchar(128);not null"`
	Email                   string   `json:"email,omitempty" gorm:"type:varchar(128);not null;unique"`
	IsActive                bool     `json:"is_active,omitempty" gorm:"type:boolean;default:true"`
	Comment                 string   `json:"comment,omitempty" gorm:"type:text"`
	IsFirstLogin            bool     `json:"is_first_login,omitempty" gorm:"type:boolean;default:true"`
	CreatedBy               string   `json:"created_by,omitempty" gorm:"type:varchar(256);not null"`
	Source                  string   `json:"source,omitempty" gorm:"type:varchar(30);not null"`
	NeedUpdatePassword      bool     `json:"need_update_password,omitempty" gorm:"type:boolean;default:true"`
	UpdatedBy               string   `json:"updated_by,omitempty" gorm:"type:varchar(256);not null"`
	MFALevel                int16    `json:"mfa_level,omitempty" gorm:"type:smallint;not null"`
	Phone                   string   `json:"phone,omitempty" gorm:"type:text"`
	Wechat                  string   `json:"wechat,omitempty" gorm:"type:varchar(256)"`
	LastLogin               *UTCTime `json:"last_login,omitempty" gorm:"type:timestamp with time zone;default null"`
	DateJoined              *UTCTime `json:"date_joined,omitempty" gorm:"type:timestamp with time zone;default null"`
	DateExpired             *UTCTime `json:"date_expired,omitempty" gorm:"type:timestamp with time zone;not null"`
	DatePasswordLastUpdated *UTCTime `json:"date_password_last_updated,omitempty" gorm:"type:timestamp with time zone;default null"`
	DateUpdated             *UTCTime `json:"date_updated,omitempty" gorm:"type:timestamp with time zone;default null"`

	Roles       []RbacRole        `json:"-" gorm:"many2many:rbac_role_bindings;joinForeignKey:UserID;joinReferences:RoleID;constraint:OnDelete:CASCADE;SaveReference:false"`
	UserGroups  []UserGroup       `json:"groups,omitempty" gorm:"many2many:users_user_groups;joinForeignKey:UserID;joinReferences:UserGroupID;constraint:OnDelete:CASCADE;SaveReference:false"`
	Permissions []AssetPermission `json:"-" gorm:"many2many:perms_assetpermission_users;joinForeignKey:user_id;joinReferences:assetpermission_id;constraint:OnDelete:CASCADE"`

	PasswordStrategy string   `json:"password_strategy,omitempty" gorm:"-"`
	Password         string   `json:"password,omitempty" gorm:"-"`
	RoleIds          []string `json:"roles,omitempty" gorm:"-"`
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
	user.RoleIds = nil
	return user
}

type JMSUser struct {
	User

	Groups      []string `json:"groups,omitempty"`
	OrgRoles    []string `json:"org_roles"`
	SystemRoles []string `json:"system_roles"`
}
