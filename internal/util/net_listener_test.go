package util

import (
	"testing"

	. "github.com/smartystreets/goconvey/convey"
)

func TestNetListener(t *testing.T) {
	Convey("auto ip & port", t, func() {
		ip, port, _ := NewAutoAddressNetListener()
		So(port, ShouldBeGreaterThan, 0)
		So(ip, ShouldNotEqual, "")
	})
}
