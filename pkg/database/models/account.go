package models

import "time"

type Account struct {
    CreatedBy    *string    `json:"created_by,omitempty" gorm:"type:varchar(128);default:null"`
    UpdatedBy    *string    `json:"updated_by,omitempty" gorm:"type:varchar(128);default:null"`
    DateCreated  *time.Time `json:"date_created,omitempty" gorm:"type:timestamp with time zone;default:null"`
    DateUpdated  time.Time  `json:"date_updated" gorm:"type:timestamp with time zone;not null"`
    Comment      string     `json:"comment" gorm:"type:text;not null"`
    ID           string     `json:"id" gorm:"type:char(36);primaryKey;not null"`
    OrgID        string     `json:"org_id" gorm:"type:varchar(36);not null;index"`
    Connectivity string     `json:"connectivity" gorm:"type:varchar(16);not null"`
    DateVerified *time.Time `json:"date_verified,omitempty" gorm:"type:timestamp with time zone;default:null"`
    Name         string     `json:"name" gorm:"type:varchar(128);not null;uniqueIndex:idx_name_asset"`
    Username     string     `json:"username" gorm:"type:varchar(128);not null;index;uniqueIndex:idx_username_asset_secret"`
    SecretType   string     `json:"secret_type" gorm:"type:varchar(16);not null;uniqueIndex:idx_username_asset_secret"`
    Secret       *string    `json:"secret,omitempty" gorm:"type:text;column:_secret"`
    Privileged   bool       `json:"privileged" gorm:"type:boolean;not null"`
    IsActive     bool       `json:"is_active" gorm:"type:boolean;not null"`
    Version      int        `json:"version" gorm:"type:integer;not null"`
    Source       string     `json:"source" gorm:"type:varchar(30);not null"`
    AssetID      string     `json:"asset_id" gorm:"type:char(36);not null;uniqueIndex:idx_name_asset;uniqueIndex:idx_username_asset_secret"`
    SuFromID     *string    `json:"su_from_id,omitempty" gorm:"type:char(36);default:null"`
    SourceID     *string    `json:"source_id,omitempty" gorm:"type:varchar(128);default:null"`
}
