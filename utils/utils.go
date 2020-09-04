package utils

import (
	"crypto/md5"
	"encoding/hex"
	"io"
	"math"
	"math/rand"
	"os"
	"time"
)

func MakeTaskId() int64 {
	rand.Seed(time.Now().Unix())
	t := time.Now().UnixNano()
	return t
}

const fileChunk = 8192

func FileMd5(filePath string) string {
	file, err := os.Open(filePath)
	if err != nil {
		return ""
	}
	defer file.Close()

	info, _ := file.Stat()
	fileSize := info.Size()
	blocks := uint64(math.Ceil(float64(fileSize) / float64(fileChunk)))
	hash := md5.New()
	for i := uint64(0); i < blocks; i++ {
		blockSize := int(math.Min(fileChunk, float64(fileSize-int64(i*fileChunk))))
		buf := make([]byte, blockSize)

		file.Read(buf)
		io.WriteString(hash, string(buf))
	}
	md5 := hash.Sum(nil)
	md5Str := hex.EncodeToString(md5)
	return md5Str
}

func GetTmpMd5(data []byte) string {
	md5Ctx := md5.New()
	md5Ctx.Write(data)
	cipherStr := md5Ctx.Sum(nil)
	return hex.EncodeToString(cipherStr)
}

func CheckAndMakeDir(folderPath string) error {
	if isFolderExist(folderPath) {
		return nil
	}
	if err := os.MkdirAll(folderPath, os.ModePerm); err != nil {
		return err
	}
	return nil
}

func isFolderExist(folderPath string) bool {
	fileinfo, err := os.Stat(folderPath)
	if err == nil && fileinfo.IsDir() {
		return true
	}
	return false
}
