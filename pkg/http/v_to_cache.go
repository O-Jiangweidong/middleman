package pkg

import (
	"fmt"
	"github.com/gin-gonic/gin"
	"middleman/pkg/database/models"
	"middleman/pkg/utils"
)

func saveOrgToCache(c *gin.Context, cache *utils.CacheManager, dbName string) (err error) {
	var newOrgs []models.Organization
	cacheKey := fmt.Sprintf("%s_organization", dbName)
	err = c.ShouldBindJSON(&newOrgs)
	if err != nil {
		return err
	}
	var rawOrgs []models.Organization
	err = cache.Get(cacheKey, &rawOrgs)
	if err == nil {
		return err
	}
	existIDs := make(map[string]interface{})
	for _, item := range rawOrgs {
		existIDs[item.ID] = item
	}

	for _, item := range newOrgs {
		if _, ok := existIDs[item.ID]; !ok {
			rawOrgs = append(rawOrgs, item)
		}
	}

	err = cache.Set(cacheKey, rawOrgs, 0)
	if err != nil {
		return err
	}
	return nil
}
