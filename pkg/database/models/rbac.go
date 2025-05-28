package models

type RbacRoleBinding struct {
	ID          string   `json:"id" gorm:"type:uuid;primaryKey;not null"`
	Scope       string   `json:"scope" gorm:"type:varchar(128);not null"`
	OrgID       string   `json:"org_id" gorm:"type:uuid;default:null;index"`
	RoleID      string   `json:"role_id" gorm:"type:uuid;not null;index"`
	UserID      string   `json:"user_id" gorm:"type:uuid;not null;index"`
	Comment     string   `json:"comment" gorm:"type:text;not null"`
	CreatedBy   string   `json:"created_by" gorm:"type:varchar(128);default:null"`
	UpdatedBy   string   `json:"updated_by" gorm:"type:varchar(128);default:null"`
	DateCreated *UTCTime `json:"date_created" gorm:"type:timestamp with time zone;default:null"`
	DateUpdated *UTCTime `json:"date_updated" gorm:"type:timestamp with time zone;default null"`

	Role RbacRole `gorm:"foreignKey:RoleID;references:ID"`
	User User     `gorm:"foreignKey:UserID;references:ID"`
}

type RbacRole struct {
	ID          string   `json:"id" gorm:"type:char(36);primaryKey;not null"`
	Name        string   `json:"name" gorm:"type:varchar(128);not null"`
	Scope       string   `json:"scope" gorm:"type:varchar(128);not null"`
	Builtin     bool     `json:"builtin" gorm:"type:boolean;not null"`
	Comment     string   `json:"comment" gorm:"type:text;not null"`
	CreatedBy   *string  `json:"created_by" gorm:"type:varchar(128);default:null"`
	UpdatedBy   *string  `json:"updated_by" gorm:"type:varchar(128);default:null"`
	DateCreated *UTCTime `json:"date_created" gorm:"type:timestamp with time zone;default:null"`
	DateUpdated *UTCTime `json:"date_updated" gorm:"type:timestamp with time zone;default:null"`

	Users []User `gorm:"many2many:rbac_role_bindings;joinForeignKey:RoleID;joinReferences:UserID;constraint:OnDelete:CASCADE"`
}
