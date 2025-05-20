package models

import "time"

type User struct {
    ID                      string     `json:"id" gorm:"type:char(32);primaryKey;not null"`
    Password                string     `json:"password" gorm:"type:varchar(128);not null"`
    LastLogin               *time.Time `json:"last_login,omitempty" gorm:"type:timestamp with time zone;default:null"`
    FirstName               string     `json:"first_name" gorm:"type:varchar(150);not null"`
    LastName                string     `json:"last_name" gorm:"type:varchar(150);not null"`
    IsActive                bool       `json:"is_active" gorm:"type:boolean;not null"`
    DateJoined              time.Time  `json:"date_joined" gorm:"type:timestamp with time zone;not null"`
    Username                string     `json:"username" gorm:"type:varchar(128);not null;unique"`
    Name                    string     `json:"name" gorm:"type:varchar(128);not null"`
    Email                   string     `json:"email" gorm:"type:varchar(128);not null;unique"`
    Role                    string     `json:"role" gorm:"type:varchar(10);not null"`
    Avatar                  *string    `json:"avatar,omitempty" gorm:"type:varchar(100);default:null"`
    Wechat                  string     `json:"wechat" gorm:"type:varchar(256);not null"`
    Phone                   *string    `json:"phone,omitempty" gorm:"type:varchar(256);default:null"`
    PrivateKey              *string    `json:"private_key,omitempty" gorm:"type:text;default:null"`
    PublicKey               *string    `json:"public_key,omitempty" gorm:"type:text;default:null"`
    Comment                 *string    `json:"comment,omitempty" gorm:"type:text;default:null"`
    IsFirstLogin            bool       `json:"is_first_login" gorm:"type:boolean;not null"`
    DateExpired             *time.Time `json:"date_expired,omitempty" gorm:"type:timestamp with time zone;default:null;index"`
    CreatedBy               string     `json:"created_by" gorm:"type:varchar(30);not null"`
    MFALevel                int16      `json:"mfa_level" gorm:"type:smallint;not null"`
    OTPSecretKey            *string    `json:"otp_secret_key,omitempty" gorm:"type:varchar(256);default:null"`
    Source                  string     `json:"source" gorm:"type:varchar(30);not null"`
    DatePasswordLastUpdated *time.Time `json:"date_password_last_updated,omitempty" gorm:"type:timestamp with time zone;default:null"`
    NeedUpdatePassword      bool       `json:"need_update_password" gorm:"type:boolean;not null"`
    DingtalkID              *string    `json:"dingtalk_id,omitempty" gorm:"type:varchar(128);default:null;unique"`
    WecomID                 *string    `json:"wecom_id,omitempty" gorm:"type:varchar(128);default:null;unique"`
    FeishuID                *string    `json:"feishu_id,omitempty" gorm:"type:varchar(128);default:null;unique"`
    IsServiceAccount        bool       `json:"is_service_account" gorm:"type:boolean;not null"`
    DateUpdated             time.Time  `json:"date_updated" gorm:"type:timestamp with time zone;not null"`
    UpdatedBy               string     `json:"updated_by" gorm:"type:varchar(30);not null"`
    DateAPIKeyLastUsed      *time.Time `json:"date_api_key_last_used,omitempty" gorm:"type:timestamp with time zone;default:null"`
    SlackID                 *string    `json:"slack_id,omitempty" gorm:"type:varchar(128);default:null;unique"`
    LarkID                  *string    `json:"lark_id,omitempty" gorm:"type:varchar(128);default:null;unique"`
}
