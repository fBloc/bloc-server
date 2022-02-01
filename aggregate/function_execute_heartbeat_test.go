package aggregate

import (
	"testing"
	"time"

	"github.com/fBloc/bloc-server/value_object"
	. "github.com/smartystreets/goconvey/convey"
)

func TestFunctionExecuteHeartBeatZero(t *testing.T) {
	Convey("IsZero hit", t, func() {
		var functionExecuteHeartBeat *FunctionExecuteHeartBeat
		So(functionExecuteHeartBeat.IsZero(), ShouldBeTrue)

		Convey("zero attrs check", func() {
			So(functionExecuteHeartBeat.IsTimeout(10), ShouldBeFalse)
		})
	})

	Convey("IsZero miss", t, func() {
		functionExecuteHeartBeat := NewFunctionExecuteHeartBeat(value_object.NewUUID())
		So(functionExecuteHeartBeat.IsZero(), ShouldBeFalse)
	})
}

func TestFunctionExecuteHeartBeat(t *testing.T) {
	functionExecuteHeartBeat := NewFunctionExecuteHeartBeat(value_object.NewUUID())
	sleepSec := 3
	time.Sleep(time.Duration(sleepSec) * time.Second)

	Convey("timeOutCheck", t, func() {
		So(functionExecuteHeartBeat.IsTimeout(float64(sleepSec+10)), ShouldBeFalse)
		So(functionExecuteHeartBeat.IsTimeout(float64(sleepSec-1)), ShouldBeTrue)
	})
}
