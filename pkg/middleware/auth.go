package middleware

import (
    "net/http"
    "strings"
    
    "middleman/pkg/database"
    "middleman/pkg/database/models"
    
    "github.com/gin-gonic/gin"
)

func AccessKeyMiddleware() gin.HandlerFunc {
    return func(c *gin.Context) {
        authHeader := c.GetHeader("Authorization")
        if authHeader == "" {
            c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{
                "error": "Missing Authorization header",
                "code":  40101,
            })
            return
        }
        
        // 格式应为: "Bearer {AccessKeyID}:{SecretAccessKey}"
        parts := strings.Split(authHeader, " ")
        if len(parts) != 2 || parts[0] != "Bearer" {
            c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{
                "error": "Invalid Authorization format",
                "code":  40102,
            })
            return
        }
        
        credentials := strings.Split(parts[1], ":")
        if len(credentials) != 2 {
            c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{
                "error": "Invalid credentials format",
                "code":  40103,
            })
            return
        }
        
        var count int64
        db := database.GetDBManager().GetDefaultDB()
        err := db.Model(&models.JumpServer{}).
            Where("access_key = ? AND secret_key = ?", credentials[0], credentials[1]).
            Count(&count).Error
        
        if err != nil {
            c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{
                "error": "Invalid access key or secret key",
                "code":  40104,
            })
            return
        }
        if count <= 0 {
            c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{
                "error": "Invalid access key or secret key",
                "code":  40105,
            })
            return
        }
        c.Next()
    }
}
