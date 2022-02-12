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
	allChars = []string{
		"a", "b", "c", "d", "e", "f",
		"g", "h", "i", "j", "k", "l", "m", "n", "o", "p", "q", "r", "s", "e",
		"u", "v", "w", "x", "y", "z"}
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

	superUserToken string
	superUserID    value_object.UUID

	nobodyName      = gofakeit.Name()
	nobodyRawPasswd = gofakeit.Password(false, false, false, false, false, 16)
	nobodyID        value_object.UUID
	nobodyToken     string
)

func superuserHeader() map[string]string {
	return map[string]string{"token": superUserToken}
}

func nobodyHeader() map[string]string {
	return map[string]string{"token": nobodyToken}
}

func notExistUserHeader() map[string]string {
	return map[string]string{"token": value_object.NewUUID().String()}
}

// function about
var (
	readeUser         = aggregate.User{ID: value_object.NewUUID()}
	executeUser       = aggregate.User{ID: value_object.NewUUID()}
	allPermissionUser = aggregate.User{ID: value_object.NewUUID()}
	fakeAggFunction   = aggregate.Function{
		Name:         "two add",
		GroupName:    "server",
		ProviderName: "server",
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
var aggFuncAdd = &aggregate.Function{
	Name:         "add",
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
					Hint:            "addends",
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

var aggFuncMultiplyFlowFunctionID = value_object.NewUUID().String()
var aggFuncMultiply = &aggregate.Function{
	Name:         "multiply",
	GroupName:    "math operation",
	ProviderName: "test",
	Description:  "test function",
	Ipts: ipt.IptSlice{
		{
			Key:     "to_multiply_ints",
			Display: "to_multiply_ints",
			Must:    true,
			Components: []*ipt.IptComponent{
				{
					ValueType:       value_type.IntValueType,
					FormControlType: value_object.InputFormControl,
					Hint:            "multipliers",
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
	return &aggregate.Flow{
		ID:           value_object.NewUUID(),
		Name:         gofakeit.Name(),
		IsDraft:      true,
		CreateUserID: value_object.NewUUID(),
		FlowFunctionIDMapFlowFunction: map[string]*aggregate.FlowFunction{
			config.FlowFunctionStartID: {
				FunctionID:                value_object.NillUUID,
				Note:                      "start node",
				UpstreamFlowFunctionIDs:   []string{},
				DownstreamFlowFunctionIDs: []string{aggFuncAddFlowFunctionID},
				ParamIpts:                 [][]aggregate.IptComponentConfig{},
			},
			aggFuncAddFlowFunctionID: {
				FunctionID:                aggFuncAdd.ID,
				Function:                  aggFuncAdd,
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
				Function:                  aggFuncMultiply,
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
				},
			},
		},
	}
}
