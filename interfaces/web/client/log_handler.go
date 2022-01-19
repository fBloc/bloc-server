package client

import (
	"encoding/json"
	"net/http"

	"github.com/fBloc/bloc-server/infrastructure/log"
	"github.com/fBloc/bloc-server/interfaces/web"
	"github.com/fBloc/bloc-server/value_object"
	"github.com/julienschmidt/httprouter"
)

func ReportLog(w http.ResponseWriter, r *http.Request, _ httprouter.Params) {
	var req FuncRunLogHttpReq
	err := json.NewDecoder(r.Body).Decode(&req)
	if err != nil {
		web.WriteBadRequestDataResp(&w, err.Error())
		return
	}

	logger := log.New(
		value_object.FuncRunRecordLog.String(),
		logBackEnd)
	for _, logItem := range req.LogData {
		logger.UploadLog(
			logItem.Time, logItem.Level,
			logItem.TagMap, logItem.Data)
	}
	logger.ForceUpload()
	web.WritePlainSucOkResp(&w)
}
