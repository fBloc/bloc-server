package util

import (
	"testing"

	. "github.com/smartystreets/goconvey/convey"
)

func TestUrlEncode(t *testing.T) {
	str := "afhas/6#12"

	Convey("test urlencode", t, func() {
		encodeStr := UrlEncode(str)
		So(encodeStr, ShouldNotContainSubstring, "#")
	})

	Convey("test md5Digest", t, func() {
		encodeStr := Md5Digest(str)
		So(encodeStr, ShouldNotEqual, "")
		So(encodeStr, ShouldNotEqual, str)
	})

	Convey("test Sha1", t, func() {
		encodeStr := Sha1([]byte(str))
		So(encodeStr, ShouldNotEqual, "")
		So(encodeStr, ShouldNotEqual, str)
	})
}
