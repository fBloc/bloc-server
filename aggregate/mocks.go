package aggregate

import (
	"github.com/brianvoe/gofakeit/v6"
	"github.com/fBloc/bloc-server/config"
	"github.com/fBloc/bloc-server/pkg/ipt"
	"github.com/fBloc/bloc-server/pkg/opt"
	"github.com/fBloc/bloc-server/pkg/value_type"
	"github.com/fBloc/bloc-server/value_object"
)

var (
	readeUser         = User{ID: value_object.NewUUID()}
	executeUser       = User{ID: value_object.NewUUID()}
	allPermissionUser = User{ID: value_object.NewUUID()}
	superUser         = User{ID: value_object.NewUUID(), IsSuper: true}
)

var functionAdd = Function{
	ID:           value_object.NewUUID(),
	Name:         "two add",
	GroupName:    "math operation",
	ProviderName: "test",
	Description:  "test function",
	Ipts: ipt.IptSlice{
		{
			Key:     "to_add_ints",
			Display: "to_add_ints",
			Must:    true,
			Components: []*ipt.IptComponent{
				{
					ValueType:       value_type.IntValueType,
					FormControlType: value_object.InputFormControl,
					Hint:            "加数",
					DefaultValue:    0,
					AllowMulti:      true,
				},
			},
		},
	},
	Opts: []*opt.Opt{
		{
			Key:         "sum",
			Description: "sum of your inputs",
			ValueType:   value_type.IntValueType,
			IsArray:     false,
		},
		{
			Key:         "describe",
			Description: "diff value type opt 4 test",
			ValueType:   value_type.StringValueType,
			IsArray:     false,
		},
	},
	ProcessStages:           []string{"parsing ipt", "finished parse ipt & start do the math", "finished"},
	ReadUserIDs:             []value_object.UUID{readeUser.ID, allPermissionUser.ID},
	ExecuteUserIDs:          []value_object.UUID{executeUser.ID, allPermissionUser.ID},
	AssignPermissionUserIDs: []value_object.UUID{allPermissionUser.ID},
}

var functionMultiply = Function{
	ID:           value_object.NewUUID(),
	Name:         "two multiply",
	GroupName:    "math operation",
	ProviderName: "test",
	Description:  "test function",
	Ipts: ipt.IptSlice{
		{
			Key:     "multiplier",
			Display: "multiplier",
			Must:    true,
			Components: []*ipt.IptComponent{
				{
					ValueType:       value_type.IntValueType,
					FormControlType: value_object.InputFormControl,
					Hint:            "乘数",
					DefaultValue:    0,
					AllowMulti:      true,
				},
			},
		},
		{
			Key:     "multiplicand",
			Display: "multiplicand",
			Must:    true,
			Components: []*ipt.IptComponent{
				{
					ValueType:       value_type.IntValueType,
					FormControlType: value_object.InputFormControl,
					Hint:            "被乘数",
					DefaultValue:    0,
					AllowMulti:      true,
				},
			},
		},
	},
	Opts: []*opt.Opt{
		{
			Key:         "result",
			Description: "result of multiply",
			ValueType:   value_type.IntValueType,
			IsArray:     false,
		},
	},
	ProcessStages:           []string{"parsing ipt", "finished parse ipt & start do the math", "finished"},
	ReadUserIDs:             []value_object.UUID{readeUser.ID, allPermissionUser.ID},
	ExecuteUserIDs:          []value_object.UUID{executeUser.ID, allPermissionUser.ID},
	AssignPermissionUserIDs: []value_object.UUID{allPermissionUser.ID},
}

var (
	secondFlowFunctionID               = value_object.NewUUID().String()
	thirdFlowFunctionID                = value_object.NewUUID().String()
	validFlowFunctionIDMapFlowFunction = map[string]*FlowFunction{
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
	fakeFlow = Flow{
		ID:                            value_object.NewUUID(),
		Name:                          gofakeit.Name(),
		IsDraft:                       true,
		OriginID:                      value_object.NewUUID(),
		CreateUserID:                  value_object.NewUUID(),
		FlowFunctionIDMapFlowFunction: validFlowFunctionIDMapFlowFunction,
	}
)
