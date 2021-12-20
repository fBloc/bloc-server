package bloc

import (
	"github.com/fBloc/bloc-server/aggregate"
	"github.com/fBloc/bloc-server/config"
	"github.com/fBloc/bloc-server/event"
	"github.com/fBloc/bloc-server/value_object"
)

func (blocApp *BlocApp) FlowTaskStartConsumer() {
	event.InjectMq(blocApp.GetOrCreateEventMQ())
	logger := blocApp.GetOrCreateConsumerLogger()
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
		logger.Infof("|--> get flow run start record id %s", flowRunRecordStr)
		flowRunRecordUuid, err := value_object.ParseToUUID(flowRunRecordStr)
		if err != nil {
			logger.Errorf(
				"parse flow_run_record_id to uuid failed: %s", flowRunRecordStr)
			continue
		}
		flowRunIns, err := flowRunRepo.GetByID(flowRunRecordUuid)
		if err != nil {
			logger.Errorf(
				"get flow_run_record by id. flow_run_record_id: %s", flowRunRecordStr)
			continue
		}
		if flowRunIns.Canceled {
			continue
		}

		flowIns, err := flowRepo.GetByID(flowRunIns.FlowID)
		if err != nil {
			logger.Errorf(
				"get flow from flow_run_record.flow_id error: %v", err)
			continue
		}
		if !flowIns.AllowParallelRun { // 若不允许同时运行，需要进行检测是不是有正在运行的
			// TODO 这里存在高并发的时候还是会并行运行的情况，后续需要处理
			isRunning, err := flowRunRepo.IsHaveRunningTask(
				flowIns.ID, flowRunRecordUuid)
			if err != nil {
				logger.Errorf(
					"filter running flow records error: %v", err)
				continue
			}
			if isRunning {
				logger.Infof(
					"this flow won't run because had been set not allowed parallel run")
				err = flowRunRepo.NotAllowedParallelRun(flowRunIns.ID)
				if err != nil {
					logger.Errorf(
						"save NotAllowedParallelRun of flow_id(%s) failed",
						flowRunIns.ID.String())
				}
				continue
			}
		}

		// 发布flow下的“第一层”functions任务
		firstLayerDownstreamFlowFunctionIDS := flowIns.FlowFunctionIDMapFlowFunction[config.FlowFunctionStartID].DownstreamFlowFunctionIDs
		flowblocidMapBlochisid := make(
			map[string]value_object.UUID, len(firstLayerDownstreamFlowFunctionIDS))
		for _, flowFunctionID := range firstLayerDownstreamFlowFunctionIDS {
			flowFunction := flowIns.FlowFunctionIDMapFlowFunction[flowFunctionID]

			functionIns := blocApp.GetFunctionByRepoID(flowFunction.FunctionID)
			if functionIns.IsZero() {
				logger.Errorf(
					"find flow's first layer function failed. function_id: %s",
					flowFunction.FunctionID.String())
				goto PubFailed
			}
			aggFunctionRunRecord := aggregate.NewFunctionRunRecordFromFlowDriven(
				*functionIns, *flowRunIns,
				flowFunctionID)
			err := functionRunRecordRepo.Create(aggFunctionRunRecord)
			if err != nil {
				logger.Errorf(
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
