package models

import (
    "database/sql/driver"
    "encoding/json"
    "fmt"
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
