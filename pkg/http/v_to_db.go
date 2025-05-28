package pkg

import (
	"encoding/json"
	"gorm.io/gorm"

	"middleman/pkg/database/models"
)

func (h *ResourcesHandler) savePlatform(db *gorm.DB, dataItems []interface{}) error {
	for _, dataItem := range dataItems {
		dataJSON, err := json.Marshal(dataItem)
		if err != nil {
			return err
		}

		var platform models.Platform
		if err = json.Unmarshal(dataJSON, &platform); err != nil {
			return err
		}

		var count int64
		if err = db.Model(platform).Where("id = ?", platform.ID).Count(&count).Error; err != nil {
			return err
		}
		if count > 0 {
			if err = db.Model(platform).Updates(&platform).Error; err != nil {
				return err
			}
		} else {
			if err = db.Create(&platform).Error; err != nil {
				return err
			}
		}
	}
	return nil
}
