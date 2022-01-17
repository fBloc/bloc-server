package influxdb

import (
	"sync"
	"time"

	"github.com/fBloc/bloc-server/infrastructure/log_collect_backend"
	"github.com/fBloc/bloc-server/internal/conns/influxdb"
	influxdb_conn "github.com/fBloc/bloc-server/internal/conns/influxdb"
)

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
	key string, tagMap map[string]string, data string,
	eventTime time.Time,
) error {
	inf.bucketClient.Write(
		key, tagMap,
		map[string]interface{}{"data": data},
		eventTime)
	return nil
}

func (inf *InfluxDBLogBackendRepository) Fetch(
	key string, tagFilterMap map[string]string, start, end time.Time,
) ([]interface{}, error) {
	inf.bucketClient.Query(key, tagFilterMap, start, end)
	return nil, nil
}

func (inf *InfluxDBLogBackendRepository) FetchAll(
	key string, tagFilterMap map[string]string,
) ([]interface{}, error) {
	inf.bucketClient.QueryAll(key, tagFilterMap)
	return nil, nil
}

func (inf *InfluxDBLogBackendRepository) ForceFlush() error {
	inf.bucketClient.Flush()
	return nil
}
