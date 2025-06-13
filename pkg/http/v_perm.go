package pkg

import (
	"fmt"
	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
	"strings"

	"middleman/pkg/database/models"
)

func (h *ResourcesHandler) savePerm(c *gin.Context) (ids []string, err error) {
	var perms []models.AssetPermission
	if err = c.ShouldBindJSON(&perms); err != nil {
		return nil, err
	}
	for _, perm := range perms {
		perm.OrgID = models.DefaultOrgID

		var users []models.User
		if len(perm.UserIds) > 0 {
			h.db.Model(&users).Where("id IN ?", perm.UserIds).Find(&users)
		}
		perm.Users = users

		var userGroups []models.UserGroup
		if len(perm.UserGroupIds) > 0 {
			h.db.Model(&userGroups).Where("id IN ?", perm.UserGroupIds).Find(&userGroups)
		}
		perm.UserGroups = userGroups

		var assets []models.Asset
		if len(perm.AssetIds) > 0 {
			h.db.Model(&assets).Where("id IN ?", perm.AssetIds).Find(&assets)
		}
		perm.Assets = assets

		var nodes []models.Node
		if len(perm.NodeIds) > 0 {
			h.db.Model(&nodes).Where("id IN ?", perm.NodeIds).Find(&nodes)
		}
		perm.Nodes = nodes

		var count int64
		if err = h.db.Model(perm).Where("id = ?", perm.ID).Count(&count).Error; err != nil {
			return nil, err
		}

		if count > 0 {
			if err = h.db.Model(perm).Omit("id").Updates(&perm).Error; err != nil {
				return nil, err
			}
		} else {
			if err = h.db.Create(&perm).Error; err != nil {
				return nil, err
			}
			ids = append(ids, perm.ID)
			go h.jmsClient.CreatePerm(perm.ToJms())
		}
	}

	return ids, nil
}

func (h *ResourcesHandler) getPerms(c *gin.Context, limit, offset int) (interface{}, int64, error) {
	var err error
	var perms []models.AssetPermission
	queryFields := map[string]bool{
		"id":        true,
		"name":      true,
		"is_active": true,
	}
	q := h.db.Model(&models.AssetPermission{}).
		Preload("Users", func(db *gorm.DB) *gorm.DB {
			return db.Select("id, name, username")
		}).
		Preload("UserGroups", func(db *gorm.DB) *gorm.DB {
			return db.Select("id, name")
		}).
		Preload("Nodes", func(db *gorm.DB) *gorm.DB {
			return db.Select("id, value, full_value")
		}).
		Preload("Assets", func(db *gorm.DB) *gorm.DB {
			return db.Select("id, name, address")
		})
	for key, values := range c.Request.URL.Query() {
		if h.processedParams[key] || !queryFields[key] {
			continue
		}

		if len(values) > 0 {
			q = q.Where(fmt.Sprintf("%s = ?", key), values[len(values)-1])
		}
	}

	searchFields := []string{"name"}
	q = h.handleSearch(c, q, searchFields)

	var count int64
	if err = q.Count(&count).Limit(limit).Offset(offset).Find(&perms).Error; err != nil {
		return nil, 0, err
	}

	for i := range perms {
		perms[i].Valid = perms[i].IsValid()
	}
	return perms, count, nil
}

func (h *ResourcesHandler) deletePerm(id, cacheKey string) (err error) {
	err = h.db.Where("id = ?", id).Delete(&models.AssetPermission{}).Error
	if err != nil {
		return err
	}
	go h.jmsClient.DeletePerm(id, cacheKey)
	return nil
}

func (h *ResourcesHandler) permRelation(
	name, permId, tableName, fieldName string, relationIds []string,
	model interface{}, db *gorm.DB,
) (err error) {
	if len(relationIds) < 1 {
		if err = db.Table(tableName).Where("assetpermission_id = ?", permId).
			Delete(nil).Error; err != nil {
			return err
		}
		return nil
	}

	var ids []string
	if len(relationIds) > 0 {
		err = h.db.Model(&model).Select("id").
			Where("id IN ?", relationIds).Pluck("id", &ids).Error
		if err != nil {
			return err
		}
	}
	if len(relationIds) != len(ids) {
		return fmt.Errorf("there are illegal ID in param %s", name)
	}

	if err = db.Table(tableName).
		Where(fmt.Sprintf("assetpermission_id = ? AND %s NOT IN (?)", fieldName), permId, ids).
		Delete(nil).Error; err != nil {
		return err
	}

	placeholders := make([]string, len(ids))
	values := make([]interface{}, 0, len(ids)*2)
	for i, id := range ids {
		placeholders[i] = "(?, ?)"
		values = append(values, permId, id)
	}
	query := fmt.Sprintf(
		"INSERT INTO %s (assetpermission_id, %s) VALUES %s ON CONFLICT DO NOTHING",
		tableName, fieldName, strings.Join(placeholders, ", "),
	)
	if err = db.Exec(query, values...).Error; err != nil {
		return err
	}

	return nil
}

func (h *ResourcesHandler) updatePerm(c *gin.Context, id string) (err error) {
	var perm models.AssetPermission
	if err = c.ShouldBindJSON(&perm); err != nil {
		return err
	}

	var count int64
	if err = h.db.Model(perm).Where("id = ?", id).Limit(1).Count(&count).Error; err != nil {
		return err
	}
	if count != 1 {
		return fmt.Errorf("permission %s not found", perm.ID)
	}

	perm.OrgID = models.DefaultOrgID
	err = h.db.Transaction(func(tx *gorm.DB) error {
		if err = h.db.Model(perm).
			Omit("id", "Users", "UserGroups", "Assets", "Nodes").
			Updates(&perm).Error; err != nil {
			return err
		}

		// User
		if err = h.permRelation(
			"users", id, "perms_assetpermission_users",
			"user_id", perm.UserIds, models.User{}, tx,
		); err != nil {
			return err
		}

		// UserGroup
		if err = h.permRelation(
			"user_groups", id, "perms_assetpermission_user_groups",
			"usergroup_id", perm.UserGroupIds, models.UserGroup{}, tx,
		); err != nil {
			return err
		}

		// Asset
		if err = h.permRelation(
			"assets", id, "perms_assetpermission_assets",
			"asset_id", perm.AssetIds, models.Asset{}, tx,
		); err != nil {
			return err
		}

		// Node
		if err = h.permRelation(
			"nodes", id, "perms_assetpermission_nodes",
			"node_id", perm.NodeIds, models.Node{}, tx,
		); err != nil {
			return err
		}

		h.jmsClient.UpdatePerm(perm.ToJms())
		return nil
	})
	return err
}
