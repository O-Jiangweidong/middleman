package pkg

import (
	"fmt"
	"net/http"
	"reflect"
	"strconv"

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
	Perm         = "Permission"
	Host         = "host"
	Device       = "device"
	Database     = "database"
	Cloud        = "cloud"
	Web          = "web"
	Gpt          = "gpt"
	Custom       = "custom"
	Organization = "organization"
)

type RegisterRequest struct {
	Name           models.NameType `json:"name" binding:"required"`
	Display        string          `json:"display" binding:"required"`
	BootstrapToken string          `json:"bootstrap_token" binding:"required"`
	Role           models.RoleType `json:"role" binding:"required"`
	IgnoreSameName bool            `json:"ignore_same_name"`
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
			Name:    registerRequest.Name,
			Display: registerRequest.Display,
			Role:    registerRequest.Role,
		},
		AccessKey: utils.GenerateRandomString(36),
		SecretKey: utils.GenerateRandomString(36),
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
	c.JSON(http.StatusOK, gin.H{"data": baseServices})
}

type ResourceRequest struct {
	Type string        `json:"type" binding:"required"`
	Data []interface{} `json:"data" binding:"required"`
}

func getResources(c *gin.Context) {
	db := c.MustGet(consts.DBContextKey).(*gorm.DB)
	if db == nil {
		c.JSON(http.StatusPreconditionFailed, gin.H{
			"error": "Database not found", "details": "Database not found",
		})
		return
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
		"user":     &[]models.User{},
		"perm":     &[]models.Permission{},
		"host":     &[]models.Host{},
		"device":   &[]models.Device{},
		"database": &[]models.Database{},
		"cloud":    &[]models.Cloud{},
		"web":      &[]models.Web{},
		"gpt":      &[]models.GPT{},
		"custom":   &[]models.Custom{},
	}

	resources, exists := resourceMap[type_]
	if !exists {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid type", "details": type_})
		return
	}

	modelType := reflect.TypeOf(resources).Elem().Elem()
	modelValue := reflect.New(modelType).Interface()

	var count int64
	result := db.Model(modelValue).
		Count(&count).Limit(limit).Offset(offset).Find(resources)

	if result.Error != nil {
		c.JSON(http.StatusPreconditionFailed, gin.H{
			"error": "Database error", "details": result.Error.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{"data": resources, "count": count})
}

func saveResources(c *gin.Context) {
	db := c.MustGet(consts.DBContextKey).(*gorm.DB)
	dbName := c.MustGet(consts.DBNameContextKey).(string)
	cache := utils.GetCache()

	if db == nil || dbName == "" {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": "Database not found", "details": "Database not found",
		})
		return
	}
	var req ResourceRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "Invalid request format", "details": err.Error(),
		})
		return
	}

	var err error
	var instance interface{}
	switch req.Type {
	case User:
		instance = models.User{}
		err = saveToDB(db, instance, req.Data)
	case Perm:
		instance = models.Permission{}
		err = saveToDB(db, &instance, req.Data)
	case Host:
		instance = models.Host{}
		err = saveToDB(db, &instance, req.Data)
	case Device:
		instance = models.Device{}
		err = saveToDB(db, &instance, req.Data)
	case Database:
		instance = models.Database{}
		err = saveToDB(db, &instance, req.Data)
	case Cloud:
		instance = models.Cloud{}
		err = saveToDB(db, &instance, req.Data)
	case Web:
		instance = models.Web{}
		err = saveToDB(db, &instance, req.Data)
	case Gpt:
		instance = models.GPT{}
		err = saveToDB(db, instance, req.Data)
	case Custom:
		instance = models.Custom{}
		err = saveToDB(db, instance, req.Data)
	case Organization:
		err = saveOrgToCache(c, cache, dbName)
	default:
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request type", "details": req.Type})
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
			"data":    instance,
		})
	}
}
