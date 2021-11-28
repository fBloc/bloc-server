package log_collect_backend

import "time"

type LogBackEnd interface {
	PersistData(key string, data []byte) error
	ListKeysBetween(prefixKey string, start, end time.Time) ([]string, error)
	PullDataBetween(prefixKey string, start, end time.Time) ([]interface{}, error)
	PullDataByKey(key string) ([]interface{}, error)
}
