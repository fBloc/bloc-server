package aggregate

import (
	"context"
	"testing"

	"github.com/fBloc/bloc-server/value_object"
	. "github.com/smartystreets/goconvey/convey"
)

func TestFlowRunRecordIsZero(t *testing.T) {
	Convey("isZero", t, func() {
		var fRR *FlowRunRecord = nil
		So(fRR.IsZero(), ShouldBeTrue)
	})
}

func TestFlowRunRecordNewUserTriggeredRunRecord(t *testing.T) {
	Convey("NewUserTriggeredFlowRunRecord not valid user", t, func() {
		nobody := User{ID: value_object.NewUUID()}
		flowRR, err := NewUserTriggeredFlowRunRecord(context.TODO(), &fakeFlow, &nobody)
		So(err, ShouldNotBeNil)
		So(flowRR.IsZero(), ShouldBeTrue)
	})

	Convey("NewUserTriggeredFlowRunRecord valid user", t, func() {
		executer := User{ID: value_object.NewUUID()}
		fakeFlow.ExecuteUserIDs = []value_object.UUID{executer.ID}
		flowRunRecord, err := NewUserTriggeredFlowRunRecord(context.TODO(), &fakeFlow, &executer)
		So(err, ShouldBeNil)
		So(flowRunRecord.IsZero(), ShouldBeFalse)
		So(flowRunRecord.IsZero(), ShouldBeFalse)
	})
}

func TestFlowRunRecordNewCrontabTriggeredRunRecord(t *testing.T) {
	Convey("NewCrontabTriggeredRunRecord", t, func() {
		flowRunRecord := NewCrontabTriggeredRunRecord(context.TODO(), &fakeFlow)
		So(flowRunRecord.IsZero(), ShouldBeFalse)
	})
}
