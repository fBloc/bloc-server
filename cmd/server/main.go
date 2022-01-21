package main

import (
	"fmt"
	"log"
	"net/url"
	"strings"

	"github.com/fBloc/bloc-server"
	flags "github.com/jessevdk/go-flags"
)

type Options struct {
	AppName         string `long:"app_name" description:"the name of this app" required:"true"`
	RabbitMQConnect string `long:"rabbitMQ_connection_str" description:"connection rabbitMQ string in format:'$user:$password@$host:$port/$vHost'" required:"true"`
	MinioConnect    string `long:"minio_connection_str" description:"connection minio string in format:'$user:$password@$host:$port'" required:"true"`
	MongoConnect    string `long:"mongo_connection_str" description:"connection mongo string in format:'[username:password@]host1[:port1][,...hostN[:portN]][/[defaultauthdb][?replicaSet=replicaSetName]]'" required:"true"`
	InfluxdbConnect string `long:"influxdb_connection_str" description:"connection influxdb string in format:'$user:$password@$host:$port?token=$token&organization=$organization'" required:"true"`
}

func ParseBasicConnection(connectionStr string) (user, password, host string, query url.Values) {
	connectionStr = "whatever4rightparse://" + connectionStr
	urlIns, err := url.Parse(connectionStr)
	if err != nil {
		panic(fmt.Sprintf(
			"connection string: %s not valid. error: %s",
			connectionStr, err.Error()))
	}
	user = urlIns.User.Username()
	password, _ = urlIns.User.Password()
	host = urlIns.Host

	query = urlIns.Query()
	return
}

func main() {
	var opts Options
	parser := flags.NewParser(&opts, flags.Default)
	_, err := parser.Parse()
	if err != nil {
		log.Fatal(err)
	}

	blocApp := &bloc.BlocApp{Name: opts.AppName}

	rabbitUser, rabbitPasswd, rabbitHost, _ := ParseBasicConnection(opts.RabbitMQConnect)
	minioUser, minioPasswd, minioHost, _ := ParseBasicConnection(opts.MinioConnect)
	mongoUser, mongoPasswd, mongoAddress, mongoQuery := ParseBasicConnection(opts.MongoConnect)
	influxdbUser, influxdbPasswd, influxdbHost, influxQuery := ParseBasicConnection(opts.InfluxdbConnect)

	blocApp.GetConfigBuilder().
		SetRabbitConfig(
			rabbitUser, rabbitPasswd, rabbitHost, "").
		SetMongoConfig(
			mongoUser, mongoPasswd, strings.Split(mongoAddress, ","),
			opts.AppName, mongoQuery.Get("replicaSet")).
		SetMinioConfig(
			opts.AppName, strings.Split(minioHost, ","), minioUser, minioPasswd).
		SetInfluxDBConfig(
			influxdbUser, influxdbPasswd, influxdbHost,
			influxQuery.Get("organization"), influxQuery.Get("token")).
		SetHttpServer(
			"0.0.0.0", 8000).
		BuildUp()

	blocApp.Run()
}
