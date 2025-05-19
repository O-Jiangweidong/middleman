package pkg

import (
    "fmt"
    "github.com/gin-gonic/gin"
    "gorm.io/gorm"
    "net/http"
    "strings"
    
    "middleman/pkg/consts"
    "middleman/pkg/database/models"
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
    db := c.MustGet(consts.DbContextKey).(*gorm.DB)
    jumpserver := models.JumpServer{
        Name: registerRequest.Name,
    }
    result := db.Create(&jumpserver)
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
}
