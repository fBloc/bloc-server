package minio

import (
	"encoding/json"
	"fmt"
	"sort"
	"strconv"
	"strings"
	"sync"
	"time"

	"github.com/fBloc/bloc-backend-go/infrastructure/log_collect_backend"
	"github.com/fBloc/bloc-backend-go/infrastructure/object_storage"
	minioInf "github.com/fBloc/bloc-backend-go/infrastructure/object_storage/minio"
	"github.com/fBloc/bloc-backend-go/value_object"
)

func init() {
	var _ log_collect_backend.LogBackEnd = &MinioLogBackendRepository{}
}

type MinioLogBackendRepository struct {
	objectStorage object_storage.ObjectStorage
	sync.Mutex
}

func New(
	bucketName string,
	addresses []string,
	key, password string,
) *MinioLogBackendRepository {
	resp := &MinioLogBackendRepository{
		objectStorage: minioInf.New(addresses, key, password, bucketName),
	}
	return resp
}

func (
	backEnd *MinioLogBackendRepository,
) PersistData(key string, data []byte) error {
	return backEnd.objectStorage.Set(key, data)
}

func (backEnd *MinioLogBackendRepository) ListKeysBetween(
	prefixKey string,
	timeStart, timeEnd time.Time,
) ([]string, error) {
	// 因为日志后缀精确到秒，故去掉nanosec信息
	timeStart = time.Date(timeStart.Year(), timeStart.Month(), timeStart.Day(), timeStart.Hour(), timeStart.Minute(), timeStart.Second(), 0, time.UTC)
	timeEnd = time.Date(timeEnd.Year(), timeEnd.Month(), timeEnd.Day(), timeEnd.Hour(), timeEnd.Minute(), timeEnd.Second(), 0, time.UTC)
	// 构建尽可能精确的prefix
	startTimeStampStr := strconv.FormatInt(timeStart.Unix(), 10)
	endTimeStampStr := strconv.FormatInt(timeEnd.Unix(), 10)
	commonTimePrefix := ""
	for startIndex, endIndex := 0, 0; startIndex < 10; {
		if startTimeStampStr[startIndex] == endTimeStampStr[endIndex] {
			commonTimePrefix += string(startTimeStampStr[startIndex])
			startIndex++
			endIndex++
			continue
		}
		break
	}
	searchKeysPrefix := prefixKey + "-" + commonTimePrefix

	// 查询出全部的key
	keys, err := backEnd.objectStorage.ListObjectKeys(
		value_object.ObjectStorageKeyFilter{
			StartWith: searchKeysPrefix})
	if err != nil {
		return []string{}, err
	}

	// 确保key有序，早的在列表开头
	// TODO maybe minio return is already sorted？？
	var validTimeStamps []int64
	for _, k := range keys {
		splitedKeys := strings.Split(k, "-")
		timeStampStr := splitedKeys[len(splitedKeys)-1]
		i, err := strconv.ParseInt(timeStampStr, 10, 64)
		if err != nil {
			panic(err)
		}
		tmp := time.Unix(i, 0).UTC()
		if tmp.Before(timeStart) || tmp.After(timeEnd) {
			continue
		}
		validTimeStamps = append(validTimeStamps, tmp.Unix())
	}

	sort.Slice(
		validTimeStamps,
		func(i, j int) bool { return validTimeStamps[i] < validTimeStamps[j] })

	var ret []string
	for _, i := range validTimeStamps {
		ret = append(ret, fmt.Sprintf("%s-%d", prefixKey, i))
	}
	return ret, nil
}

func (backEnd *MinioLogBackendRepository) PullDataByKey(
	key string,
) ([]interface{}, error) {
	data, err := backEnd.objectStorage.Get(key)
	if err != nil {
		return []interface{}{}, err
	}
	var ret []interface{}
	err = json.Unmarshal(data, &ret)
	if err != nil {
		return []interface{}{}, err
	}

	return ret, nil
}

func (backEnd *MinioLogBackendRepository) PullDataBetween(
	prefixKey string,
	timeStart, timeEnd time.Time,
) ([]string, error) {
	sortedLogKeys, err := backEnd.ListKeysBetween(prefixKey, timeStart, timeEnd)
	if err != nil {
		return nil, err
	}

	var wg sync.WaitGroup
	wg.Add(len(sortedLogKeys))
	type respLogByteWithIndex struct {
		logString string
		index     int
	}
	logSlice := make([]respLogByteWithIndex, 0, len(sortedLogKeys))
	var mu sync.Mutex
	for index, key := range sortedLogKeys {
		go func(wg *sync.WaitGroup, index int, key string) {
			defer wg.Done()
			thisLog, err := backEnd.objectStorage.Get(key)
			if err != nil {
				return
			}
			mu.Lock()
			logSlice = append(logSlice, respLogByteWithIndex{
				index:     index,
				logString: string(thisLog)})
			defer mu.Unlock()
		}(&wg, index, key)
	}
	wg.Wait()

	sort.Slice(logSlice[:], func(i, j int) bool {
		return logSlice[i].index < logSlice[j].index
	})

	var resp []string
	for _, j := range logSlice {
		resp = append(resp, j.logString)
	}

	return resp, nil
}
