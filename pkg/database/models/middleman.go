package models

import "time"

type BaseJumpServer struct {
    ID        uint      `json:"id" gorm:"primaryKey"`
    Name      string    `json:"name" gorm:"not null;size=128;uniqueIndex"`
    CreatedAt time.Time `json:"created_at" gorm:"autoCreateTime"`
    UpdatedAt time.Time `json:"updated_at" gorm:"autoUpdateTime"`
}

type JumpServer struct {
    BaseJumpServer
    AccessKey string `json:"access_key" gorm:"type:varchar(36);not null"`
    SecretKey string `json:"secret_key" gorm:"type:varchar(36);not null"`
}
