package http_server

import (
	"fmt"

	"github.com/fBloc/bloc-server/aggregate"
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
		GroupName:    "math operation",
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
