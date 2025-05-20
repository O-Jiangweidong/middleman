package models

import "time"

type Organization struct {
    ID          string     `json:"id" gorm:"type:char(36);primaryKey;not null"`
    Name        string     `json:"name" gorm:"type:varchar(128);not null;unique"`
    CreatedBy   *string    `json:"created_by,omitempty" gorm:"type:varchar(128);default:null"`
    DateCreated *time.Time `json:"date_created,omitempty" gorm:"type:timestamp with time zone;default:null"`
    Comment     string     `json:"comment" gorm:"type:text;not null"`
    Builtin     bool       `json:"builtin" gorm:"type:boolean;not null"`
    DateUpdated time.Time  `json:"date_updated" gorm:"type:timestamp with time zone;not null"`
    UpdatedBy   *string    `json:"updated_by,omitempty" gorm:"type:varchar(128);default:null"`
}
