package service

import (
	"github.com/aviddiviner/gin-limit"
	"github.com/gin-gonic/gin"
)

func setUpRouter() *gin.Engine {
	gin.SetMode(gin.ReleaseMode)
	r := gin.Default()
	r.Use(limit.MaxAllowed(20))
	r.POST("/api/upload/new/task", NewTaskHandler)
	r.GET("/api/upload/files/blocks/info", GetBlocksInfoHandler)
	r.POST("/api/upload/files/block", MultiPartHandler)
	r.POST("/api/upload/files/state", UpdateTaskStatus)
	return r
}
