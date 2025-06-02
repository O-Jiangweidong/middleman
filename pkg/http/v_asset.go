package pkg

import (
	"fmt"
	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"gorm.io/gorm"

	"middleman/pkg/database/models"
)

func (h *ResourcesHandler) savePlatform(c *gin.Context, db *gorm.DB) (err error) {
	var platforms []models.Platform
	if err = c.ShouldBindJSON(&platforms); err != nil {
		return err
	}
	for _, platform := range platforms {
		var count int64
		if err = db.Model(platform).Where("id = ?", platform.ID).Count(&count).Error; err != nil {
			return err
		}
		if count > 0 {
			if err = db.Model(platform).Omit("id").Updates(&platform).Error; err != nil {
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

func (h *ResourcesHandler) saveHost(c *gin.Context, db *gorm.DB) (ids []string, err error) {
	var hosts []models.Host
	if err = c.ShouldBindJSON(&hosts); err != nil {
		return nil, err
	}
	for _, host := range hosts {
		var count int64
		if err = db.Model(host).Where("asset_ptr_id = ?", host.Asset.ID).Count(&count).Error; err != nil {
			return nil, err
		}

		var newAccounts []models.Account
		for _, account := range host.Asset.Accounts {
			if account.ID == "" {
				account.ID = uuid.New().String()
			}
			account.Connectivity = "-"
			account.OrgID = models.DefaultOrgID
			newAccounts = append(newAccounts, account)
		}
		host.Asset.Accounts = newAccounts

		if count > 0 {
			if err = db.Model(host).Omit("id").Updates(&host).Error; err != nil {
				return nil, err
			}
		} else {
			if err = db.Create(&host).Error; err != nil {
				return nil, err
			}
			if err = h.jmsClient.CreateAsset(host); err != nil {
				return nil, err
			}
			ids = append(ids, host.AssetPtrID)
		}
	}

	return ids, nil
}

func (h *ResourcesHandler) getPlatforms(c *gin.Context, db *gorm.DB, limit, offset int) (interface{}, int64, error) {
	var err error
	var platforms []models.Platform
	queryFields := map[string]bool{
		"name":     true,
		"type":     true,
		"category": true,
	}
	q := db.Model(&models.Platform{})
	for key, values := range c.Request.URL.Query() {
		if h.processedParams[key] || !queryFields[key] {
			continue
		}

		if len(values) > 0 {
			q = q.Where(fmt.Sprintf("%s = ?", key), values[len(values)-1])
		}
	}

	searchFields := []string{"name", "type", "category"}
	q = h.handleSearch(c, q, searchFields)

	var count int64
	if err = q.Debug().Count(&count).Limit(limit).Offset(offset).Find(&platforms).Error; err != nil {
		return nil, 0, err
	}
	return platforms, count, nil
}

func (h *ResourcesHandler) getAssets(c *gin.Context, db *gorm.DB, limit, offset int) (interface{}, int64, error) {
	var err error
	var assets []models.Asset
	queryFields := map[string]bool{
		"id":        true,
		"address":   true,
		"name":      true,
		"is_active": true,
	}
	q := db.Model(&models.Asset{}).Preload("Platform")
	for key, values := range c.Request.URL.Query() {
		if h.processedParams[key] || !queryFields[key] {
			continue
		}

		if len(values) > 0 {
			q = q.Where(fmt.Sprintf("%s = ?", key), values[len(values)-1])
		}
	}

	searchFields := []string{"address", "name"}
	q = h.handleSearch(c, q, searchFields)

	var count int64
	if err = q.Count(&count).Limit(limit).Offset(offset).Find(&assets).Error; err != nil {
		return nil, 0, err
	}

	var newAssets []models.Asset
	for _, asset := range assets {
		p := asset.Platform
		asset.Category = models.LabelValue{Label: p.Category, Value: p.Category}
		asset.Type = models.LabelValue{Label: p.Type, Value: p.Type}
		asset.Accounts = nil
		newAssets = append(newAssets, asset)
	}
	return newAssets, count, nil
}

func (h *ResourcesHandler) getAccounts(c *gin.Context, db *gorm.DB, limit, offset int) (interface{}, int64, error) {
	var err error
	var accounts []models.Account
	queryFields := map[string]bool{
		"id":          true,
		"name":        true,
		"username":    true,
		"secret_type": true,
	}
	q := db.Model(&models.Account{}).Preload("Asset")
	for key, values := range c.Request.URL.Query() {
		if h.processedParams[key] || !queryFields[key] {
			continue
		}

		if len(values) > 0 {
			q = q.Where(fmt.Sprintf("%s = ?", key), values[len(values)-1])
		}
	}

	searchFields := []string{"username", "name"}
	q = h.handleSearch(c, q, searchFields)

	var count int64
	if err = q.Count(&count).Limit(limit).Offset(offset).Find(&accounts).Error; err != nil {
		return nil, 0, err
	}
	return accounts, count, nil
}

func (h *ResourcesHandler) deleteAsset(id string, db *gorm.DB) (err error) {
	return db.Where("id = ?", id).Delete(&models.Asset{}).Error
}
