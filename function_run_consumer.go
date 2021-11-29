package bloc

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"github.com/fBloc/bloc-backend-go/aggregate"
	"github.com/fBloc/bloc-backend-go/config"
	"github.com/fBloc/bloc-backend-go/event"
	"github.com/fBloc/bloc-backend-go/pkg/value_type"
	function_execute_heartbeat_repository "github.com/fBloc/bloc-backend-go/repository/function_execute_heartbeat"
	"github.com/fBloc/bloc-backend-go/value_object"
)

func reportHeartBeat(
	heartBeatRepo function_execute_heartbeat_repository.FunctionExecuteHeartbeatRepository,
	heartBeatRecordID value_object.UUID,
	done chan bool) {
	ticker := time.NewTicker(aggregate.HeartBeatReportInterval)
	defer ticker.Stop()

	for {
		select {
		case <-done: // 任务已完成，删除心跳
			heartBeatRepo.Delete(heartBeatRecordID)
			return
		case <-ticker.C: // 上报心跳
			heartBeatRepo.AliveReport(heartBeatRecordID)
		}
	}
}

// FunctionRunConsumer 消费并执行具体的function节点，若有下游，发布下游待执行functions
func (blocApp *BlocApp) FunctionRunConsumer() {
	event.InjectMq(blocApp.GetOrCreateEventMQ())
	event.InjectFutureEventStorageImplement(blocApp.GetOrCreateFutureEventStorage())
	logger := blocApp.GetOrCreateConsumerLogger()
	funcRunRecordRepo := blocApp.GetOrCreateFunctionRunRecordRepository()
	heartBeatRepo := blocApp.GetOrCreateFuncRunHBeatRepository()
	flowRepo := blocApp.GetOrCreateFlowRepository()
	objectStorage := blocApp.GetOrCreateConsumerObjectStorage()
	flowRunRecordRepo := blocApp.GetOrCreateFlowRunRecordRepository()

	funcToRunEventChan := make(chan event.DomainEvent)
	err := event.ListenEvent(
		&event.FunctionToRun{}, "run_function_consumer", funcToRunEventChan)
	if err != nil {
		panic(err)
	}

	for functionToRunEvent := range funcToRunEventChan {
		functionRunRecordIDStr := functionToRunEvent.Identity()
		funcRunRecordUuid, err := value_object.ParseToUUID(functionRunRecordIDStr)
		if err != nil {
			logger.Errorf(
				"parse received function_run_record_id(%s) to uuid failed.",
				functionRunRecordIDStr)
			continue
		}

		logger.Infof("|--> get function run record id %s", functionRunRecordIDStr)
		functionRecordIns, err := funcRunRecordRepo.GetByID(funcRunRecordUuid)
		if err != nil {
			logger.Errorf(
				"|--> function run record id %s match error: %s", functionRunRecordIDStr, err.Error())
			continue
		}
		if functionRecordIns.IsZero() {
			logger.Errorf("|--> function run record id %s match no ins", functionRunRecordIDStr)
			continue
		}

		flowIns, err := flowRepo.GetByID(functionRecordIns.FlowID)
		if err != nil {
			logger.Errorf(
				"get flow by flow_id failed. flow_id: %s",
				functionRecordIns.FlowID.String())
			continue
		}

		flowRunRecordIns, err := flowRunRecordRepo.GetByID(functionRecordIns.FlowRunRecordID)
		if err != nil {
			logger.Errorf(
				"get flow_run_record_ins from flow_run_record_id(%s) failed",
				functionRecordIns.FlowRunRecordID.String())
			continue
		}
		flowFuncIDMapFuncRunRecordID := flowRunRecordIns.FlowFuncIDMapFuncRunRecordID
		if flowFuncIDMapFuncRunRecordID == nil {
			flowFuncIDMapFuncRunRecordID = make(map[string]value_object.UUID)
		}

		// 装配bloc_history对应bloc的IPT
		flowFunction := flowIns.FlowFunctionIDMapFlowFunction[functionRecordIns.FlowFunctionID]
		// 需确保所有上游节点都已经运行完成了
		upstreamAllSucFinished := true
		upstreamFunctionIntercepted := false
		for _, i := range flowFunction.UpstreamFlowFunctionIDs {
			upstreamFunctionRunRecordID, ok := flowRunRecordIns.FlowFuncIDMapFuncRunRecordID[i]
			if !ok { // 不存在表示没有运行完
				logger.Infof(
					"upstream function not run finished. upstream flow_function_id: %s", i)
				upstreamAllSucFinished = false
				break
			}
			upstreamFunctionRunRecordIns, err := funcRunRecordRepo.GetByID(upstreamFunctionRunRecordID)
			if err != nil {
				logger.Errorf(
					"get upstream function run record ins error:%v. upstream_function_run_record_id: %s",
					err, upstreamFunctionRunRecordID.String())
				upstreamAllSucFinished = false
				break
			}
			if upstreamFunctionRunRecordIns.IsZero() {
				logger.Errorf(
					"get upstream function run record ins nil. upstream_function_run_record_id: %s",
					upstreamFunctionRunRecordID.String())
				upstreamAllSucFinished = false
				break
			}
			if !upstreamFunctionRunRecordIns.Finished() {
				logger.Infof(
					"upstream function is not finished. upstream_function_run_record_id: %s",
					upstreamFunctionRunRecordID.String())
				upstreamAllSucFinished = false
				break
			}
			if !upstreamFunctionRunRecordIns.Pass {
				logger.Infof(
					"upstream function intercepted. breakout. upstream_function_run_record_id: %s",
					upstreamFunctionRunRecordID.String())
				upstreamFunctionIntercepted = true
				break
			}
		}
		if !upstreamAllSucFinished {
			logger.Infof("upstream not all finished. break out")
			event.PubEventAtCertainTime(functionToRunEvent, time.Now().Add(5*time.Second))
			continue
		}
		if upstreamFunctionIntercepted {
			// 上游有节点明确表示拦截了，不能继续往下执行。
			flowRunRecordRepo.Intercepted(flowRunRecordIns.ID, "TODO")
			flowRunRecordRepo.Suc(flowRunRecordIns.ID)
			event.PubEvent(&event.FlowRunFinished{
				FlowRunRecordID: flowRunRecordIns.ID,
			})
			continue
		}
		functionIns := blocApp.GetFunctionByRepoID(functionRecordIns.FunctionID)
		logger.Infof("|----> function id %s", functionIns.ID)
		if functionIns.IsZero() {
			logger.Errorf(
				"get nil function_ins by function_id: %s",
				functionRecordIns.FunctionID.String())
			continue
		}

		// 装配输入参数到blocHis实例【从flowBloc中配置的输入参数的来源（manual/connection）获得】
		functionRecordIns.Ipts = make([][]interface{}, len(flowFunction.ParamIpts))
		for paramIndex, paramIpt := range flowFunction.ParamIpts {
			functionRecordIns.Ipts[paramIndex] = make([]interface{}, len(paramIpt))
			for componentIndex, componentIpt := range paramIpt {
				var value interface{}
				if componentIpt.IptWay == value_object.UserIpt {
					value = componentIpt.Value
				} else if componentIpt.IptWay == value_object.Connection {
					// 找到上游对应节点的运行记录并从其opt中取出要的数据
					upstreamBlocHisID := flowFuncIDMapFuncRunRecordID[componentIpt.FlowFunctionID]
					upstreamFuncRunRecordIns, err := funcRunRecordRepo.GetByID(upstreamBlocHisID)
					if err != nil {
						funcRunRecordRepo.SaveFail(
							functionRecordIns.ID,
							"ipt value get from upstream connection failed")
						functionRecordIns.Ipts[paramIndex][componentIndex] = err.Error()
						continue
					}
					if upstreamFuncRunRecordIns.IsZero() {
						funcRunRecordRepo.SaveFail(
							functionRecordIns.ID,
							"ipt value get from upstream connection failed") // TODO 记录
						functionRecordIns.Ipts[paramIndex][componentIndex] = "not valid"
						continue
					}
					optValue, ok := upstreamFuncRunRecordIns.Opt[componentIpt.Key]
					if !ok {
						funcRunRecordRepo.SaveFail(
							functionRecordIns.ID, "ipt value get from upstream connection failed")
						functionRecordIns.Ipts[paramIndex][componentIndex] = "not valid"
						continue
					}
					tmp, err := objectStorage.Get(optValue.(string))
					if err != nil {
						funcRunRecordRepo.SaveFail(
							functionRecordIns.ID,
							"ipt value get from upstream connection failed")
						functionRecordIns.Ipts[paramIndex][componentIndex] = "not valid"
						continue
					}
					json.Unmarshal(tmp, &value)
				}
				if value == nil && componentIpt.Blank {
					// 非必需参数 且 用户没有填写
					continue
				} else {
					// 有效性检查
					dataValid := value_type.CheckValueTypeValueValid(componentIpt.ValueType, value)
					if !dataValid {
						failMsg := fmt.Sprintf(
							"ipt value not valid. ipt_index: %d; component_indxe: %d, value: %v",
							paramIndex, componentIndex, value)
						funcRunRecordRepo.SaveFail(functionRecordIns.ID, failMsg)
						functionRecordIns.Ipts[paramIndex][componentIndex] = "not valid"
						continue
					}
				}
				functionRecordIns.Ipts[paramIndex][componentIndex] = value
				functionIns.Ipts[paramIndex].Components[componentIndex].Value = value
			}
		}

		err = funcRunRecordRepo.SaveIptBrief(
			funcRunRecordUuid, functionRecordIns.Ipts,
			objectStorage)
		if err != nil {
			// TODO 修改为更为合适的处理
			panic(err)
		}

		// > ipt装配完成，先保存输入
		// 若装配IPT失败
		if functionRecordIns.ErrorMsg != "" {
			logger.Errorf(
				"|----> function run record id %s assemble ipt failed, err: %s",
				functionRunRecordIDStr, functionRecordIns.ErrorMsg)
			funcRunRecordRepo.SaveFail(
				functionRecordIns.ID,
				"装配IPT失败")
			continue
		}

		// > 装配IPT成功
		// > 调用exeBloc开始实际运行代码
		executeFunc := blocApp.GetExecuteFunctionByRepoID(functionIns.ID)

		// 实际开始运行
		timeOutChan := make(chan struct{})
		if flowIns.TimeoutInSeconds > 0 { // 设置了整体运行的超时时长
			thisFlowTaskAllFunctionTasks, err := funcRunRecordRepo.FilterByFlowRunRecordID(flowRunRecordIns.ID)
			var thisFlowTaskUsedSeconds float64
			if err != nil {
				for _, i := range thisFlowTaskAllFunctionTasks {
					if !i.End.IsZero() {
						thisFlowTaskUsedSeconds += i.End.Sub(i.Start).Seconds()
					}
				}
			}
			leftSeconds := flowIns.TimeoutInSeconds - uint32(thisFlowTaskUsedSeconds)
			if leftSeconds > 0 { // 未超时
				timer := time.After(time.Duration(leftSeconds) * time.Second)
				go func() {
					for range timer {
						timeOutChan <- struct{}{}
					}
				}()
			} else { // 已超时
				logger.Infof(
					"|----> func run record id %s timeout canceled", functionRunRecordIDStr)
				funcRunRecordRepo.SaveCancel(funcRunRecordUuid)
				flowRunRecordRepo.TimeoutCancel(flowRunRecordIns.ID)
				continue
			}
		}

		// 开启当前正在运行的function的心跳上报
		done := make(chan bool)
		heartBeatIns := aggregate.NewFunctionExecuteHeartBeat(funcRunRecordUuid)
		err = heartBeatRepo.Create(heartBeatIns)
		if err != nil {
			logger.Errorf("create heart_beat failed: %v", err)
		} else {
			go reportHeartBeat(heartBeatRepo, funcRunRecordUuid, done)
		}

		cancelCheckTimer := time.NewTicker(6 * time.Second)
		progressReportChan := make(chan value_object.FunctionRunStatus)
		functionRunOptChan := make(chan *value_object.FunctionRunOpt)
		var funcRunOpt *value_object.FunctionRunOpt
		ctx := context.Background()
		ctx, cancelFunctionExecute := context.WithCancel(ctx)
		funcRunRecordLogger := blocApp.CreateFunctionRunLogger(funcRunRecordUuid)
		go func() {
			executeFunc.Run(ctx, functionIns.Ipts, progressReportChan, functionRunOptChan, funcRunRecordLogger)
		}()
		for {
			select {
			// 1. 超时
			case <-timeOutChan:
				logger.Infof("|----> function run record id %s timeout canceled", functionRunRecordIDStr)
				funcRunRecordRepo.SaveCancel(funcRunRecordUuid)
				flowRunRecordRepo.TimeoutCancel(flowRunRecordIns.ID)
				goto FunctionNodeRunFinished
			// 2. flow被用户在前端取消
			case <-cancelCheckTimer.C:
				isCanceled := flowRunRecordRepo.ReGetToCheckIsCanceled(flowRunRecordIns.ID)
				if isCanceled {
					logger.Infof("|----> function run record id %s user canceled", functionRunRecordIDStr)
					funcRunRecordRepo.SaveCancel(funcRunRecordUuid)
					goto FunctionNodeRunFinished
				}
			// 3. function运行进度上报
			case runningStatus := <-progressReportChan:
				if runningStatus.Progress > 0 {
					funcRunRecordRepo.PatchProgress(
						funcRunRecordUuid, runningStatus.Progress)
				}
				if runningStatus.Msg != "" {
					funcRunRecordRepo.PatchProgressMsg(
						funcRunRecordUuid, runningStatus.Msg)
				}
				if runningStatus.ProcessStageIndex > 0 {
					funcRunRecordRepo.PatchStageIndex(
						funcRunRecordUuid, runningStatus.ProcessStageIndex)
				}
			// 4. 运行成功完成
			case funcRunOpt = <-functionRunOptChan:
				logger.Infof("|----> function run record id %s run suc", functionRunRecordIDStr)
				goto FunctionNodeRunFinished
			}
		}
	FunctionNodeRunFinished:
		cancelFunctionExecute()
		close(progressReportChan)
		cancelCheckTimer.Stop()
		funcRunRecordLogger.ForceUpload()
		done <- true
		if funcRunOpt.Suc { // 若运行成功，需要将每个输出保存到oss中
			logger.Infof("|----> function run record id %s suc", functionRunRecordIDStr)
			// 将blocOpt的具体每个opt保存到oss并且替换值value, 输出的前50个字符保存到brief中方便前端展示
			for optKey, optVal := range funcRunOpt.Detail {
				uploadByte, _ := json.Marshal(optVal)
				ossKey := functionRunRecordIDStr + "_" + optKey
				err = objectStorage.Set(ossKey, uploadByte)
				if err == nil {
					optInRune := []rune(string(uploadByte))
					minLength := 51
					if len(optInRune) < minLength {
						minLength = len(optInRune)
					}
					if funcRunOpt.Brief == nil {
						funcRunOpt.Brief = make(map[string]string, len(funcRunOpt.Detail))
					}
					funcRunOpt.Brief[optKey] = string(optInRune[:minLength-1])
					funcRunOpt.Detail[optKey] = ossKey
				} else {
					funcRunOpt.Brief[optKey] = "存储运行输出到对象存储失败"
				}
			}
		} else {
			logger.Errorf("|----> function run record id %s run failed", functionRunRecordIDStr)
		}

		if funcRunOpt.Suc {
			funcRunRecordRepo.SaveSuc(
				funcRunRecordUuid, funcRunOpt.Description,
				funcRunOpt.Detail, funcRunOpt.Brief, funcRunOpt.Pass)

			if funcRunOpt.Pass { // 运行通过
				if len(flowFunction.DownstreamFlowFunctionIDs) > 0 { // 若有下游待运行的function节点
					// 创建并发布下游节点
					for _, downStreamFlowFunctionID := range flowFunction.DownstreamFlowFunctionIDs {
						downStreamFlowFunction := flowIns.FlowFunctionIDMapFlowFunction[downStreamFlowFunctionID]
						downStreamFunctionIns := blocApp.GetFunctionByRepoID(downStreamFlowFunction.FunctionID)

						downStreamFunctionRunRecord := aggregate.NewFunctionRunRecordFromFlowDriven(
							downStreamFunctionIns, *flowRunRecordIns,
							downStreamFlowFunctionID)
						_ = funcRunRecordRepo.Create(downStreamFunctionRunRecord)

						err = flowRunRecordRepo.AddFlowFuncIDMapFuncRunRecordID(
							flowRunRecordIns.ID,
							downStreamFlowFunctionID,
							downStreamFunctionRunRecord.ID)
						if err != nil {
							logger.Errorf(
								`flowRunRecordRepo.AddFlowFuncIDMapFuncRunRecordID error: %v.
								flow_run_record_id:%s`,
								flowRunRecordIns.ID.String(), err)
							// TODO 咋办？
						}
						flowRunRecordIns.FlowFuncIDMapFuncRunRecordID[downStreamFlowFunctionID] = downStreamFunctionRunRecord.ID
					}
				} else { // 此函数节点没有下游了。进行检查flow是否全部运行成功从而完成了，
					/*
						检测要点：
							因为 FlowFunctionIDMapFlowFunction 有此flow每个function运行历史的对应记录
							从而检查是不是都有效完成了
					*/
					toCheckFlowFunctionIDMapDone := make(map[string]bool, len(flowIns.FlowFunctionIDMapFlowFunction))
					for flowFunctionID := range flowIns.FlowFunctionIDMapFlowFunction {
						toCheckFlowFunctionIDMapDone[flowFunctionID] = false
					}
					delete(toCheckFlowFunctionIDMapDone, config.FlowFunctionStartID)
					flowFinished := true
					for toCheckFlowFunctionID := range toCheckFlowFunctionIDMapDone {
						funcRunRecordID, ok := flowFuncIDMapFuncRunRecordID[toCheckFlowFunctionID]
						if !ok { // 表示此flow_bloc还没有运行记录
							flowFinished = false
							break
						}
						functionRunRecordIns, err := funcRunRecordRepo.GetByID(funcRunRecordID)
						if err != nil {
							logger.Errorf(
								"get function_run_record by function_run_record_id(%s) error: %s",
								funcRunRecordID.String(), err.Error())
							// 先保守处理为未完成运行
							flowFinished = false
							break
						}
						if !functionRunRecordIns.Finished() {
							flowFinished = false
							break
						}
					}
					// 已检测到全部完成
					if flowFinished {
						flowRunRecordRepo.Suc(flowRunRecordIns.ID)
						event.PubEvent(&event.FlowRunFinished{
							FlowRunRecordID: flowRunRecordIns.ID,
						})
						logger.Infof("|------> pub finished flow_task__id from all finished: %s", flowRunRecordIns.ID)
					}
				}
			} else { // 运行拦截，此function节点以下的节点不用再运行了，此步骤拦截
				flowRunRecordRepo.Intercepted(flowRunRecordIns.ID, "TODO")
				flowRunRecordRepo.Suc(flowRunRecordIns.ID)
				event.PubEvent(&event.FlowRunFinished{
					FlowRunRecordID: flowRunRecordIns.ID,
				})
				logger.Infof("|------> pub finished flow_run_record__id from intercepted: %s", flowRunRecordIns.ID)
			}
		} else { // function节点运行失败, 处理有重试的情况
			if !flowRunRecordIns.IsFromArrangement() { // 非arrangement触发
				// 无重试策略
				if !flowIns.HaveRetryStrategy() || flowRunRecordIns.RetriedAmount >= flowIns.RetryAmount {
					funcRunRecordRepo.SaveFail(funcRunRecordUuid, funcRunOpt.ErrorMsg)
					flowRunRecordRepo.Fail(flowRunRecordIns.ID, "have function failed")
				} else { // 有重试策略
					flowRunRecordRepo.PatchDataForRetry(flowRunRecordIns.ID, flowRunRecordIns.RetriedAmount)

					retryGapSeconds := 3
					if flowIns.RetryIntervalInSecond > 0 {
						retryGapSeconds = int(flowIns.RetryIntervalInSecond)
					}
					futureTime := time.Now().Add(time.Duration(retryGapSeconds) * time.Second)
					event.PubEventAtCertainTime(
						&event.FunctionToRun{FunctionRunRecordID: funcRunRecordUuid},
						futureTime,
					)
				}
			}
		}
	}
}
