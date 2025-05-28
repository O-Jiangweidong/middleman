package pkg

import (
	"fmt"
	"net/http"
	"strconv"
	"strings"

	"middleman/pkg/config"
	"middleman/pkg/consts"
	"middleman/pkg/database"
	"middleman/pkg/database/models"
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

func getAllowedFields(resourceType string) map[string]bool {
	switch resourceType {
	case User:
		return map[string]bool{
			"id":       true,
			"username": true,
			"email":    true,
		}
	case Role:
		return map[string]bool{
			"id":      true,
			"name":    true,
			"scope":   true,
			"builtin": true,
		}
	default:
		return make(map[string]bool)
	}
}

func getSearchFields(resourceType string) []string {
	switch resourceType {
	case User:
		return []string{"username", "name", "email"}
	default:
		return []string{}
	}
}

func getResources(c *gin.Context) {
	db := c.MustGet(consts.DBContextKey).(*gorm.DB)
	dbName := c.MustGet(consts.DBInfoContextKey).(models.JumpServer).Name
	fmt.Println("DB Name:", dbName)
	if db == nil {
		c.JSON(http.StatusPreconditionFailed, gin.H{
			"error": "Database not found", "details": "Database not found",
		})
		return
	}

	processedParams := map[string]bool{
		"offset": true, "limit": true, "type": true, "search": true,
	}

	offset, err := strconv.Atoi(c.DefaultQuery("offset", "0"))
	if err != nil || offset < 0 {
		offset = 0
	}

	limit, err := strconv.Atoi(c.DefaultQuery("limit", "15"))
	if err != nil || limit < 1 || limit > 100 {
		limit = 15
	}

	type_ := c.Query("type")
	resourceMap := map[string]interface{}{
		User:     &[]models.User{},
		Asset:    &[]models.Asset{},
		Platform: &[]models.Platform{},
		Perm:     &[]models.Permission{},
		Host:     &[]models.Host{},
		Device:   &[]models.Device{},
		Database: &[]models.Database{},
		Cloud:    &[]models.Cloud{},
		Web:      &[]models.Web{},
		Gpt:      &[]models.GPT{},
		Custom:   &[]models.Custom{},
		Role:     &[]models.RbacRole{},
	}

	resources, exists := resourceMap[type_]
	if !exists {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid type", "details": type_})
		return
	}

	allowedFields := getAllowedFields(type_)

	q := db.Model(resources)
	for key, values := range c.Request.URL.Query() {
		if processedParams[key] || !allowedFields[key] {
			continue
		}

		if len(values) > 0 {
			q = q.Where(fmt.Sprintf("%s = ?", key), values[len(values)-1])
		}
	}

	search := c.Query("search")
	searchFields := getSearchFields(type_)

	if search != "" && len(searchFields) > 0 {
		var conditions []string
		var args []interface{}
		for _, f := range searchFields {
			args = append(args, search)
			conditions = append(conditions, fmt.Sprintf("%s = ?", f))
		}

		q = q.Where("("+strings.Join(conditions, " OR ")+")", args...)
	}

	var count int64
	result := q.Count(&count).Limit(limit).Offset(offset).Find(resources)
	if result.Error != nil {
		c.JSON(http.StatusPreconditionFailed, gin.H{
			"error": "Database error", "details": result.Error.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{"results": resources, "count": count})
}

type ResourcesHandler struct {
	jmsClient *utils.JumpServer
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
