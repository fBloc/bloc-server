package object_storage

import (
	"net/http"

	"github.com/fBloc/bloc-server/interfaces/web"
	"github.com/julienschmidt/httprouter"
)

// GetValueByKeyReturnString return full value in string by object storage key
func GetValueByKeyReturnString(w http.ResponseWriter, r *http.Request, ps httprouter.Params) {
	logTags := web.GetTraceAboutFields(r.Context())
	logTags["business"] = "get object storage string value"

	key := ps.ByName("key")
	if key == "" {
		logger.Warningf(logTags, "key in get path cannot be blank")
		web.WriteBadRequestDataResp(&w, r, "key in get path cannot be blank")
		return
	}
	logTags["key"] = key

	keyExist, dataBytes, err := objectStorage.Get(key)
	if !keyExist {
		logger.Warningf(logTags, "key not exist")
		web.WriteBadRequestDataResp(&w, r, "key not exist")
		return
	}
	if err != nil {
		logger.Errorf(logTags, "visit object storage failed: %v", err)
		web.WriteInternalServerErrorResp(&w, r, err, "visit object storage infrastructure failed")
		return
	}

	logger.Infof(logTags, "finished")
	web.WriteSucResp(&w, r, string(dataBytes))
}

// GetValueByKeyReturnByte return full raw value by object storage key
func GetValueByKeyReturnByte(w http.ResponseWriter, r *http.Request, ps httprouter.Params) {
	logTags := web.GetTraceAboutFields(r.Context())
	logTags["business"] = "get object storage byte value"

	key := ps.ByName("key")
	if key == "" {
		logger.Warningf(logTags, "key in get path cannot be blank")
		web.WriteBadRequestDataResp(&w, r, "key in get path cannot be blank")
		return
	}

	keyExist, dataBytes, err := objectStorage.Get(key)
	if !keyExist {
		logger.Warningf(logTags, "key not exist")
		web.WriteBadRequestDataResp(&w, r, "key not exist")
		return
	}
	if err != nil {
		logger.Errorf(logTags, "visit object storage failed: %v", err)
		web.WriteInternalServerErrorResp(&w, r, err, "visit object storage infrastructure failed")
		return
	}

	logger.Infof(logTags, "finished")
	web.WriteSucResp(&w, r, dataBytes)
}
