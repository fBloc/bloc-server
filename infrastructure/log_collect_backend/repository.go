package log_collect_backend

import "time"

type LogBackEnd interface {
	PersistData(key string, data []byte) error
	PullDataBetween(prefixKey string, start, end time.Time) ([]string, error)
}
