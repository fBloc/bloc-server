package object_storage

import (
	"github.com/fBloc/bloc-backend-go/value_object"
)

type ObjectStorage interface {
	Set(key string, data []byte) error
	Get(key string) ([]byte, error)
	GetPartial(key string, amount int64) ([]byte, error)
	ListObjectKeys(value_object.ObjectStorageKeyFilter) ([]string, error)
}
