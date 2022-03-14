package http_server

import (
	"testing"

	"github.com/fBloc/bloc-server/interfaces/web"
	"github.com/fBloc/bloc-server/internal/http_util"
	"github.com/fBloc/bloc-server/value_object"
	. "github.com/smartystreets/goconvey/convey"
)

func TestReportClientHeartBeat(t *testing.T) {
	Convey("report client heartbeat", t, func() {
		fakeFuncRunRecordID := value_object.NewUUID()
		resp := struct {
			web.RespMsg
			Data string `json:"data"`
		}{}
		_, err := http_util.Get(
			superuserHeader(),
			serverAddress+"/api/v1/client/report_functionExecute_heartbeat/"+fakeFuncRunRecordID.String(),
			http_util.BlankGetParam,
			&resp)
		So(err, ShouldBeNil)
		So(resp.Data, ShouldEqual, "ok")
	})
}
