package http_server

import (
	"encoding/json"
	"log"
	"net/http"
	"testing"
	"time"

	"github.com/brianvoe/gofakeit/v6"
	"github.com/fBloc/bloc-server/config"
	"github.com/fBloc/bloc-server/interfaces/web"
	"github.com/fBloc/bloc-server/interfaces/web/client"
	"github.com/fBloc/bloc-server/interfaces/web/flow"
	"github.com/fBloc/bloc-server/interfaces/web/flow_run_record"
	"github.com/fBloc/bloc-server/interfaces/web/function_run_record"
	"github.com/fBloc/bloc-server/internal/http_util"
	"github.com/fBloc/bloc-server/value_object"
	. "github.com/smartystreets/goconvey/convey"
	"github.com/spf13/cast"
)

func TestRunFlow(t *testing.T) {
	// create a draft flow
	clientFlow := getFakeAggFlow()
	reqFlow := aggFlowToWebFlow(clientFlow)
	reqBody, _ := json.Marshal(reqFlow)
	resp := struct {
		web.RespMsg
		DraftFlow *flow.Flow `json:"data"`
	}{}
	http_util.Post(
		superuserHeader(),
		serverAddress+"/api/v1/draft_flow",
		http_util.BlankGetParam,
		reqBody, &resp)
	if resp.DraftFlow.IsZero() {
		log.Fatalf("create draft flow failed")
	}
	if resp.DraftFlow.ID.IsNil() {
		log.Fatalf("create draft flow's ID is nil")
	}

	// pub draft to online flow
	pubResp := struct {
		web.RespMsg
		OnlineFlow *flow.Flow `json:"data"`
	}{}
	_, err := http_util.Get(
		superuserHeader(),
		serverAddress+"/api/v1/draft_flow/commit_by_id/"+resp.DraftFlow.ID.String(),
		http_util.BlankGetParam,
		&pubResp)
	if err != nil {
		log.Fatalf("pub draft flow to online failed: %v", err)
	}
	if pubResp.OnlineFlow.IsZero() {
		log.Fatalf("pub draft flow to online returned zero flow. resp: %v", pubResp)
	}
	onlineFlow := pubResp.OnlineFlow

	Convey("trigger flow to run", t, func() {
		var resp web.RespMsg
		_, err := http_util.Get(
			superuserHeader(),
			serverAddress+"/api/v1/flow/run/by_origin_id/"+onlineFlow.OriginID.String(),
			http_util.BlankGetParam, &resp)
		So(err, ShouldBeNil)
		So(resp.Code, ShouldEqual, http.StatusOK)
		time.Sleep(time.Second)

		flowRunRecordResp := struct {
			web.RespMsg
			Data struct {
				Total int64                                 `json:"total"`
				Items []*flow_run_record.FlowFunctionRecord `json:"items"`
			} `json:"data"`
		}{}

		Convey("flow_run_record lifecycle!!!", func() {
			_, err := http_util.Get(
				superuserHeader(),
				serverAddress+"/api/v1/flow_run_record/",
				map[string]string{"flow_origin_id": onlineFlow.OriginID.String()},
				&flowRunRecordResp)
			So(err, ShouldBeNil)
			So(flowRunRecordResp.Data.Total, ShouldEqual, 1)

			theFlowRunRecord := flowRunRecordResp.Data.Items[0]
			So(theFlowRunRecord.FlowID, ShouldEqual, onlineFlow.ID)
			So(theFlowRunRecord.FlowOriginID, ShouldEqual, onlineFlow.OriginID)
			So( // checkout first layer's functions are all published
				len(theFlowRunRecord.FlowFuncIDMapFuncRunRecordID),
				ShouldEqual,
				len(clientFlow.FlowFunctionIDMapFlowFunction[config.FlowFunctionStartID].DownstreamFlowFunctionIDs),
			)
			So(theFlowRunRecord.StartTime.IsZero(), ShouldBeFalse)
			So(theFlowRunRecord.EndTime, ShouldEqual, nil)

			addFunctionRunRecordID, ok := theFlowRunRecord.FlowFuncIDMapFuncRunRecordID[aggFuncAddFlowFunctionID]
			So(ok, ShouldBeTrue)
			So(addFunctionRunRecordID.IsNil(), ShouldBeFalse)

			// fetch the unfinished function_run_record ins
			addFunctionRunRecordresp := struct {
				web.RespMsg
				FunctionRunRecord *function_run_record.FunctionRunRecord `json:"data"`
			}{}
			_, err = http_util.Get(
				superuserHeader(),
				serverAddress+"/api/v1/function_run_record/get_by_id/"+addFunctionRunRecordID.String(),
				http_util.BlankGetParam, &addFunctionRunRecordresp)
			So(err, ShouldBeNil)
			So(addFunctionRunRecordresp.Code, ShouldEqual, http.StatusOK)
			So(addFunctionRunRecordresp.FunctionRunRecord.ID, ShouldEqual, addFunctionRunRecordID)
			So(addFunctionRunRecordresp.FunctionRunRecord.End, ShouldBeNil)
			So(addFunctionRunRecordresp.FunctionRunRecord.Suc, ShouldBeFalse)
			So(addFunctionRunRecordresp.FunctionRunRecord.ProgressMilestoneIndex, ShouldEqual, nil)
			So(addFunctionRunRecordresp.FunctionRunRecord.Progress, ShouldEqual, 0)
			So(len(addFunctionRunRecordresp.FunctionRunRecord.ProgressMsg), ShouldEqual, 0)

			// log
			toLogData := gofakeit.RandomString(allChars)
			logReq := &client.FuncRunLogHttpReq{
				LogData: []*client.Msg{
					{
						Level: value_object.Info,
						TagMap: map[string]string{
							"function_run_record_id": addFunctionRunRecordID.String()},
						Data: toLogData,
						Time: time.Now(),
					},
				},
			}
			logReqBody, err := json.Marshal(logReq)
			So(err, ShouldBeNil)
			var logReportResp web.RespMsg
			_, err = http_util.Post(
				superuserHeader(),
				serverAddress+"/api/v1/client/report_log",
				http_util.BlankGetParam, logReqBody, &logReportResp)
			So(err, ShouldBeNil)

			pullLogResp := struct {
				web.RespMsg
				Logs []interface{} `json:"data"`
			}{}
			_, err = http_util.Get(
				superuserHeader(),
				serverAddress+"/api/v1/function_run_record/pull_log_by_id/"+addFunctionRunRecordID.String(),
				http_util.BlankGetParam, &pullLogResp)
			So(err, ShouldBeNil)
			So(pullLogResp.Code, ShouldEqual, http.StatusOK)
			So(len(pullLogResp.Logs), ShouldEqual, 1)
			logMap := cast.ToStringMap(pullLogResp.Logs[0])
			fetchedLogData := cast.ToString(logMap["data"])
			So(fetchedLogData, ShouldEqual, toLogData)

			// report_progress
			ProgressMilestoneIndex := 1
			var progressPercent float32 = 20
			progressMsg := "suc parsed ipt params"
			progressReport := client.ProgressReportHttpReq{
				FunctionRunRecordID: addFunctionRunRecordID.String(),
				FuncRunProgress: client.HighReadableFunctionRunProgress{
					ProgressMilestoneIndex: &ProgressMilestoneIndex,
					Progress:               progressPercent,
					Msg:                    progressMsg},
			}
			progressReportBody, err := json.Marshal(progressReport)
			So(err, ShouldBeNil)
			var progressReportResp web.RespMsg
			_, err = http_util.Post(
				superuserHeader(),
				serverAddress+"/api/v1/client/report_progress",
				http_util.BlankGetParam, progressReportBody, &progressReportResp)
			So(err, ShouldBeNil)

			http_util.Get(
				superuserHeader(),
				serverAddress+"/api/v1/function_run_record/get_by_id/"+addFunctionRunRecordID.String(),
				http_util.BlankGetParam, &addFunctionRunRecordresp)
			So(*addFunctionRunRecordresp.FunctionRunRecord.ProgressMilestoneIndex, ShouldEqual, ProgressMilestoneIndex)
			So(addFunctionRunRecordresp.FunctionRunRecord.Progress, ShouldEqual, progressPercent)
			So(addFunctionRunRecordresp.FunctionRunRecord.ProgressMsg[0], ShouldEqual, progressMsg)

			// upload add-function run opt
			addFunctionOptSumFieldPersistReq := &client.PersistFuncRunOptFieldHttpReq{
				FunctionRunRecordID: addFunctionRunRecordID,
				OptKey:              "sum",
				Data:                12,
			}
			addFunctionOptSumFieldPersistBody, _ := json.Marshal(addFunctionOptSumFieldPersistReq)
			addFunctionOptSumFieldPersistResp := struct {
				web.RespMsg
				Data *client.PersistFuncRunOptFieldHttpResp `json:"data"`
			}{}
			_, err = http_util.Post(
				superuserHeader(),
				serverAddress+"/api/v1/client/persist_certain_function_run_opt_field",
				http_util.BlankGetParam, addFunctionOptSumFieldPersistBody, &addFunctionOptSumFieldPersistResp)
			So(err, ShouldBeNil)
			So(addFunctionOptSumFieldPersistResp.Code, ShouldEqual, http.StatusOK)

			// check the upload is valid
			objectStorageResp := struct {
				web.RespMsg
				Data []byte `json:"data"`
			}{}
			_, err = http_util.Get(
				superuserHeader(),
				serverAddress+"/api/v1/client/get_byte_value_by_key/"+addFunctionOptSumFieldPersistResp.Data.ObjectStorageKey,
				http_util.BlankGetParam, &objectStorageResp)
			So(err, ShouldBeNil)
			var sumRespInOBS int
			err = json.Unmarshal(objectStorageResp.Data, &sumRespInOBS)
			So(err, ShouldBeNil)
			So(sumRespInOBS, ShouldEqual, 12)

			// report add-function finished
			reportFinished := &client.FuncRunFinishedHttpReq{
				FunctionRunRecordID:       addFunctionRunRecordID.String(),
				Suc:                       true,
				OptKeyMapBriefData:        map[string]string{"sum": addFunctionOptSumFieldPersistResp.Data.Brief},
				OptKeyMapObjectStorageKey: map[string]string{"sum": addFunctionOptSumFieldPersistResp.Data.ObjectStorageKey},
			}
			reportFinishedBody, _ := json.Marshal(reportFinished)
			var reportFinishedResp web.RespMsg
			_, err = http_util.Post(
				superuserHeader(),
				serverAddress+"/api/v1/client/function_run_finished",
				http_util.BlankGetParam, reportFinishedBody, &reportFinishedResp)
			So(err, ShouldBeNil)
			So(reportFinishedResp.Code, ShouldEqual, http.StatusOK)

			// check report finished is valid
			_, err = http_util.Get(
				superuserHeader(),
				serverAddress+"/api/v1/function_run_record/get_by_id/"+addFunctionRunRecordID.String(),
				http_util.BlankGetParam, &addFunctionRunRecordresp)
			So(err, ShouldBeNil)
			So(addFunctionRunRecordresp.Code, ShouldEqual, http.StatusOK)
			So(addFunctionRunRecordresp.FunctionRunRecord.ID, ShouldEqual, addFunctionRunRecordID)
			So(addFunctionRunRecordresp.FunctionRunRecord.End.IsZero(), ShouldBeFalse)
			So(addFunctionRunRecordresp.FunctionRunRecord.Suc, ShouldBeTrue)

			// as add function suc reported finished. it's next layer's function multiply should be triggered
			time.Sleep(time.Second)
			_, err = http_util.Get(
				superuserHeader(),
				serverAddress+"/api/v1/flow_run_record/",
				map[string]string{"id": theFlowRunRecord.ID.String()},
				&flowRunRecordResp)
			So(err, ShouldBeNil)
			So(flowRunRecordResp.Data.Total, ShouldEqual, 1)

			theFlowRunRecord = flowRunRecordResp.Data.Items[0]
			So(theFlowRunRecord.FlowID, ShouldEqual, onlineFlow.ID)
			So(theFlowRunRecord.FlowOriginID, ShouldEqual, onlineFlow.OriginID)
			// 2 stand for the second function is suc published
			So(len(theFlowRunRecord.FlowFuncIDMapFuncRunRecordID), ShouldEqual, 2)
			So(theFlowRunRecord.Status, ShouldEqual, value_object.Running)
			So(theFlowRunRecord.EndTime, ShouldBeNil)

			theSecondMultiplyFunctionRunRecordID, ok := theFlowRunRecord.FlowFuncIDMapFuncRunRecordID[aggFuncMultiplyFlowFunctionID]
			So(ok, ShouldBeTrue)

			// get multiply's function record to check it's suc created and data is right
			multiplyFunctionRunRecordresp := struct {
				web.RespMsg
				FunctionRunRecord *function_run_record.FunctionRunRecord `json:"data"`
			}{}
			_, err = http_util.Get(
				superuserHeader(),
				serverAddress+"/api/v1/function_run_record/get_by_id/"+theSecondMultiplyFunctionRunRecordID.String(),
				http_util.BlankGetParam, &multiplyFunctionRunRecordresp)
			So(err, ShouldBeNil)
			So(multiplyFunctionRunRecordresp.Code, ShouldEqual, http.StatusOK)
			So(multiplyFunctionRunRecordresp.FunctionRunRecord.ID, ShouldEqual, theSecondMultiplyFunctionRunRecordID)
			So(multiplyFunctionRunRecordresp.FunctionRunRecord.End, ShouldBeNil)
			So(multiplyFunctionRunRecordresp.FunctionRunRecord.Suc, ShouldBeFalse)

			// upload multiply-function run opt
			multiplyFunctionOptSumFieldPersistReq := &client.PersistFuncRunOptFieldHttpReq{
				FunctionRunRecordID: theSecondMultiplyFunctionRunRecordID,
				OptKey:              "result",
				Data:                128,
			}
			multiplyFunctionOptSumFieldPersistBody, _ := json.Marshal(multiplyFunctionOptSumFieldPersistReq)
			multiplyFunctionOptSumFieldPersistResp := struct {
				web.RespMsg
				Data *client.PersistFuncRunOptFieldHttpResp `json:"data"`
			}{}
			_, err = http_util.Post(
				superuserHeader(),
				serverAddress+"/api/v1/client/persist_certain_function_run_opt_field",
				http_util.BlankGetParam, multiplyFunctionOptSumFieldPersistBody,
				&multiplyFunctionOptSumFieldPersistResp)
			So(err, ShouldBeNil)
			So(multiplyFunctionOptSumFieldPersistResp.Code, ShouldEqual, http.StatusOK)

			// report multiply function run finished
			reportFinished = &client.FuncRunFinishedHttpReq{
				FunctionRunRecordID:       theSecondMultiplyFunctionRunRecordID.String(),
				Suc:                       true,
				OptKeyMapBriefData:        map[string]string{"result": multiplyFunctionOptSumFieldPersistResp.Data.Brief},
				OptKeyMapObjectStorageKey: map[string]string{"result": multiplyFunctionOptSumFieldPersistResp.Data.ObjectStorageKey},
			}
			reportFinishedBody, _ = json.Marshal(reportFinished)
			_, err = http_util.Post(
				superuserHeader(),
				serverAddress+"/api/v1/client/function_run_finished",
				http_util.BlankGetParam, reportFinishedBody, &reportFinishedResp)
			So(err, ShouldBeNil)
			So(reportFinishedResp.Code, ShouldEqual, http.StatusOK)

			// as all the function are reported finished. the flow_run_record should be marked as finished
			time.Sleep(time.Second)
			http_util.Get(
				superuserHeader(),
				serverAddress+"/api/v1/flow_run_record/",
				map[string]string{"id": theFlowRunRecord.ID.String()},
				&flowRunRecordResp)
			theFlowRunRecord = flowRunRecordResp.Data.Items[0]
			So(theFlowRunRecord.Status, ShouldEqual, value_object.Suc)
			So(theFlowRunRecord.EndTime.IsZero(), ShouldBeFalse)
		})
	})
}
