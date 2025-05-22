package models

import (
	"database/sql/driver"
	"encoding/json"
	"errors"
	"fmt"
	"time"
)

const utcFormat = "2006-01-02 15:04:05 -0700"

var (
	supportedTimeFormat = []string{
		"2006/01/02 15:04:05 -0700",
		utcFormat,
		time.RFC3339,
	}
)

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

type UTCTime struct {
	time.Time
}

func (t *UTCTime) MarshalJSON() ([]byte, error) {
	return []byte(fmt.Sprintf(`"%s"`, t.Format(utcFormat))), nil
}

func (t *UTCTime) UnmarshalJSON(data []byte) (err error) {
	for _, format := range supportedTimeFormat {
		if parseTime, err := time.Parse(fmt.Sprintf(
			`"%s"`, format), string(data)); err == nil {
			t.Time = parseTime
			return nil
		}
	}
	return fmt.Errorf("%w: %s", errors.New("unsupported time format"), data)
}

func (t *UTCTime) Value() (driver.Value, error) {
	if t.IsZero() {
		return nil, nil
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
