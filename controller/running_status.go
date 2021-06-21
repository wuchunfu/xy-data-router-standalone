package controller

import (
	"net/http"

	"github.com/gin-gonic/gin"

	"github.com/fufuok/xy-data-router/service"
)

func runningStatusHandler(c *gin.Context) {
	c.IndentedJSON(http.StatusOK, service.RunningStatus())
}

func runningQueueStatusHandler(c *gin.Context) {
	c.IndentedJSON(http.StatusOK, service.RunningQueueStatus())
}
