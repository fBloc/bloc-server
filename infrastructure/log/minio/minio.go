package minio

import (
	"encoding/json"
	"fmt"
	"sync"
	"time"

	"github.com/fBloc/bloc-backend-go/infrastructure/log"
	"github.com/fBloc/bloc-backend-go/infrastructure/object_storage"
	minioInf "github.com/fBloc/bloc-backend-go/infrastructure/object_storage/minio"
	"github.com/fBloc/bloc-backend-go/value_object"
)

func init() {
	var _ log.Logger = &MinioLogRepository{}
}

type msg struct {
	Level value_object.LogLevel `json:"level"`
	Data  string                `json:"data"`
}

type MinioLogRepository struct {
	name           string
	data           []*msg
	lastUpdateTime time.Time
	objectStorage  object_storage.ObjectStorage
	sync.Mutex
}

func New(
	logName string,
	bucketName string,
	addresses []string,
	key, password string,
) *MinioLogRepository {
	resp := &MinioLogRepository{
		name:           logName,
		lastUpdateTime: time.Now(),
		objectStorage:  minioInf.New(addresses, key, password, bucketName),
	}
	go resp.upload()
	return resp
}

func (logger *MinioLogRepository) SetName(name string) {
	logger.name = name
}

func (
	logger *MinioLogRepository,
) Infof(format string, a ...interface{}) {
	logger.Lock()
	defer logger.Unlock()

	logger.data = append(logger.data, &msg{
		Level: value_object.Info,
		Data:  fmt.Sprintf(format, a...),
	})
}

func (
	logger *MinioLogRepository,
) Warningf(format string, a ...interface{}) {
	logger.Lock()
	defer logger.Unlock()

	logger.data = append(logger.data, &msg{
		Level: value_object.Warning,
		Data:  fmt.Sprintf(format, a...),
	})
}

func (
	logger *MinioLogRepository,
) Errorf(format string, a ...interface{}) {
	logger.Lock()
	defer logger.Unlock()

	logger.data = append(logger.data, &msg{
		Level: value_object.Error,
		Data:  fmt.Sprintf(format, a...),
	})
}

func (logger *MinioLogRepository) ForceUpload() {
	logger.upload()
}

func (logger *MinioLogRepository) upload() {
	logger.Lock()
	defer logger.Unlock()

	timeFlag := time.Now().Add(-30 * time.Second)
	if logger.lastUpdateTime.After(timeFlag) {
		return
	}

	// 上传日志
	data, err := json.Marshal(logger.data)
	logger.data = make([]*msg, 0, 100)
	if err != nil {
		return
	}
	// TODO 要不要panic？
	_ = logger.objectStorage.Set(
		fmt.Sprintf("%s-%d", logger.name, timeFlag.Unix()),
		data)
}
