package middleware

import (
	"fmt"
	"github.com/gin-gonic/gin"
	"net/http"

	"middleman/pkg/consts"
	"middleman/pkg/database"
	mm "middleman/pkg/middleware/models"
)

func DatabaseMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		server := c.MustGet(consts.AuthDBInfoContextKey).(mm.JumpServer)
		if server.Role == mm.RoleMaster {
			dbName := c.GetHeader("SLAVE-NAME")
			defaultDB := database.GetDBManager().GetDefaultDB()
			defaultDB.Model(&mm.JumpServer{}).
				Where("name = ? AND role = ?", dbName, mm.RoleSlave).Find(&server)
		}

		if server.Name == "" {
			c.AbortWithStatusJSON(http.StatusBadRequest, gin.H{
				"error": fmt.Sprintf("Invalid branch node name"),
				"code":  40002,
			})
			return
		}

		_, err := database.GetDBManager().GetDB(string(server.Name))
		if err != nil {
			c.AbortWithStatusJSON(http.StatusBadRequest, gin.H{
				"error": fmt.Sprintf("Failed to get database name: %v", err),
				"code":  40001,
			})
			return
		}

		c.Set(consts.DBInfoContextKey, server)
		c.Next()
	}
}
