package client

import (
	"encoding/json"
	"fmt"
	"net/http"
	"sync"
	"time"

	"github.com/fBloc/bloc-server/aggregate"
	"github.com/fBloc/bloc-server/config"
	"github.com/fBloc/bloc-server/event"
	"github.com/fBloc/bloc-server/interfaces/web"
	"github.com/fBloc/bloc-server/value_object"
	"github.com/julienschmidt/httprouter"
)

func FunctionRunStart(w http.ResponseWriter, r *http.Request, _ httprouter.Params) {
	logTags := web.GetTraceAboutFields(r.Context())
	logTags["business"] = "function_run_start report api"
	scheduleLogger.Infof(logTags, "start")

	var req FuncRunStartHttpReq
	err := json.NewDecoder(r.Body).Decode(&req)
	if err != nil {
		scheduleLogger.Warningf(logTags, "unmarshal body failed: %v", err)
		web.WriteBadRequestDataResp(&w, r, err.Error())
		return
	}
	logTags["function_run_record_id"] = req.FunctionRunRecordID

	funcRunRecordUUID, err := value_object.ParseToUUID(req.FunctionRunRecordID)
	if err != nil {
		scheduleLogger.Warningf(logTags, "parse to uuid failed: %v", err)
		web.WriteBadRequestDataResp(&w, r, "parse function_id to uuid failed: %v", err)
		return
	}

	err = fRRService.FunctionRunRecords.SaveStart(funcRunRecordUUID)
	if err != nil {
		scheduleLogger.Errorf(logTags, "save function_run_record start failed: %v", err)
		web.WriteInternalServerErrorResp(&w, r, err, "update function_run_record start failed")
		return
	}

	scheduleLogger.Infof(logTags, "finished")
	web.WritePlainSucOkResp(&w, r)
}

func FunctionRunFinished(w http.ResponseWriter, r *http.Request, _ httprouter.Params) {
	logTags := web.GetTraceAboutFields(r.Context())
	logTags["business"] = "function_run_finshed report api"
	scheduleLogger.Infof(logTags, "start")

	var req FuncRunFinishedHttpReq
	err := json.NewDecoder(r.Body).Decode(&req)
	if err != nil {
		scheduleLogger.Warningf(logTags, "unmarshal body failed: %v", err)
		web.WriteBadRequestDataResp(&w, r, err.Error())
		return
	}
	logTags["function_run_record_id"] = req.FunctionRunRecordID
	scheduleLogger.Infof(logTags, `status of whether suc: %t`, req.Suc)

	funcRunRecordUUID, err := value_object.ParseToUUID(req.FunctionRunRecordID)
	if err != nil {
		scheduleLogger.Warningf(logTags, "parse to uuid failed: %v", err)
		web.WriteBadRequestDataResp(&w, r, "parse function_id to uuid failed: %v", err)
		return
	}

	// 兜底逻辑、删除此function_run_record_id对应的heartbeat。理论上应该通过明确的删除接口来完成
	heartBeatDeleteAmount, err := heartbeatService.HeartBeatRepo.DeleteByFunctionRunRecordID(funcRunRecordUUID)
	if err != nil {
		scheduleLogger.Errorf(logTags,
			"delete heartbeat failed: %v", err)
	}
	scheduleLogger.Infof(logTags,
		"delete heartbeat amount: %d", heartBeatDeleteAmount)

	fRRIns, err := fRRService.FunctionRunRecords.GetByID(funcRunRecordUUID)
	if err != nil {
		scheduleLogger.Errorf(logTags, "get functionRunRecord by id failed: %v", err)
		web.WriteInternalServerErrorResp(&w, r, err, "find function_run_record_ins by it's id failed")
		return
	}
	if fRRIns.IsZero() {
		scheduleLogger.Warningf(logTags, "get functionRunRecord by id match no record")
		web.WriteBadRequestDataResp(&w, r, "find no function_run_record_ins by this function_id")
		return
	}

	flowIns, err := flowService.Flow.GetByID(fRRIns.FlowID)
	if err != nil {
		scheduleLogger.Errorf(logTags, "get flowIns by id failed: %v", err)
		web.WriteInternalServerErrorResp(&w, r, err, "find flow_ins failed")
		return
	}
	if flowIns.IsZero() {
		scheduleLogger.Warningf(logTags, "get flowIns by id match no record")
		web.WriteBadRequestDataResp(&w, r, "find no flow_ins")
		return
	}
	logTags["flow_id"] = fRRIns.FlowID.String()

	flowRunRecordIns, err := flowRunRecordService.FlowRunRecord.GetByID(fRRIns.FlowRunRecordID)
	if err != nil {
		scheduleLogger.Errorf(logTags, "get flowRunRecordIns by id failed: %v", err)
		web.WriteInternalServerErrorResp(&w, r, err, "find flow_run_record_ins failed")
		return
	}
	if flowRunRecordIns.IsZero() {
		scheduleLogger.Warningf(logTags, "get flowRunRecordIns by id match no record")
		web.WriteBadRequestDataResp(&w, r, "find no flow_run_record_ins")
		return
	}
	traceCtx := value_object.SetTraceIDToContext(flowRunRecordIns.TraceID)
	logTags["flow_run_record_id"] = fRRIns.FlowRunRecordID.String()

	functionIns := reported.idMapFunc[fRRIns.FunctionID]
	if req.Suc {
		err := fRRService.FunctionRunRecords.SaveSuc(
			funcRunRecordUUID, req.Description,
			functionIns.OptKeyMapValueType(),
			functionIns.OptKeyMapIsArray(),
			req.OptKeyMapObjectStorageKey,
			req.OptKeyMapBriefData,
			req.InterceptBelowFunctionRun,
		)
		if err != nil {
			scheduleLogger.Errorf(logTags, "function_run_record save suc failed: %v", err)
			// 如果保存运行成功失败，由于下游运行前会检测上游是否成功，如果不成功就不会运行
			// 故就算发布了下游也不会实际运行。所以应该这里就保持拦截

			// flowRunRecord保存为失败
			err = flowRunRecordService.FlowRunRecord.Fail(
				flowRunRecordIns.ID, "function_run_record save suc failed")
			if err != nil {
				scheduleLogger.Errorf(logTags,
					"save flow_run_record run fail failed: %v", err)
			}

			// 直接返回
			web.WriteInternalServerErrorResp(&w, r, err, "function_run_record save suc failed")
			return
		}

		// 检测其下的所有下游function是否活跃、如果有不活跃的直接停止flow继续运行
		handledFunctionIDMap := make(map[value_object.UUID]bool)
		var iterDowns func(flowFunctionID string)
		iterDowns = func(flowFunctionID string) {
			for _, downFlowFunctionID := range flowIns.FlowFunctionIDMapFlowFunction[flowFunctionID].DownstreamFlowFunctionIDs {
				handledFunctionIDMap[flowIns.FlowFunctionIDMapFlowFunction[downFlowFunctionID].FunctionID] = true
				iterDowns(downFlowFunctionID)
			}
		}
		var checkAllFuncAliveMutex sync.WaitGroup
		checkAllFuncAliveMutex.Add(len(handledFunctionIDMap))
		notAliveFuncs := make([]*aggregate.Function, 0, 2)
		var notAliveFuncMutex sync.Mutex
		for functionID := range handledFunctionIDMap {
			go func(
				functionID value_object.UUID,
				wg *sync.WaitGroup,
			) {
				defer wg.Done()
				functionIns, err := fService.Function.GetByIDForCheckAliveTime(functionID)
				if err != nil {
					return
				}
				if time.Since(functionIns.LastAliveTime) > config.FunctionReportTimeout {
					notAliveFuncMutex.Lock()
					defer notAliveFuncMutex.Unlock()
					notAliveFuncs = append(notAliveFuncs, functionIns)
					return
				}
			}(functionID, &checkAllFuncAliveMutex)
		}
		checkAllFuncAliveMutex.Wait()
		if len(notAliveFuncs) > 0 {
			errorMsg := fmt.Sprintf("have %d functions dead. wont run!", len(notAliveFuncs))
			for _, i := range notAliveFuncs {
				scheduleLogger.Warningf(logTags,
					"function id-%s name-%s provider-%s is dead. last alive time %v",
					i.ID.String(), i.Name, i.ProviderName, i.LastAliveTime)
			}
			err = flowService.FlowRunRecord.FunctionDead(fRRIns.FlowRunRecordID, errorMsg)
			if err != nil {
				scheduleLogger.Errorf(logTags, "save flowRunRepo.FunctionDead failed: %v", err)
			}
			goto Final
		}
		scheduleLogger.Infof(logTags, "all downs functions are alive")

		flowFunction := flowIns.FlowFunctionIDMapFlowFunction[fRRIns.FlowFunctionID]
		if !req.InterceptBelowFunctionRun { // 成功运行完成且不拦截
			if len(flowFunction.DownstreamFlowFunctionIDs) > 0 { // 若有下游待运行的function节点
				// 创建并发布下游function节点
				for _, downStreamFlowFunctionID := range flowFunction.DownstreamFlowFunctionIDs {
					downStreamFlowFunction := flowIns.FlowFunctionIDMapFlowFunction[downStreamFlowFunctionID]
					downStreamFunctionIns := reported.idMapFunc[downStreamFlowFunction.FunctionID]

					downStreamFunctionRunRecord := aggregate.NewFunctionRunRecordFromFlowDriven(
						traceCtx, downStreamFunctionIns, *flowRunRecordIns, downStreamFlowFunctionID)
					downStreamFunctionRunRecordMsg := fmt.Sprintf(
						`function_id: %s, function_run_record_id: %s.`,
						downStreamFunctionRunRecord.FunctionID.String(), downStreamFunctionRunRecord.ID.String())

					err = fRRService.FunctionRunRecords.Create(downStreamFunctionRunRecord)
					if err != nil {
						scheduleLogger.Errorf(logTags,
							"create downstream funcRunRecord to repository failed. downstream function info: %s, err: %v",
							downStreamFunctionRunRecordMsg, err)
						goto Final
					}

					err = flowRunRecordService.FlowRunRecord.AddFlowFuncIDMapFuncRunRecordID(
						flowRunRecordIns.ID, downStreamFlowFunctionID, downStreamFunctionRunRecord.ID)
					if err != nil {
						scheduleLogger.Errorf(logTags,
							"add downstream funcRunRecord to flow_run_record failed: %v. downstream function info: %s",
							err, downStreamFunctionRunRecordMsg)
						goto Final
					}
					flowRunRecordIns.FlowFuncIDMapFuncRunRecordID[downStreamFlowFunctionID] = downStreamFunctionRunRecord.ID

					err = event.PubEvent(&event.FunctionToRun{FunctionRunRecordID: downStreamFunctionRunRecord.ID})
					if err != nil {
						scheduleLogger.Errorf(logTags,
							`pub downstream function to run event failed. downstream func info: %s. error: %v`,
							downStreamFunctionRunRecordMsg, err)

						err = flowRunRecordService.FlowRunRecord.Fail(
							flowRunRecordIns.ID, "pub downstream run event failed")
						if err != nil {
							scheduleLogger.Errorf(logTags, "save flow_run_record run fail failed: %v", err)
						}
						goto Final
					}
				}
			} else { // 此函数节点没有下游了。进行检查flow是否全部运行成功从而完成了，
				goto CheckWhetherFlowRunFinished
			}
		} else { // 运行拦截，此function节点以下的节点不用再运行了，此步骤拦截
			goto CheckWhetherFlowRunFinished
		}
	CheckWhetherFlowRunFinished:
		/*
			检测要点：
				因为 FlowFunctionIDMapFlowFunction 有此flow每个function运行历史的对应记录
				从而检查是不是都有效完成了
		*/

		var checkFlowRunWhetherFinished func(flowRunRecordID string, finished *bool, flag string)
		checkFlowRunWhetherFinished = func(flowFunctionID string, finished *bool, flag string) {
			if !*finished {
				return
			}
			interceptBelow := false
			if flowFunctionID != config.FlowFunctionStartID {
				funcRunRecordID, ok := flowRunRecordIns.FlowFuncIDMapFuncRunRecordID[flowFunctionID]
				if !ok { // 表示此flow_function还没有运行记录
					*finished = false
					return
				}
				functionRunRecordIns, err := fRRService.FunctionRunRecords.GetByID(funcRunRecordID)
				if err != nil {
					scheduleLogger.Errorf(logTags,
						"get function_run_record by function_run_record_id(%s) failed: %s",
						funcRunRecordID.String(), err.Error())
					// 先保守处理为未完成运行
					*finished = false
					return
				}
				if functionRunRecordIns.InterceptBelowFunctionRun {
					interceptBelow = true
				}
				if !functionRunRecordIns.Finished() {
					*finished = false
					return
				}
			}
			if !interceptBelow {
				flowFunc := flowIns.FlowFunctionIDMapFlowFunction[flowFunctionID]
				for _, downstreamFlowFuncID := range flowFunc.DownstreamFlowFunctionIDs {
					checkFlowRunWhetherFinished(downstreamFlowFuncID, finished, flag)
				}
			}
		}
		finished := true
		checkFlowRunWhetherFinished(config.FlowFunctionStartID, &finished, functionIns.Name)
		if finished {
			err = flowRunRecordService.FlowRunRecord.Suc(flowRunRecordIns.ID)
			if err != nil {
				scheduleLogger.Errorf(logTags, "save flow_run_record suc failed: %v", err)
			}
		}
		goto Final
	} else { // function节点运行失败, 处理有重试的情况
		// 无重试策略
		if !flowIns.HaveRetryStrategy() || flowRunRecordIns.RetriedAmount >= flowIns.RetryAmount {
			scheduleLogger.Infof(logTags,
				"no retry set in flow. just save flow_run_record as failed")

			err = fRRService.FunctionRunRecords.SaveFail(funcRunRecordUUID, req.ErrorMsg)
			if err != nil {
				scheduleLogger.Errorf(logTags,
					"save function_run_record run fail failed: %v", err)
			}

			err = flowRunRecordService.FlowRunRecord.Fail(flowRunRecordIns.ID, "have function failed")
			if err != nil {
				scheduleLogger.Errorf(logTags,
					"save flow_run_record run fail failed: %v", err)
			}
		} else { // 有重试策略
			scheduleLogger.Infof(logTags, "retry")

			flowRunRecordService.FlowRunRecord.PatchDataForRetry(
				flowRunRecordIns.ID, flowRunRecordIns.RetriedAmount)

			retryGapSeconds := 3 // default set as 3 second
			if flowIns.RetryIntervalInSecond > 0 {
				retryGapSeconds = int(flowIns.RetryIntervalInSecond)
			}
			futureTime := time.Now().Add(time.Duration(retryGapSeconds) * time.Second)
			scheduleLogger.Infof(logTags,
				"the retry event will be executed after: %s", futureTime.Format(time.RFC3339))

			err = event.PubEventAtCertainTime(
				&event.FunctionToRun{FunctionRunRecordID: funcRunRecordUUID},
				futureTime)
			if err != nil {
				scheduleLogger.Errorf(logTags, "pub event at certain time failed: %v", err)
			}
		}
	}

Final:
	scheduleLogger.Infof(logTags, "finished")
	web.WritePlainSucOkResp(&w, r)
}
