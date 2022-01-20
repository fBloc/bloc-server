package log_collect_backend

import "time"

type LogBackEnd interface {
	Write(logName string, tagMap map[string]string, data string, logTime time.Time) error
	Fetch(logName string, tagFilterMap map[string]string, start, end time.Time) ([]interface{}, error)
	FetchAll(logName string, tagFilterMap map[string]string) ([]interface{}, error)
	ForceFlush() error
}
