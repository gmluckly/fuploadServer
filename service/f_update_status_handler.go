package service

import (
	"fmt"
	"github.com/gin-gonic/gin"
	"net/http"
	"strings"
)

func UpdateTaskStatus(c *gin.Context) {
	fmt.Println("update status")
	type request struct {
		TaskId int64  `json:"taskId"`
		Status string `json:"status"`
	}
	var req request
	if err := c.BindJSON(&req); err == nil {
		task := tasks.getTask(req.TaskId)
		if task == nil {
			c.JSON(http.StatusOK, ResponseErrNotTask)
			return
		}
		if strings.EqualFold(req.Status, task.status) {
			c.JSON(http.StatusOK, ResponseErrTaskStatus)
			return
		} else {
			task.statusChan <- req.Status
			c.JSON(http.StatusOK, ResponseOK)
		}
	} else {
		c.JSON(http.StatusOK, ResponseErrParameter)
	}
}
