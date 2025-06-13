package pkg

import (
	"fmt"
	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"gorm.io/gorm"
	"middleman/pkg/database/models"
)

func (h *ResourcesHandler) savePlatform(c *gin.Context) (err error) {
	var platforms []models.Platform
	if err = c.ShouldBindJSON(&platforms); err != nil {
		return err
	}
	for _, platform := range platforms {
		var count int64
		if err = h.db.Model(platform).Where("id = ?", platform.ID).
			Count(&count).Error; err != nil {
			return err
		}
		if count > 0 {
			if err = h.db.Model(platform).Omit("id").Updates(&platform).Error; err != nil {
				return err
			}
		} else {
			if err = h.db.Create(&platform).Error; err != nil {
				return err
			}
		}
	}

	return nil
}

func (h *ResourcesHandler) saveHost(c *gin.Context) (ids []string, err error) {
	var hosts []models.Host
	if err = c.ShouldBindJSON(&hosts); err != nil {
		return nil, err
	}
	for _, host := range hosts {
		var count int64
		if err = h.db.Model(host).Where("asset_ptr_id = ?", host.Asset.ID).
			Count(&count).Error; err != nil {
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
			if err = h.db.Model(host).Omit("id").Updates(&host).Error; err != nil {
				return nil, err
			}
		} else {
			err = h.db.Transaction(func(tx *gorm.DB) error {
				var txErr error

				var nameCount int64
				if txErr = tx.Model(models.Asset{}).
					Where("name = ? AND org_id = ?", host.Asset.Name, host.Asset.OrgID).
					Limit(1).Count(&nameCount).Error; txErr != nil {
					return txErr
				}

				if nameCount > 0 {
					return fmt.Errorf("name %s already exists", host.Asset.Name)
				}

				if txErr = tx.Create(&host).Error; txErr != nil {
					return txErr
				}

				type Relation struct {
					AssetID string `gorm:"column:asset_id"`
					NodeID  string `gorm:"column:node_id"`
				}
				var relations []Relation
				for _, nodeID := range host.Asset.NodeIds {
					relations = append(relations, Relation{
						NodeID: nodeID, AssetID: host.AssetPtrID,
					})
				}
				if txErr = tx.Table("assets_asset_nodes").
					CreateInBatches(relations, 100).Error; txErr != nil {
					return txErr
				}
				return nil
			})
			if err != nil {
				return nil, err
			}

			ids = append(ids, host.AssetPtrID)
			go h.jmsClient.CreateAsset(host)
		}
	}

	return ids, nil
}

func (h *ResourcesHandler) getPlatforms(c *gin.Context, limit, offset int) (interface{}, int64, error) {
	var err error
	var platforms []models.Platform
	queryFields := map[string]bool{
		"name":     true,
		"type":     true,
		"category": true,
	}
	q := h.db.Model(&models.Platform{})
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
	if err = q.Count(&count).Limit(limit).Offset(offset).Find(&platforms).Error; err != nil {
		return nil, 0, err
	}
	return platforms, count, nil
}

func (h *ResourcesHandler) getAssets(c *gin.Context, limit, offset int, category string) (interface{}, int64, error) {
	var err error
	var assets []models.Asset
	queryFields := map[string]bool{
		"id":        true,
		"address":   true,
		"name":      true,
		"is_active": true,
	}

	nodeID := c.Query("node_id")
	if nodeID == "" {
		nodeID = c.Query("node_id")
	}

	q := h.db.Model(&models.Asset{}).Preload("Platform")
	if nodeID != "" {
		q = q.Where("? IN (SELECT node_id FROM assets_asset_nodes WHERE asset_id = assets.id)", nodeID)
	}
	q.Preload("Nodes", func(db *gorm.DB) *gorm.DB {
		return db.Select("id", "value", "full_value")
	})

	if category != "" {
		q = q.Joins("JOIN platforms ON platforms.id = assets.platform_id").
			Where("platforms.category = ?", category)
	}

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

	type respAsset struct {
		models.Asset

		Nodes        []models.SimpleNode `json:"nodes"`
		NodesDisplay []string            `json:"nodes_display"`
	}
	newAssets := make([]respAsset, 0, len(assets))
	for _, asset := range assets {
		p := asset.Platform
		asset.Category = models.LabelValue{Label: p.Category, Value: p.Category}
		asset.Type = models.LabelValue{Label: p.Type, Value: p.Type}
		asset.Accounts = nil

		nodesDisplay := make([]string, 0, len(asset.Nodes))
		newNodes := make([]models.SimpleNode, 0, len(asset.Nodes))
		for _, n := range asset.Nodes {
			nodesDisplay = append(nodesDisplay, n.FullValue)
			newNodes = append(newNodes, models.SimpleNode{ID: n.ID, Name: n.Value})
		}
		newAssets = append(newAssets, respAsset{
			Asset: asset, NodesDisplay: nodesDisplay, Nodes: newNodes,
		})
	}
	return newAssets, count, nil
}

func (h *ResourcesHandler) getAccounts(c *gin.Context, limit, offset int) (interface{}, int64, error) {
	var err error
	var accounts []models.Account
	queryFields := map[string]bool{
		"id":          true,
		"name":        true,
		"username":    true,
		"secret_type": true,
	}
	q := h.db.Model(&models.Account{}).Preload("Asset")
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

func (h *ResourcesHandler) deleteAsset(id, cacheKey string) (err error) {
	err = h.db.Where("id = ?", id).Delete(&models.Asset{}).Error
	if err != nil {
		return err
	}
	go h.jmsClient.RemoveAsset(id, cacheKey)
	return nil
}
