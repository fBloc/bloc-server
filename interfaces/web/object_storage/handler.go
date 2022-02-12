package object_storage

import (
	"net/http"

	"github.com/fBloc/bloc-server/interfaces/web"
	"github.com/julienschmidt/httprouter"
)

// GetValueByKeyReturnString return full value in string by object storage key
func GetValueByKeyReturnString(w http.ResponseWriter, r *http.Request, ps httprouter.Params) {
	key := ps.ByName("key")
	if key == "" {
		web.WriteBadRequestDataResp(&w, "key in get path cannot be blank")
		return
	}

	keyExist, dataBytes, err := objectStorage.Get(key)
	if !keyExist {
		web.WriteBadRequestDataResp(&w, "key not exist")
		return
	}
	if err != nil {
		web.WriteInternalServerErrorResp(&w, err, "visit object storage infrastructure failed")
		return
	}

	web.WriteSucResp(&w, string(dataBytes))
}

// GetValueByKeyReturnByte return full raw value by object storage key
func GetValueByKeyReturnByte(w http.ResponseWriter, r *http.Request, ps httprouter.Params) {
	key := ps.ByName("key")
	if key == "" {
		web.WriteBadRequestDataResp(&w, "key in get path cannot be blank")
		return
	}

	keyExist, dataBytes, err := objectStorage.Get(key)
	if !keyExist {
		web.WriteBadRequestDataResp(&w, "key not exist")
		return
	}
	if err != nil {
		web.WriteInternalServerErrorResp(&w, err, "visit object storage infrastructure failed")
		return
	}

	web.WriteSucResp(&w, dataBytes)
}
