package models

import (
    "fmt"
    "regexp"
    "time"
)

type RoleType string
type NameType string

const (
    RoleMaster RoleType = "master"
    RoleSlave  RoleType = "slave"
)

func (r RoleType) IsValid() bool {
    return r == RoleMaster || r == RoleSlave
}

func (n NameType) IsValid() bool {
    match, _ := regexp.MatchString(`^[a-zA-Z_]+$`, string(n))
    return match
}

type BaseJumpServer struct {
    ID        uint      `json:"id" gorm:"primaryKey"`
    Name      NameType  `json:"name" gorm:"not null;size=128;uniqueIndex"`
    Display   string    `json:"display" gorm:"not null;size=128"`
    CreatedAt time.Time `json:"created_at" gorm:"autoCreateTime"`
    UpdatedAt time.Time `json:"updated_at" gorm:"autoUpdateTime"`
    Role      RoleType  `json:"role" gorm:"not null"`
}

type JumpServer struct {
    BaseJumpServer
    AccessKey string `json:"access_key" gorm:"type:varchar(36);not null"`
    SecretKey string `json:"secret_key" gorm:"type:varchar(36);not null"`
}

func (js *JumpServer) Validate() error {
    if !js.Role.IsValid() {
        return fmt.Errorf("无效的角色类型: %s", js.Role)
    }
    if !js.Name.IsValid() {
        return fmt.Errorf("name 只能是大小写字母及下划线组成")
    }
    return nil
}
