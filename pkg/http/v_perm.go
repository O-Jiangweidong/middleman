package pkg

import (
	"fmt"
	"github.com/gin-gonic/gin"
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
		Preload("Users").Preload("UserGroups").
		Preload("Nodes").Preload("Assets")
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

func (h *ResourcesHandler) deletePerm(id string) (err error) {
	err = h.db.Where("id = ?", id).Delete(&models.AssetPermission{}).Error
	if err != nil {
		return err
	}
	go h.jmsClient.DeletePerm(id)
	return nil
}

func (h *ResourcesHandler) updatePerm(c *gin.Context, id string) (err error) {
	type reqPerm struct {
	}
	var req reqPerm
	if err = c.ShouldBindJSON(&req); err != nil {
		return err
	}
	return nil
}
