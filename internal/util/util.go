package util

import (
	"crypto/md5"
	"crypto/sha1"
	"encoding/hex"
	"net/url"
)

func EncodeString(str string) string {
	return url.QueryEscape(str)
}

func Md5Digest(str string) string {
	hash := md5.Sum([]byte(str))
	return hex.EncodeToString(hash[:])
}

func Sha1(data []byte) string {
	_sha1 := sha1.New()
	_sha1.Write(data)
	return hex.EncodeToString(_sha1.Sum([]byte("")))
}
