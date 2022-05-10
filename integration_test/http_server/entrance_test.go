package http_server

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"os"
	"sync"
	"testing"
	"time"

	"github.com/fBloc/bloc-server"
	"github.com/fBloc/bloc-server/config"
	"github.com/fBloc/bloc-server/interfaces/web"
	"github.com/fBloc/bloc-server/interfaces/web/client"
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

	var waitContainerAllReady sync.WaitGroup

	var influxdbResource *dockertest.Resource
	waitContainerAllReady.Add(1)
	go func(wg *sync.WaitGroup) {
		defer wg.Done()
		// influxdb
		influxdbOptions := &dockertest.RunOptions{
			Repository: "influxdb",
			Tag:        "2.1.1"}
		influxdbResource, err = pool.RunWithOptions(influxdbOptions)
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
	}(&waitContainerAllReady)

	var minioResource *dockertest.Resource
	waitContainerAllReady.Add(1)
	go func(wg *sync.WaitGroup) {
		defer wg.Done()
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
		minioResource, err = pool.RunWithOptions(minioOptions)
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
	}(&waitContainerAllReady)

	var mongoResource *dockertest.Resource
	waitContainerAllReady.Add(1)
	go func(wg *sync.WaitGroup) {
		defer wg.Done()
		// mongodb
		mongoResource, err = pool.RunWithOptions(&dockertest.RunOptions{
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
	}(&waitContainerAllReady)

	var rabbitResource *dockertest.Resource
	waitContainerAllReady.Add(1)
	go func(wg *sync.WaitGroup) {
		defer wg.Done()
		// rabbitMQ
		rabbitResource, err = pool.RunWithOptions(&dockertest.RunOptions{
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
	}(&waitContainerAllReady)

	waitContainerAllReady.Wait()

	// start bloc server
	go func() {
		appName := "test"
		blocApp := &bloc.BlocApp{Name: appName}

		blocApp.GetConfigBuilder().
			SetRabbitConfig(
				rabbitConf.User, rabbitConf.Password, rabbitConf.Host, "").
			SetMongoConfig(
				mongoConf.User, mongoConf.Password, mongoConf.Addresses,
				appName, "", "").
			SetMinioConfig(
				appName, minioConf.Addresses, minioConf.AccessKey, minioConf.AccessPassword).
			SetInfluxDBConfig(
				influxDBConf.UserName, influxDBConf.Password, influxDBConf.Address,
				influxDBConf.Organization, influxDBConf.Token).
			SetHttpServer(
				serverHost, serverPort).
			BuildUp()

		blocApp.Run()
	}()

	// wait until server is ready
	checkTicker := time.NewTicker(2 * time.Second)
	for range checkTicker.C {
		var resp web.RespMsg
		_, err := http_util.Get(
			http_util.BlankHeader,
			serverAddress+"/api/v1/bloc",
			http_util.BlankGetParam, &resp)
		if err != nil {
			continue
		}
		if resp.Code == http.StatusOK {
			checkTicker.Stop()
			break
		}
	}

	// initial superuser about
	superUser := user.User{
		Name:        config.DefaultUserName,
		RaWPassword: config.DefaultUserPassword}
	superUserPostBody, _ := json.Marshal(superUser)
	superUserLoginResp := struct {
		web.RespMsg
		Data user.User `json:"data"`
	}{}
	http_util.Post(
		http_util.BlankHeader,
		serverAddress+"/api/v1/login",
		http_util.BlankGetParam, superUserPostBody, &superUserLoginResp)
	if superUserLoginResp.Data.Token.IsNil() {
		log.Fatalf("login to get token error:" + err.Error())
	}
	superUserToken = superUserLoginResp.Data.Token.String()

	// initial a have no permission of anything user
	addNobody := user.User{
		Name:        nobodyName,
		RaWPassword: nobodyRawPasswd}
	addNobodyBody, _ := json.Marshal(addNobody)
	var addNobodyResp web.RespMsg
	http_util.Post(
		superuserHeader(),
		serverAddress+"/api/v1/user",
		http_util.BlankGetParam, addNobodyBody, &addNobodyResp)

	nobody := user.User{
		Name:        nobodyName,
		RaWPassword: nobodyRawPasswd}
	nobodyPostBody, _ := json.Marshal(nobody)
	nobodyLoginResp := struct {
		web.RespMsg
		Data user.User `json:"data"`
	}{}
	http_util.Post(
		http_util.BlankHeader,
		serverAddress+"/api/v1/login",
		http_util.BlankGetParam, nobodyPostBody, &nobodyLoginResp)
	if nobodyLoginResp.Data.Token.IsNil() {
		log.Fatalf("login to get nobody's token error:" + err.Error())
	}
	nobodyToken = nobodyLoginResp.Data.Token.String()
	nobodyID = nobodyLoginResp.Data.ID

	// register a function for function test
	registerFunction := client.RegisterFuncReq{
		Who: fakeAggFunction.ProviderName,
		GroupNameMapFunctions: map[string][]*client.HttpFunction{
			fakeAggFunction.GroupName: []*client.HttpFunction{
				{
					Name:               fakeAggFunction.Name,
					GroupName:          fakeAggFunction.GroupName,
					Description:        fakeAggFunction.Description,
					Ipts:               fakeAggFunction.Ipts,
					Opts:               fakeAggFunction.Opts,
					ProgressMilestones: fakeAggFunction.ProgressMilestones,
				},
			},
		},
	}
	registerFunctionBody, _ := json.Marshal(registerFunction)
	registerFunctionResp := struct {
		web.RespMsg
		Data client.RegisterFuncReq `json:"data"`
	}{}
	_, err = http_util.Post(
		http_util.BlankHeader,
		serverAddress+"/api/v1/client/register_functions",
		http_util.BlankGetParam, registerFunctionBody, &registerFunctionResp)
	if err != nil {
		log.Fatalf("register function error: %v", err)
	}
	if registerFunctionResp.Code != http.StatusOK {
		log.Fatalf("register function failed: %v", registerFunctionResp)
	}
	fakeAggFunction.ID = registerFunctionResp.Data.GroupNameMapFunctions[fakeAggFunction.GroupName][0].ID

	// register the two function 4 flow test
	register4FlowTestFunctions := client.RegisterFuncReq{
		Who: fakeAggFunction.ProviderName,
		GroupNameMapFunctions: map[string][]*client.HttpFunction{
			fakeAggFunction.GroupName: []*client.HttpFunction{
				{
					Name:               aggFuncAdd.Name,
					GroupName:          aggFuncAdd.GroupName,
					Description:        aggFuncAdd.Description,
					Ipts:               aggFuncAdd.Ipts,
					Opts:               aggFuncAdd.Opts,
					ProgressMilestones: aggFuncAdd.ProgressMilestones,
				},
				{
					Name:               aggFuncMultiply.Name,
					GroupName:          aggFuncMultiply.GroupName,
					Description:        aggFuncMultiply.Description,
					Ipts:               aggFuncMultiply.Ipts,
					Opts:               aggFuncMultiply.Opts,
					ProgressMilestones: aggFuncMultiply.ProgressMilestones,
				},
			},
		},
	}
	register4FlowTestFunctionsBody, _ := json.Marshal(register4FlowTestFunctions)
	register4FlowTestFunctionsResp := struct {
		web.RespMsg
		Data client.RegisterFuncReq `json:"data"`
	}{}
	_, err = http_util.Post(
		http_util.BlankHeader,
		serverAddress+"/api/v1/client/register_functions",
		http_util.BlankGetParam, register4FlowTestFunctionsBody, &register4FlowTestFunctionsResp)
	if err != nil {
		log.Fatalf("register function error: %v", err)
	}
	if registerFunctionResp.Code != http.StatusOK {
		log.Fatalf("register functions failed: %v", registerFunctionResp)
	}
	for _, function := range register4FlowTestFunctionsResp.Data.GroupNameMapFunctions[fakeAggFunction.GroupName] {
		if function.Name == aggFuncAdd.Name {
			aggFuncAdd.ID = function.ID
		} else if function.Name == aggFuncMultiply.Name {
			aggFuncMultiply.ID = function.ID
		}
	}

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
