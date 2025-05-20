package pkg

import (
    "net/http"
    
    "github.com/gin-gonic/gin"
    
    "middleman/pkg/database/models"
)

func handlePermission(c *gin.Context) interface{} {
    var permission models.Permission
    if err := c.ShouldBindJSON(&permission); err != nil {
        c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid permission format", "details": err.Error()})
        return nil
    }
    return &permission
}
