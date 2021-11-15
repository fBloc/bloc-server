package object_storage

import (
	"net/http"

	"github.com/fBloc/bloc-backend-go/interfaces/web"
	"github.com/julienschmidt/httprouter"
)

// GetValueByKeyReturnString return full value by object storage key
func GetValueByKeyReturnString(w http.ResponseWriter, r *http.Request, ps httprouter.Params) {
	key := ps.ByName("key")
	if key == "" {
		web.WriteBadRequestDataResp(&w, "key in get path cannot be blank")
		return
	}

	dataBytes, err := objectStorage.Get(key)
	if err != nil { // TODO this error may be caused by key not exist
		web.WriteInternalServerErrorResp(&w, err, "visit object storage infrastructure failed")
		return
	}

	web.WriteSucResp(&w, string(dataBytes))
}