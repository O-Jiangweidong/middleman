package models

import (
	"database/sql/driver"
	"encoding/json"
	"fmt"
)

type Platform struct {
	ID          uint     `json:"id" gorm:"primaryKey;autoIncrement"`
	Name        string   `json:"name" gorm:"type:varchar(50);not null;unique"`
	Type        string   `json:"type" gorm:"type:varchar(32);not null"`
	Category    string   `json:"category" gorm:"type:varchar(32);not null"`
	CreatedBy   string   `json:"created_by" gorm:"type:varchar(128);default:null"`
	UpdatedBy   string   `json:"updated_by" gorm:"type:varchar(128);default:null"`
	DateCreated *UTCTime `json:"date_created" gorm:"type:timestamp with time zone;default:null"`
	DateUpdated *UTCTime `json:"date_updated" gorm:"type:timestamp with time zone;not null"`
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

type ProtocolList []Protocol

func (p ProtocolList) Value() (driver.Value, error) {
	return json.Marshal(p)
}

func (p ProtocolList) Scan(value interface{}) error {
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
	ID           string       `json:"id" gorm:"type:char(36);primaryKey;not null"`
	Address      string       `json:"address" gorm:"type:varchar(767);not null"`
	Name         string       `json:"name" gorm:"type:varchar(128);not null"`
	IsActive     bool         `json:"is_active" gorm:"type:boolean;not null"`
	CreatedBy    string       `json:"created_by,omitempty" gorm:"type:varchar(128)"`
	DateCreated  *UTCTime     `json:"date_created,omitempty" gorm:"type:timestamp with time zone;default:null"`
	Comment      string       `json:"comment" gorm:"type:text;not null"`
	OrgID        string       `json:"org_id" gorm:"type:varchar(36);not null;index"`
	PlatformID   int          `json:"platform_id" gorm:"type:integer;not null"`
	Connectivity string       `json:"connectivity" gorm:"type:varchar(16);not null"`
	DateVerified *UTCTime     `json:"date_verified,omitempty" gorm:"type:timestamp with time zone;default:null"`
	DateUpdated  *UTCTime     `json:"date_updated" gorm:"type:timestamp with time zone;default:null"`
	UpdatedBy    string       `json:"updated_by,omitempty" gorm:"type:varchar(128)"`
	Protocols    ProtocolList `json:"protocols" gorm:"type:jsonb;not null"`

	Platform string    `json:"platform"`
	Accounts []Account `json:"accounts"`

	Host     *Host     `gorm:"foreignKey:AssetPtrID"`
	Web      *Web      `gorm:"foreignKey:AssetPtrID"`
	Device   *Device   `gorm:"foreignKey:AssetPtrID"`
	Database *Database `gorm:"foreignKey:AssetPtrID"`
	Cloud    *Cloud    `gorm:"foreignKey:AssetPtrID"`
	GPT      *GPT      `gorm:"foreignKey:AssetPtrID"`
	Custom   *Custom   `gorm:"foreignKey:AssetPtrID"`
}

type Web struct {
	AssetPtrID       string  `json:"asset_ptr_id" gorm:"primaryKey;type:char(36);not null"`
	Asset            Asset   `json:"asset" gorm:"foreignKey:AssetPtrID;references:ID"`
	Autofill         string  `json:"autofill" gorm:"type:varchar(16);not null"`
	PasswordSelector string  `json:"password_selector" gorm:"type:varchar(128);not null"`
	SubmitSelector   string  `json:"submit_selector" gorm:"type:varchar(128);not null"`
	UsernameSelector string  `json:"username_selector" gorm:"type:varchar(128);not null"`
	Script           JSONMap `json:"script" gorm:"type:jsonb;not null"`
}

type Host struct {
	AssetPtrID string `json:"asset_ptr_id" gorm:"primaryKey;type:char(36);not null"`
	Asset      Asset  `json:"asset" gorm:"foreignKey:AssetPtrID;references:ID"`
}

func (h *Host) UnmarshalJSON(data []byte) error {
	var simpleAsset SimpleAsset
	if err := json.Unmarshal(data, &simpleAsset); err != nil {
		return err
	}

	var asset Asset
	if err := json.Unmarshal(data, &asset); err != nil {
		return err
	}
	h.AssetPtrID = simpleAsset.AssetPtrID
	h.Asset = asset
	return nil
}

type Device struct {
	AssetPtrID string `gorm:"primaryKey;type:char(36);not null"`
	Asset      Asset  `gorm:"foreignKey:AssetPtrID;references:ID"`
}

type Database struct {
	AssetPtrID       string `gorm:"primaryKey;type:char(36);not null"`
	Asset            Asset  `gorm:"foreignKey:AssetPtrID;references:ID"`
	DBName           string `json:"db_name" gorm:"type:varchar(1024);not null"`
	AllowInvalidCert bool   `json:"allow_invalid_cert" gorm:"type:boolean;not null"`
	CACert           string `json:"ca_cert" gorm:"type:text;not null"`
	ClientCert       string `json:"client_cert" gorm:"type:text;not null"`
	ClientKey        string `json:"client_key" gorm:"type:text;not null"`
	UseSSL           bool   `json:"use_ssl" gorm:"type:boolean;not null"`
}

type Cloud struct {
	AssetPtrID string `gorm:"primaryKey;type:char(36);not null"`
	Asset      Asset  `gorm:"foreignKey:AssetPtrID;references:ID"`
}

type GPT struct {
	AssetPtrID string `gorm:"primaryKey;type:char(36);not null"`
	Asset      Asset  `gorm:"foreignKey:AssetPtrID;references:ID"`
	Proxy      string `json:"proxy" gorm:"type:varchar(128);not null"`
}

type Custom struct {
	AssetPtrID string `gorm:"primaryKey;type:char(36);not null"`
	Asset      Asset  `gorm:"foreignKey:AssetPtrID;references:ID"`
}
