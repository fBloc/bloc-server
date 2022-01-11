package influxdb

import (
	"context"
	"fmt"
	"strings"
	"sync"
	"time"

	influxdb2 "github.com/influxdata/influxdb-client-go/v2"
	"github.com/influxdata/influxdb-client-go/v2/api"
)

var (
	client influxdb2.Client
	once   sync.Once
)

type InfluxDBConfig struct {
	Address      string
	Token        string
	Organization string
}

type Client struct {
	organization string
	client       influxdb2.Client
	queryApi     api.QueryAPI
	sync.Mutex
}

func NewClient(conf InfluxDBConfig) *Client {
	initClientFunc := func() {
		client = influxdb2.NewClientWithOptions(
			conf.Address, conf.Token,
			influxdb2.DefaultOptions().
				SetUseGZip(true).
				SetPrecision(time.Millisecond))
	}
	once.Do(initClientFunc)

	ctx, _ := context.WithTimeout(context.TODO(), 2*time.Second)
	serverAlive, err := client.Ping(ctx) // Ping just check server alive. not check auth
	if !serverAlive || err != nil {
		panic("ping influxDB server failed.err:" + err.Error())
	}

	return &Client{
		organization: conf.Organization,
		client:       client,
		queryApi:     client.QueryAPI(conf.Organization),
	}
}

type BucketClient struct {
	bucketName string
	writeApi   api.WriteAPI
	client     *Client
	sync.Mutex
}

func (c *Client) NewBucketClient(bucket string) *BucketClient {
	// writeApi需要每个org-bucket对用一个
	writeApi := c.client.WriteAPI(c.organization, bucket)
	// TODO check valid

	return &BucketClient{
		bucketName: bucket,
		writeApi:   writeApi,
		client:     c,
	}
}

func (bC *BucketClient) Write(
	measurement string,
	tags map[string]string,
	fields map[string]interface{},
	occurTime time.Time,
) {
	p := influxdb2.NewPoint(measurement, tags, fields, occurTime)
	bC.writeApi.WritePoint(p)
	// TODO:
	// WriteAPI can set SetWriteFailedCallback
	// as this write is async, need to handle fai!
}

func (bC *BucketClient) Query(
	measurement string, tagFilterMap map[string]string, start, end time.Time,
) {
	filters := []string{fmt.Sprintf(`r._measurement == "%s"`, measurement)}
	for tag, tagVal := range tagFilterMap {
		filters = append(filters, fmt.Sprintf(`r.%s == "%s"`, tag, tagVal))
	}
	filterStr := fmt.Sprintf(
		`filter(fn: (r) => %s)`,
		strings.Join(filters, " and "))

	querys := []string{
		fmt.Sprintf(`from(bucket:"%s")`, bC.bucketName),
		fmt.Sprintf(
			`range(start: %s, end: %s)`,
			start.Format("2006-01-02T15:04:05Z"),
			end.Format("2006-01-02T15:04:05Z")),
		filterStr}

	queryStr := strings.Join(querys, " |> ")

	// `from(bucket:"test-initial") |> range(start: -30m) |> filter(fn: (r) => r._measurement == "intime" and r.type == "album" and r._field == "105505801")`,
	result, err := bC.client.queryApi.Query(context.Background(), queryStr)
	if err == nil {
		// Iterate over query response
		for result.Next() {
			// Notice when group key has changed
			if result.TableChanged() {
				fmt.Printf("table: %s\n", result.TableMetadata().String())
			}
			// Access data
			fmt.Printf("value: %+v\n", result.Record().Time())
		}
		// Check for an error
		if result.Err() != nil {
			fmt.Printf("query parsing error: %s\n", result.Err().Error())
		}
	} else {
		panic(err)
	}
}
