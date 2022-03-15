package bloc

import (
	"time"

	"github.com/fBloc/bloc-server/aggregate"
	"github.com/fBloc/bloc-server/event"
	"github.com/fBloc/bloc-server/value_object"
)

// RePubDeadRuns 重发运行中断的任务
func (blocApp *BlocApp) RePubDeadRuns() {
	logger := blocApp.GetOrCreateScheduleLogger()
	heartBeatRepo := blocApp.GetOrCreateFuncRunHBeatRepository()
	funcRunRecordRepo := blocApp.GetOrCreateFunctionRunRecordRepository()

	ticker := time.NewTicker(aggregate.HeartBeatDeadThreshold)
	defer ticker.Stop()
	for range ticker.C {
		deads, err := heartBeatRepo.AllDeads(aggregate.HeartBeatDeadThreshold)
		if err != nil {
			logger.Errorf(
				map[string]string{},
				"heartbeat watcher error: %s", err.Error())
			continue
		}
		if len(deads) <= 0 {
			continue
		}

		for _, d := range deads {
			logTags := map[string]string{
				string(value_object.SpanID): value_object.NewSpanID(),
				"business":                  "heartbeat watcher",
				"function_run_record_id":    d.FunctionRunRecordID.String(),
			}

			// 查询对应的function run record是否存在
			funcRunRecord, err := funcRunRecordRepo.GetByID(d.FunctionRunRecordID)
			if err != nil {
				logger.Errorf(logTags,
					"get funcRunRecordRepo by id failed. : %v", err)
				continue
			}
			if funcRunRecord.IsZero() {
				logger.Warningf(logTags,
					"get funcRunRecordRepo by id match no record")
				continue
			}
			logTags[string(value_object.TraceID)] = funcRunRecord.TraceID
			logger.Infof(logTags, "start handle dead heatbeat")

			err = funcRunRecordRepo.ClearProgress(funcRunRecord.ID)
			if err != nil {
				logger.Errorf(logTags,
					"funcRunRecordRepo.ClearProgress failed: %v", err)
			}

			// 立即进行删除此条信息（利用mongo通过ID删除的原子性保障来确保不会「重复重发」）
			deleteAmount, err := heartBeatRepo.DeleteByFunctionRunRecordID(d.FunctionRunRecordID)
			if err != nil {
				logger.Errorf(logTags,
					"heartBeatRepo.Delete failed, error: %s", err.Error())
				continue
			}
			if deleteAmount != 1 { // 避免并发watch重复发布
				logger.Infof(logTags, "already repubed. breakout")
				continue
			}

			// 再次进行发布
			err = event.PubEvent(&event.FunctionToRun{
				FunctionRunRecordID: funcRunRecord.ID,
			})
			if err != nil {
				logger.Errorf(logTags,
					"pub func event failed: %v", err)
			} else {
				logger.Infof(logTags,
					"re-pub function run record suc")
			}
		}
	}
}
