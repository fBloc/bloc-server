package minio

import (
	"bytes"
	"context"
	"errors"
	"fmt"
	"sync"

	"github.com/minio/minio-go/v7"
	"github.com/minio/minio-go/v7/pkg/credentials"
)

var (
	client         *minio.Client
	config         *MinioConfig
	setClientMutex sync.Mutex
)

type MinioConfig struct {
	BucketName     string
	AccessKey      string
	AccessPassword string
	Addresses      []string
}

func (mF *MinioConfig) IsNil() bool {
	if mF == nil {
		return true
	}
	return mF.BucketName == "" || mF.AccessKey == "" ||
		mF.AccessPassword == "" || len(mF.Addresses) == 0
}

func Init(conf *MinioConfig) *minio.Client {
	config = conf
	setClientMutex.Lock()
	client = getValidClient()
	setClientMutex.Unlock()
	if client == nil {
		panic("connect to minio fialed, check your minio service")
	}

	// 确保bucket已经存在了
	err := client.MakeBucket(context.Background(), config.BucketName, minio.MakeBucketOptions{})
	if err != nil { // 有可能是bucket已经存在了而爆出的error
		exists, errBucketExists := client.BucketExists(context.Background(), config.BucketName)
		if errBucketExists == nil && exists {
			// already has this bucket
		} else {
			panic(fmt.Sprintf("minio initial bucket failed: %s", err.Error()))
		}
	}
	return client
}

func getValidClient() *minio.Client {
	for i := 0; i < len(config.Addresses); i++ {
		// this seconds return error make no sense.
		// as i tested, event the address not exist server. error stands nil
		minioClient, _ := minio.New(config.Addresses[i], &minio.Options{
			Creds:  credentials.NewStaticV4(config.AccessKey, config.AccessPassword, ""),
			Secure: false,
		})

		// make sure client is valid
		_, err := minioClient.ListBuckets(context.TODO())
		if err == nil {
			return minioClient
		}
	}
	return nil
}

func NewClient(addresses []string, key, password, bucketName string) *minio.Client {
	return Init(&MinioConfig{
		BucketName:     bucketName,
		AccessKey:      key,
		AccessPassword: password,
		Addresses:      addresses})
}

func GetClient() *minio.Client {
	return client
}

func setWithRetry(key string, byteData []byte) error {
	// 两次尝试，若第一次失败，切换client，再次进行尝试，若依然失败，返回error
	for i := 0; i < 2; i++ {
		minioClient := GetClient()
		objJsonIOReader := bytes.NewReader(byteData)
		_, err := minioClient.PutObject(
			context.Background(),
			config.BucketName,
			key,
			objJsonIOReader,
			objJsonIOReader.Size(),
			minio.PutObjectOptions{ContentType: "application/octet-stream"})
		if err == nil {
			return nil
		}
		validClient := getValidClient()
		if validClient == nil {
			return errors.New("get no valid minio client")
		}
		setClientMutex.Lock()
		client = validClient
		setClientMutex.Unlock()
	}
	return errors.New("save to oss failed")
}

func Set(key string, byteData []byte) error {
	return setWithRetry(key, byteData)
}

func getWithRetry(key string) ([]byte, error) {
	// 两次尝试，若第一次失败，切换client，再次进行尝试，若依然失败，返回error
	for i := 0; i < 2; i++ {
		minioClient := GetClient()

		reader, err := minioClient.GetObject(
			context.Background(), config.BucketName, key, minio.GetObjectOptions{})
		if err != nil {
			validClient := getValidClient()
			if validClient == nil {
				continue
			}
			setClientMutex.Lock()
			client = validClient
			setClientMutex.Unlock()
			continue
		}
		defer reader.Close()

		stat, err := reader.Stat()
		if err != nil {
			validClient := getValidClient()
			if validClient == nil {
				continue
			}
			setClientMutex.Lock()
			client = validClient
			setClientMutex.Unlock()
			continue
		}
		data := make([]byte, stat.Size)
		reader.Read(data)
		return data, nil
	}
	return nil, errors.New("get failed")
}

func Get(key string) ([]byte, error) {
	return getWithRetry(key)
}

func GetPartial(key string, amount int64) (string, error) {
	minioClient := GetClient()

	reader, err := minioClient.GetObject(
		context.Background(), config.BucketName, key, minio.GetObjectOptions{})
	if err != nil {
		return "", err
	}
	defer reader.Close()

	_, err = reader.Stat()
	if err != nil {
		return "", err
	}

	data := make([]byte, amount)
	reader.Read(data)
	return string(data), nil
}
