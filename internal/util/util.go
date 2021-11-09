package util

import (
	"crypto/md5"
	"crypto/sha1"
	"encoding/hex"
	"fmt"
	"io"
	"net/url"
)

func EncodeString(str string) string {
	return url.QueryEscape(str)
}

func Md5Digest(str string) string {
	w := md5.New()
	io.WriteString(w, str)
	md5str2 := fmt.Sprintf("%x", w.Sum(nil))
	return md5str2
}

func Sha1(data []byte) string {
	_sha1 := sha1.New()
	_sha1.Write(data)
	return hex.EncodeToString(_sha1.Sum([]byte("")))
}
