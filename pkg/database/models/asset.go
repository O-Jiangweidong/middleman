package models

import (
    "time"
)

type Asset struct {
    ID           string     `json:"id" gorm:"type:char(36);primaryKey;not null"`
    Address      string     `json:"address" gorm:"type:varchar(767);not null"`
    Name         string     `json:"name" gorm:"type:varchar(128);not null"`
    IsActive     bool       `json:"is_active" gorm:"type:boolean;not null"`
    CreatedBy    *string    `json:"created_by,omitempty" gorm:"type:varchar(128);default:null"`
    DateCreated  *time.Time `json:"date_created,omitempty" gorm:"type:timestamp with time zone;default:null"`
    Comment      string     `json:"comment" gorm:"type:text;not null"`
    DomainID     *string    `json:"domain_id,omitempty" gorm:"type:char(36);default:null"`
    OrgID        string     `json:"org_id" gorm:"type:varchar(36);not null;index"`
    PlatformID   int        `json:"platform_id" gorm:"type:integer;not null"`
    Connectivity string     `json:"connectivity" gorm:"type:varchar(16);not null"`
    DateVerified *time.Time `json:"date_verified,omitempty" gorm:"type:timestamp with time zone;default:null"`
    DateUpdated  time.Time  `json:"date_updated" gorm:"type:timestamp with time zone;not null"`
    UpdatedBy    *string    `json:"updated_by,omitempty" gorm:"type:varchar(128);default:null"`
    CustomInfo   JSONMap    `json:"custom_info" gorm:"type:jsonb;not null"`
    GatheredInfo JSONMap    `json:"gathered_info" gorm:"type:jsonb;not null"`
}

type Web struct {
    Asset            `gorm:"embedded"` // 继承 Asset 结构体的所有字段
    AssetPtrID       string  `json:"id" gorm:"type:char(36);primaryKey;not null;column:asset_ptr_id"`
    Autofill         string  `json:"autofill" gorm:"type:varchar(16);not null"`
    PasswordSelector string  `json:"password_selector" gorm:"type:varchar(128);not null"`
    SubmitSelector   string  `json:"submit_selector" gorm:"type:varchar(128);not null"`
    UsernameSelector string  `json:"username_selector" gorm:"type:varchar(128);not null"`
    Script           JSONMap `json:"script" gorm:"type:jsonb;not null"`
}

type Host struct {
    Asset      `gorm:"embedded"` // 继承 Asset 结构体的所有字段
    AssetPtrID string `json:"id" gorm:"type:char(36);primaryKey;not null;column:asset_ptr_id"`
}

type Device struct {
    Asset      `gorm:"embedded"` // 继承 Asset 结构体的所有字段
    AssetPtrID string `json:"id" gorm:"type:char(36);primaryKey;not null;column:asset_ptr_id"`
}

type Database struct {
    Asset            `gorm:"embedded"`
    AssetPtrID       string `json:"id" gorm:"type:char(36);primaryKey;not null;column:asset_ptr_id"`
    DBName           string `json:"db_name" gorm:"type:varchar(1024);not null"`
    AllowInvalidCert bool   `json:"allow_invalid_cert" gorm:"type:boolean;not null"`
    CACert           string `json:"ca_cert" gorm:"type:text;not null"`
    ClientCert       string `json:"client_cert" gorm:"type:text;not null"`
    ClientKey        string `json:"client_key" gorm:"type:text;not null"`
    UseSSL           bool   `json:"use_ssl" gorm:"type:boolean;not null"`
}

type Cloud struct {
    Asset      `gorm:"embedded"` // 继承 Asset 结构体的所有字段
    AssetPtrID string `json:"id" gorm:"type:char(36);primaryKey;not null;column:asset_ptr_id"`
}

type GPT struct {
    Asset      `gorm:"embedded"`
    AssetPtrID string `json:"id" gorm:"type:char(36);primaryKey;not null;column:asset_ptr_id"`
    Proxy      string `json:"proxy" gorm:"type:varchar(128);not null"`
}

type Custom struct {
    Asset      `gorm:"embedded"`
    AssetPtrID string `json:"id" gorm:"type:char(36);primaryKey;not null;column:asset_ptr_id"`
}
