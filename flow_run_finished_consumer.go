package bloc

import (
	"github.com/fBloc/bloc-server/event"
	"github.com/fBloc/bloc-server/value_object"
)

// FlowTaskFinishedConsumer receive flow run finished event
func (blocApp *BlocApp) FlowTaskFinishedConsumer() {
	event.InjectMq(blocApp.GetOrCreateEventMQ())
	logger := blocApp.GetOrCreateScheduleLogger()
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
		logTag := map[string]string{
			"business":           "flow run finished consumer",
			"flow_run_record_id": flowRunRecordStr}

		logger.Infof(
			logTag,
			"FlowRunFinishedConsumer flow_run_record__id %s finished", flowRunRecordStr)
		flowRunRecordUuid, err := value_object.ParseToUUID(flowRunRecordStr)
		if err != nil {
			logger.Errorf(
				logTag, "cannot parse identity:%s to uuid! error:%v",
				flowRunRecordStr, err)
			continue
		}
		flowRunIns, err := flowRunRepo.GetByID(flowRunRecordUuid)
		if err != nil {
			logger.Errorf(logTag, "flow_run_record_id find no record!")
			continue
		}

		// 更新此flow_run_record的状态为成功
		err = flowRunRepo.Suc(flowRunIns.ID)
		if err != nil {
			logger.Errorf(logTag, "save suc of flowRunRecord failed: %v", err)
		} else {
			logger.Infof(logTag, "finished")
		}
	}
}
