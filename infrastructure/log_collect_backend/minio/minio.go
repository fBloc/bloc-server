package minio

import (
	"errors"
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

func (backEnd *MinioLogBackendRepository) PullDataBetween(
	prefixKey string,
	timeStart, timeEnd time.Time,
) ([]string, error) {
	if !timeStart.Before(timeEnd) {
		return []string{}, errors.New("time_start is not before time_end")
	}
	if timeEnd.Sub(timeStart) > 24*time.Hour {
		return []string{}, errors.New("not support gap duration > 1 day")
	}
	startTimeStampStr := strconv.FormatInt(timeStart.Unix(), 10)
	endTimeStampStr := strconv.FormatInt(timeEnd.Unix(), 10)
	commonTimePrefix := ""
	startIndex, endIndex := 0, 0
	for {
		if startTimeStampStr[startIndex] == endTimeStampStr[endIndex] {
			commonTimePrefix += string(startTimeStampStr[startIndex])
			startIndex++
			endIndex++
			continue
		}
		break
	}

	searchKeysPrefix := prefixKey + "-" + commonTimePrefix
	keys, err := backEnd.objectStorage.ListObjectKeys(
		value_object.ObjectStorageKeyFilter{
			StartWith: searchKeysPrefix})
	if err != nil {
		return nil, err
	}
	var validTimeStamps []int64
	for _, k := range keys {
		splitedKeys := strings.Split(k, "-")
		timeStampStr := splitedKeys[len(splitedKeys)-1]
		i, err := strconv.ParseInt(timeStampStr, 10, 64)
		if err != nil {
			panic(err)
		}
		tmp := time.Unix(i, 0)
		if tmp.Before(timeStart) || tmp.After(timeEnd) {
			continue
		}
		validTimeStamps = append(validTimeStamps, tmp.Unix())
	}

	sort.Slice(
		validTimeStamps,
		func(i, j int) bool { return validTimeStamps[i] < validTimeStamps[j] })

	var wg sync.WaitGroup
	wg.Add(len(validTimeStamps))
	type respLogByteWithIndex struct {
		logString string
		index     int
	}
	logSlice := make([]respLogByteWithIndex, 0, len(validTimeStamps))
	var mu sync.Mutex
	for index, timeStamp := range validTimeStamps {
		go func(wg *sync.WaitGroup, index int, timeStamp int64) {
			defer wg.Done()
			thisLog, err := backEnd.objectStorage.Get(
				fmt.Sprintf("%s-%d", prefixKey, timeStamp))
			if err != nil {
				return
			}
			mu.Lock()
			logSlice = append(logSlice, respLogByteWithIndex{
				index:     index,
				logString: string(thisLog)})
			defer mu.Unlock()
		}(&wg, index, timeStamp)
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
