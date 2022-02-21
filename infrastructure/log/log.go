package log

import (
	"fmt"
	"sync"
	"time"

	"github.com/fBloc/bloc-server/infrastructure/log_collect_backend"
	"github.com/fBloc/bloc-server/value_object"
)

type Logger struct {
	name       string
	logBackend log_collect_backend.LogBackEnd
	sync.Mutex
}

func (logger *Logger) IsZero() bool {
	if logger == nil {
		return true
	}
	return logger.name == "" || logger.logBackend == nil
}

func New(
	name string,
	collectBackend log_collect_backend.LogBackEnd,
) *Logger {
	l := &Logger{
		name:       name,
		logBackend: collectBackend,
	}
	return l
}

func (
	logger *Logger,
) UploadLog(
	happenTime time.Time,
	logLevel value_object.LogLevel,
	tagMap map[string]string,
	format string,
	a ...interface{},
) {
	tagMap["log_level"] = string(logLevel)
	logger.logBackend.Write(
		logger.name,
		tagMap,
		fmt.Sprintf(format, a...),
		happenTime)
}

func (
	logger *Logger,
) Infof(
	tagMap map[string]string,
	format string, a ...interface{},
) {
	tagMap["log_level"] = string(value_object.Info)
	logger.logBackend.Write(
		logger.name, tagMap,
		fmt.Sprintf(format, a...),
		time.Now())
}

func (
	logger *Logger,
) Warningf(
	tagMap map[string]string,
	format string, a ...interface{},
) {
	tagMap["log_level"] = string(value_object.Warning)
	logger.logBackend.Write(
		logger.name, tagMap,
		fmt.Sprintf(format, a...), time.Now())
}

func (
	logger *Logger,
) Errorf(
	tagMap map[string]string,
	format string, a ...interface{},
) {
	tagMap["log_level"] = string(value_object.Error)
	logger.logBackend.Write(
		logger.name, tagMap,
		fmt.Sprintf(format, a...), time.Now())
}

func (logger *Logger) ForceUpload() {
	logger.logBackend.ForceFlush()
}

func (logger *Logger) PullLogBetweenTime(
	tagFilter map[string]string,
	timeStart, timeEnd time.Time,
) ([]interface{}, error) {
	return logger.logBackend.Fetch(logger.name, tagFilter, timeStart, timeEnd)
}
