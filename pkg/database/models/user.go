package models

type User struct {
	ID                      string   `json:"id" gorm:"type:uuid;primaryKey;not null"`
	Username                string   `json:"username" gorm:"type:varchar(128);not null;unique"`
	Name                    string   `json:"name" gorm:"type:varchar(128);not null"`
	Email                   string   `json:"email" gorm:"type:varchar(128);not null;unique"`
	IsActive                bool     `json:"is_active" gorm:"type:boolean;not null"`
	Comment                 string   `json:"comment,omitempty" gorm:"type:text"`
	LastLogin               *UTCTime `json:"last_login,omitempty" gorm:"type:timestamp with time zone;not null"`
	DateJoined              *UTCTime `json:"date_joined" gorm:"type:timestamp with time zone;not null"`
	IsFirstLogin            bool     `json:"is_first_login" gorm:"type:boolean;not null"`
	DateExpired             *UTCTime `json:"date_expired,omitempty" gorm:"type:timestamp with time zone;not null"`
	CreatedBy               string   `json:"created_by" gorm:"type:varchar(30);not null"`
	Source                  string   `json:"source" gorm:"type:varchar(30);not null"`
	DatePasswordLastUpdated *UTCTime `json:"date_password_last_updated,omitempty" gorm:"type:timestamp with time zone;not null"`
	NeedUpdatePassword      bool     `json:"need_update_password" gorm:"type:boolean;not null"`
	DateUpdated             *UTCTime `json:"date_updated" gorm:"type:timestamp with time zone;not null"`
	UpdatedBy               string   `json:"updated_by" gorm:"type:varchar(30);not null"`
}
