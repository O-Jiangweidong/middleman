package models

import "time"

type Organization struct {
	ID          string     `json:"id"`
	Name        string     `json:"name"`
	CreatedBy   string     `json:"created_by,omitempty"`
	DateCreated *time.Time `json:"date_created,omitempty"`
	Comment     string     `json:"comment"`
	Builtin     bool       `json:"builtin"`
	DateUpdated time.Time  `json:"date_updated"`
	UpdatedBy   *string    `json:"updated_by,omitempty"`
}

var DefaultOrg = Organization{
	ID:        "00000000000000000000000000000002",
	Name:      "Default",
	CreatedBy: "System",
	Comment:   "",
	Builtin:   true,
	UpdatedBy: nil,
}

var SystemOrg = Organization{
	ID:        "00000000000000000000000000000004",
	Name:      "SYSTEM",
	CreatedBy: "System",
	Comment:   "",
	Builtin:   true,
	UpdatedBy: nil,
}
