package main

import (
	"fmt"
	"net/http"
	"os"
	"strconv"

	"path/filepath"

	"encoding/json"
	"mime/multipart"

	"bytes"
	"io/ioutil"
	"strings"

	"fuploadServer/utils"
)

var uploadFileSize = 1000000

func sendFile(userId int64, targetContent string) {

	fileName := filepath.Base(targetContent)
	localFilePath := targetContent
	fileMd5 := utils.FileMd5(localFilePath)

	fileInfo, _ := os.Stat(localFilePath)
	//fmt.Println("fileInfo:", fileInfo)
	fileSize := fileInfo.Size()

	url := "http://127.0.0.1:8090/api/business/new/task"
	body := make(map[string]interface{})
	body["fileName"] = fileName
	body["fileSize"] = fileSize
	body["fileMd5"] = fileMd5
	body["userId"] = userId

	bytesData, err := json.Marshal(body)
	if err != nil {
		fmt.Println("err:", err)
		return
	}
	reader := bytes.NewReader(bytesData)
	request, err := http.NewRequest("POST", url, reader)
	if err != nil {
		fmt.Println(err.Error())
		return
	}
	request.Header.Set("Content-Type", "application/json")
	client := http.Client{}
	resp, err := client.Do(request)
	if err != nil {
		fmt.Println("err:", err)
	} else {
		type taskInfo struct {
			TaskId    int64  `json:"taskId"`
			UploadUrl string `json:"uploadUrl"`
			BlocksUrl string `json:"blocksUrl"`
			StateUrl  string `json:"stateUrl"`
		}
		type respBody struct {
			ReturnCode string   `json:"retCode"`
			Status     string   `json:"status"`
			Data       taskInfo `json:"data"`
		}
		var r respBody
		respBytes, _ := ioutil.ReadAll(resp.Body)
		err := json.Unmarshal(respBytes, &r)
		fmt.Println("Unmarshal err:", err, "respBytes:", resp.Body)
		uploadUrl := r.Data.UploadUrl
		fmt.Println("r: ", r)
		doUpload(uploadUrl, localFilePath, fileSize)
	}
}

func doUpload(url, filePath string, fileSize int64) {
	fmt.Println("url:", url, "filePath:", filePath)
	file, err := os.Open(filePath)
	defer file.Close()
	if err != nil {
		fmt.Println("err:", err)
	} else {
		var totalBklock int
		var tmp int
		tmp = int(fileSize) / uploadFileSize
		//fmt.Println("tmp:", tmp)
		totalBklock = tmp + 1
		leftSize := int(fileSize)
		fmt.Println("222 totalBklock:", totalBklock)
		//syncChan := make(chan bool)
		for i := 0; i < totalBklock; i++ {
			var size int
			if leftSize > uploadFileSize {
				size = uploadFileSize
			} else {
				size = leftSize
			}

			data := make([]byte, size)
			offset := int64(i * uploadFileSize)
			//fmt.Println("offset:", offset)
			n, err := file.ReadAt(data, offset)
			if err != nil {
				fmt.Println("read file error:", err)
			}
			leftSize = leftSize - n
			file.Seek(int64(uploadFileSize+n), 0)
			var start, end int
			start = i * uploadFileSize
			end = start + n - 1
			r := sendFileData(url, start, end, data)
			if !r {
				//TODO add the block to the failureList
			}
		}
	}
}

func sendFileData(url string, start, end int, data []byte) bool {
	//fmt.Println("block url: ", url, " start:", start, " end:", end)
	partMd5 := utils.GetTmpMd5(data)
	b := new(bytes.Buffer)
	w := multipart.NewWriter(b)

	pw, err := w.CreateFormField("partName")
	pw.Write([]byte("1.tmp"))

	pw1, _ := w.CreateFormField("startByte")
	pw1.Write([]byte(strconv.Itoa(start)))

	pw2, _ := w.CreateFormField("endByte")
	pw2.Write([]byte(strconv.Itoa(end)))

	pw3, _ := w.CreateFormField("partMd5")
	pw3.Write([]byte(partMd5))
	fileName := strconv.Itoa(start) + ".tmp"
	fileData, err := w.CreateFormFile("file", fileName)

	if err != nil {
		fmt.Println("err:", err)
	}

	_, err = fileData.Write(data)
	if err != nil {
		fmt.Println("fileData write error:", err)
	}

	w.Close()

	req, err := http.NewRequest("POST", url, b)

	if err != nil {
		fmt.Println("NewRequest error:", err)
	}

	//w.SetBoundary("------------" + partMd5)

	req.Header.Set("Content-Type", w.FormDataContentType())

	resp, err := http.DefaultClient.Do(req)

	defer resp.Body.Close()

	type result struct {
		ReturnCode string `json:"retCode"`
		Status     string `json:"status"`
	}
	var r result
	bodyDatas, _ := ioutil.ReadAll(resp.Body)
	err = json.Unmarshal(bodyDatas, &r)
	if err != nil {
		fmt.Println("err:", err)
	}
	//fmt.Println("return_code:", r.ReturnCode)
	return strings.EqualFold(r.ReturnCode, "0")
}

func main() {
	var userId int64 = 123456
	//var token string = "12345"
	targetContent := "/home/test.mp4"
	//storePath := "/tmp/fupload"
	sendFile(userId, targetContent)
}
