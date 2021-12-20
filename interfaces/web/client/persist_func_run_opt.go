package client

import (
	"github.com/fBloc/bloc-server/infrastructure/object_storage"
	"github.com/fBloc/bloc-server/value_object"
)

var objectStorage object_storage.ObjectStorage

func InjectObjectStorageImplement(oSImplement object_storage.ObjectStorage) {
	objectStorage = oSImplement
}

type PersistFuncRunOptFieldHttpReq struct {
	FunctionRunRecordID value_object.UUID `json:"function_run_record_id"`
	OptKey              string            `json:"opt_key"`
	Data                interface{}       `json:"data"`
}

type PersistFuncRunOptFieldHttpResp struct {
	ObjectStorageKey string `json:"object_storage_key"`
	Brief            string `json:"brief"`
}
