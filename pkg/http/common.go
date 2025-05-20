package pkg

import (
    "fmt"
    "gorm.io/gorm"
    "middleman/pkg/consts"
    "middleman/pkg/database"
    "middleman/pkg/utils"
    "net/http"
    "strings"
    
    "middleman/pkg/database/models"
    
    "github.com/gin-gonic/gin"
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

func handleResources(c *gin.Context) {
    db := c.MustGet(consts.DbContextKey).(*gorm.DB)
    if db == nil {
        c.JSON(http.StatusPreconditionFailed, gin.H{
            "error": "Database not found", "details": "Database not found",
        })
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
    case "user":
        instance = handleUser(c)
    case "permission":
        instance = handlePermission(c)
    case "host":
        instance = handleHost(c)
    case "device":
        instance = handleDevice(c)
    case "database":
        instance = handleDatabase(c)
    case "cloud":
        instance = handleCloud(c)
    case "web":
        instance = handleWeb(c)
    case "gpt":
        instance = handleGPT(c)
    case "custom":
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
