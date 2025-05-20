package pkg

import (
    "fmt"
    "gorm.io/gorm"
    "middleman/pkg/consts"
    "middleman/pkg/database"
    "middleman/pkg/utils"
    "net/http"
    "reflect"
    "strconv"
    "strings"
    
    "middleman/pkg/database/models"
    
    "github.com/gin-gonic/gin"
)

const (
    User     = "user"
    Perm     = "Permission"
    Host     = "host"
    Device   = "device"
    Database = "database"
    Cloud    = "cloud"
    Web      = "web"
    Gpt      = "gpt"
    Custom   = "custom"
)

type RegisterRequest struct {
    Name string `json:"name"`
}

func handleRegister(c *gin.Context) {
    var registerRequest RegisterRequest
    if err := c.ShouldBindJSON(&registerRequest); err != nil {
        c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
        return
    }
    db := database.DBManager.GetDefaultDB()
    server := models.JumpServer{
        BaseJumpServer: models.BaseJumpServer{
            Name: registerRequest.Name,
        },
        AccessKey: utils.GenerateRandomString(36),
        SecretKey: utils.GenerateRandomString(36),
    }
    result := db.Create(&server)
    if result.Error != nil {
        var msg string
        if strings.Contains(result.Error.Error(), "unique constraint") {
            msg = "存在相同名称的堡垒机，请修改配置文件中的堡垒机名称，并重启重新注册"
        } else {
            msg = fmt.Sprintf("创建用户失败: %v", result.Error)
        }
        c.JSON(http.StatusBadRequest, gin.H{"error": msg})
        return
    }
    c.JSON(http.StatusOK, gin.H{"data": server})
}

func getJumpServers(c *gin.Context) {
    db := database.DBManager.GetDefaultDB()
    var servers []models.BaseJumpServer
    db.Find(&servers)
    c.JSON(http.StatusOK, gin.H{"data": servers})
}

type ResourceRequest struct {
    Type string `json:"type" binding:"required"`
}

func getResources(c *gin.Context) {
    db := c.MustGet(consts.DbContextKey).(*gorm.DB)
    orgID := c.MustGet(consts.OrgContextKey).(string)
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
    
    var total int64
    query := db.Model(modelValue).Where("org_id = ?", orgID)
    result := query.Count(&total).Limit(limit).Offset(offset).Find(resources)
    
    if result.Error != nil {
        c.JSON(http.StatusPreconditionFailed, gin.H{
            "error": "Database error", "details": result.Error.Error(),
        })
        return
    }
    
    c.JSON(http.StatusOK, gin.H{"data": resources, "total": total})
}

func saveResources(c *gin.Context) {
    db := c.MustGet(consts.DbContextKey).(*gorm.DB)
    if db == nil {
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
    
    var instance interface{}
    switch req.Type {
    case User:
        instance = handleUser(c)
    case Perm:
        instance = handlePermission(c)
    case Host:
        instance = handleHost(c)
    case Device:
        instance = handleDevice(c)
    case Database:
        instance = handleDatabase(c)
    case Cloud:
        instance = handleCloud(c)
    case Web:
        instance = handleWeb(c)
    case Gpt:
        instance = handleGPT(c)
    case Custom:
        instance = handleCustom(c)
    default:
        c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request type", "details": req.Type})
        return
    }
    
    if instance != nil {
        result := db.Create(instance)
        if result.Error != nil {
            c.JSON(http.StatusInternalServerError, gin.H{
                "error":   fmt.Sprintf("Failed to create permission: %v", result.Error),
                "details": "Database operation failed",
            })
            return
        }
        c.JSON(http.StatusCreated, gin.H{
            "message": "Permission created successfully",
            "data":    instance,
        })
    }
    c.JSON(http.StatusBadRequest, gin.H{"error": "Unknown error", "details": "Unknown error"})
}
