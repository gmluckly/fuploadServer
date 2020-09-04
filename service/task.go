package service

import (
	"bytes"
	"encoding/json"
	"fmt"
	"fuploadServer/config"
	"fuploadServer/utils"
	"net/http"
	"os"
	"os/exec"
	"sort"
	"strconv"
	"strings"
	"sync"
	"time"
)

const (
	retCodeMd5     string = "-2"
	retCodeTimeout string = "-1"
	retCodeFinish  string = "0"
)

var mDiff int64 = 1800

type Tasks struct {
	mutex   *sync.RWMutex
	taskMap map[int64]*task
}

var tasks Tasks

// when the task's tmpSize is more than mValue,we try to write the buffer to temp file,20M
var mValue int64 = 20000000

func init() {
	taskMap := make(map[int64]*task)
	tasks.taskMap = taskMap
	tasks.mutex = new(sync.RWMutex)
}

type block struct {
	start  int64
	end    int64
	wg     *sync.WaitGroup
	buffer *bytes.Buffer
}

type expectByte struct {
	start int64
	end   int64
}
type task struct {
	taskId            int64
	status            string // the task'status:pause|uploading|finish
	cTime             int64  // create time
	mTime             int64  // the time that last get block
	uploadUrl         string
	stateUrl          string
	blocksUrl         string
	userId            int64
	token             string
	storePath         string
	fileName          string
	fileSize          int64
	fileMd5           string
	feedBackParameter interface{} // send the parameter back to b server
	blockChan         chan *block // receive data from the multipart
	done              chan struct{}
	statusChan        chan string
	blocks            []*block
	outOrderBlocks    []*block
	file              *os.File
	tmpSize           int64
	tmpPath           string // merge file path in local
	fCount            int    // failure of temp file counter
	expect            []expectByte
	writeFileIndex    int64
}

func (ts Tasks) getTask(taskId int64) *task {
	ts.mutex.Lock()
	defer ts.mutex.Unlock()
	if v, ok := ts.taskMap[taskId]; ok {
		return v
	}
	return nil
}

func (ts *Tasks) getTaskByKey(key string) *task {
	var result *task
	for _, v := range tasks.taskMap {
		tmpKey := strconv.FormatInt(v.userId, 10) + v.fileMd5
		if strings.EqualFold(key, tmpKey) {
			result = v
			break
		}
	}
	return result
}

func (ts *Tasks) setTask(t *task) {
	ts.mutex.Lock()
	defer ts.mutex.Unlock()
	ts.taskMap[t.taskId] = t
}

func (ts *Tasks) updateTask(t *task) {
	ts.mutex.Lock()
	defer ts.mutex.Unlock()
	ts.taskMap[t.taskId] = t
}
func (ts *Tasks) deleteTaskById(taskId int64) {
	ts.mutex.Lock()
	defer ts.mutex.Unlock()
	delete(ts.taskMap, taskId)
}
func (ts *Tasks) deleteTask(t *task) {
	ts.mutex.Lock()
	defer ts.mutex.Unlock()
	delete(ts.taskMap, t.taskId)
}

func mergeTaskBlock(taskId int64, blockChan chan *block, done chan struct{}, statusChan chan string) {
	for {
		select {
		case <-done:
			fmt.Println("finish the task")
			return
		case newStatus := <-statusChan:
			//TODO update task status
			t := tasks.getTask(taskId)
			t.status = newStatus
			tasks.updateTask(t)
		case b := <-blockChan:
			// when we get a new block,try to find out ahead of the tmp info
			// and merge them;
			// if not, append the blocks
			t := tasks.getTask(taskId)
			blocks := t.blocks
			// check if the block is exist
			check := checkBlock(b, blocks)
			if check {
				if len(blocks) > 5 {
					sort.Slice(blocks, func(i, j int) bool {
						return blocks[i].start < blocks[j].start
					})
					var newB *block
					for i := 0; i < len(blocks); i++ {
						tmp := blocks[i]
						if tmp.end == b.start-1 {
							buffer := tmp.buffer
							_, err := buffer.Write(b.buffer.Bytes())
							if err != nil {
								fmt.Println("merge buffer error:", err)
							}
							newB = &block{
								start:  tmp.start,
								end:    b.end,
								buffer: buffer,
							}
							if i == 0 {
								blocks = append(blocks[1:], newB)
							} else {
								blocks = append(blocks[:i], blocks[i+1:]...)
								//blocks = blocks[:i+copy(blocks[i:], blocks[i+1:])]
								blocks = append(blocks, newB)
							}
							break
						}
					}

					if newB == nil {
						blocks = append(blocks, b)
					}
				} else {
					blocks = append(blocks, b)
				}
				t.blocks = blocks
				t.tmpSize = t.tmpSize + b.end - b.start + 1
				//TODO  delete expect block if the expect list is exist
				var eByte expectByte
				eByte.start = b.start
				eByte.end = b.end
				deleteExpectBlock(t, eByte)
				b.wg.Done()
				mergeTaskBlock2File(t)
			} else {
				// TODO
				/*
				 mark that they send a wrong block,
				 if the client do this more than 5 times,forbid the task,
				 and the client must initiate a request that get blocks info
				 of the task.
				*/
				fCount := t.fCount + 1
				if fCount >= 5 {
					t.status = "pause"
				}
				t.fCount = fCount
				tasks.updateTask(t)
				b.wg.Done()
			}
		case <-time.After(config.GetConfig().TaskTimeout * time.Second):
			// if the task do not finish after xxx seconds, and the client do not send
			// block anymore, stop the task, delete it!
			fmt.Println("time s up  and do not finish the task")
			//send result to b server
			t := tasks.getTask(taskId)
			if time.Now().Unix()-t.mTime > mDiff {
				doAfterTimeout(t)
				return
			}
		}
	}
}

//if the file is finished that merge buffer to the tmpStorePath,
//or else check the tmpSize of buffer when it is more than mValue.
//we use the idea that 'write ahread log'(WAL)
func mergeTaskBlock2File(t *task) {
	blocks := t.blocks
	f := t.file
	//fmt.Println("tmpSize:", t.tmpSize)
	if t.tmpSize >= t.fileSize {
		fmt.Println("merge file")
		sort.Slice(blocks, func(i, j int) bool {
			return blocks[i].start < blocks[j].start
		})
		for i := 0; i < len(blocks); i++ {
			start := blocks[i].start
			buffer := blocks[i].buffer
			_, err := f.WriteAt(buffer.Bytes(), start)
			if err != nil {
				fmt.Println("merge file error:", err)
			}
			//fmt.Println("n: ", n, " err:", err)
		}
		f.Close()
		//TODO check file's md5 and copy to the storePath
		doAfterMerge(t)
	} else {
		newBlocks := make([]*block, 0, 10)
		//var expect expectByte
		ex := t.expect
		index := t.writeFileIndex
		if t.tmpSize >= mValue {
			sort.Slice(blocks, func(i, j int) bool {
				return blocks[i].start < blocks[j].start
			})
			for i := 0; i < len(blocks); i++ {
				start := blocks[i].start
				if index == start {
					buffer := blocks[i].buffer
					_, err := f.WriteAt(buffer.Bytes(), start)
					//fmt.Println("n: ", n, "  :", err)
					if err != nil {
						newBlocks = append(newBlocks, blocks[i])
					} else {
						index = blocks[i].end + 1
					}
				} else {
					var e expectByte
					e.start = index
					e.end = blocks[i].start - 1
					ex = append(ex, e)
					break
				}
			}
		} else {
			fmt.Println("tmpSize is smaller than mValue ")
			newBlocks = append(newBlocks, blocks...)
		}
		t.blocks = newBlocks
		t.writeFileIndex = index
		t.expect = ex
		t.mTime = time.Now().Unix()
		tasks.updateTask(t)
	}
}

//delete the tmp file and send the result to the business server
func doAfterMerge(t *task) {
	tmpFileName := t.tmpPath + t.fileName
	finalMd5 := utils.FileMd5(tmpFileName)
	fmt.Println("finalMd5 :", finalMd5)
	if strings.EqualFold(finalMd5, t.fileMd5) {
		argStr := "cp " + t.tmpPath + t.fileName + " " + t.storePath
		cmd := exec.Command("/bin/sh", "-c", argStr)
		err := cmd.Run()
		if err != nil {
			fmt.Println("copy file to store path failure:", err)
		} else {
			// delete the temp file
			deleteTempfile(t.tmpPath)
			t.done <- struct{}{}
			tasks.deleteTask(t)
			//TODO send the result to business server
			sendResult2BServer(t.token, retCodeFinish, t.feedBackParameter)
		}
	} else {
		fmt.Println("error,finalMd5 != fileMd5 ")
		t.done <- struct{}{}
		deleteTempfile(t.tmpPath)
		tasks.deleteTask(t)
		sendResult2BServer(t.token, retCodeMd5, t.feedBackParameter)
	}
}
func deleteTempfile(tmpPath string) {
	argStr2 := "rm  -rf " + tmpPath
	cmd := exec.Command("/bin/sh", "-c", argStr2)
	cmd.Run()
}
func doAfterTimeout(t *task) {
	deleteTempfile(t.tmpPath)
	tasks.deleteTask(t)
	//TODO send the result to business server
	sendResult2BServer(t.token, retCodeTimeout, t.feedBackParameter)
}

func sendResult2BServer(token string, retCode string, feedBackParameter interface{}) {
	config := config.GetConfig()
	var url string
	bServer := config.BServer
	//for i := 0; i < len(bServer); i++ {
	//	if strings.EqualFold(token, bServer[i].Token) {
	//		url = bServer[i].NotifyUrl
	//	}
	//}
	if v, ok := bServer[token]; ok {
		url = v.NotifyUrl
	}

	if url == "" {
		fmt.Println("cant not find notify url ")
		return
	}
	body := make(map[string]interface{})
	body["retCode"] = retCode
	var status string
	switch retCode {
	case retCodeFinish:
		status = "ok"
	case retCodeMd5:
		status = "md5 error"
	case retCodeTimeout:
		status = "time out"
	default:
		status = "system error"
	}
	body["status"] = status
	body["date"] = feedBackParameter
	bytesData, err := json.Marshal(body)
	if err != nil {
		fmt.Println("json Marshal err:", err)
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
	resp, err2 := client.Do(request)
	defer resp.Body.Close()
	if err != nil {
		fmt.Println("http request err:", err2)
	}
}

func checkBlock(b *block, blocks []*block) bool {
	if len(blocks) == 0 {
		return true
	} else {
		var result bool
		for i := 0; i < len(blocks); i++ {
			if b.start < blocks[i].end && b.end > blocks[i].start {
				result = false
				break
			} else {
				result = true
			}
		}
		return result
	}
}

func deleteExpectBlock(t *task, eByte expectByte) {
	expect := t.expect
	index := -1
	for i := 0; i < len(expect); i++ {
		e := expect[i]
		if e.start == eByte.start {
			index = i
			break
		}
	}
	if index >= 0 {
		expect = append(expect[:index], expect[index+1:]...)
	}
	t.expect = expect
}
