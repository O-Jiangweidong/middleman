package middleware

import (
	"fmt"
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"

	"middleman/pkg/consts"
	"middleman/pkg/database"
	mm "middleman/pkg/middleware/models"
)

func DatabaseMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		var dbName string
		var server mm.JumpServer

		// TODO master 角色可以指定数据库,先注释，功能完善放开
		//if server.Role == models.RoleMaster {
		//    dbName = c.GetHeader("SLAVE-NAME")
		//}
		defaultDB := database.GetDBManager().GetDefaultDB()
		dbName = c.GetHeader("SLAVE-NAME") // TODO 测试完了之后删除
		if dbName == "" {
			authHeader := c.GetHeader("Authorization")
			parts := strings.Split(authHeader, " ")
			credentials := strings.Split(parts[1], ":")
			defaultDB.Model(&mm.JumpServer{}).
				Where("access_key = ? AND secret_key = ?", credentials[0], credentials[1]).
				Find(&server)
		} else {
			defaultDB.Model(&mm.JumpServer{}).
				Where("name = ?", dbName).Find(&server)
		}

		dbName = string(server.Name)
		if dbName == "" {
			dbName = database.DefaultDBName
		}
		if dbName != database.DefaultDBName {
			var count int64
			defaultDB.Model(&mm.JumpServer{}).
				Where("name = ?", dbName).Count(&count).Find(&server)
			if count < 1 {
				c.AbortWithStatusJSON(http.StatusBadRequest, gin.H{
					"error": fmt.Sprintf("Invalid branch node name"),
					"code":  40002,
				})
				return
			}
		}
		db, err := database.GetDBManager().GetDB(dbName)
		if err != nil {
			c.AbortWithStatusJSON(http.StatusBadRequest, gin.H{
				"error": fmt.Sprintf("Failed to get database name: %v", err),
				"code":  40001,
			})
			return
		}

		c.Set(consts.DBContextKey, db)
		c.Set(consts.DBInfoContextKey, server)
		c.Next()
	}
}
