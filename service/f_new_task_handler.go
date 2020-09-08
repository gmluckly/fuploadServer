package service

import (
	"fmt"
	"fuploadServer/config"
	"fuploadServer/utils"
	"net/http"
	"os"
	"strings"
	"time"
	//"os"
	"strconv"

	"github.com/gin-gonic/gin"
)

func NewTaskHandler(c *gin.Context) {
	type request struct {
		Token             string      `json:"token"`
		UserId            int64       `json:"userId"`
		StorePath         string      `json:"storePath"`
		FileName          string      `json:"fileName"`
		FileSize          int64       `json:"fileSize"`
		FileMd5           string      `json:"fileMd5"`
		FeedBackParameter interface{} `json:"feedbackParameter"`
	}
	var req request
	if err := c.BindJSON(&req); err == nil {
		config := config.GetConfig()
		bServers := config.BServer
		var checkToken bool
		if _, ok := bServers[req.Token]; ok {
			checkToken = true
		}
		if checkToken {
			//TODO find if there is an old task before
			key := strconv.FormatInt(req.UserId, 10) + req.FileMd5
			type response struct {
				TaskId    int64  `json:"taskId"`
				UploadUrl string `json:"uploadUrl"`
				BlocksUrl string `json:"blocksUrl"`
				StateUrl  string `json:"stateUrl"`
			}
			oldTask := tasks.getTaskByKey(key)
			if oldTask != nil {
				//TODO here there is an old task
				var result response
				result.TaskId = oldTask.taskId
				result.UploadUrl = oldTask.uploadUrl
				result.StateUrl = oldTask.stateUrl
				result.BlocksUrl = oldTask.blocksUrl
				c.JSON(http.StatusOK, ResponseData("0", "ok", result))

			} else {
				// new task
				//fmt.Println("req: ", req)
				//var task taskInfo
				id := utils.MakeTaskId()
				tmpPath := config.TmpDir + strconv.FormatInt(id, 10) + "/"
				tmpFile := tmpPath + req.FileName
				utils.CheckAndMakeDir(tmpPath)
				f, err := os.Create(tmpFile)
				if err != nil {
					fmt.Println(" Create err:", err)
					c.JSON(http.StatusOK, ResponseErrDisk)
					return
				}
				idStr := strconv.FormatInt(id, 10)
				var uploadUrl string
				var stateUrl string
				var blocksUrl string
				uUrl := "/api/upload/files/block?taskId="
				sUrl := "/api/upload/files/state"
				bUrl := "/api/upload/files/blocks/info?taskId="
				if strings.EqualFold(config.Server.ProxyAddr, "") {
					Port := strconv.Itoa(config.Server.Port)
					uploadUrl = config.Server.Addr + ":" + Port + uUrl + idStr
					stateUrl = config.Server.Addr + ":" + Port + sUrl
					blocksUrl = config.Server.Addr + ":" + Port + bUrl + idStr
				} else {
					uploadUrl = config.Server.ProxyAddr + uUrl + idStr
					stateUrl = config.Server.ProxyAddr + sUrl
					blocksUrl = config.Server.ProxyAddr + bUrl + idStr
				}
				t := &task{
					taskId:            id,
					status:            "uploading",
					cTime:             time.Now().Unix(),
					uploadUrl:         uploadUrl,
					stateUrl:          stateUrl,
					blocksUrl:         blocksUrl,
					userId:            req.UserId,
					token:             req.Token,
					storePath:         req.StorePath,
					fileSize:          req.FileSize,
					fileName:          req.FileName,
					fileMd5:           req.FileMd5,
					feedBackParameter: req.FeedBackParameter,
					blockChan:         make(chan *block),
					done:              make(chan struct{}),
					statusChan:        make(chan string),
					blocks:            make([]*block, 0, 10),
					file:              f,
					tmpPath:           tmpPath,
				}
				tasks.setTask(t)

				var result response
				result.TaskId = t.taskId
				result.UploadUrl = uploadUrl
				result.BlocksUrl = blocksUrl
				result.StateUrl = stateUrl
				go mergeTaskBlock(id, t.blockChan, t.done, t.statusChan)

				c.JSON(http.StatusOK, ResponseData("0", "ok", result))
			}
		} else {
			c.JSON(http.StatusOK, ResponseErrNotAuth)
		}
	} else {
		c.JSON(http.StatusOK, ResponseErrParameter)
	}
}
