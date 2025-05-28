package models

import (
	"crypto/sha256"
	"encoding/base64"
	"fmt"
	"gorm.io/gorm"
	"middleman/pkg/config"
	"middleman/pkg/utils"
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
	Endpoint  string    `json:"endpoint" gorm:"not null"`
}

type JumpServer struct {
	BaseJumpServer
	PrivateToken string `json:"private_token" gorm:"not null"`
	AccessKey    string `json:"access_key" gorm:"type:varchar(36);not null"`
	SecretKey    string `json:"secret_key" gorm:"not null"`
}

func (jms *JumpServer) GetKey() []byte {
	conf := config.GetConf()
	hashed := sha256.Sum256([]byte(conf.BootstrapToken))
	return hashed[:]
}

func (jms *JumpServer) BeforeSave(tx *gorm.DB) error {
	key := jms.GetKey()
	ciphertext, err := utils.Encrypt([]byte(jms.PrivateToken), key)
	if err != nil {
		return fmt.Errorf("加密 Private token 失败: %w", err)
	}
	jms.PrivateToken = base64.StdEncoding.EncodeToString(ciphertext)
	ciphertext, err = utils.Encrypt([]byte(jms.SecretKey), key)
	if err != nil {
		return fmt.Errorf("加密 Secret key 失败: %w", err)
	}
	jms.SecretKey = base64.StdEncoding.EncodeToString(ciphertext)
	return nil
}

func (jms *JumpServer) AfterFind(tx *gorm.DB) error {
	key := jms.GetKey()
	text, err := base64.StdEncoding.DecodeString(jms.PrivateToken)
	if err != nil {
		return fmt.Errorf("解码失败: %w", err)
	}
	plaintext, err := utils.Decrypt(text, key)
	if err != nil {
		return fmt.Errorf("解密失败: %w", err)
	}
	jms.PrivateToken = plaintext

	text, err = base64.StdEncoding.DecodeString(jms.SecretKey)
	if err != nil {
		return fmt.Errorf("解码失败: %w", err)
	}
	plaintext, err = utils.Decrypt(text, key)
	if err != nil {
		return fmt.Errorf("解密失败: %w", err)
	}
	jms.AccessKey = plaintext
	return nil
}

func (jms *JumpServer) Validate() error {
	if !jms.Role.IsValid() {
		return fmt.Errorf("无效的角色类型: %s", jms.Role)
	}
	if !jms.Name.IsValid() {
		return fmt.Errorf("name 只能是大小写字母及下划线组成")
	}
	return nil
}
