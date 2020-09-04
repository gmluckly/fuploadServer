package utils

import "testing"
import "fmt"

func TestMakeTaskId(t *testing.T) {
	result := MakeTaskId()
	fmt.Println("result:", result)
}
