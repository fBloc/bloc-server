package log_collect_backend

import "time"

type LogBackEnd interface {
	Write(key string, tagMap map[string]string, data string, logTime time.Time) error
	Fetch(key string, tagFilterMap map[string]string, start, end time.Time) ([]interface{}, error)
	FetchAll(key string, tagFilterMap map[string]string) ([]interface{}, error)
	ForceFlush() error
}
