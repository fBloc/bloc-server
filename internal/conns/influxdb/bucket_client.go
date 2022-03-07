package influxdb

import (
	"context"
	"fmt"
	"sort"
	"strings"
	"time"

	influxdb2 "github.com/influxdata/influxdb-client-go/v2"
	"github.com/influxdata/influxdb-client-go/v2/api"
	"github.com/influxdata/influxdb-client-go/v2/domain"
	"github.com/pkg/errors"
)

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
	writeApi := c.client.WriteAPI(c.organization.Name, bucketName)

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
	bucket, measurement string, tagFilterMap map[string]string, start, end time.Time,
) string {
	filters := []string{}
	if measurement != "" {
		filters = append(filters,
			fmt.Sprintf(`r._measurement == "%s"`, measurement))
	}
	for tagK, tagV := range tagFilterMap {
		filters = append(filters,
			fmt.Sprintf(`r.%s == "%s"`, tagK, tagV))
	}

	fromStr := fmt.Sprintf(`from(bucket:"%s")`, bucket)
	totalSQL := []string{fromStr}

	var rangeStr string
	if end.IsZero() {
		rangeStr = fmt.Sprintf(
			`range(start: %s)`,
			start.Format(time.RFC3339))
	} else {
		rangeStr = fmt.Sprintf(
			`range(start: %s, stop: %s)`,
			start.Format(time.RFC3339), end.Format(time.RFC3339))
	}
	totalSQL = append(totalSQL, rangeStr)

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
	measurement string, tagFilterMap map[string]string, start, end time.Time,
) ([]map[string]interface{}, error) {
	filterStr := buildFilterString(
		bC.bucketName, measurement, tagFilterMap, start, end)

	result, err := bC.client.queryAPI.Query(context.Background(), filterStr)
	if err != nil {
		return nil, err
	}
	if result.Err() != nil {
		return nil, errors.Wrap(result.Err(), "query parsing error")
	}

	happenTimes := make([]time.Time, 0, 100)
	happenTimeMapRecord := make(map[time.Time]map[string]interface{}, 100)
	for result.Next() {
		hT := result.Record().Time()
		if !start.IsZero() && hT.Before(start) {
			continue
		}
		tmp := map[string]interface{}{
			"time": hT,
			"data": result.Record().Value(),
		}
		happenTimes = append(happenTimes, hT)
		for i, j := range result.Record().Values() {
			if strings.HasPrefix(i, "_") {
				continue
			}
			if i == "result" || i == "table" {
				continue
			}
			tmp[i] = j
		}
		happenTimeMapRecord[hT] = tmp
	}
	sort.SliceStable(
		happenTimes,
		func(i, j int) bool {
			return happenTimes[i].Before(happenTimes[j])
		})

	ret := make([]map[string]interface{}, len(happenTimes))
	for i, j := range happenTimes {
		ret[i] = happenTimeMapRecord[j]
	}
	return ret, nil
}

func (bC *BucketClient) QueryAll(
	measurement string, tagFilterMap map[string]string,
) ([]map[string]interface{}, error) {
	return bC.Query(measurement, tagFilterMap, time.Time{}, time.Now())
}

func (bC *BucketClient) Flush() {
	bC.writeApi.Flush()
}
