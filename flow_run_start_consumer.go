package bloc

import (
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
		logTags := map[string]string{"flow_run_record_id": flowRunRecordStr}
		logger.Infof(logTags,
			"get flow run start record id %s", flowRunRecordStr)

		flowRunRecordUuid, err := value_object.ParseToUUID(flowRunRecordStr)
		if err != nil {
			logger.Errorf(logTags, "parse to uuid failed: $v", err)
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
			pubFuncLogTags["function_id"] = flowFunction.FunctionID.String()
			if functionIns.IsZero() {
				// TODO 处理function注册的心跳过期问题
				logger.Errorf(pubFuncLogTags, "find no function from blocApp")
				goto PubFailed
			}

			aggFunctionRunRecord := aggregate.NewFunctionRunRecordFromFlowDriven(
				traceCtx, *functionIns, *flowRunIns, flowFunctionID)
			err = event.PubEvent(&event.FunctionToRun{FunctionRunRecordID: aggFunctionRunRecord.ID})
			if err != nil {
				logger.Errorf(pubFuncLogTags, "pub FunctionToRun event failed: %v", err)
				goto PubFailed
			}
			err = functionRunRecordRepo.Create(aggFunctionRunRecord)
			if err != nil {
				logger.Errorf(
					map[string]string{
						"flow_run_record_id": flowRunRecordStr,
						"flow_id":            flowRunIns.FlowID.String(),
						"function_id":        flowFunction.FunctionID.String(),
					},
					"create flow's first layer function_run_record failed. function_id: %s, err: %v",
					flowFunction.FunctionID.String(), err)
				goto PubFailed
			}
			flowblocidMapBlochisid[flowFunctionID] = aggFunctionRunRecord.ID
		}
		flowRunIns.FlowFuncIDMapFuncRunRecordID = flowblocidMapBlochisid
		err = flowRunRepo.PatchFlowFuncIDMapFuncRunRecordID(
			flowRunIns.ID, flowRunIns.FlowFuncIDMapFuncRunRecordID)
		if err != nil {
			logger.Errorf(
				map[string]string{"flow_run_record_id": flowRunRecordStr},
				"update flow_run_record's flowFuncID_map_funcRunRecordID field failed: %s",
				err.Error())
			goto PubFailed
		}
		flowRunRepo.Start(flowRunIns.ID)
		continue
	PubFailed:
		flowRunRepo.Fail(flowRunIns.ID, "pub flow's first lay functions failed")
	}
}
