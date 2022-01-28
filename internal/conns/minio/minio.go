package minio

import (
	"bytes"
	"context"
	"sync"

	"github.com/minio/minio-go/v7"
	"github.com/minio/minio-go/v7/pkg/credentials"
	"github.com/pkg/errors"
)

var confSignatureMapClient = make(map[confSignature]*MinioCon)
var confSignatureMapClientMutex sync.Mutex

func getValidClient(config *MinioConfig) *minio.Client {
	for i := 0; i < len(config.Addresses); i++ {
		// this seconds return error make no sense.
		// as i tested, event the address not exist server. error stands nil
		minioClient, _ := minio.New(config.Addresses[i], &minio.Options{
			Creds:  credentials.NewStaticV4(config.AccessKey, config.AccessPassword, ""),
			Secure: false,
		})

		// make sure client is valid
		_, err := minioClient.ListBuckets(context.TODO())
		if err != nil {
			continue
		}

		// create bucket if not exist
		err = minioClient.MakeBucket(
			context.Background(), config.BucketName, minio.MakeBucketOptions{})
		if err != nil { // maybe cause by bucket already exist
			exists, errBucketExists := minioClient.BucketExists(
				context.Background(), config.BucketName)
			if errBucketExists == nil && exists {
				// already has this bucket
				return minioClient
			}
		}
	}
	return nil
}

func Connect(conf *MinioConfig) (*MinioCon, error) {
	sig := conf.signature()
	con, ok := confSignatureMapClient[sig]
	if ok && con.client != nil {
		return con, nil
	}

	config := *conf

	conn := MinioCon{
		conf:      config,
		signature: sig,
		client:    nil}

	err := conn.initialClient()
	if err != nil {
		return nil, err
	}

	confSignatureMapClientMutex.Lock()
	confSignatureMapClient[sig] = &conn
	confSignatureMapClientMutex.Unlock()

	return &conn, err
}

type MinioCon struct {
	conf           MinioConfig
	signature      confSignature
	client         *minio.Client
	getClientMutex sync.Mutex
}

func (con *MinioCon) initialClient() error {
	return con.switchToAValidClient()
}

func (con *MinioCon) switchToAValidClient() error {
	con.getClientMutex.Lock()
	defer con.getClientMutex.Unlock()

	if con.client != nil {
		return nil
	}

	client := getValidClient(&con.conf)
	if client == nil {
		return errors.New("no valid client")
	}
	con.client = client
	return nil
}

func (con *MinioCon) Set(key string, byteData []byte) (err error) {
	// have one more change to switch client to set
	for i := 0; i < 2; i++ {
		objJsonIOReader := bytes.NewReader(byteData)
		_, err = con.client.PutObject(
			context.Background(),
			con.conf.BucketName,
			key,
			objJsonIOReader,
			objJsonIOReader.Size(),
			minio.PutObjectOptions{ContentType: "application/octet-stream"})
		if err == nil {
			return nil
		}

		changeClientErr := con.switchToAValidClient()
		if changeClientErr != nil {
			return errors.Wrap(changeClientErr, "no valid client")
		}
	}
	return errors.Wrap(err, "save to object storage error:")
}

func (con *MinioCon) Get(key string) ([]byte, error) {
	// have one more change to switch client to get
	for i := 0; i < 2; i++ {
		reader, err := con.client.GetObject(
			context.Background(), con.conf.BucketName, key, minio.GetObjectOptions{})
		if err != nil {
			err = con.switchToAValidClient()
			if err != nil {
				return []byte{}, errors.Wrap(err, "no valid client")
			} else {
				continue
			}
		}
		defer reader.Close()

		stat, err := reader.Stat()
		if err != nil {
			err = con.switchToAValidClient()
			if err != nil {
				return []byte{}, errors.Wrap(err, "no valid client")
			} else {
				continue
			}
		}

		data := make([]byte, stat.Size)
		reader.Read(data)
		return data, nil
	}
	return nil, errors.New("get failed")
}

// GetPartial current no use
// func (con *MinioCon) GetPartial(key string, amount int64) (string, error) {
// 	reader, err := con.client.GetObject(
// 		context.Background(), con.conf.BucketName, key, minio.GetObjectOptions{})
// 	if err != nil {
// 		return "", err
// 	}
// 	defer reader.Close()

// 	_, err = reader.Stat()
// 	if err != nil {
// 		return "", err
// 	}

// 	data := make([]byte, amount)
// 	reader.Read(data)
// 	return string(data), nil
// }
