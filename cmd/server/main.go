package main

import (
	"fmt"
	"log"
	"net/url"
	"strconv"

	"github.com/fBloc/bloc-server"
	flags "github.com/jessevdk/go-flags"
)

type Options struct {
	AppName         string `long:"app_name" description:"the name of this app" required:"true"`
	RabbitMQConnect string `long:"rabbitMQ_connection_str" description:"connection rabbitMQ string in format:'$user:$password@$host:$port/$vHost'" required:"true"`
	MinioConnect    string `long:"minio_connection_str" description:"connection minio string in format:'$user:$password@$host:$port'" required:"true"`
	MongoConnect    string `long:"mongo_connection_str" description:"connection mongo string in format:'$user:$password@$host:$port'" required:"true"`
	InfluxdbConnect string `long:"influxdb_connection_str" description:"connection influxdb string in format:'$user:$password@$host:$port?token=$token&organization=$organization'" required:"true"`
}

func ParseBasicConnection(connectionStr string) (user, password, host string, port int, query url.Values) {
	connectionStr = "whatever4rightparse://" + connectionStr
	urlIns, err := url.Parse(connectionStr)
	if err != nil {
		panic(fmt.Sprintf(
			"connection string: %s not valid. error: %s",
			connectionStr, err.Error()))
	}
	user = urlIns.User.Username()
	password, _ = urlIns.User.Password()
	host = urlIns.Hostname()

	portStr := urlIns.Port()
	if portStr == "" {
		portStr = "80"
	}
	port, err = strconv.Atoi(portStr)
	if err != nil {
		panic(
			fmt.Sprintf("connection string: %s port not valid", connectionStr))
	}

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

	rabbitUser, rabbitPasswd, rabbitHost, rabbitPort, _ := ParseBasicConnection(opts.RabbitMQConnect)
	minioUser, minioPasswd, minioHost, minioPort, _ := ParseBasicConnection(opts.MinioConnect)
	mongoUser, mongoPasswd, mongoHost, mongoPort, _ := ParseBasicConnection(opts.MongoConnect)
	influxdbUser, influxdbPasswd, influxdbHost, influxdbPort, influxQuery := ParseBasicConnection(opts.InfluxdbConnect)

	blocApp.GetConfigBuilder().
		SetRabbitConfig(
			rabbitUser, rabbitPasswd, rabbitHost, rabbitPort, "").
		SetMongoConfig(
			[]string{mongoHost}, mongoPort, opts.AppName, mongoUser, mongoPasswd).
		SetMinioConfig(
			opts.AppName, []string{fmt.Sprintf("%s:%d", minioHost, minioPort)}, minioUser, minioPasswd).
		SetHttpServer(
			"0.0.0.0", 8000).
		SetInfluxDBConfig(
			influxdbUser, influxdbPasswd, fmt.Sprintf("%s:%d", influxdbHost, influxdbPort),
			influxQuery.Get("organization"), influxQuery.Get("token")).
		BuildUp()

	blocApp.Run()
}
