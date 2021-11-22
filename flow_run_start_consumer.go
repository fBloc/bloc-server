package bloc

import (
	"github.com/fBloc/bloc-backend-go/aggregate"
	"github.com/fBloc/bloc-backend-go/config"
	"github.com/fBloc/bloc-backend-go/event"
	"github.com/fBloc/bloc-backend-go/value_object"
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
			// TODO 不应该panic
			panic(err)
		}
		flowRunIns, err := flowRunRepo.GetByID(flowRunRecordUuid)
		if err != nil {
			// TODO
			panic(err)
		}
		if flowRunIns.Canceled {
			continue
		}

		flowIns, err := flowRepo.GetByID(flowRunIns.FlowID)
		if err != nil {
			// TODO
			panic(err)
		}

		// 发布flow下的“第一层”functions任务
		firstLayerDownstreamFlowFunctionIDS := flowIns.FlowFunctionIDMapFlowFunction[config.FlowFunctionStartID].DownstreamFlowFunctionIDs
		flowblocidMapBlochisid := make(
			map[string]value_object.UUID, len(firstLayerDownstreamFlowFunctionIDS))
		for _, flowFunctionID := range firstLayerDownstreamFlowFunctionIDS {
			flowFunction := flowIns.FlowFunctionIDMapFlowFunction[flowFunctionID]

			functionIns := blocApp.GetFunctionByRepoID(flowFunction.FunctionID)
			if functionIns.IsZero() {
				// TODO 不应该有此情况
			}

			aggFunctionRunRecord := aggregate.NewFunctionRunRecordFromFlowDriven(
				functionIns, *flowRunIns,
				flowFunctionID)
			err := functionRunRecordRepo.Create(aggFunctionRunRecord)
			if err != nil {
				// TODO
			}
			flowblocidMapBlochisid[flowFunctionID] = aggFunctionRunRecord.ID
		}
		flowRunIns.FlowFuncIDMapFuncRunRecordID = flowblocidMapBlochisid
		err = flowRunRepo.PatchFlowFuncIDMapFuncRunRecordID(
			flowRunIns.ID, flowRunIns.FlowFuncIDMapFuncRunRecordID)
		if err != nil {
			// TODO
		}
		flowRunRepo.Start(flowRunIns.ID)
	}
}
