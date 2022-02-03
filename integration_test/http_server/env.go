package http_server

import (
	"fmt"

	"github.com/fBloc/bloc-server/internal/conns/influxdb"
	"github.com/fBloc/bloc-server/internal/conns/minio"
	"github.com/fBloc/bloc-server/internal/conns/mongodb"
	"github.com/fBloc/bloc-server/internal/conns/rabbit"
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

func loginedHeader() map[string]string {
	return map[string]string{"token": loginedToken}
}
