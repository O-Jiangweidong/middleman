package middleware

import (
    "fmt"
    "net/http"
    
    "github.com/gin-gonic/gin"
    
    "middleman/pkg/consts"
    "middleman/pkg/database"
)

func DatabaseMiddleware() gin.HandlerFunc {
    return func(c *gin.Context) {
        name := c.Request.Header.Get("JUMPSERVER-NAME")
        if name == "" {
            name = database.DefaultDBName
        }
        db, err := database.GetDBManager().GetDB(name)
        if err != nil {
            c.AbortWithStatusJSON(http.StatusBadRequest, gin.H{
                "error": fmt.Sprintf("Failed to get database name: %v", err),
                "code":  40001,
            })
            return
        }
        
        c.Set(consts.DbContextKey, db)
        c.Next()
    }
}
