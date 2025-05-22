package pkg

import (
	"encoding/json"
	"fmt"
	"reflect"

	"gorm.io/gorm"
)

func saveToDB(db *gorm.DB, instance interface{}, dataItems []interface{}) error {
	instanceType := reflect.TypeOf(instance)
	for _, dataItem := range dataItems {
		newInstance := reflect.New(instanceType).Interface()
		dataJSON, err := json.Marshal(dataItem)
		if err != nil {
			return err
		}

		if err = json.Unmarshal(dataJSON, newInstance); err != nil {
			return err
		}

		idField := reflect.ValueOf(newInstance).Elem().FieldByName("ID")
		if !idField.IsValid() {
			return fmt.Errorf("模型 %v 缺少ID字段", instanceType.Name())
		}

		var count int64
		query := db.Model(newInstance)
		if err = query.Where("id = ?", idField.Interface()).
			Count(&count).Error; err != nil {
			return err
		}

		if count > 0 {
			if err = db.Model(newInstance).Updates(newInstance).Error; err != nil {
				return err
			}
		} else {
			if err = db.Create(newInstance).Error; err != nil {
				return err
			}
		}
	}
	return nil
}
