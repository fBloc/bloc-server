package aggregate

import (
	"context"
	"testing"

	. "github.com/smartystreets/goconvey/convey"
)

func TestFunctionRunRecordIsZero(t *testing.T) {
	Convey("IsZero hit", t, func() {
		var functionRunRecord *FunctionRunRecord = nil
		So(functionRunRecord.IsZero(), ShouldBeTrue)
	})

	Convey("IsZero miss", t, func() {
		flowRunRecord := NewCrontabTriggeredRunRecord(context.TODO(), &fakeFlow)
		functionRunRecord := NewFunctionRunRecordFromFlowDriven(
			context.TODO(), functionAdd, *flowRunRecord, secondFlowFunctionID)
		So(functionRunRecord.IsZero(), ShouldBeFalse)
	})

	Convey("nil attrs check", t, func() {
		var funcRunRecord *FunctionRunRecord = nil
		So(funcRunRecord.UsedSeconds(), ShouldEqual, 0)
		So(funcRunRecord.Failed(), ShouldBeFalse)
		So(funcRunRecord.Finished(), ShouldBeFalse)
		So(func() { funcRunRecord.SetSuc() }, ShouldNotPanic)
		So(func() { funcRunRecord.SetFail("") }, ShouldNotPanic)
	})
}

func TestFunctionRunRecord(t *testing.T) {
	flowRunRecord := NewCrontabTriggeredRunRecord(context.TODO(), &fakeFlow)
	functionRunRecord := NewFunctionRunRecordFromFlowDriven(
		context.TODO(), functionAdd, *flowRunRecord, secondFlowFunctionID)

	Convey("initial FunctionRunRecord attrs check", t, func() {
		So(functionRunRecord.UsedSeconds(), ShouldBeGreaterThan, 0)
		So(functionRunRecord.Failed(), ShouldBeFalse)
		So(functionRunRecord.Finished(), ShouldBeFalse)
		So(functionRunRecord.Suc, ShouldBeFalse)
	})

	Convey("set", t, func() {
		So(functionRunRecord.Suc, ShouldEqual, false)
		functionRunRecord.SetSuc()
		So(functionRunRecord.Suc, ShouldBeTrue)
		So(functionRunRecord.Failed(), ShouldBeFalse)
		So(functionRunRecord.Finished(), ShouldBeTrue)

		functionRunRecord.SetFail("")
		So(functionRunRecord.Suc, ShouldEqual, false)
	})
}
