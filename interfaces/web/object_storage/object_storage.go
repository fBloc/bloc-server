package object_storage

import (
	"github.com/fBloc/bloc-server/infrastructure/object_storage"
)

var objectStorage object_storage.ObjectStorage

func InjectObjectStorageImplement(oSImplement object_storage.ObjectStorage) {
	objectStorage = oSImplement
}
