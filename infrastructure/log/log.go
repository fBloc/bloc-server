package log

import (
	"encoding/json"
	"fmt"
	"sync"
	"time"

	"github.com/fBloc/bloc-backend-go/infrastructure/log_collect_backend"
	"github.com/fBloc/bloc-backend-go/value_object"
)

type msg struct {
	Level value_object.LogLevel `json:"level"`
	Data  string                `json:"data"`
	Time  time.Time             `json:"time"`
}

type Logger struct {
	name       string
	data       []*msg
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
	go l.upload()
	return l
}

func (
	logger *Logger,
) Infof(format string, a ...interface{}) {
	logger.Lock()
	defer logger.Unlock()

	logger.data = append(logger.data, &msg{
		Time:  time.Now(),
		Level: value_object.Info,
		Data:  fmt.Sprintf(format, a...),
	})
}

func (
	logger *Logger,
) Warningf(format string, a ...interface{}) {
	logger.Lock()
	defer logger.Unlock()

	logger.data = append(logger.data, &msg{
		Time:  time.Now(),
		Level: value_object.Warning,
		Data:  fmt.Sprintf(format, a...),
	})
}

func (
	logger *Logger,
) Errorf(format string, a ...interface{}) {
	logger.Lock()
	defer logger.Unlock()

	logger.data = append(logger.data, &msg{
		Time:  time.Now(),
		Level: value_object.Error,
		Data:  fmt.Sprintf(format, a...),
	})
}

func (logger *Logger) ForceUpload() {
	if len(logger.data) <= 0 {
		return
	}

	logger.Lock()
	data, err := json.Marshal(logger.data)
	logger.data = logger.data[:0]
	logger.Unlock()
	if err != nil {
		panic(err)
		// return
	}

	// TODO 要不要panic？
	err = logger.logBackend.PersistData(
		fmt.Sprintf("%s-%d", logger.name, time.Now().Unix()),
		data)
	if err != nil {
		panic(err)
	}
}

func (logger *Logger) upload() {
	ticker := time.NewTicker(30 * time.Second)
	defer ticker.Stop()

	for range ticker.C {
		if len(logger.data) <= 0 {
			continue
		}

		logger.Lock()

		// 上传日志
		data, err := json.Marshal(logger.data)
		logger.data = logger.data[:0]
		logger.Unlock()

		if err != nil {
			panic(err)
			// return
		}
		// TODO 要不要panic？
		timeFlag := time.Now().Add(-30 * time.Second)
		err = logger.logBackend.PersistData(
			fmt.Sprintf("%s-%d", logger.name, timeFlag.Unix()),
			data)
		if err != nil {
			panic(err)
		}
	}
}

func (logger *Logger) PullLogBetweenTime(
	timeStart, timeEnd time.Time,
) ([]interface{}, error) {
	return logger.logBackend.PullDataBetween(logger.name, timeStart, timeEnd)
}
