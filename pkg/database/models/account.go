package models

type Account struct {
	ID           string   `json:"id" gorm:"type:uuid;primaryKey;not null"`
	OrgID        string   `json:"org_id" gorm:"type:uuid;not null;index"`
	Connectivity string   `json:"connectivity" gorm:"type:varchar(16);not null"`
	Name         string   `json:"name" gorm:"type:varchar(128);not null;uniqueIndex:idx_name_asset"`
	Comment      string   `json:"comment,omitempty" gorm:"type:text;not null"`
	Username     string   `json:"username" gorm:"type:varchar(128);not null"`
	SecretType   string   `json:"secret_type" gorm:"type:varchar(16);not null"`
	Privileged   bool     `json:"privileged" gorm:"type:boolean;not null"`
	IsActive     bool     `json:"is_active" gorm:"type:boolean;not null"`
	Version      int      `json:"version,omitempty" gorm:"type:integer;not null"`
	Source       string   `json:"source,omitempty" gorm:"type:varchar(30);not null"`
	AssetID      string   `json:"asset_id,omitempty" gorm:"type:uuid;not null;uniqueIndex:idx_name_asset"`
	SuFromID     string   `json:"su_from_id,omitempty" gorm:"type:uuid;default:null"`
	SourceID     string   `json:"source_id,omitempty" gorm:"type:varchar(128);default:null"`
	CreatedBy    string   `json:"created_by,omitempty" gorm:"type:varchar(128);default:null"`
	UpdatedBy    string   `json:"updated_by,omitempty" gorm:"type:varchar(128);default:null"`
	DateVerified *UTCTime `json:"date_verified,omitempty" gorm:"type:timestamp with time zone;default:null"`
	DateCreated  *UTCTime `json:"date_created,omitempty" gorm:"type:timestamp with time zone;default:null"`
	DateUpdated  *UTCTime `json:"date_updated,omitempty" gorm:"type:timestamp with time zone;default:null"`

	PushNow bool   `json:"push_now" gorm:"-"`
	Secret  string `json:"secret,omitempty" gorm:"-"`

	Asset Asset `json:"-" gorm:"foreignKey:AssetID;references:ID"`
}
