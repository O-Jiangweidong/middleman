package pkg

import (
    "net/http"
    
    "github.com/gin-gonic/gin"
    
    "middleman/pkg/database/models"
)

func handleUser(c *gin.Context) interface{} {
    var user models.User
    if err := c.ShouldBindJSON(&user); err != nil {
        c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid user format", "details": err.Error()})
        return nil
    }
    return &user
}
