package pkg

import (
	"fmt"
	"net/http"
	"strconv"
	"strings"

	"middleman/pkg/config"
	"middleman/pkg/consts"
	"middleman/pkg/database"
	"middleman/pkg/middleware/models"
	"middleman/pkg/utils"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

const (
	User         = "user"
	Asset        = "asset"
	Platform     = "platform"
	Perm         = "Permission"
	Host         = "host"
	Device       = "device"
	Database     = "database"
	Cloud        = "cloud"
	Web          = "web"
	Gpt          = "gpt"
	Custom       = "custom"
	Organization = "organization"
	Role         = "role"
	UserGroup    = "user_group"
)

type RegisterRequest struct {
	Name           models.NameType `json:"name" binding:"required"`
	Display        string          `json:"display" binding:"required"`
	BootstrapToken string          `json:"bootstrap_token" binding:"required"`
	Role           models.RoleType `json:"role" binding:"required"`
	IgnoreSameName bool            `json:"ignore_same_name"`
	Endpoint       string          `json:"endpoint" binding:"required"`
	PrivateToken   string          `json:"private_token" binding:"required"`
}

func handleRegister(c *gin.Context) {
	conf := config.GetConf()
	var registerRequest RegisterRequest
	if err := c.ShouldBindJSON(&registerRequest); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	if conf.BootstrapToken != registerRequest.BootstrapToken {
		c.JSON(http.StatusForbidden, gin.H{"error": "BootstrapToken does not match"})
		return
	}
	db := database.GetDBManager().GetDefaultDB()
	server := models.JumpServer{
		BaseJumpServer: models.BaseJumpServer{
			Name:     registerRequest.Name,
			Display:  registerRequest.Display,
			Role:     registerRequest.Role,
			Endpoint: registerRequest.Endpoint,
		},
		AccessKey:    utils.GenerateRandomString(36),
		SecretKey:    utils.GenerateRandomString(36),
		PrivateToken: registerRequest.PrivateToken,
	}
	if err := server.Validate(); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	var count int64
	db.Model(&models.JumpServer{}).Where("name = ?", registerRequest.Name).Count(&count)
	if count > 0 && !registerRequest.IgnoreSameName {
		msg := fmt.Sprintf(
			"Name: %s 重复，请修改配置文件 MIDDLEMAN_SERVICE_NAME，并重启重新注册",
			server.Name,
		)
		c.JSON(http.StatusBadRequest, gin.H{"error": msg})
		return
	}
	if count == 0 {
		if err := db.Create(&server).Error; err != nil {
			c.JSON(http.StatusBadRequest, gin.H{
				"error": fmt.Sprintf("创建节点失败: %v", err),
			})
			return
		}

		if server.Role == models.RoleSlave {
			_, err := database.GetDBManager().GetDB(string(server.Name))
			if err != nil {
				c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
				return
			}
		}
	} else {
		err := db.Model(&models.JumpServer{}).
			Where("name = ?", registerRequest.Name).
			Updates(map[string]interface{}{
				"access_key": server.AccessKey,
				"secret_key": server.SecretKey,
				"display":    server.Display,
			}).Error
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}
	}
	c.JSON(http.StatusOK, gin.H{"data": server})
}

func getSlaveNodes(c *gin.Context) {
	db := database.GetDBManager().GetDefaultDB()
	var services []models.JumpServer
	db.Where("role = ?", models.RoleSlave).Find(&services)

	var baseServices []models.BaseJumpServer
	for _, s := range services {
		baseServices = append(baseServices, s.BaseJumpServer)
	}
	c.JSON(http.StatusOK, gin.H{"data": baseServices, "total": len(baseServices)})
}

func getResources(c *gin.Context) {
	db := c.MustGet(consts.DBContextKey).(*gorm.DB)
	dbInfo := c.MustGet(consts.DBInfoContextKey).(models.JumpServer)
	fmt.Println("DB Name:", dbInfo.Name)
	if db == nil {
		c.JSON(http.StatusPreconditionFailed, gin.H{
			"error": "Database not found", "details": "Database not found",
		})
		return
	}

	handle := newResourcesHandler(dbInfo)
	handle.processedParams = map[string]bool{
		"offset": true, "limit": true, "m_type": true, "search": true,
	}

	offset, err := strconv.Atoi(c.DefaultQuery("offset", "0"))
	if err != nil || offset < 0 {
		offset = 0
	}

	limit, err := strconv.Atoi(c.DefaultQuery("limit", "15"))
	if err != nil || limit < 1 || limit > 200 {
		c.JSON(http.StatusBadRequest, gin.H{
			"error":   "Invalid param limit",
			"details": "Param limit must be between 1 and 200",
		})
		return
	}

	var resources interface{}
	var count int64
	resourceType := c.Query("m_type")
	switch resourceType {
	case User:
		resources, count, err = handle.getUsers(c, db, limit, offset)
	case Platform:
		resources, count, err = handle.getPlatforms(c, db, limit, offset)
	default:
		c.JSON(http.StatusBadRequest, gin.H{
			"error":   "Invalid request type",
			"details": fmt.Sprintf("Invalid request type: %s", resourceType),
		})
		return
	}

	if err != nil {
		c.JSON(http.StatusPreconditionFailed, gin.H{
			"error": "Database error", "details": err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{"results": resources, "count": count})
}

type ResourcesHandler struct {
	jmsClient *utils.JumpServer

	processedParams map[string]bool
}

func (h *ResourcesHandler) handleSearch(c *gin.Context, q *gorm.DB, searchFields []string) *gorm.DB {
	search := c.Query("search")
	if search != "" && len(searchFields) > 0 {
		var conditions []string
		var args []interface{}
		for _, f := range searchFields {
			args = append(args, search)
			conditions = append(conditions, fmt.Sprintf("%s = ?", f))
		}

		q = q.Where("("+strings.Join(conditions, " OR ")+")", args...)
	}
	return q
}

func newResourcesHandler(dbInfo models.JumpServer) *ResourcesHandler {
	return &ResourcesHandler{
		jmsClient: utils.NewJumpServer(dbInfo.Endpoint, dbInfo.PrivateToken),
	}
}

func saveResources(c *gin.Context) {
	db := c.MustGet(consts.DBContextKey).(*gorm.DB)
	dbInfo := c.MustGet(consts.DBInfoContextKey).(models.JumpServer)

	handler := newResourcesHandler(dbInfo)

	if db == nil || dbInfo.Name == "" {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": "Database not found", "details": "Database not found",
		})
		return
	}

	resourceType := c.Query("type")

	var err error
	switch resourceType {
	case User:
		err = handler.saveUser(c, db)
	case Role:
		err = handler.saveRole(c, db)
	case UserGroup:
		err = handler.saveUserGroup(c, db)
	case Platform:
		err = handler.savePlatform(c, db)
	case Host:
		err = handler.saveHost(c, db)
	default:
		c.JSON(http.StatusBadRequest, gin.H{
			"error":   "Invalid request type",
			"details": fmt.Sprintf("Invalid request type: %s", resourceType),
		})
		return
	}

	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error":   fmt.Sprintf("Failed to save resource: %v", err.Error()),
			"details": "Database operation failed",
		})
	} else {
		c.JSON(http.StatusCreated, gin.H{
			"message": "Permission created successfully",
		})
	}
}
