package bloc

import (
	"fmt"
	"sync"
	"time"

	"github.com/fBloc/bloc-server/aggregate"
	"github.com/fBloc/bloc-server/config"
	"github.com/fBloc/bloc-server/event"
	"github.com/fBloc/bloc-server/value_object"
)

func (blocApp *BlocApp) FlowTaskStartConsumer() {
	event.InjectMq(blocApp.GetOrCreateEventMQ())
	logger := blocApp.GetOrCreateScheduleLogger()
	flowRunRepo := blocApp.GetOrCreateFlowRunRecordRepository()
	flowRepo := blocApp.GetOrCreateFlowRepository()
	functionRepo := blocApp.GetOrCreateFunctionRepository()
	functionRunRecordRepo := blocApp.GetOrCreateFunctionRunRecordRepository()

	flowToRunEventChan := make(chan event.DomainEvent)
	err := event.ListenEvent(
		&event.FlowToRun{}, "flow_to_run_consumer",
		flowToRunEventChan)
	if err != nil {
		panic(err)
	}

	for flowToRunEvent := range flowToRunEventChan {
		flowRunRecordStr := flowToRunEvent.Identity()
		logTags := map[string]string{
			string(value_object.SpanID): value_object.NewSpanID(),
			"business":                  "flow run start consumer",
			"flow_run_record_id":        flowRunRecordStr}
		logger.Infof(logTags,
			"get flow run start record id %s", flowRunRecordStr)

		flowRunRecordUuid, err := value_object.ParseToUUID(flowRunRecordStr)
		if err != nil {
			logger.Errorf(logTags, "parse to uuid failed: %v", err)
			continue
		}

		flowRunIns, err := flowRunRepo.GetByID(flowRunRecordUuid)
		if err != nil {
			logger.Errorf(logTags, "get flow_run_record by id failed: %v", err)
			continue
		}
		if flowRunIns.Canceled {
			logger.Infof(logTags, "flow already canceled")
			continue
		}
		logTags[string(value_object.TraceID)] = flowRunIns.TraceID
		logger.Infof(logTags, "received msg and suc get flow_run_record ins")
		if flowRunIns.Finished() {
			logger.Errorf(logTags, "flow already finished. actual should not into here!")
			continue
		}

		flowIns, err := flowRepo.GetByID(flowRunIns.FlowID)
		if err != nil {
			logger.Errorf(logTags,
				"get flow from flow_run_record.flow_id error: %v", err)
			continue
		}
		logTags["flow_id"] = flowRunIns.FlowID.String()
		if !flowIns.AllowParallelRun { // 若不允许同时运行，需要进行检测是不是有正在运行的
			// TODO 这里存在高并发的时候还是会并行运行的情况，后续需要处理
			isRunning, err := flowRunRepo.IsHaveRunningTask(
				flowIns.ID, flowRunRecordUuid)
			if err != nil {
				logger.Errorf(logTags,
					"filter running flow records error: %v", err)
				continue
			}
			if isRunning {
				logger.Infof(logTags, "won't run because not allowed parallel run")
				err = flowRunRepo.NotAllowedParallelRun(flowRunIns.ID)
				if err != nil {
					logger.Errorf(logTags, "save flowRunRepo.NotAllowedParallelRun failed: %v", err)
				}
				continue
			}
		}

		// 检测其下的functions的心跳是否存在过期的，如果存在就直接不要发布保存失败就是了
		var checkAllFuncAliveMutex sync.WaitGroup
		checkAllFuncAliveMutex.Add(len(flowIns.FlowFunctionIDMapFlowFunction) - 1)
		notAliveFuncs := make([]*aggregate.Function, 0, 2)
		var notAliveFuncMutex sync.Mutex
		for flowFunctionID, flowFunc := range flowIns.FlowFunctionIDMapFlowFunction {
			if flowFunctionID == config.FlowFunctionStartID {
				continue
			}
			go func(
				functionID value_object.UUID,
				wg *sync.WaitGroup,
			) {
				defer wg.Done()
				functionIns, err := functionRepo.GetByIDForCheckAliveTime(functionID)
				if err != nil {
					return
				}
				if time.Since(functionIns.LastAliveTime) > config.FunctionReportTimeout {
					notAliveFuncMutex.Lock()
					defer notAliveFuncMutex.Unlock()
					notAliveFuncs = append(notAliveFuncs, functionIns)
					return
				}
			}(flowFunc.FunctionID, &checkAllFuncAliveMutex)
		}
		checkAllFuncAliveMutex.Wait()
		if len(notAliveFuncs) > 0 {
			errorMsg := fmt.Sprintf("have %d functions dead. wont run!", len(notAliveFuncs))
			for _, i := range notAliveFuncs {
				logger.Warningf(logTags,
					"function id-%s name-%s provider-%s is dead. last alive time %v",
					i.ID.String(), i.Name, i.ProviderName, i.LastAliveTime)
			}
			err = flowRunRepo.FunctionDead(flowRunIns.ID, errorMsg)
			if err != nil {
				logger.Errorf(logTags, "save flowRunRepo.FunctionDead failed: %v", err)
			}
			continue
		}
		logger.Infof(logTags, "all downs functions are alive")

		// 发布flow下的“第一层”functions任务
		// ToEnhance 目前第一层的发布在这里，而后续的发布在client api，这种重复不好。应当解决
		firstLayerDownstreamFlowFunctionIDS := flowIns.FlowFunctionIDMapFlowFunction[config.FlowFunctionStartID].DownstreamFlowFunctionIDs
		flowblocidMapBlochisid := make(
			map[string]value_object.UUID, len(firstLayerDownstreamFlowFunctionIDS))
		pubFuncLogTags := logTags
		traceCtx := value_object.SetTraceIDToContext(flowRunIns.TraceID)
		for _, flowFunctionID := range firstLayerDownstreamFlowFunctionIDS {
			pubFuncLogTags["flow_function_id"] = flowFunctionID

			flowFunction := flowIns.FlowFunctionIDMapFlowFunction[flowFunctionID]
			functionIns := blocApp.GetFunctionByRepoID(flowFunction.FunctionID)
			pubFuncLogTags["downstream function_id"] = flowFunction.FunctionID.String()
			if functionIns.IsZero() {
				// TODO 处理function注册的心跳过期问题
				logger.Errorf(pubFuncLogTags, "find no function from blocApp")
				goto PubFailed
			}

			aggFunctionRunRecord := aggregate.NewFunctionRunRecordFromFlowDriven(
				traceCtx, *functionIns, *flowRunIns, flowFunctionID)
			err = functionRunRecordRepo.Create(aggFunctionRunRecord)
			if err != nil {
				logger.Errorf(pubFuncLogTags,
					"create flow's first layer function_run_record failed. function_id: %s, err: %v",
					flowFunction.FunctionID.String(), err)
				goto PubFailed
			}

			err = event.PubEvent(&event.FunctionToRun{FunctionRunRecordID: aggFunctionRunRecord.ID})
			pubFuncLogTags["downstream function_run_record_id"] = aggFunctionRunRecord.ID.String()
			if err != nil {
				logger.Errorf(pubFuncLogTags, "pub FunctionToRun event failed: %v", err)
				goto PubFailed
			}

			logger.Infof(pubFuncLogTags, "suc pubed downstream function to run event")
			flowblocidMapBlochisid[flowFunctionID] = aggFunctionRunRecord.ID
		}
		flowRunIns.FlowFuncIDMapFuncRunRecordID = flowblocidMapBlochisid
		err = flowRunRepo.PatchFlowFuncIDMapFuncRunRecordID(
			flowRunIns.ID, flowRunIns.FlowFuncIDMapFuncRunRecordID)
		if err != nil {
			logger.Errorf(logTags,
				"update flow_run_record's flowFuncID_map_funcRunRecordID field failed: %s",
				err.Error())
			goto PubFailed
		}
		err = flowRunRepo.Start(flowRunIns.ID)
		if err != nil {
			logger.Errorf(logTags,
				"flowRunRecord save start failed: %v", err.Error())
		} else {
			logger.Infof(logTags, "finished(suc)")
		}
		continue
	PubFailed:
		err = flowRunRepo.Fail(flowRunIns.ID, "pub flow's first lay functions failed")
		if err != nil {
			logger.Errorf(logTags,
				"flowRunRecord save failed(due to pub flow's first lay functions failed) failed: %v", err.Error())
		} else {
			logger.Infof(logTags, "finished(fail)")
		}
	}
}
