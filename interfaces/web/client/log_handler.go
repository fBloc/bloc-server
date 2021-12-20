package client

import (
	"encoding/json"
	"net/http"

	"github.com/fBloc/bloc-server/infrastructure/log"
	"github.com/fBloc/bloc-server/interfaces/web"
	"github.com/julienschmidt/httprouter"
)

func ReportLog(w http.ResponseWriter, r *http.Request, _ httprouter.Params) {
	var req LogHttpReq
	err := json.NewDecoder(r.Body).Decode(&req)
	if err != nil {
		web.WriteBadRequestDataResp(&w, err.Error())
		return
	}

	logger := log.New(req.Name, logBackEnd)
	logger.ForceUpload()

	web.WritePlainSucOkResp(&w)
}
