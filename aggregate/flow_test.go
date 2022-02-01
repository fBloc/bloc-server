package aggregate

import (
	"testing"

	"github.com/brianvoe/gofakeit/v6"
	"github.com/fBloc/bloc-server/config"
	"github.com/fBloc/bloc-server/pkg/value_type"
	"github.com/fBloc/bloc-server/value_object"
	. "github.com/smartystreets/goconvey/convey"
)

func TestCheckFlowValid(t *testing.T) {
	Convey("valid hit check", t, func() {
		for flowFuncID, flowFunction := range validFlowFunctionIDMapFlowFunction {
			valid, err := flowFunction.CheckValid(
				flowFuncID,
				validFlowFunctionIDMapFlowFunction)
			So(err, ShouldBeNil)
			So(valid, ShouldBeTrue)
		}
	})

	Convey("upstream check", t, func() {
		flowFunctionIDMapFlowFunction := map[string]*FlowFunction{
			config.FlowFunctionStartID: {
				FunctionID:                value_object.NillUUID,
				Note:                      "start node",
				UpstreamFlowFunctionIDs:   []string{},
				DownstreamFlowFunctionIDs: []string{secondFlowFunctionID},
				ParamIpts:                 [][]IptComponentConfig{},
			},
			secondFlowFunctionID: {
				FunctionID:                functionAdd.ID,
				Note:                      "add",
				UpstreamFlowFunctionIDs:   []string{},
				DownstreamFlowFunctionIDs: []string{},
				ParamIpts: [][]IptComponentConfig{
					{
						{
							Blank:     false,
							IptWay:    value_object.UserIpt,
							ValueType: value_type.StringValueType,
							Value:     []int{1, 2, 3},
						},
					},
				},
			},
		}
		for flowFuncID, flowFunction := range flowFunctionIDMapFlowFunction {
			valid, err := flowFunction.CheckValid(
				flowFuncID, validFlowFunctionIDMapFlowFunction)
			if flowFuncID == secondFlowFunctionID {
				So(err, ShouldNotBeNil)
				So(valid, ShouldBeFalse)
			}
		}
	})

	Convey("start node mismatch amount check", t, func() {
		flowFunctionIDMapFlowFunction := map[string]*FlowFunction{
			config.FlowFunctionStartID: {
				FunctionID:                value_object.NillUUID,
				Note:                      "start node",
				UpstreamFlowFunctionIDs:   []string{},
				DownstreamFlowFunctionIDs: []string{},
				ParamIpts:                 [][]IptComponentConfig{},
			},
			secondFlowFunctionID: {
				FunctionID:                functionAdd.ID,
				Note:                      "add",
				UpstreamFlowFunctionIDs:   []string{config.FlowFunctionStartID},
				DownstreamFlowFunctionIDs: []string{},
				ParamIpts: [][]IptComponentConfig{
					{
						{
							Blank:     false,
							IptWay:    value_object.UserIpt,
							ValueType: value_type.StringValueType,
							Value:     []int{1, 2, 3},
						},
					},
				},
			},
		}
		for flowFuncID, flowFunction := range flowFunctionIDMapFlowFunction {
			valid, err := flowFunction.CheckValid(
				flowFuncID, validFlowFunctionIDMapFlowFunction)
			if flowFuncID == config.FlowFunctionStartID {
				So(err, ShouldNotBeNil)
				So(valid, ShouldBeFalse)
			}
		}
	})

	Convey("start node wrong downstream id check", t, func() {
		flowFunctionIDMapFlowFunction := map[string]*FlowFunction{
			config.FlowFunctionStartID: {
				FunctionID:                value_object.NillUUID,
				Note:                      "start node",
				UpstreamFlowFunctionIDs:   []string{},
				DownstreamFlowFunctionIDs: []string{secondFlowFunctionID + "wrong"},
				ParamIpts:                 [][]IptComponentConfig{},
			},
			secondFlowFunctionID: {
				FunctionID:                functionAdd.ID,
				Note:                      "add",
				UpstreamFlowFunctionIDs:   []string{config.FlowFunctionStartID},
				DownstreamFlowFunctionIDs: []string{},
				ParamIpts: [][]IptComponentConfig{
					{
						{
							Blank:     false,
							IptWay:    value_object.UserIpt,
							ValueType: value_type.StringValueType,
							Value:     []int{1, 2, 3},
						},
					},
				},
			},
		}
		for flowFuncID, flowFunction := range flowFunctionIDMapFlowFunction {
			valid, err := flowFunction.CheckValid(
				flowFuncID, validFlowFunctionIDMapFlowFunction)
			if flowFuncID == config.FlowFunctionStartID {
				So(err, ShouldNotBeNil)
				So(valid, ShouldBeFalse)
			}
		}
	})

	Convey("upstream wrong flow function id check", t, func() {
		flowFunctionIDMapFlowFunction := map[string]*FlowFunction{
			config.FlowFunctionStartID: {
				FunctionID:                value_object.NillUUID,
				Note:                      "start node",
				UpstreamFlowFunctionIDs:   []string{},
				DownstreamFlowFunctionIDs: []string{secondFlowFunctionID},
				ParamIpts:                 [][]IptComponentConfig{},
			},
			secondFlowFunctionID: {
				FunctionID:                functionAdd.ID,
				Note:                      "add",
				UpstreamFlowFunctionIDs:   []string{config.FlowFunctionStartID + "miss"},
				DownstreamFlowFunctionIDs: []string{},
				ParamIpts: [][]IptComponentConfig{
					{
						{
							Blank:     false,
							IptWay:    value_object.UserIpt,
							ValueType: value_type.StringValueType,
							Value:     []int{1, 2, 3},
						},
					},
				},
			},
		}
		for flowFuncID, flowFunction := range flowFunctionIDMapFlowFunction {
			valid, err := flowFunction.CheckValid(
				flowFuncID, validFlowFunctionIDMapFlowFunction)
			if flowFuncID == secondFlowFunctionID {
				So(err, ShouldNotBeNil)
				So(valid, ShouldBeFalse)
			}
		}
	})

	Convey("must set function for ipt param check", t, func() {
		flowFunctionIDMapFlowFunction := map[string]*FlowFunction{
			config.FlowFunctionStartID: {
				FunctionID:                value_object.NillUUID,
				Note:                      "start node",
				UpstreamFlowFunctionIDs:   []string{},
				DownstreamFlowFunctionIDs: []string{secondFlowFunctionID},
				ParamIpts:                 [][]IptComponentConfig{},
			},
			secondFlowFunctionID: {
				FunctionID: functionAdd.ID,
				// there should set function
				Note:                      "add",
				UpstreamFlowFunctionIDs:   []string{config.FlowFunctionStartID},
				DownstreamFlowFunctionIDs: []string{},
				ParamIpts: [][]IptComponentConfig{
					{
						{
							Blank: true, // this cannot be true
						},
					},
				},
			},
		}
		for flowFuncID, flowFunction := range flowFunctionIDMapFlowFunction {
			valid, err := flowFunction.CheckValid(
				flowFuncID, validFlowFunctionIDMapFlowFunction)
			if flowFuncID == secondFlowFunctionID {
				So(err, ShouldNotBeNil)
				So(valid, ShouldBeFalse)
			}
		}
	})

	Convey("must set but blank check", t, func() {
		flowFunctionIDMapFlowFunction := map[string]*FlowFunction{
			config.FlowFunctionStartID: {
				FunctionID:                value_object.NillUUID,
				Note:                      "start node",
				UpstreamFlowFunctionIDs:   []string{},
				DownstreamFlowFunctionIDs: []string{secondFlowFunctionID},
				ParamIpts:                 [][]IptComponentConfig{},
			},
			secondFlowFunctionID: {
				FunctionID:                functionAdd.ID,
				Function:                  &functionAdd,
				Note:                      "add",
				UpstreamFlowFunctionIDs:   []string{config.FlowFunctionStartID},
				DownstreamFlowFunctionIDs: []string{},
				ParamIpts: [][]IptComponentConfig{
					{
						{
							Blank: true, // this cannot be true
						},
					},
				},
			},
		}
		for flowFuncID, flowFunction := range flowFunctionIDMapFlowFunction {
			valid, err := flowFunction.CheckValid(
				flowFuncID, validFlowFunctionIDMapFlowFunction)
			if flowFuncID == secondFlowFunctionID {
				So(err, ShouldNotBeNil)
				So(valid, ShouldBeFalse)
			}
		}
	})

	Convey("connection flow_function_id not valid check", t, func() {
		flowFunctionIDMapFlowFunction := map[string]*FlowFunction{
			config.FlowFunctionStartID: {
				FunctionID:                value_object.NillUUID,
				Note:                      "start node",
				UpstreamFlowFunctionIDs:   []string{},
				DownstreamFlowFunctionIDs: []string{secondFlowFunctionID},
				ParamIpts:                 [][]IptComponentConfig{},
			},
			secondFlowFunctionID: {
				FunctionID:                functionAdd.ID,
				Function:                  &functionAdd,
				Note:                      "add",
				UpstreamFlowFunctionIDs:   []string{config.FlowFunctionStartID},
				DownstreamFlowFunctionIDs: []string{thirdFlowFunctionID},
				ParamIpts: [][]IptComponentConfig{
					{
						{
							Blank:     false,
							IptWay:    value_object.UserIpt,
							ValueType: value_type.StringValueType,
							Value:     []int{1, 2, 3},
						},
					},
				},
			},
			thirdFlowFunctionID: {
				FunctionID:                functionMultiply.ID,
				Function:                  &functionMultiply,
				Note:                      "multiply",
				UpstreamFlowFunctionIDs:   []string{secondFlowFunctionID},
				DownstreamFlowFunctionIDs: []string{},
				ParamIpts: [][]IptComponentConfig{
					{
						{
							Blank:          false,
							IptWay:         value_object.Connection,
							ValueType:      value_type.IntValueType,
							FlowFunctionID: secondFlowFunctionID + "missmatch", // here wrong
							Key:            "sum",
						},
					},
					{
						{
							Blank:     false,
							IptWay:    value_object.UserIpt,
							ValueType: value_type.IntValueType,
							Value:     10,
						},
					},
				},
			},
		}
		for flowFuncID, flowFunction := range flowFunctionIDMapFlowFunction {
			valid, err := flowFunction.CheckValid(
				flowFuncID, validFlowFunctionIDMapFlowFunction)
			if flowFuncID == secondFlowFunctionID {
				So(err, ShouldBeNil)
				So(valid, ShouldBeTrue)
			}
			if flowFuncID == thirdFlowFunctionID {
				So(err, ShouldNotBeNil)
				So(valid, ShouldBeFalse)
			}
		}
	})

	Convey("connection function must be direct upstream", t, func() {
		flowFunctionIDMapFlowFunction := map[string]*FlowFunction{
			config.FlowFunctionStartID: {
				FunctionID:                value_object.NillUUID,
				Note:                      "start node",
				UpstreamFlowFunctionIDs:   []string{},
				DownstreamFlowFunctionIDs: []string{secondFlowFunctionID},
				ParamIpts:                 [][]IptComponentConfig{},
			},
			secondFlowFunctionID: {
				FunctionID:                functionAdd.ID,
				Function:                  &functionAdd,
				Note:                      "add",
				UpstreamFlowFunctionIDs:   []string{config.FlowFunctionStartID},
				DownstreamFlowFunctionIDs: []string{thirdFlowFunctionID},
				ParamIpts: [][]IptComponentConfig{
					{
						{
							Blank:     false,
							IptWay:    value_object.UserIpt,
							ValueType: value_type.StringValueType,
							Value:     []int{1, 2, 3},
						},
					},
				},
			},
			thirdFlowFunctionID: {
				FunctionID:                functionMultiply.ID,
				Function:                  &functionMultiply,
				Note:                      "multiply",
				UpstreamFlowFunctionIDs:   []string{secondFlowFunctionID},
				DownstreamFlowFunctionIDs: []string{},
				ParamIpts: [][]IptComponentConfig{
					{
						{
							Blank:          false,
							IptWay:         value_object.Connection,
							ValueType:      value_type.IntValueType,
							FlowFunctionID: config.FlowFunctionStartID, // here wrong
							Key:            "sum",
						},
					},
					{
						{
							Blank:     false,
							IptWay:    value_object.UserIpt,
							ValueType: value_type.IntValueType,
							Value:     10,
						},
					},
				},
			},
		}
		for flowFuncID, flowFunction := range flowFunctionIDMapFlowFunction {
			valid, err := flowFunction.CheckValid(
				flowFuncID, validFlowFunctionIDMapFlowFunction)
			if flowFuncID == secondFlowFunctionID {
				So(err, ShouldBeNil)
				So(valid, ShouldBeTrue)
			}
			if flowFuncID == thirdFlowFunctionID {
				So(err, ShouldNotBeNil)
				So(valid, ShouldBeFalse)
			}
		}
	})

	Convey("wrong user-ipt value type", t, func() {
		flowFunctionIDMapFlowFunction := map[string]*FlowFunction{
			config.FlowFunctionStartID: {
				FunctionID:                value_object.NillUUID,
				Note:                      "start node",
				UpstreamFlowFunctionIDs:   []string{},
				DownstreamFlowFunctionIDs: []string{secondFlowFunctionID},
				ParamIpts:                 [][]IptComponentConfig{},
			},
			secondFlowFunctionID: {
				FunctionID:                functionAdd.ID,
				Function:                  &functionAdd,
				Note:                      "add",
				UpstreamFlowFunctionIDs:   []string{config.FlowFunctionStartID},
				DownstreamFlowFunctionIDs: []string{thirdFlowFunctionID},
				ParamIpts: [][]IptComponentConfig{
					{
						{
							Blank:     false,
							IptWay:    value_object.UserIpt,
							ValueType: value_type.IntValueType,
							Value:     []string{"1", "2", gofakeit.Name()}, // wrong type
						},
					},
				},
			},
		}
		for flowFuncID, flowFunction := range flowFunctionIDMapFlowFunction {
			valid, err := flowFunction.CheckValid(
				flowFuncID, validFlowFunctionIDMapFlowFunction)
			if flowFuncID == secondFlowFunctionID {
				So(err, ShouldNotBeNil)
				So(valid, ShouldBeFalse)
			}
		}
	})

	Convey("wrong connection value type", t, func() {
		flowFunctionIDMapFlowFunction := map[string]*FlowFunction{
			config.FlowFunctionStartID: {
				FunctionID:                value_object.NillUUID,
				Note:                      "start node",
				UpstreamFlowFunctionIDs:   []string{},
				DownstreamFlowFunctionIDs: []string{secondFlowFunctionID},
				ParamIpts:                 [][]IptComponentConfig{},
			},
			secondFlowFunctionID: {
				FunctionID:                functionAdd.ID,
				Function:                  &functionAdd,
				Note:                      "add",
				UpstreamFlowFunctionIDs:   []string{config.FlowFunctionStartID},
				DownstreamFlowFunctionIDs: []string{thirdFlowFunctionID},
				ParamIpts: [][]IptComponentConfig{
					{
						{
							Blank:     false,
							IptWay:    value_object.UserIpt,
							ValueType: value_type.StringValueType,
							Value:     []int{1, 2, 3},
						},
					},
				},
			},
			thirdFlowFunctionID: {
				FunctionID:                functionMultiply.ID,
				Function:                  &functionMultiply,
				Note:                      "multiply",
				UpstreamFlowFunctionIDs:   []string{secondFlowFunctionID},
				DownstreamFlowFunctionIDs: []string{},
				ParamIpts: [][]IptComponentConfig{
					{
						{
							Blank:          false,
							IptWay:         value_object.Connection,
							ValueType:      value_type.IntValueType,
							FlowFunctionID: secondFlowFunctionID,
							Key:            "describe", // wrong type
						},
					},
					{
						{
							Blank:     false,
							IptWay:    value_object.UserIpt,
							ValueType: value_type.IntValueType,
							Value:     10,
						},
					},
				},
			},
		}
		for flowFuncID, flowFunction := range flowFunctionIDMapFlowFunction {
			valid, err := flowFunction.CheckValid(
				flowFuncID, validFlowFunctionIDMapFlowFunction)
			if flowFuncID == secondFlowFunctionID {
				So(err, ShouldBeNil)
				So(valid, ShouldBeTrue)
			}
			if flowFuncID == thirdFlowFunctionID {
				So(err, ShouldNotBeNil)
				So(valid, ShouldBeFalse)
			}
		}
	})
}

func TestFlowIsZero(t *testing.T) {
	Convey("should be zero", t, func() {
		var flow *Flow = nil
		So(flow.IsZero(), ShouldBeTrue)
	})

	Convey("should not be zero", t, func() {
		So(fakeFlow.IsZero(), ShouldBeFalse)
	})
}

func TestFlowHaveRetryStrategy(t *testing.T) {
	Convey("retry strategy", t, func() {
		var nilFlow *Flow = nil
		So(nilFlow.HaveRetryStrategy(), ShouldBeFalse)
		So(fakeFlow.HaveRetryStrategy(), ShouldBeFalse)

		fakeFlow.RetryIntervalInSecond = 100
		So(fakeFlow.HaveRetryStrategy(), ShouldBeFalse)

		Convey("have strategy", func() {
			fakeFlow.RetryAmount = 3
			So(fakeFlow.HaveRetryStrategy(), ShouldBeTrue)
		})
	})
}

func TestPermission(t *testing.T) {
	readeUser := User{ID: value_object.NewUUID()}
	writeUser := User{ID: value_object.NewUUID()}
	executeUser := User{ID: value_object.NewUUID()}
	deleteUser := User{ID: value_object.NewUUID()}
	assignPermissionUser := User{ID: value_object.NewUUID()}
	superUser := User{ID: value_object.NewUUID(), IsSuper: true}

	fakeFlow.ReadUserIDs = []value_object.UUID{readeUser.ID}
	fakeFlow.WriteUserIDs = []value_object.UUID{writeUser.ID}
	fakeFlow.ExecuteUserIDs = []value_object.UUID{executeUser.ID}
	fakeFlow.DeleteUserIDs = []value_object.UUID{deleteUser.ID}
	fakeFlow.AssignPermissionUserIDs = []value_object.UUID{assignPermissionUser.ID}

	Convey("read", t, func() {
		So(fakeFlow.UserCanRead(&readeUser), ShouldBeTrue)
		So(fakeFlow.UserCanRead(&writeUser), ShouldBeFalse)
		So(fakeFlow.UserCanRead(&executeUser), ShouldBeFalse)
		So(fakeFlow.UserCanRead(&deleteUser), ShouldBeFalse)
		So(fakeFlow.UserCanRead(&assignPermissionUser), ShouldBeFalse)
		So(fakeFlow.UserCanRead(&superUser), ShouldBeTrue)
	})

	Convey("write", t, func() {
		So(fakeFlow.UserCanWrite(&writeUser), ShouldBeTrue)
		So(fakeFlow.UserCanWrite(&readeUser), ShouldBeFalse)
		So(fakeFlow.UserCanWrite(&executeUser), ShouldBeFalse)
		So(fakeFlow.UserCanWrite(&deleteUser), ShouldBeFalse)
		So(fakeFlow.UserCanWrite(&assignPermissionUser), ShouldBeFalse)
		So(fakeFlow.UserCanWrite(&superUser), ShouldBeTrue)
	})

	Convey("execute", t, func() {
		So(fakeFlow.UserCanExecute(&executeUser), ShouldBeTrue)
		So(fakeFlow.UserCanExecute(&readeUser), ShouldBeFalse)
		So(fakeFlow.UserCanExecute(&writeUser), ShouldBeFalse)
		So(fakeFlow.UserCanExecute(&deleteUser), ShouldBeFalse)
		So(fakeFlow.UserCanExecute(&assignPermissionUser), ShouldBeFalse)
		So(fakeFlow.UserCanExecute(&superUser), ShouldBeTrue)
	})

	Convey("delete", t, func() {
		So(fakeFlow.UserCanDelete(&deleteUser), ShouldBeTrue)
		So(fakeFlow.UserCanDelete(&readeUser), ShouldBeFalse)
		So(fakeFlow.UserCanDelete(&writeUser), ShouldBeFalse)
		So(fakeFlow.UserCanDelete(&executeUser), ShouldBeFalse)
		So(fakeFlow.UserCanDelete(&assignPermissionUser), ShouldBeFalse)
		So(fakeFlow.UserCanDelete(&superUser), ShouldBeTrue)
	})

	Convey("assign permission", t, func() {
		So(fakeFlow.UserCanAssignPermission(&assignPermissionUser), ShouldBeTrue)
		So(fakeFlow.UserCanAssignPermission(&readeUser), ShouldBeFalse)
		So(fakeFlow.UserCanAssignPermission(&writeUser), ShouldBeFalse)
		So(fakeFlow.UserCanAssignPermission(&executeUser), ShouldBeFalse)
		So(fakeFlow.UserCanAssignPermission(&deleteUser), ShouldBeFalse)
		So(fakeFlow.UserCanAssignPermission(&superUser), ShouldBeTrue)
	})
}
