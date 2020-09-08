package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"github.com/gin-gonic/gin"
	"io/ioutil"
	"net/http"
)

func main() {
	r := setUpRouter()
	r.Run(":8090")
}

func setUpRouter() *gin.Engine {
	gin.SetMode(gin.DebugMode)
	r := gin.Default()

	r.POST("/api/business/new/task", UploadFile)
	r.POST("/api/business/fupload/callback", UploadCallBack)

	return r
}

type request struct {
	UserId   int64  `json:"userId"`
	FileName string `json:"fileName"`
	FileSize int64  `json:"fileSize"`
	FileMd5  string `json:"fileMd5"`
}

func UploadFile(c *gin.Context) {
	var req request
	if err := c.BindJSON(&req); err != nil {
		c.JSON(http.StatusOK, gin.H{"retCode": "-1", "status": "parameter error", "data": ""})
		return
	}
	body := make(map[string]interface{})
	body["fileName"] = req.FileName
	body["fileSize"] = req.FileSize
	body["fileMd5"] = req.FileMd5
	body["userId"] = req.UserId
	body["storePath"] = "/tmp/business/"
	body["feedbackParameter"] = ""
	body["token"] = "cp00000010103010301"
	bytesData, err := json.Marshal(body)
	if err != nil {
		fmt.Println("err:", err)
		return
	}
	reader := bytes.NewReader(bytesData)
	url := "http://127.0.0.1:8080/api/upload/new/task"
	request, err := http.NewRequest("POST", url, reader)
	if err != nil {
		fmt.Println(err.Error())
		return
	}
	request.Header.Set("Content-Type", "application/json")
	client := http.Client{}
	resp, err := client.Do(request)
	if err != nil {
		c.JSON(http.StatusOK, gin.H{"retCode": "-1", "status": "parameter error", "data": ""})
		return
	} else {
		type taskInfo struct {
			TaskId    int64  `json:"taskId"`
			UploadUrl string `json:"uploadUrl"`
			BlocksUrl string `json:"blocksUrl"`
			StateUrl  string `json:"stateUrl"`
		}
		type respBody struct {
			RetCode string   `json:"retCode"`
			Status  string   `json:"status"`
			Data    taskInfo `json:"data"`
		}
		var r respBody
		respBytes, _ := ioutil.ReadAll(resp.Body)
		err := json.Unmarshal(respBytes, &r)
		fmt.Println("Unmarshal err:", err)
		fmt.Println("r: ", r)
		var result taskInfo
		result.TaskId = r.Data.TaskId
		result.BlocksUrl = r.Data.BlocksUrl
		result.StateUrl = r.Data.StateUrl
		result.UploadUrl = r.Data.UploadUrl

		c.JSON(http.StatusOK, gin.H{"retCode": "0", "status": "ok", "data": result})
	}
}

func UploadCallBack(c *gin.Context) {
	type callBack struct {
		RetCode string      `json:"retCode"`
		Status  string      `json:"status"`
		Data    interface{} `json:"data"`
	}
	var calReq callBack
	if err := c.BindJSON(&calReq); err != nil {
		c.JSON(http.StatusOK, gin.H{"retCode": "-1", "status": "parameter error", "data": ""})
		return
	}
	fmt.Println("upload finish", calReq)
	//TODO you business
	c.JSON(http.StatusOK, gin.H{"retCode": "0", "status": "ok", "data": ""})
}
