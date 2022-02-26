package object_storage

import (
	"github.com/fBloc/bloc-server/infrastructure/log"
	"github.com/fBloc/bloc-server/infrastructure/object_storage"
)

var (
	objectStorage object_storage.ObjectStorage
	logger        *log.Logger
)

func InjectObjectStorageImplement(
	oSImplement object_storage.ObjectStorage,
) {
	objectStorage = oSImplement
}

func InjectLogger(
	l *log.Logger,
) {
	logger = l
}
