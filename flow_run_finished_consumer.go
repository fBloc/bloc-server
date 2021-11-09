package bloc

import (
	"github.com/fBloc/bloc-backend-go/event"

	"github.com/google/uuid"
)

// FlowTaskConsumer 接收到flow完成的任务，若是arr_flow，继续发布下一层的arr_flow的任务
func (blocApp *BlocApp) FlowTaskFinishedConsumer() {
	event.InjectMq(blocApp.GetOrCreateEventMQ())
	logger := blocApp.GetOrCreateConsumerLogger()
	flowRunRepo := blocApp.GetOrCreateFlowRunRecordRepository()

	flowRunFinishedEventChan := make(chan event.DomainEvent)
	err := event.ListenEvent(
		&event.FlowRunFinished{}, "flow_run_finished_consumer",
		flowRunFinishedEventChan)
	if err != nil {
		panic(err)
	}

	for flowRunFinishedEvent := range flowRunFinishedEventChan {
		flowRunRecordStr := flowRunFinishedEvent.Identity()
		logger.Infof("> FlowRunFinishedConsumer flow_run_record__id %s finished", flowRunRecordStr)
		flowRunRecordUuid, err := uuid.Parse(flowRunRecordStr)
		if err != nil {
			// TODO 不应该panic
			panic(err)
		}
		flowRunIns, err := flowRunRepo.GetByID(flowRunRecordUuid)
		if err != nil {
			// TODO
			panic(err)
		}

		// 更新此flow_run_record的状态为成功
		flowRunRepo.Suc(flowRunIns.ID)

		// TODO 如果是arrangement出发的，需要发布下游的flow任务
	}
}
