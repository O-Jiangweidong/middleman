package models

import (
    "time"
)

type JumpServer struct {
    ID        uint      `gorm:"primaryKey"`
    Name      string    `gorm:"not null;size=128;uniqueIndex"`
    CreatedAt time.Time `gorm:"autoCreateTime"`
    UpdatedAt time.Time `gorm:"autoUpdateTime"`
}
