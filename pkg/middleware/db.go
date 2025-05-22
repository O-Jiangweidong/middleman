package middleware

import (
	"fmt"
	"middleman/pkg/database/models"
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"

	"middleman/pkg/consts"
	"middleman/pkg/database"
)

func DatabaseMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		var dbName string
		var server models.JumpServer
		authHeader := c.GetHeader("Authorization")
		parts := strings.Split(authHeader, " ")
		credentials := strings.Split(parts[1], ":")
		defaultDB := database.GetDBManager().GetDefaultDB()

		defaultDB.Model(&models.JumpServer{}).
			Where("access_key = ? AND secret_key = ?", credentials[0], credentials[1]).
			Find(&server)
		dbName = string(server.Name)
		// TODO master 角色可以指定数据库,先注释，功能完善放开
		//if server.Role == models.RoleMaster {
		//    dbName = c.GetHeader("SLAVE-NAME")
		//}
		dbName = c.GetHeader("SLAVE-NAME") // TODO 测试完了之后删除
		if dbName == "" {
			dbName = database.DefaultDBName
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
		c.Set(consts.DBNameContextKey, dbName)
		c.Next()
	}
}
