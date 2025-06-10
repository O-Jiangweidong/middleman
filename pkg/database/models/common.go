package models

import (
	"database/sql/driver"
	"encoding/json"
	"fmt"
	"time"
)

const (
	DefaultOrgID = "00000000-0000-0000-0000-000000000002"
)

const utcFormat = "2006-01-02 15:04:05.000000-07:00"

var (
	supportedTimeFormat = []string{
		utcFormat, time.RFC3339, "2006/01/02 15:04:05 +0800",
	}
)

type OnlyID struct {
	ID string `json:"id"`
}

type LabelValue struct {
	Value interface{} `json:"value,omitempty"`
	Label string      `json:"label,omitempty"`
}

type JSONMap map[string]interface{}

func (j *JSONMap) Value() (driver.Value, error) {
	if j == nil {
		return nil, nil
	}
	return json.Marshal(j)
}

func (j *JSONMap) Scan(value interface{}) error {
	if value == nil {
		*j = make(JSONMap)
		return nil
	}
	var bytes []byte
	switch v := value.(type) {
	case []byte:
		bytes = v
	case string:
		bytes = []byte(v)
	default:
		return fmt.Errorf("不支持的类型: %T", value)
	}
	if len(bytes) == 0 {
		*j = make(JSONMap)
		return nil
	}
	return json.Unmarshal(bytes, j)
}

type StringArray []string

func (sa StringArray) Value() (driver.Value, error) {
	if sa == nil {
		return nil, nil
	}
	return json.Marshal(sa)
}

func (sa *StringArray) Scan(value interface{}) error {
	return json.Unmarshal(value.([]byte), &sa)
}

type UTCTime struct {
	time.Time
}

func (t *UTCTime) MarshalJSON() ([]byte, error) {
	return []byte(fmt.Sprintf(`"%s"`, t.Format(utcFormat))), nil
}

func (t *UTCTime) UnmarshalJSON(data []byte) (err error) {
	for _, format := range supportedTimeFormat {
		if parseTime, err := time.Parse(
			fmt.Sprintf(`"%s"`, format), string(data),
		); err == nil {
			t.Time = parseTime
			return nil
		}
	}
	return nil
}

func (t *UTCTime) Value() (driver.Value, error) {
	if t == nil || t.IsZero() {
		return time.Now().UTC().Format(time.RFC3339), nil
	}
	return t.UTC().Format(time.RFC3339), nil
}

func (t *UTCTime) Scan(value interface{}) error {
	if value == nil {
		t.Time = time.Time{}
		return nil
	}

	var err error
	switch v := value.(type) {
	case time.Time:
		t.Time = v.UTC()
	case string:
		t.Time, err = time.Parse("2006-01-02 15:04:05 +0000 UTC", v)
		if err != nil {
			return err
		}
		t.Time = t.UTC()
	default:
		return fmt.Errorf("不支持的时间类型: %T", value)
	}
	return nil
}
