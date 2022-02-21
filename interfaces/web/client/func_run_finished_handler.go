package client

import (
	"encoding/json"
	"fmt"
	"net/http"
	"time"

	"github.com/fBloc/bloc-server/aggregate"
	"github.com/fBloc/bloc-server/config"
	"github.com/fBloc/bloc-server/event"
	"github.com/fBloc/bloc-server/interfaces/web"
	"github.com/fBloc/bloc-server/value_object"
	"github.com/julienschmidt/httprouter"
)

func FunctionRunFinished(w http.ResponseWriter, r *http.Request, _ httprouter.Params) {
	var req FuncRunFinishedHttpReq
	err := json.NewDecoder(r.Body).Decode(&req)
	if err != nil {
		web.WriteBadRequestDataResp(&w, err.Error())
		return
	}

	fRRService.Logger.Infof(
		map[string]string{"function_run_record_id": req.FunctionRunRecordID},
		`received function_run_finished report.function_run_record id: %s, suc: %t`,
		req.FunctionRunRecordID, req.Suc)
	funcRunRecordUUID, err := value_object.ParseToUUID(req.FunctionRunRecordID)
	if err != nil {
		web.WriteBadRequestDataResp(&w, "parse function_id to uuid failed: %v", err)
		return
	}

	fRRIns, err := fRRService.FunctionRunRecords.GetByID(funcRunRecordUUID)
	if err != nil {
		web.WriteInternalServerErrorResp(&w, err, "find function_run_record_ins by it's id failed")
		return
	}
	if fRRIns.IsZero() {
		web.WriteBadRequestDataResp(&w, "find no function_run_record_ins by this function_id")
		return
	}

	flowIns, err := flowService.Flow.GetByID(fRRIns.FlowID)
	if err != nil {
		web.WriteInternalServerErrorResp(&w, err, "find flow_ins failed")
		return
	}
	if flowIns.IsZero() {
		web.WriteBadRequestDataResp(&w, "find no flow_ins")
		return
	}

	flowRunRecordIns, err := flowRunRecordService.FlowRunRecord.GetByID(fRRIns.FlowRunRecordID)
	if err != nil {
		web.WriteInternalServerErrorResp(&w, err, "find flow_run_record_ins failed")
		return
	}
	if flowRunRecordIns.IsZero() {
		web.WriteBadRequestDataResp(&w, "find no flow_run_record_ins")
		return
	}

	functionIns := reported.idMapFunc[fRRIns.FunctionID]
	if req.Suc {
		fRRService.FunctionRunRecords.SaveSuc(
			funcRunRecordUUID, req.Description,
			functionIns.OptKeyMapValueType(),
			functionIns.OptKeyMapIsArray(),
			req.OptKeyMapObjectStorageKey,
			req.OptKeyMapBriefData,
			req.InterceptBelowFunctionRun,
		)

		flowFunction := flowIns.FlowFunctionIDMapFlowFunction[fRRIns.FlowFunctionID]
		if !req.InterceptBelowFunctionRun { // 成功运行完成且不拦截
			if len(flowFunction.DownstreamFlowFunctionIDs) > 0 { // 若有下游待运行的function节点
				// 创建并发布下游function节点
				for _, downStreamFlowFunctionID := range flowFunction.DownstreamFlowFunctionIDs {
					downStreamFlowFunction := flowIns.FlowFunctionIDMapFlowFunction[downStreamFlowFunctionID]
					downStreamFunctionIns := reported.idMapFunc[downStreamFlowFunction.FunctionID]

					downStreamFunctionRunRecord := aggregate.NewFunctionRunRecordFromFlowDriven(
						downStreamFunctionIns, *flowRunRecordIns, downStreamFlowFunctionID)
					err = event.PubEvent(&event.FunctionToRun{FunctionRunRecordID: downStreamFunctionRunRecord.ID})
					if err != nil {
						consumerLogger.Errorf(
							map[string]string{"flow_run_record_id": flowRunRecordIns.ID.String()},
							`pub function to run event failed. flow_run_record_id: %s, function_id: %s, function_run_record_id: %s. error: %v`,
							flowRunRecordIns.ID.String(), downStreamFunctionRunRecord.FunctionID.String(),
							downStreamFunctionRunRecord.ID.String(), err)
					}

					_ = fRRService.FunctionRunRecords.Create(downStreamFunctionRunRecord)

					err = flowRunRecordService.FlowRunRecord.AddFlowFuncIDMapFuncRunRecordID(
						flowRunRecordIns.ID, downStreamFlowFunctionID, downStreamFunctionRunRecord.ID)
					if err != nil {
						consumerLogger.Errorf(
							map[string]string{"flow_run_record_id": flowRunRecordIns.ID.String()},
							`flowRunRecordRepo.AddFlowFuncIDMapFuncRunRecordID error: %v.flow_run_record_id:%s`,
							flowRunRecordIns.ID.String(), err,
						)
						err := flowRunRecordService.FlowRunRecord.Fail(
							flowRunRecordIns.ID,
							fmt.Sprintf(
								"failed because of AddFlowFuncIDMapFuncRunRecordID to repository error: %v",
								err))
						if err != nil {
							consumerLogger.Errorf(
								map[string]string{"flow_run_record_id": flowRunRecordIns.ID.String()},
								"FlowRunRecord.Fail to repository failed")
							break
						}
					}
					flowRunRecordIns.FlowFuncIDMapFuncRunRecordID[downStreamFlowFunctionID] = downStreamFunctionRunRecord.ID
				}
			} else { // 此函数节点没有下游了。进行检查flow是否全部运行成功从而完成了，
				/*
					检测要点：
						因为 FlowFunctionIDMapFlowFunction 有此flow每个function运行历史的对应记录
						从而检查是不是都有效完成了
				*/
				toCheckFlowFunctionIDMapDone := make(map[string]bool, len(flowIns.FlowFunctionIDMapFlowFunction))
				for flowFunctionID := range flowIns.FlowFunctionIDMapFlowFunction {
					toCheckFlowFunctionIDMapDone[flowFunctionID] = false
				}
				delete(toCheckFlowFunctionIDMapDone, config.FlowFunctionStartID)
				flowFinished := true
				for toCheckFlowFunctionID := range toCheckFlowFunctionIDMapDone {
					funcRunRecordID, ok := flowRunRecordIns.FlowFuncIDMapFuncRunRecordID[toCheckFlowFunctionID]
					if !ok { // 表示此flow_function还没有运行记录
						flowFinished = false
						break
					}
					functionRunRecordIns, err := fRRService.FunctionRunRecords.GetByID(funcRunRecordID)
					if err != nil {
						consumerLogger.Errorf(
							map[string]string{"function_run_record_id": funcRunRecordID.String()},
							"get function_run_record by function_run_record_id(%s) error: %s",
							funcRunRecordID.String(), err.Error())
						// 先保守处理为未完成运行
						flowFinished = false
						break
					}
					if !functionRunRecordIns.Finished() {
						flowFinished = false
						break
					}
				}
				// 已检测到全部完成
				if flowFinished {
					flowRunRecordService.FlowRunRecord.Suc(flowRunRecordIns.ID)
					event.PubEvent(&event.FlowRunFinished{
						FlowRunRecordID: flowRunRecordIns.ID,
					})
					consumerLogger.Infof(
						map[string]string{"flow_run_record_id": flowRunRecordIns.ID.String()},
						"pub finished flow_task__id from all finished: %s",
						flowRunRecordIns.ID)
				}
			}
		} else { // 运行拦截，此function节点以下的节点不用再运行了，此步骤拦截
			flowRunRecordService.FlowRunRecord.Intercepted(
				flowRunRecordIns.ID,
				fmt.Sprintf(
					"intercepted by function: %s-%s",
					functionIns.GroupName, functionIns.Name))
			event.PubEvent(&event.FlowRunFinished{
				FlowRunRecordID: flowRunRecordIns.ID,
			})
			consumerLogger.Infof(
				map[string]string{"flow_run_record_id": flowRunRecordIns.ID.String()},
				"pub finished flow_run_record__id from intercepted: %s",
				flowRunRecordIns.ID)
		}
	} else { // function节点运行失败, 处理有重试的情况
		// 无重试策略
		if !flowIns.HaveRetryStrategy() || flowRunRecordIns.RetriedAmount >= flowIns.RetryAmount {
			fRRService.FunctionRunRecords.SaveFail(funcRunRecordUUID, req.ErrorMsg)
			flowRunRecordService.FlowRunRecord.Fail(flowRunRecordIns.ID, "have function failed")
		} else { // 有重试策略
			flowRunRecordService.FlowRunRecord.PatchDataForRetry(
				flowRunRecordIns.ID, flowRunRecordIns.RetriedAmount)

			retryGapSeconds := 3
			if flowIns.RetryIntervalInSecond > 0 {
				retryGapSeconds = int(flowIns.RetryIntervalInSecond)
			}
			futureTime := time.Now().Add(time.Duration(retryGapSeconds) * time.Second)
			event.PubEventAtCertainTime(
				&event.FunctionToRun{
					FunctionRunRecordID: funcRunRecordUUID},
				futureTime)
		}
	}
	fRRService.Logger.Infof(
		map[string]string{"function_run_record_id": req.FunctionRunRecordID},
		`function_run_finished report finshed.function_run_record id: %s`,
		req.FunctionRunRecordID)
	web.WritePlainSucOkResp(&w)
}
