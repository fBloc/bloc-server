package main

import (
	"fmt"
	"log"
	"strconv"
	"strings"

	"github.com/fBloc/bloc-server"
	flags "github.com/jessevdk/go-flags"
)

type Options struct {
	AppName         string `long:"app_name" description:"the name of this app" required:"true"`
	RabbitMQConnect string `long:"rabbitMQ_connection_str" description:"connection rabbitMQ string in format:'$user:$password@$host:$port/$vHost'" required:"true"`
	MinioConnect    string `long:"minio_connection_str" description:"connection minio string in format:'$user:$password@$host:$port'" required:"true"`
	MongoConnect    string `long:"mongo_connection_str" description:"connection mongo string in format:'$user:$password@$host:$port'" required:"true"`
}

func ParseBasicConnection(connectionStr string) (user, password, host string, port int) {
	tmp := strings.Split(connectionStr, "/")
	leftStr := tmp[0]
	tmp = strings.Split(leftStr, "@")
	auth, socketAddress := tmp[0], tmp[1]

	tmp = strings.Split(auth, ":")
	user, password = tmp[0], tmp[1]

	tmp = strings.Split(socketAddress, ":")
	host, portStr := tmp[0], tmp[1]

	port, err := strconv.Atoi(portStr)
	if err != nil {
		panic(fmt.Sprintf("input port(%s) is not a valid int", portStr))
	}
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

	rabbitUser, rabbitPasswd, rabbitHost, rabbitPort := ParseBasicConnection(opts.RabbitMQConnect)
	minioUser, minioPasswd, minioHost, minioPort := ParseBasicConnection(opts.MinioConnect)
	mongoUser, mongoPasswd, mongoHost, mongoPort := ParseBasicConnection(opts.MongoConnect)

	blocApp.GetConfigBuilder().SetRabbitConfig(
		rabbitUser, rabbitPasswd, rabbitHost, rabbitPort, "",
	).SetMongoConfig(
		[]string{mongoHost}, mongoPort, opts.AppName, mongoUser, mongoPasswd,
	).SetMinioConfig(
		opts.AppName, []string{fmt.Sprintf("%s:%d", minioHost, minioPort)}, minioUser, minioPasswd,
	).SetHttpServer(
		"0.0.0.0", 8000,
	).BuildUp()

	blocApp.Run()
}
