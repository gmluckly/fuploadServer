package example

import (
	"testing"
)

func TestSendFile(t *testing.T) {

	var userId int64 = 123456
	var uToken string = "12345"
	var bToken string = "12345"
	targetContent := "/home/mpr/1.exe"
	storePath := "/tmp/fupload"
	sendFile(userId, uToken, bToken, targetContent, storePath)

}
