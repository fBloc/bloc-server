package bloc

import (
	"time"

	"github.com/fBloc/bloc-server/aggregate"
	"github.com/fBloc/bloc-server/event"
	"github.com/fBloc/bloc-server/value_object"
)

// CrontabWatcher 分钟接别的观测配置了crontab的flow并进行发起
func (blocApp *BlocApp) CrontabWatcher() {
	flowRepo := blocApp.GetOrCreateFlowRepository()
	flowRunRecordRepo := blocApp.GetOrCreateFlowRunRecordRepository()
	logger := blocApp.GetOrCreateScheduleLogger()

	ticker := time.NewTicker(time.Minute)
	defer ticker.Stop()
	for range ticker.C {
		crontabFlows, err := flowRepo.FilterCrontabFlows()
		if err != nil {
			logger.Errorf(
				map[string]string{},
				"Error on filter crontab flows: %s", err.Error())
			continue
		}

		now := time.Now()
		for _, flowIns := range crontabFlows {
			// ToEnhance 池化，避免要发布的太多、一下子起的goroutine太多。（优先级-低）
			go func(flowIns aggregate.Flow, crontabTrigTime time.Time) {
				// 时间若不符合crontab规则，则不发布运行任务
				if !flowIns.Crontab.TimeMatched(now) {
					logger.Infof(
						map[string]string{},
						"flow: %s's crontab config: %s not match the time: %s",
						flowIns.Name, flowIns.Crontab.CrontabStr, now.Format(time.RFC3339))
					return
				}

				traceID := value_object.NewTraceID()
				logTags := map[string]string{
					string(value_object.TraceID): traceID,
					"business":                   "crontab publish flow to run",
					"flow_id":                    flowIns.ID.String()}
				logger.Infof(
					logTags,
					"flow:%s's crontab config: %s match time: %s",
					flowIns.Name, flowIns.Crontab.CrontabStr, now.Format(time.RFC3339))

				// 符合就发布运行任务
				ctx := value_object.SetTraceIDToContext(traceID)
				flowRunRecord := aggregate.NewCrontabTriggeredRunRecord(ctx, &flowIns)
				created, err := flowRunRecordRepo.CrontabFindOrCreate(flowRunRecord, crontabTrigTime)
				if err != nil {
					logger.Errorf(logTags, "create flow_run_record failed: %v", err)
					return
				}
				if created { // 并发安全、避免重复发布任务
					logger.Infof(logTags, "already created")
					return
				}

				err = event.PubEvent(&event.FlowToRun{FlowRunRecordID: flowRunRecord.ID})
				if err != nil {
					logger.Errorf(logTags, "pub flow to run event failed: %v", err)
				} else {
					logger.Infof(logTags, "suc pub flow_run_record. id: %s", flowRunRecord.ID.String())
				}
			}(flowIns, now)
		}
	}
}
