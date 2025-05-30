package models

import (
	"database/sql/driver"
	"encoding/json"
	"fmt"
)

type Platform struct {
	ID          uint     `json:"id" gorm:"primaryKey;autoIncrement"`
	Name        string   `json:"name,omitempty" gorm:"type:varchar(50);not null;unique"`
	Type        string   `json:"type,omitempty" gorm:"type:varchar(32);not null"`
	Internal    bool     `json:"internal,omitempty" gorm:"type:boolean;default:false"`
	Comment     string   `json:"comment,omitempty" gorm:"type:text;"`
	Category    string   `json:"category,omitempty" gorm:"type:varchar(32);not null"`
	CreatedBy   string   `json:"created_by,omitempty" gorm:"type:varchar(128);default:null"`
	UpdatedBy   string   `json:"updated_by,omitempty" gorm:"type:varchar(128);default:null"`
	DateCreated *UTCTime `json:"date_created,omitempty" gorm:"type:timestamp with time zone;default:null"`
	DateUpdated *UTCTime `json:"date_updated,omitempty" gorm:"type:timestamp with time zone;default:null"`

	Assets []Asset `json:"-" gorm:"foreignKey:PlatformID"`
}

type PlatformProtocol struct {
	ID         int    `json:"id"`
	Name       string `json:"name"`
	Port       int    `json:"port"`
	PlatformID int    `json:"platform_id"`
	Default    bool   `json:"default"`
	Required   bool   `json:"required"`
	Primary    bool   `json:"primary"`
	Public     bool   `json:"public"`
}

type Protocol struct {
	Name string `json:"name"`
	Port int64  `json:"port"`
}

type ProtocolArray []Protocol

func (p ProtocolArray) Value() (driver.Value, error) {
	return json.Marshal(p)
}

func (p *ProtocolArray) Scan(value interface{}) error {
	if value == nil {
		return nil
	}

	var bytes []byte
	switch v := value.(type) {
	case []byte:
		bytes = v
	case string:
		bytes = []byte(v)
	default:
		return fmt.Errorf("不支持的类型: %T", value)
	}
	return json.Unmarshal(bytes, &p)
}

type SimpleAsset struct {
	AssetPtrID string `json:"asset_ptr_id"`
}

type Asset struct {
	ID           string        `json:"id" gorm:"type:uuid;primaryKey;not null"`
	Address      string        `json:"address,omitempty" gorm:"type:varchar(767);not null"`
	Name         string        `json:"name,omitempty" gorm:"type:varchar(128);not null"`
	IsActive     bool          `json:"is_active,omitempty" gorm:"type:boolean;default:true"`
	CreatedBy    string        `json:"created_by,omitempty" gorm:"type:varchar(128)"`
	UpdatedBy    string        `json:"updated_by,omitempty" gorm:"type:varchar(128)"`
	Comment      string        `json:"comment,omitempty" gorm:"type:text"`
	OrgID        string        `json:"org_id,omitempty" gorm:"type:uuid;not null;index"`
	PlatformID   uint          `json:"platform_id,omitempty" gorm:"type:integer;not null"`
	Connectivity string        `json:"connectivity,omitempty" gorm:"type:varchar(16);not null"`
	DateCreated  *UTCTime      `json:"date_created,omitempty" gorm:"type:timestamp with time zone;default:null"`
	DateVerified *UTCTime      `json:"date_verified,omitempty" gorm:"type:timestamp with time zone;default:null"`
	DateUpdated  *UTCTime      `json:"date_updated,omitempty" gorm:"type:timestamp with time zone;default:null"`
	Protocols    ProtocolArray `json:"protocols,omitempty" gorm:"type:jsonb;not null"`

	Platform Platform  `json:"platform,omitempty" gorm:"foreignKey:PlatformID;references:ID"`
	Accounts []Account `json:"accounts,omitempty" gorm:"foreignKey:AssetID;constraint:OnDelete:CASCADE"`

	Permissions []AssetPermission `json:"-" gorm:"many2many:perms_assetpermission_assets;joinForeignKey:asset_id;joinReferences:assetpermission_id;constraint:OnDelete:CASCADE"`
	Nodes       []Node            `json:"-" gorm:"many2many:assets_asset_nodes;joinForeignKey:node_id;joinReferences:asset_id;constraint:OnDelete:CASCADE"`

	Host     *Host     `json:"-" gorm:"foreignKey:AssetPtrID"`
	Web      *Web      `json:"-" gorm:"foreignKey:AssetPtrID"`
	Device   *Device   `json:"-" gorm:"foreignKey:AssetPtrID"`
	Database *Database `json:"-" gorm:"foreignKey:AssetPtrID"`
	Cloud    *Cloud    `json:"-" gorm:"foreignKey:AssetPtrID"`
	GPT      *GPT      `json:"-" gorm:"foreignKey:AssetPtrID"`
	Custom   *Custom   `json:"-" gorm:"foreignKey:AssetPtrID"`

	Category LabelValue `json:"category,omitempty" gorm:"-"`
	Type     LabelValue `json:"type,omitempty" gorm:"-"`
}

type JmsAsset struct {
	Asset
}

func (a Asset) ToJms() JmsAsset {
	a.Platform.ID = a.PlatformID
	a.OrgID = ""
	a.PlatformID = 0
	var accounts []Account
	for _, account := range a.Accounts {
		account.AssetID = ""
		accounts = append(accounts, account)
	}
	a.Accounts = accounts
	return JmsAsset{Asset: a}
}

type Web struct {
	AssetPtrID       string  `json:"asset_ptr_id" gorm:"primaryKey;type:uuid;not null"`
	Asset            Asset   `json:"asset" gorm:"foreignKey:AssetPtrID;references:ID"`
	Autofill         string  `json:"autofill" gorm:"type:varchar(16);not null"`
	PasswordSelector string  `json:"password_selector" gorm:"type:varchar(128);not null"`
	SubmitSelector   string  `json:"submit_selector" gorm:"type:varchar(128);not null"`
	UsernameSelector string  `json:"username_selector" gorm:"type:varchar(128);not null"`
	Script           JSONMap `json:"script" gorm:"type:jsonb;not null"`
}

type Host struct {
	AssetPtrID string `json:"asset_ptr_id" gorm:"primaryKey;type:uuid;not null"`
	Asset      Asset  `json:"asset" gorm:"foreignKey:AssetPtrID;references:ID;constraint:OnDelete:CASCADE"`
}

func (h *Host) UnmarshalJSON(data []byte) error {
	var asset Asset
	if err := json.Unmarshal(data, &asset); err != nil {
		return err
	}
	h.AssetPtrID = asset.ID
	h.Asset = asset
	h.Asset.OrgID = DefaultOrgID
	return nil
}

type Device struct {
	AssetPtrID string `gorm:"primaryKey;type:uuid;not null"`
	Asset      Asset  `gorm:"foreignKey:AssetPtrID;references:ID"`
}

type Database struct {
	AssetPtrID       string `gorm:"primaryKey;type:uuid;not null"`
	Asset            Asset  `gorm:"foreignKey:AssetPtrID;references:ID"`
	DBName           string `json:"db_name" gorm:"type:varchar(1024);not null"`
	AllowInvalidCert bool   `json:"allow_invalid_cert" gorm:"type:boolean;not null"`
	CACert           string `json:"ca_cert" gorm:"type:text;not null"`
	ClientCert       string `json:"client_cert" gorm:"type:text;not null"`
	ClientKey        string `json:"client_key" gorm:"type:text;not null"`
	UseSSL           bool   `json:"use_ssl" gorm:"type:boolean;not null"`
}

type Cloud struct {
	AssetPtrID string `gorm:"primaryKey;type:uuid;not null"`
	Asset      Asset  `gorm:"foreignKey:AssetPtrID;references:ID"`
}

type GPT struct {
	AssetPtrID string `gorm:"primaryKey;type:uuid;not null"`
	Asset      Asset  `gorm:"foreignKey:AssetPtrID;references:ID"`
	Proxy      string `json:"proxy" gorm:"type:varchar(128);not null"`
}

type Custom struct {
	AssetPtrID string `gorm:"primaryKey;type:uuid;not null"`
	Asset      Asset  `gorm:"foreignKey:AssetPtrID;references:ID"`
}

type Node struct {
	ID           string   `json:"id" gorm:"type:uuid;primaryKey"`
	Key          string   `json:"key" gorm:"type:varchar(64);not null;unique"`
	Value        string   `json:"value" gorm:"type:varchar(128);not null"`
	ChildMark    int      `json:"child_mark" gorm:"type:int;not null"`
	OrgID        string   `json:"org_id" gorm:"type:uuid;not null;index"`
	AssetsAmount int      `json:"assets_amount" gorm:"type:int;default:0"`
	ParentKey    string   `json:"parent_key" gorm:"type:varchar(64);not null;index"`
	FullValue    string   `json:"full_value" gorm:"type:varchar(4096);not null"`
	Comment      string   `json:"comment" gorm:"type:text"`
	CreatedBy    string   `json:"created_by" gorm:"type:varchar(128);default null"`
	UpdatedBy    string   `json:"updated_by" gorm:"type:varchar(128);default null"`
	DateCreate   *UTCTime `json:"date_create" gorm:"type:timestamp with time zone;default:null"`
	DateCreated  *UTCTime `json:"date_created" gorm:"type:timestamp with time zone;default:null"`
	DateUpdated  *UTCTime `json:"date_updated" gorm:"type:timestamp with time zone;default:null"`

	Permissions []AssetPermission `json:"-" gorm:"many2many:perms_assetpermission_nodes;joinForeignKey:node_id;joinReferences:assetpermission_id;constraint:OnDelete:CASCADE"`
	Assets      []Asset           `json:"-" gorm:"many2many:assets_asset_nodes;joinForeignKey:node_id;joinReferences:asset_id;constraint:OnDelete:CASCADE"`
}

func (Node) TableName() string {
	return "assets_node"
}
