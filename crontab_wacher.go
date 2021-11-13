package bloc

import (
	"time"

	"github.com/fBloc/bloc-backend-go/aggregate"
	"github.com/fBloc/bloc-backend-go/event"
)

// CrontabWatcher 分钟接别的观测配置了crontab的flow并进行发起
func (blocApp *BlocApp) CrontabWatcher() {
	flowRepo := blocApp.GetOrCreateFlowRepository()
	flowRunRecordRepo := blocApp.GetOrCreateFlowRunRecordRepository()
	logger := blocApp.GetOrCreateConsumerLogger()

	ticker := time.NewTicker(time.Minute)
	defer ticker.Stop()
	for range ticker.C {
		crontabFlows, err := flowRepo.FilterCrontabFlows()
		if err != nil {
			logger.Errorf("Error on filter crontab flows:", err.Error())
			continue
		}

		now := time.Now()
		for _, flowIns := range crontabFlows {
			// ToEnhance 池化，避免要发布的太多、一下子起的goroutine太多。（优先级-低）
			go func(flowIns *aggregate.Flow, crontabTrigTime time.Time) {
				flowRunRecord := aggregate.NewCrontabTriggeredRunRecord(*flowIns)
				created, err := flowRunRecordRepo.CrontabFindOrCreate(flowRunRecord, crontabTrigTime)
				if err != nil {
					logger.Errorf("error create flow run record", err.Error())
					return
				}
				if created { // 并发安全、避免重复发布任务
					return
				}
				event.PubEvent(&event.FlowToRun{FlowRunRecordID: flowRunRecord.ID})
			}(&flowIns, now)
		}
	}
}
