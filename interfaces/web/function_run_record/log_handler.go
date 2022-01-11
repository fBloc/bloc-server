package function_run_record

import (
	"net/http"

	"github.com/fBloc/bloc-server/interfaces/web"

	"github.com/julienschmidt/httprouter"
)

func GetLogByKey(w http.ResponseWriter, r *http.Request, ps httprouter.Params) {
	logKey := ps.ByName("log_key")
	if logKey == "" {
		web.WriteBadRequestDataResp(&w, "log_key cannot be null")
		return
	}

	data, err := logBackend.FetchAll(logKey, map[string]string{})
	if err != nil {
		web.WriteInternalServerErrorResp(&w, err, "pull log by key failed")
		return
	}
	web.WriteSucResp(&w, data)
}
