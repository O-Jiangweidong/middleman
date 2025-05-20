package pkg

import (
    "net/http"
    
    "github.com/gin-gonic/gin"
    
    "middleman/pkg/database/models"
)

func handleHost(c *gin.Context) interface{} {
    var host models.Host
    if err := c.ShouldBindJSON(&host); err != nil {
        c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid host format", "details": err.Error()})
        return nil
    }
    return &host
}

func handleDevice(c *gin.Context) interface{} {
    var device models.Device
    if err := c.ShouldBindJSON(&device); err != nil {
        c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid device format", "details": err.Error()})
        return nil
    }
    return &device
}

func handleDatabase(c *gin.Context) interface{} {
    var database models.Database
    if err := c.ShouldBindJSON(&database); err != nil {
        c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid database format", "details": err.Error()})
        return nil
    }
    return &database
}

func handleCloud(c *gin.Context) interface{} {
    var cloud models.Cloud
    if err := c.ShouldBindJSON(&cloud); err != nil {
        c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid cloud format", "details": err.Error()})
        return nil
    }
    return &cloud
}

func handleWeb(c *gin.Context) interface{} {
    var web models.Web
    if err := c.ShouldBindJSON(&web); err != nil {
        c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid web format", "details": err.Error()})
        return nil
    }
    return &web
}

func handleGPT(c *gin.Context) interface{} {
    var gpt models.GPT
    if err := c.ShouldBindJSON(&gpt); err != nil {
        c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid gpt format", "details": err.Error()})
        return nil
    }
    return &gpt
}

func handleCustom(c *gin.Context) interface{} {
    var custom models.Device
    if err := c.ShouldBindJSON(&custom); err != nil {
        c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid custom format", "details": err.Error()})
        return nil
    }
    return &custom
}
