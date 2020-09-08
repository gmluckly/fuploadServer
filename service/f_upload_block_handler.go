package service

import (
	"bytes"
	"encoding/json"
	"fuploadServer/config"
	"fuploadServer/utils"
	"io"
	"io/ioutil"
	"strings"
	"sync"

	//"os"
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
)

func MultiPartHandler(c *gin.Context) {
	reader, err := c.Request.MultipartReader()
	if err != nil {
		c.JSON(http.StatusOK, ResponseErrSystem)
	}
	var start int64
	var end int64
	var partMd5 string
	taskIdStr := c.DefaultQuery("taskId", "0")
	taskId, _ := strconv.ParseInt(taskIdStr, 10, 64)
	task := tasks.getTask(taskId)
	if strings.EqualFold(task.status, "pause") {
		c.JSON(http.StatusOK, ResponseErrTaskPause)
		return
	}
	buf := new(bytes.Buffer)
	for {
		part, err := reader.NextPart()
		if err == io.EOF {
			//log.Printf("==== EOF\n")
			break
		} else if err != nil {
			c.JSON(http.StatusOK, ResponseErrParameter)
			return
		}
		formname := part.FormName()
		if formname == "file" {
			content, err := ioutil.ReadAll(part)
			if err != nil {
				c.JSON(http.StatusOK, ResponseErrSystem)
				return
			}
			_, err1 := buf.Write(content)
			if err1 != nil {
				c.JSON(http.StatusOK, ResponseErrSystem)
				return
			}

		} else {
			var err error
			var data []byte
			data, err = ioutil.ReadAll(part)
			if err != nil {
				c.JSON(http.StatusOK, ResponseErrParameter)
				return
			}
			switch formname {
			case "startByte":
				start, _ = strconv.ParseInt(string(data), 10, 64)
			case "endByte":
				end, _ = strconv.ParseInt(string(data), 10, 64)
			case "partMd5":
				partMd5 = string(data)
			}
		}
	}
	if start > end {
		c.JSON(http.StatusOK, ResponseErrTmpIndex)
		return
	}

	config := config.GetConfig()
	if config.CheckTmpMd5 {
		bufByte := buf.Bytes()
		bufferMd5 := utils.GetTmpMd5(bufByte)
		//fmt.Println("partMd5:", partMd5, "check bufferMd5:", bufferMd5)
		if bufferMd5 != partMd5 {
			c.JSON(http.StatusOK, ResponseErrTmpMd5)
			return
		}
	}
	var wg sync.WaitGroup
	wg.Add(1)
	b := &block{
		start:  start,
		end:    end,
		wg:     &wg,
		buffer: buf,
	}

	task.blockChan <- b
	wg.Wait()
	expect := task.expect
	if len(expect) != 0 {
		type response struct {
			Start int64 `json:"startByte"`
			End   int64 `json:"endByte"`
		}
		result := make([]response, 0, 5)
		for i := 0; i < len(expect); i++ {
			var r response
			r.Start = expect[i].start
			r.End = expect[i].end
			result = append(result, r)
		}
		resultByte, _ := json.Marshal(result)
		c.JSON(http.StatusOK, ResponseData("1", "ok", resultByte))
		return
	}
	c.JSON(http.StatusOK, ResponseOK)
}
