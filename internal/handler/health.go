package handler

import (
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
)

type HealthResponse struct {
	Status    string `json:"status"`
	Timestamp string `json:"timestamp"`
	Version   string `json:"version"`
}

var AppVersion = "1.0.0"

func Health(c *gin.Context) {
	response := HealthResponse{
		Status:    "ok",
		Timestamp: time.Now().UTC().Format(time.RFC3339),
		Version:   AppVersion,
	}
	c.JSON(http.StatusOK, response)
}

func Ready(c *gin.Context) {
	c.JSON(http.StatusOK, gin.H{"status": "ready"})
}

func Live(c *gin.Context) {
	c.JSON(http.StatusOK, gin.H{"status": "alive"})
}
