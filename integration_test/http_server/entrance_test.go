package http_server

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"os"
	"testing"
	"time"

	"github.com/fBloc/bloc-server"
	"github.com/fBloc/bloc-server/config"
	"github.com/fBloc/bloc-server/interfaces/web"
	"github.com/fBloc/bloc-server/interfaces/web/user"
	"github.com/fBloc/bloc-server/internal/conns/influxdb"
	"github.com/fBloc/bloc-server/internal/conns/minio"
	"github.com/fBloc/bloc-server/internal/conns/mongodb"
	"github.com/fBloc/bloc-server/internal/conns/rabbit"
	"github.com/fBloc/bloc-server/internal/http_util"
	"github.com/ory/dockertest/v3"
	"github.com/ory/dockertest/v3/docker"
)

func TestMain(m *testing.M) {
	pool, err := dockertest.NewPool("")
	if err != nil {
		log.Fatalf("Could not connect to docker: %s", err)
	}

	// influxdb
	influxdbOptions := &dockertest.RunOptions{
		Repository: "influxdb",
		Tag:        "2.1.1"}
	influxdbResource, err := pool.RunWithOptions(influxdbOptions)
	if err != nil {
		log.Fatalf("Could not start influxdbResource: %s", err)
	}
	influxDBConf.Address = "localhost:" + influxdbResource.GetPort("8086/tcp")
	if err := pool.Retry(func() error {
		_, err = influxdb.Connect(influxDBConf)
		if err != nil {
			return err
		}
		log.Println("influxdb ready")
		return nil
	}); err != nil {
		log.Fatalf("Could not connect to docker influxdb: %s", err)
	}

	// minio
	minioOptions := &dockertest.RunOptions{
		Repository: "minio/minio",
		Tag:        "RELEASE.2021-11-24T23-19-33Z",
		Cmd:        []string{"server", "/data"},
		Env: []string{
			"MINIO_ROOT_USER=" + minioConf.AccessKey,
			"MINIO_ROOT_PASSWORD=" + minioConf.AccessPassword,
		},
	}
	minioResource, err := pool.RunWithOptions(minioOptions)
	if err != nil {
		log.Fatalf("Could not start minioResource: %s", err)
	}
	minioConf.Addresses = []string{
		fmt.Sprintf("localhost:%s", minioResource.GetPort("9000/tcp"))}
	if err := pool.Retry(func() error {
		_, err = minio.Connect(minioConf)
		if err != nil {
			return err
		}
		log.Println("minio ready")
		return nil
	}); err != nil {
		log.Fatalf("Could not connect to minio docker: %s", err)
	}

	// mongodb
	mongoResource, err := pool.RunWithOptions(&dockertest.RunOptions{
		Repository: "mongo",
		Tag:        "5.0.5",
		Env: []string{
			"MONGO_INITDB_ROOT_USERNAME=" + mongoConf.User,
			"MONGO_INITDB_ROOT_PASSWORD=" + mongoConf.Password},
	}, func(config *docker.HostConfig) {
		// set AutoRemove to true so that stopped container goes away by itself
		config.AutoRemove = true
	})
	if err != nil {
		log.Fatalf("Could not start mongoResource: %s", err)
	}

	mongoConf.Addresses = []string{
		fmt.Sprintf("localhost:%s", mongoResource.GetPort("27017/tcp"))}

	err = pool.Retry(func() error {
		var err error
		_, err = mongodb.InitClient(mongoConf)
		if err != nil {
			return err
		}
		log.Println("mongo ready")
		return nil
	})
	if err != nil {
		log.Fatalf("Could not connect to mongo docker: %s", err)
	}

	// rabbitMQ
	rabbitResource, err := pool.RunWithOptions(&dockertest.RunOptions{
		Repository: "rabbitmq",
		Tag:        "3.9.11-alpine",
		Env: []string{
			"RABBITMQ_DEFAULT_USER=" + rabbitConf.User,
			"RABBITMQ_DEFAULT_PASS=" + rabbitConf.Password,
		},
	}, func(config *docker.HostConfig) {
		config.AutoRemove = true
		config.RestartPolicy = docker.RestartPolicy{Name: "no"}
	})
	if err != nil {
		log.Fatalf("Could not start rabbitResource: %s", err)
	}

	rabbitConf.Host = []string{
		fmt.Sprintf("localhost:%s", rabbitResource.GetPort("5672/tcp"))}

	err = pool.Retry(func() error {
		var err error
		_, err = rabbit.InitChannel(rabbitConf)
		if err != nil {
			return err
		}
		log.Println("rabbit ready")
		return nil
	})

	if err != nil {
		log.Fatalf("Could not connect to rabbit docker: %s", err)
	}

	go func() {
		appName := "test"
		blocApp := &bloc.BlocApp{Name: appName}

		blocApp.GetConfigBuilder().
			SetRabbitConfig(
				rabbitConf.User, rabbitConf.Password, rabbitConf.Host, "").
			SetMongoConfig(
				mongoConf.User, mongoConf.Password, mongoConf.Addresses,
				appName, "").
			SetMinioConfig(
				appName, minioConf.Addresses, minioConf.AccessKey, minioConf.AccessPassword).
			SetInfluxDBConfig(
				influxDBConf.UserName, influxDBConf.Password, influxDBConf.Address,
				influxDBConf.Organization, influxDBConf.Token).
			SetHttpServer(
				"localhost", 8080).
			BuildUp()

		blocApp.Run()
	}()

	// wait until server is ready
	checkTicker := time.NewTicker(2 * time.Second)
	for range checkTicker.C {
		var resp web.RespMsg
		_, err := http_util.Get("localhost:8080/api/v1/bloc", http_util.BlankHeader, &resp)
		if err != nil {
			continue
		}
		if resp.Code == http.StatusOK {
			checkTicker.Stop()
			break
		}
	}

	// initial the user token
	loginUser := user.User{
		Name:        config.DefaultUserName,
		RaWPassword: config.DefaultUserPassword}
	loginPostBody, _ := json.Marshal(loginUser)
	loginResp := struct {
		Code int       `json:"status_code"`
		Msg  string    `json:"status_msg"`
		Data user.User `json:"data"`
	}{}
	http_util.Post(
		serverAddress+"/api/v1/login",
		http_util.BlankHeader, loginPostBody, &loginResp)
	if loginResp.Data.Token.IsNil() {
		log.Panicf("login to get token error:" + err.Error())
	}
	loginedToken = loginResp.Data.Token.String()

	// run tests
	code := m.Run()

	// remove container
	if err = pool.Purge(influxdbResource); err != nil {
		log.Fatalf("Could not purge influxdbResource: %s", err)
	}

	if err = pool.Purge(minioResource); err != nil {
		log.Fatalf("Could not purge minioResource: %s", err)
	}

	if err = pool.Purge(mongoResource); err != nil {
		log.Fatalf("Could not purge mongoResource: %s", err)
	}

	if err = pool.Purge(rabbitResource); err != nil {
		log.Fatalf("Could not purge rabbitResource: %s", err)
	}

	os.Exit(code)
}
