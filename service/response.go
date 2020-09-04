package service

import (
	"github.com/gin-gonic/gin"
)

var ResponseErrNotAuth = gin.H{"retCode": "1", "status": "not authorized", "data": ""}
var ResponseErrNotTask = gin.H{"retCode": "2", "status": "can not find task", "data": ""}
var ResponseErrDisk = gin.H{"retCode": "3", "status": "can't create temp dir", "data": ""}
var ResponseErrTaskPause = gin.H{"retCode": "4", "status": "task was pause", "data": ""}
var ResponseErrTaskStatus = gin.H{"retCode": "5", "status": "same as the old state", "data": ""}
var ResponseErrParameter = gin.H{"retCode": "6", "status": "parameter error", "data": ""}
var ResponseErrTmpMd5 = gin.H{"retCode": "7", "status": "temp part md5 error", "data": ""}
var ResponseErrTmpIndex = gin.H{"retCode": "8", "status": "temp part index error", "data": ""}
var ResponseErrSystem = gin.H{"retCode": "20", "status": "unknown system error", "data": ""}
var ResponseOK = gin.H{"retCode": "0", "status": "ok", "data": ""}

func ResponseData(returnCode, status string, data interface{}) gin.H {
	return gin.H{"retCode": returnCode, "status": status, "data": data}
}
