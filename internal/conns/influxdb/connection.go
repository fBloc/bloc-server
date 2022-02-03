package influxdb

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"net/url"
	"path"
	"strings"
	"sync"
	"time"

	"github.com/fBloc/bloc-server/internal/http_util"
	influxdb2 "github.com/influxdata/influxdb-client-go/v2"
	"github.com/influxdata/influxdb-client-go/v2/api"
	"github.com/influxdata/influxdb-client-go/v2/domain"
)

var (
	confSignatureMapConnection  = make(map[confSignature]*Connection)
	confSignatureMapClientMutex sync.Mutex
)

type Connection struct {
	organization *domain.Organization
	client       influxdb2.Client
	queryAPI     api.QueryAPI
	bucketAPI    api.BucketsAPI
}

func Connect(
	conf *InfluxDBConfig,
) (*Connection, error) {
	if !conf.Valid() {
		return nil, errors.New("influxDB connection str not valid!")
	}

	if !strings.HasPrefix(conf.Address, "http") {
		conf.Address = "http://" + conf.Address
	}

	confSignatureMapClientMutex.Lock()
	defer confSignatureMapClientMutex.Unlock()

	sig := conf.signature()
	con, ok := confSignatureMapConnection[sig]
	if ok && con.client != nil {
		return con, nil
	}

	// setup
	needSetup, err := canSetup(conf)
	if err != nil {
		return nil, err
	}
	if needSetup {
		setUpSuc, err := setup(conf)
		if err != nil {
			return nil, err
		}
		if !setUpSuc {
			return nil, errors.New("influxdb setup failed")
		}
	}

	client := influxdb2.NewClientWithOptions(
		conf.Address,
		conf.Token,
		influxdb2.DefaultOptions().SetUseGZip(true))

	ctx, _ := context.WithTimeout(
		context.Background(), 2*time.Second,
	)
	serverRunning, err := client.Ping(ctx)
	if !serverRunning || err != nil {
		return nil, errors.New(
			"ping influxDB server failed. error:" + err.Error())
	}

	// make sure organization exist. otherwise create it
	var orgIns *domain.Organization
	orgApi := client.OrganizationsAPI()
	orgIns, err = orgApi.FindOrganizationByName(
		context.Background(), conf.Organization)
	if err != nil {
		return nil, errors.New(
			"InfluxDB FindOrganizationByName error:" + err.Error())
	}
	if orgIns == nil {
		orgIns, err = orgApi.CreateOrganizationWithName(
			context.Background(), conf.Organization)
		if err != nil {
			return nil, errors.New(
				"InfluxDB CreateOrganizationWithName error:" + err.Error())
		}
	}

	connection := &Connection{
		organization: orgIns,
		client:       client,
		queryAPI:     client.QueryAPI(conf.Organization),
		bucketAPI:    client.BucketsAPI()}
	confSignatureMapConnection[sig] = connection
	return connection, nil
}

const setupPath = "/api/v2/setup"

// canSetup visit server to check whether initialed setup
func canSetup(conf *InfluxDBConfig) (bool, error) {
	u, err := url.Parse(conf.Address)
	if err != nil {
		return false, err
	}
	u.Path = path.Join(u.Path, setupPath)

	var resp struct {
		Allowed bool `json:"allowed"`
	}
	_, err = http_util.Get(
		http_util.BlankHeader, u.String(),
		http_util.BlankGetParam, &resp)
	if err != nil {
		return false, err
	}
	return resp.Allowed, nil
}

// setup do initial setup
func setup(conf *InfluxDBConfig) (bool, error) {
	if conf.UserName == "" || conf.Password == "" {
		panic("setup influxDB must need UserName & Password")
	}
	req := struct {
		UserName string `json:"username"`
		Password string `json:"password"`
		Token    string `json:"token"`
		Org      string `json:"org"`
		Bucket   string `json:"bucket"`
	}{
		UserName: conf.UserName,
		Password: conf.Password,
		Org:      conf.Organization,
		Token:    conf.Token,
		Bucket:   "bloc",
	}

	reqBody, err := json.Marshal(req)
	if err != nil {
		return false, err
	}

	failResp := struct { // suc & fail has diff resp
		Code    string `json:"code"`
		Message string `json:"message"`
	}{}
	u, err := url.Parse(conf.Address)
	if err != nil {
		return false, err
	}

	u.Path = path.Join(u.Path, setupPath)
	statusCode, err := http_util.Post(
		http_util.BlankHeader,
		u.String(), http_util.BlankGetParam,
		reqBody, &failResp)
	if err != nil {
		return false, err
	}
	if statusCode == http.StatusCreated {
		return true, nil
	}
	return false, fmt.Errorf(
		"setup influxDB failed:%s - %s", failResp.Code, failResp.Message)
}
