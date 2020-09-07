package main

import (
	"testing"
)

func TestSendFile(t *testing.T) {

	var userId int64 = 123456
	token := "cp00000010103010301"
	targetContent := "/home/mpr/1.exe"
	storePath := "/tmp/fupload"
	sendFile(userId, token, targetContent, storePath)

}
