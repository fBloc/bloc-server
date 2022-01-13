package influxdb

import (
	"context"
	"encoding/json"
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
	client            influxdb2.Client
	writeApi          api.WriteAPI
	initialClientOnce sync.Once
)

const setupPath = "/api/v2/setup"

type InfluxDBConfig struct {
	Address      string
	UserName     string
	Password     string
	Token        string
	Organization string
}

type canSetupResp struct {
	Allowed bool `json:"allowed"`
}

type setupPost struct {
	UserName string `json:"username"`
	Password string `json:"password"`
	Token    string `json:"token"`
	Org      string `json:"org"`
	Bucket   string `json:"bucket"`
}

func (conf *InfluxDBConfig) Valid() bool {
	return conf.Address != "" && conf.UserName != "" &&
		conf.Password != "" && conf.Token != "" && conf.Organization != ""
}

func (conf *InfluxDBConfig) canSetup() (bool, error) {
	u, err := url.Parse(conf.Address)
	if err != nil {
		return false, err
	}
	u.Path = path.Join(u.Path, setupPath)

	var resp canSetupResp
	_, err = http_util.Get(u.String(), http_util.BlankHeader, &resp)
	if err != nil {
		return false, err
	}
	return resp.Allowed, nil
}

func (conf *InfluxDBConfig) setup() (bool, error) {
	if conf.UserName == "" || conf.Password == "" {
		panic("setup influxDB must need UserName & Password")
	}
	req := setupPost{
		UserName: conf.UserName,
		Password: conf.Password,
		Org:      conf.Organization,
		Token:    conf.Token,
		Bucket:   "bloc"}
	reqBody, err := json.Marshal(req)
	if err != nil {
		return false, err
	}

	var resp interface{}
	u, err := url.Parse(conf.Address)
	if err != nil {
		return false, err
	}

	u.Path = path.Join(u.Path, setupPath)
	statusCode, err := http_util.Post(
		u.String(), http_util.BlankHeader, reqBody, &resp)
	if err != nil {
		return false, err
	}
	return statusCode == http.StatusCreated, nil
}

type Connection struct {
	organization *domain.Organization
	client       influxdb2.Client
	queryAPI     api.QueryAPI
	bucketAPI    api.BucketsAPI
}

func NewConnection(conf *InfluxDBConfig) *Connection {
	if !conf.Valid() {
		panic("influxDB connection str not valid!")
	}

	if !strings.HasPrefix(conf.Address, "http") {
		conf.Address = "http://" + conf.Address
	}

	// setup
	needSetup, err := conf.canSetup()
	if err != nil {
		panic(err)
	}
	if needSetup {
		setUpSuc, err := conf.setup()
		if err != nil {
			panic(err)
		}
		if !setUpSuc {
			panic("influxdb setup failed")
		}
	}

	initConnectionFunc := func() {
		client = influxdb2.NewClientWithOptions(
			conf.Address,
			conf.Token,
			influxdb2.DefaultOptions().SetUseGZip(true))
	}
	initialClientOnce.Do(initConnectionFunc)

	ctx, _ := context.WithTimeout(
		context.Background(), 2*time.Second,
	)
	serverRunning, err := client.Ping(ctx)
	if !serverRunning || err != nil {
		panic("ping influxDB server failed. error")
	}

	// make sure organization exist
	// otherwise create it
	var orgIns *domain.Organization
	orgApi := client.OrganizationsAPI()
	orgIns, err = orgApi.FindOrganizationByName(
		context.Background(), conf.Organization)
	if err != nil {
		panic("InfluxDB FindOrganizationByName error:" + err.Error())
	}
	if orgIns == nil {
		orgIns, err = orgApi.CreateOrganizationWithName(
			context.Background(), conf.Organization)
		if err != nil {
			panic("InfluxDB CreateOrganizationWithName error:" + err.Error())
		}
	}

	return &Connection{
		organization: orgIns,
		client:       client,
		queryAPI:     client.QueryAPI(conf.Organization),
		bucketAPI:    client.BucketsAPI(),
	}
}

type BucketClient struct {
	client     *Connection
	bucketName string
	writeApi   api.WriteAPI
}

func (c *Connection) NewBucketClient(
	bucketName string,
	keepDuration time.Duration,
) (*BucketClient, error) {
	// make sure bucket exist. otherwise create it
	bucketIns, _ := c.bucketAPI.FindBucketByName(
		context.TODO(), bucketName)
	if bucketIns == nil {
		_, err := c.bucketAPI.CreateBucketWithName(
			context.Background(),
			c.organization,
			bucketName,
			domain.RetentionRule{
				EverySeconds: int64(keepDuration.Seconds())},
		)
		if err != nil {
			return nil, err
		}
	}

	// WriteAPI returns the asynchronous, non-blocking, Write client.
	// Ensures using a single WriteAPI instance for each org/bucket pair.
	writeApi = client.WriteAPI(c.organization.Name, bucketName)

	return &BucketClient{
		client:     c,
		bucketName: bucketName,
		writeApi:   writeApi,
	}, nil
}

func (bC *BucketClient) Write(
	measurement string,
	tags map[string]string,
	fields map[string]interface{},
	occurTime time.Time,
) {
	p := influxdb2.NewPoint(measurement, tags, fields, occurTime)
	bC.writeApi.WritePoint(p)
}

func buildFilterString(
	key string, tagFilterMap map[string]string, start, end time.Time,
) string {
	var filters []string
	for tagK, tagV := range tagFilterMap {
		filters = append(
			filters,
			fmt.Sprintf(`r.%s == "%s"`, tagK, tagV))
	}

	fromStr := fmt.Sprintf(`from(bucket:"%s")`, "http-server")
	totalSQL := []string{fromStr}

	if !start.IsZero() {
		var rangeStr string
		if !end.IsZero() {
			rangeStr = fmt.Sprintf(
				`range(start: %s, stop: %s)`,
				start.Format(`2006-01-02T15:04:05Z`), end.Format(`2006-01-02T15:04:05Z`))
		} else {
			rangeStr = fmt.Sprintf(`range(start: %s)`, start.Format(`2006-01-02T15:04:05Z`))
		}
		totalSQL = append(totalSQL, rangeStr)
	}

	if len(filters) > 0 {
		filterStr := fmt.Sprintf(
			`filter(fn: (r) => %s)`,
			strings.Join(filters, " and "))
		totalSQL = append(totalSQL, filterStr)
	}

	totalSQLStr := strings.Join(totalSQL, " |> ")
	return totalSQLStr
}

func (bC *BucketClient) Query(
	key string, tagFilterMap map[string]string, start, end time.Time,
) {
	filterStr := buildFilterString(key, tagFilterMap, start, end)

	result, err := bC.client.queryAPI.Query(context.Background(), filterStr)
	if err == nil {
		for result.Next() {
			fmt.Printf("===> :%+v\t%+v\n\n", result.Record().Time().Add(8*time.Hour), result.Record().Measurement())
		}
		if result.Err() != nil {
			fmt.Printf("query parsing error: %s\n", result.Err().Error())
		}
	} else {
		panic(err)
	}
}

func (bC *BucketClient) QueryAll(
	key string, tagFilterMap map[string]string,
) {
	filterStr := buildFilterString(key, tagFilterMap, time.Time{}, time.Time{})

	result, err := bC.client.queryAPI.Query(context.Background(), filterStr)
	if err == nil {
		for result.Next() {
			fmt.Printf("===> :%+v - %s\n", result.Record().Time(), result.Record().Value())
		}
		if result.Err() != nil {
			fmt.Printf("query parsing error: %s\n", result.Err().Error())
		}
	} else {
		panic(err)
	}
}

func (bC *BucketClient) Flush() {
	bC.writeApi.Flush()
}
