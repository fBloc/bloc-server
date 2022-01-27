package minio

import (
	"sync"

	"github.com/fBloc/bloc-server/infrastructure/object_storage"
	minioConn "github.com/fBloc/bloc-server/internal/conns/minio"
)

func init() {
	var _ object_storage.ObjectStorage = &ObjectStorageMinioRepository{}
}

type ObjectStorageMinioRepository struct {
	bucketName string
	conn       *minioConn.MinioCon
	sync.Mutex
}

func New(minioConf *minioConn.MinioConfig) (*ObjectStorageMinioRepository, error) {
	minioClient, err := minioConn.Connect(minioConf)
	if err != nil {
		return nil, err
	}

	oSMR := &ObjectStorageMinioRepository{
		bucketName: minioConf.BucketName,
		conn:       minioClient,
	}
	return oSMR, nil
}

func (oSMR *ObjectStorageMinioRepository) Set(key string, byteData []byte) error {
	return oSMR.conn.Set(key, byteData)
}

func (oSMR *ObjectStorageMinioRepository) Get(key string) ([]byte, error) {
	return oSMR.conn.Get(key)
}
