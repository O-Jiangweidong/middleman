package models

import (
	"encoding/json"
	"time"
)

type AssetPermission struct {
	ID          string      `json:"id" gorm:"type:uuid;primaryKey"`
	Name        string      `json:"name" gorm:"type:varchar(128);not null"`
	IsActive    bool        `json:"is_active" gorm:"type:boolean;not null"`
	CreatedBy   string      `json:"created_by" gorm:"type:varchar(128);default:null"`
	UpdatedBy   string      `json:"updated_by" gorm:"type:varchar(128);default:null"`
	Comment     string      `json:"comment" gorm:"type:text"`
	OrgID       string      `json:"org_id" gorm:"type:uuid;default null"`
	Actions     int         `json:"actions" gorm:"type:int;not null"`
	FromTicket  bool        `json:"from_ticket" gorm:"type:boolean;default:false"`
	Accounts    StringArray `json:"accounts" gorm:"type:jsonb;not null"`
	Protocols   StringArray `json:"protocols" gorm:"type:jsonb;not null"`
	DateExpired *UTCTime    `json:"date_expired" gorm:"type:timestamp with time zone;default:null"`
	DateStart   *UTCTime    `json:"date_start" gorm:"type:timestamp with time zone;default:null"`
	DateCreated *UTCTime    `json:"date_created" gorm:"type:timestamp with time zone;default:null"`
	DateUpdated *UTCTime    `json:"date_updated" gorm:"type:timestamp with time zone;default:null"`

	Users      []User      `json:"users" gorm:"many2many:perms_assetpermission_users;joinForeignKey:assetpermission_id;joinReferences:user_id;constraint:OnDelete:CASCADE;autoCreate:false;autoUpdate:false"`
	UserGroups []UserGroup `json:"user_groups" gorm:"many2many:perms_assetpermission_user_groups;joinForeignKey:assetpermission_id;joinReferences:usergroup_id;constraint:OnDelete:CASCADE;autoCreate:false;autoUpdate:false"`
	Assets     []Asset     `json:"assets" gorm:"many2many:perms_assetpermission_assets;joinForeignKey:assetpermission_id;joinReferences:asset_id;constraint:OnDelete:CASCADE;autoCreate:false;autoUpdate:false"`
	Nodes      []Node      `json:"nodes" gorm:"many2many:perms_assetpermission_nodes;joinForeignKey:assetpermission_id;joinReferences:node_id;constraint:OnDelete:CASCADE;autoCreate:false;autoUpdate:false"`

	UserIds        []string `json:"user_ids,omitempty" gorm:"-"`
	UserGroupIds   []string `json:"user_group_ids,omitempty" gorm:"-"`
	AssetIds       []string `json:"asset_ids,omitempty" gorm:"-"`
	NodeIds        []string `json:"node_ids,omitempty" gorm:"-"`
	ActionsDisplay []string `json:"actions_display,omitempty" gorm:"-"`
	Valid          bool     `json:"is_valid" gorm:"-"`
}

func (AssetPermission) TableName() string {
	return "perms_assetpermission"
}

func (p AssetPermission) IsValid() bool {
	now := time.Now().UTC()
	if p.DateExpired.After(now) && p.DateStart.Before(now) && p.IsActive {
		return true
	}
	return false
}

type StringSlice []string

func (s StringSlice) MarshalJSON() ([]byte, error) {
	if s == nil {
		return []byte("[]"), nil
	}
	return json.Marshal([]string(s))
}

func (s *StringSlice) UnmarshalJSON(data []byte) error {
	var items []string
	if err := json.Unmarshal(data, &items); err != nil {
		return err
	}
	*s = items
	return nil
}

type JmsAssetPermission struct {
	AssetPermission

	Users      StringSlice `json:"users"`
	UserGroups StringSlice `json:"user_groups"`
	Assets     StringSlice `json:"assets"`
	Nodes      StringSlice `json:"nodes"`
	Actions    []string    `json:"actions"`
}

func (p AssetPermission) ToJms() JmsAssetPermission {
	obj := JmsAssetPermission{AssetPermission: p}

	obj.Users = p.UserIds
	obj.UserGroups = p.UserGroupIds
	obj.Assets = p.AssetIds
	obj.Nodes = p.NodeIds
	obj.Actions = p.ActionsDisplay

	obj.UserIds = nil
	obj.UserGroupIds = nil
	obj.AssetIds = nil
	obj.NodeIds = nil
	obj.ActionsDisplay = nil
	return obj
}
