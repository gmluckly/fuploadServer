/*
get the task's block info ,such as {start,end};and return the expect_list{start,end}
so the client can upload the block that the server needs
*/

package service

import (
	"fmt"
	"sort"
	"strconv"
	"net/http"

	"github.com/gin-gonic/gin"
)

func GetBlocksInfoHandler(c *gin.Context) {
	taskIdStr := c.DefaultQuery("taskId", "")
	taskId, _ := strconv.ParseInt(taskIdStr, 10, 64)
	task := tasks.getTask(taskId)
	if task != nil {
		c.JSON(http.StatusOK, ResponseErrNotTask)
		return
	}
	fileSize := task.fileSize
	expects := getExpectList(task.blocks, fileSize)
	type response struct {
		TotalSize      int64        `json:"totalSize"`
		Status         string       `json:"status"`
		ExpectedBlocks []expectInfo `json:"expectBlocks"`
	}
	var resp response
	resp.TotalSize = fileSize
	resp.Status = task.status
	resp.ExpectedBlocks = expects
	c.JSON(http.StatusOK, ResponseData("0", "ok", resp))
}

type expectInfo struct {
	StartByte int64 `json:"startByte"`
	EndByte   int64 `json:"endByte"`
}

func getExpectList(blocks []*block, fileSize int64) []expectInfo {
	sort.Slice(blocks, func(i, j int) bool {
		return blocks[i].start < blocks[j].start
	})
	result := make([]expectInfo, 0, 10)
	var start int64 = 0
	if len(blocks) == 0 {
		var ex expectInfo
		ex.StartByte = 0
		ex.EndByte = fileSize - 1
		result = append(result, ex)
	} else {
		for i := 0; i < len(blocks); i++ {
			b := blocks[i]
			if b.start > start {
				var tmp expectInfo
				tmp.StartByte = start
				tmp.EndByte = b.start - 1
				result = append(result, tmp)
			}
			if b.end != fileSize {
				start = b.end + 1
			}
		}
	}
	fmt.Println("result: ", result)

	return result
}
