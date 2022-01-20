package influxdb

import (
	"sync"
	"time"

	"github.com/fBloc/bloc-server/infrastructure/log_collect_backend"
	"github.com/fBloc/bloc-server/internal/conns/influxdb"
	influxdb_conn "github.com/fBloc/bloc-server/internal/conns/influxdb"
)

/*
influxDB save log about:
1. org: bloc
2. bucket: log
3. _measurement: http-server | schedule | func-run-record (value_object.LogType)
*/

func init() {
	var _ log_collect_backend.LogBackEnd = &InfluxDBLogBackendRepository{}
}

type InfluxDBLogBackendRepository struct {
	bucketClient *influxdb.BucketClient
	sync.Mutex
}

func New(
	influxDBClient *influxdb_conn.Connection,
	expireDuration time.Duration,
) (*InfluxDBLogBackendRepository, error) {
	writeApi, err := influxDBClient.NewBucketClient("log", expireDuration)
	if err != nil {
		return nil, err
	}
	return &InfluxDBLogBackendRepository{
		bucketClient: writeApi,
	}, nil
}

func (inf *InfluxDBLogBackendRepository) Write(
	logName string, tagMap map[string]string, data string,
	eventTime time.Time,
) error {
	inf.bucketClient.Write(
		logName, tagMap,
		map[string]interface{}{"data": data},
		eventTime)
	return nil
}

func (inf *InfluxDBLogBackendRepository) Fetch(
	logName string, tagFilterMap map[string]string, start, end time.Time,
) ([]interface{}, error) {
	tmp, err := inf.bucketClient.Query(logName, tagFilterMap, start, end)
	if err != nil {
		return nil, err
	}
	ret := make([]interface{}, 0, len(tmp))
	for _, i := range tmp {
		ret = append(ret, i)
	}
	return ret, nil
}

func (inf *InfluxDBLogBackendRepository) FetchAll(
	logName string, tagFilterMap map[string]string,
) ([]interface{}, error) {
	tmp, err := inf.bucketClient.QueryAll(logName, tagFilterMap)
	if err != nil {
		return nil, err
	}
	ret := make([]interface{}, 0, len(tmp))
	for _, i := range tmp {
		ret = append(ret, i)
	}
	return ret, nil
}

func (inf *InfluxDBLogBackendRepository) ForceFlush() error {
	inf.bucketClient.Flush()
	return nil
}
