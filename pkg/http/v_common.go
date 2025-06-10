package pkg

import (
	"errors"
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
	User          = "user"
	Asset         = "asset"
	Node          = "node"
	ChildrenNode  = "children_node"
	NodeWithAsset = "node_with_assets"
	Account       = "account"
	Platform      = "platform"
	Permission    = "perm"
	Host          = "host"
	Device        = "device"
	Database      = "database"
	Cloud         = "cloud"
	Web           = "web"
	Gpt           = "gpt"
	Custom        = "custom"
	Organization  = "organization"
	Role          = "role"
	UserGroup     = "user_group"
	UserUnblock   = "user_unblock"
	UserResetMFA  = "user_reset_mfa"
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
	dbInfo := c.MustGet(consts.DBInfoContextKey).(models.JumpServer)
	fmt.Println("DB Name:", dbInfo.Name)

	handle, err := newResourcesHandler(dbInfo)
	if err != nil {
		c.JSON(http.StatusPreconditionFailed, gin.H{
			"error": "Database init failed", "details": "Database init failed",
		})
		return
	}
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
		resources, count, err = handle.getUsers(c, limit, offset)
	case Platform:
		resources, count, err = handle.getPlatforms(c, limit, offset)
	case Asset:
		resources, count, err = handle.getAssets(c, limit, offset, "")
	case Host, Web, Device, Database, Custom, Gpt, Cloud:
		resources, count, err = handle.getAssets(c, limit, offset, resourceType)
	case Account:
		resources, count, err = handle.getAccounts(c, limit, offset)
	case Permission:
		resources, count, err = handle.getPerms(c, limit, offset)
	case Node:
		resources, count, err = handle.getNodes(c, limit, offset)
	case ChildrenNode:
		resources, count, err = handle.getChildrenNodes(c)
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

	if count == -1 {
		c.JSON(http.StatusOK, resources)
		return
	}

	c.JSON(http.StatusOK, gin.H{"results": resources, "count": count})
}

type ResourcesHandler struct {
	jmsClient *utils.JumpServer
	db        *gorm.DB
	dbName    string

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

func newResourcesHandler(dbInfo models.JumpServer) (*ResourcesHandler, error) {
	db, err := database.GetDBManager().GetDB(string(dbInfo.Name))
	if err != nil {
		return nil, err
	}
	return &ResourcesHandler{
		jmsClient: utils.NewJumpServer(dbInfo.Endpoint, dbInfo.PrivateToken),
		db:        db, dbName: string(dbInfo.Name),
	}, nil
}

func saveResources(c *gin.Context) {
	var err error
	dbInfo := c.MustGet(consts.DBInfoContextKey).(models.JumpServer)
	cache := utils.GetCache()

	handler, err := newResourcesHandler(dbInfo)
	if err != nil {
		c.JSON(http.StatusPreconditionFailed, gin.H{
			"error": "Database init failed", "details": "Database init failed",
		})
		return
	}

	resourceType := c.Query("m_type")
	var ids []string
	switch resourceType {
	case User:
		ids, err = handler.saveUser(c)
	case Role:
		err = handler.saveRole(c)
	case UserGroup:
		ids, err = handler.saveUserGroup(c)
	case Platform:
		err = handler.savePlatform(c)
	case Host:
		ids, err = handler.saveHost(c)
	case Permission:
		ids, err = handler.savePerm(c)
	case ChildrenNode:
		ids, err = handler.saveChildrenNode(c)
	case Node:
		err = handler.saveNode(c)
	case NodeWithAsset:
		err = handler.assetNodeRelation(c)
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
		return
	}

	for _, id := range ids {
		_ = cache.Set(fmt.Sprintf("%s-%s", resourceType, id), "", 0)
	}

	c.JSON(http.StatusCreated, gin.H{
		"message": fmt.Sprintf("Resource[%s] created successfully", resourceType),
	})
}

func updateResources(c *gin.Context) {
	var err error

	dbInfo := c.MustGet(consts.DBInfoContextKey).(models.JumpServer)
	handler, err := newResourcesHandler(dbInfo)
	if err != nil {
		c.JSON(http.StatusPreconditionFailed, gin.H{
			"error": "Database init failed", "details": "Database init failed",
		})
		return
	}

	resourceType := c.Query("m_type")
	id := c.Param("id")
	switch resourceType {
	case Node:
		err = handler.updateNode(c, id)
	case UserUnblock:
		err = handler.unblockUser(id)
	case UserResetMFA:
		err = handler.resetUserMFA(id)
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
		return
	}

	c.JSON(http.StatusCreated, gin.H{
		"message": fmt.Sprintf("Resource[%s] update successfully", resourceType),
	})
}

func deleteResource(c *gin.Context) {
	var err error
	var dbName string
	var handler *ResourcesHandler
	var handlers []*ResourcesHandler

	validResourceTypes := map[string]bool{
		Permission: true,
		Asset:      true,
	}

	resourceType := c.Query("m_type")
	if !validResourceTypes[resourceType] {
		c.JSON(http.StatusBadRequest, gin.H{
			"error":   "Invalid request type",
			"details": fmt.Sprintf("Invalid request type: %s", resourceType),
		})
		return
	}

	cache := utils.GetCache()
	id := c.Param("id")
	cacheKey := fmt.Sprintf("%s-%s", resourceType, id)
	err = cache.Get(cacheKey, &dbName)
	defaultDB := database.GetDBManager().GetDefaultDB()
	if err != nil || dbName == "" {
		var servers []models.JumpServer
		defaultDB.Model(models.JumpServer{}).Where("role = ?", models.RoleSlave).Find(&servers)
		for _, server := range servers {
			handler, err = newResourcesHandler(server)
			if err != nil {
				c.JSON(http.StatusPreconditionFailed, gin.H{
					"error": "Database init failed", "details": "Database init failed",
				})
				return
			}
			handlers = append(handlers, handler)
		}
	} else {
		var server models.JumpServer
		defaultDB.Model(models.JumpServer{}).Where("name = ?", dbName).Find(&server)
		handler, err = newResourcesHandler(server)
		if err != nil {
			c.JSON(http.StatusPreconditionFailed, gin.H{
				"error": "Database init failed", "details": "Database init failed",
			})
			return
		}
		handlers = append(handlers, handler)
	}

	go func(handlers []*ResourcesHandler, cacheKey string, cache *utils.CacheManager) {
		for _, handler = range handlers {
			switch resourceType {
			case Permission:
				err = handler.deletePerm(id)
			case Asset:
				err = handler.deleteAsset(id)
			}
			if errors.Is(err, consts.NotFoundError) {
				continue
			} else if err == nil {
				_ = cache.Delete(cacheKey)
				break
			} else {
				// TODO 记录报错，等待轮回程序执行,这里后续写
			}
		}
	}(handlers, cacheKey, cache)
	c.JSON(http.StatusAccepted, gin.H{
		"message": fmt.Sprintf("Resource[%s] deleted successfully", resourceType),
	})
}
