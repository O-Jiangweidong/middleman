package models

import "time"

type Permission struct {
    ID          string     `json:"id" gorm:"type:char(36);primaryKey;not null"`
    Name        string     `json:"name" gorm:"type:varchar(128);not null"`
    IsActive    bool       `json:"is_active" gorm:"type:boolean;not null"`
    DateExpired time.Time  `json:"date_expired" gorm:"type:timestamp with time zone;not null;index"`
    CreatedBy   *string    `json:"created_by,omitempty" gorm:"type:varchar(128);default:null"`
    DateCreated *time.Time `json:"date_created,omitempty" gorm:"type:timestamp with time zone;default:null"`
    Comment     string     `json:"comment" gorm:"type:text;not null"`
    DateStart   time.Time  `json:"date_start" gorm:"type:timestamp with time zone;not null;index"`
    OrgID       string     `json:"org_id" gorm:"type:varchar(36);not null;index;uniqueIndex:idx_org_name"`
    Actions     int        `json:"actions" gorm:"type:integer;not null"`
    FromTicket  bool       `json:"from_ticket" gorm:"type:boolean;not null"`
    Accounts    JSONMap    `json:"accounts" gorm:"type:jsonb;not null"`
    DateUpdated time.Time  `json:"date_updated" gorm:"type:timestamp with time zone;not null"`
    UpdatedBy   *string    `json:"updated_by,omitempty" gorm:"type:varchar(128);default:null"`
    Protocols   JSONMap    `json:"protocols" gorm:"type:jsonb;not null"`
}

type PermissionUser struct {
    ID                uint   `json:"id" gorm:"type:serial;primaryKey;autoIncrement"`
    AssetPermissionID string `json:"assetpermission_id" gorm:"type:char(36);not null;uniqueIndex:idx_asset_user"`
    UserID            string `json:"user_id" gorm:"type:char(36);not null;uniqueIndex:idx_asset_user"`
}
