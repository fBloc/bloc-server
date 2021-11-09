package util

import (
	"fmt"
	"testing"
)

func TestUtil(t *testing.T) {
	str := "afhas/6#12"
	encodeStr := EncodeString(str)

	fmt.Println(encodeStr)
}
