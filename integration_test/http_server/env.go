package http_server

import (
	"fmt"

	"github.com/brianvoe/gofakeit/v6"
	"github.com/fBloc/bloc-server/aggregate"
	"github.com/fBloc/bloc-server/config"
	"github.com/fBloc/bloc-server/internal/conns/influxdb"
	"github.com/fBloc/bloc-server/internal/conns/minio"
	"github.com/fBloc/bloc-server/internal/conns/mongodb"
	"github.com/fBloc/bloc-server/internal/conns/rabbit"
	"github.com/fBloc/bloc-server/pkg/ipt"
	"github.com/fBloc/bloc-server/pkg/opt"
	"github.com/fBloc/bloc-server/pkg/value_type"
	"github.com/fBloc/bloc-server/value_object"
)

var (
	influxDBConf = &influxdb.InfluxDBConfig{
		UserName:     "testUser",
		Password:     "password",
		Token:        "1326143eba0c8e1a408a014e9a63d767",
		Organization: "bloc-test",
	}

	minioConf = &minio.MinioConfig{
		BucketName:     "test",
		AccessKey:      "blocMinio",
		AccessPassword: "blocMinioPasswd",
	}

	mongoConf = &mongodb.MongoConfig{
		Db:       "bloc-test-mongo",
		User:     "root",
		Password: "password",
	}

	rabbitConf = &rabbit.RabbitConfig{
		User:     "blocRabbit",
		Password: "blocRabbitPasswd",
	}

	serverHost    = "localhost"
	serverPort    = 8484
	serverAddress = fmt.Sprintf("%s:%d", serverHost, serverPort)

	loginedToken = ""
	superuserID  value_object.UUID
)

func superuserHeader() map[string]string {
	return map[string]string{"token": loginedToken}
}

func nobodyHeader() map[string]string {
	return map[string]string{"token": value_object.NewUUID().String()}
}

// function about
var (
	readeUser         = aggregate.User{ID: value_object.NewUUID()}
	executeUser       = aggregate.User{ID: value_object.NewUUID()}
	allPermissionUser = aggregate.User{ID: value_object.NewUUID()}
	fakeAggFunction   = aggregate.Function{
		Name:         "two add",
		GroupName:    "test",
		ProviderName: "test",
		Description:  "",
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
		},
		ProcessStages:           []string{"parsing ipt", "finished parse ipt & start do the math", "finished"},
		ReadUserIDs:             []value_object.UUID{readeUser.ID, allPermissionUser.ID},
		ExecuteUserIDs:          []value_object.UUID{executeUser.ID, allPermissionUser.ID},
		AssignPermissionUserIDs: []value_object.UUID{allPermissionUser.ID},
	}
)

var aggFuncAddFlowFunctionID = value_object.NewUUID().String()
var aggFuncAdd = aggregate.Function{
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

var aggFuncMultiplyFlowFunctionID = value_object.NewUUID().String()
var aggFuncMultiply = aggregate.Function{
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

func getFakeAggFlow() *aggregate.Flow {
	validFlowFunctionIDMapFlowFunction := map[string]*aggregate.FlowFunction{
		config.FlowFunctionStartID: {
			FunctionID:                value_object.NillUUID,
			Note:                      "start node",
			UpstreamFlowFunctionIDs:   []string{},
			DownstreamFlowFunctionIDs: []string{aggFuncAddFlowFunctionID},
			ParamIpts:                 [][]aggregate.IptComponentConfig{},
		},
		aggFuncAddFlowFunctionID: {
			FunctionID:                aggFuncAdd.ID,
			Function:                  &aggFuncAdd,
			Note:                      "add",
			UpstreamFlowFunctionIDs:   []string{config.FlowFunctionStartID},
			DownstreamFlowFunctionIDs: []string{aggFuncMultiplyFlowFunctionID},
			ParamIpts: [][]aggregate.IptComponentConfig{
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
		aggFuncMultiplyFlowFunctionID: {
			FunctionID:                aggFuncMultiply.ID,
			Function:                  &aggFuncMultiply,
			Note:                      "multiply",
			UpstreamFlowFunctionIDs:   []string{aggFuncAddFlowFunctionID},
			DownstreamFlowFunctionIDs: []string{},
			ParamIpts: [][]aggregate.IptComponentConfig{
				{
					{
						Blank:          false,
						IptWay:         value_object.Connection,
						ValueType:      value_type.IntValueType,
						FlowFunctionID: aggFuncAddFlowFunctionID,
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
	return &aggregate.Flow{
		Name:                          gofakeit.Name(),
		IsDraft:                       true,
		CreateUserID:                  value_object.NewUUID(),
		FlowFunctionIDMapFlowFunction: validFlowFunctionIDMapFlowFunction,
	}
}
