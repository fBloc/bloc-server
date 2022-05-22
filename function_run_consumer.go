package bloc

import (
	"encoding/json"
	"fmt"
	"time"

	"github.com/fBloc/bloc-server/event"
	"github.com/fBloc/bloc-server/pkg/value_type"
	"github.com/fBloc/bloc-server/value_object"
)

// FunctionRunConsumer 接收到要运行的function，主要有以下预操作：
// 1. 装配ipt具体值
// 2. 检测是否已超时
// 3. 都没问题发布client能识别的的运行消息
func (blocApp *BlocApp) FunctionRunConsumer() {
	event.InjectMq(blocApp.GetOrCreateEventMQ())
	event.InjectFutureEventStorageImplement(blocApp.GetOrCreateFutureEventStorage())
	logger := blocApp.GetOrCreateScheduleLogger()
	funcRunRecordRepo := blocApp.GetOrCreateFunctionRunRecordRepository()
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

		logTags := map[string]string{
			string(value_object.SpanID): value_object.NewSpanID(),
			"business":                  "function run consumer",
			"function_run_record_id":    functionRunRecordIDStr}
		logger.Infof(logTags, "received function_run_record: %s", functionRunRecordIDStr)

		funcRunRecordUuid, err := value_object.ParseToUUID(functionRunRecordIDStr)
		if err != nil {
			logger.Errorf(logTags, "to uuid failed: %v", err)
			continue
		}

		functionRecordIns, err := funcRunRecordRepo.GetByID(funcRunRecordUuid)
		if err != nil {
			logger.Errorf(logTags,
				"get func_run_record by id failed: %v", err)
			continue
		}
		if functionRecordIns.IsZero() {
			logger.Errorf(logTags,
				"get func_run_record by id match no record")
			continue
		}
		logTags[string(value_object.TraceID)] = functionRecordIns.TraceID
		logger.Infof(logTags, "received msg and suc get function_run_record ins")
		if functionRecordIns.Finished() {
			logger.Warningf(logTags, "should not pub already finished function!")
			continue
		}
		if !functionRecordIns.Start.IsZero() {
			logger.Warningf(logTags, "should not pub already started function!")
			continue
		}

		flowIns, err := flowRepo.GetByID(functionRecordIns.FlowID)
		logTags["flow_id"] = functionRecordIns.FlowID.String()
		if err != nil {
			logger.Errorf(logTags, "get flow by flow_id failed: %v", err)
			continue
		}

		flowRunRecordIns, err := flowRunRecordRepo.GetByID(functionRecordIns.FlowRunRecordID)
		logTags["flow_run_record_id"] = functionRecordIns.FlowRunRecordID.String()
		if err != nil {
			logger.Errorf(logTags, "get flow_run_record_ins failed: %v", err)
			continue
		}
		flowFuncIDMapFuncRunRecordID := flowRunRecordIns.FlowFuncIDMapFuncRunRecordID
		if flowFuncIDMapFuncRunRecordID == nil {
			flowFuncIDMapFuncRunRecordID = make(map[string]value_object.UUID)
		}

		// 装配function_run_record对应function的具体输入参数值
		flowFunction := flowIns.FlowFunctionIDMapFlowFunction[functionRecordIns.FlowFunctionID]
		// 需确保所有上游节点都已经运行完成了
		upstreamAllSucFinished := true
		upstreamFunctionIntercepted := false
		if len(flowFunction.UpstreamFlowFunctionIDs) > 1 { // 在只有一个上游节点的情况下，不需要检测
			for _, i := range flowFunction.UpstreamFlowFunctionIDs {
				upstreamFunctionRunRecordID, ok := flowRunRecordIns.FlowFuncIDMapFuncRunRecordID[i]
				if !ok { // 不存在表示没有运行完
					logger.Warningf(logTags,
						"upstream function not run finished. upstream flow_function_id: %s", i)
					upstreamAllSucFinished = false
					break
				}
				upstreamFunctionRunRecordIns, err := funcRunRecordRepo.GetByID(upstreamFunctionRunRecordID)
				if err != nil {
					logger.Errorf(logTags,
						"get upstream function run record ins error:%v. upstream_function_run_record_id: %s",
						err, upstreamFunctionRunRecordID.String())
					upstreamAllSucFinished = false
					break
				}
				if upstreamFunctionRunRecordIns.IsZero() {
					logger.Errorf(logTags,
						"get upstream function run record ins nil. upstream_function_run_record_id: %s",
						upstreamFunctionRunRecordID.String())
					upstreamAllSucFinished = false
					break
				}
				if !upstreamFunctionRunRecordIns.Finished() {
					logger.Infof(logTags,
						"upstream function is not finished. upstream_function_run_record_id: %s",
						upstreamFunctionRunRecordID.String())
					upstreamAllSucFinished = false
					break
				}
				if upstreamFunctionRunRecordIns.InterceptBelowFunctionRun {
					// 为什么会出现这种情况的说明：可能有两个上游节点，其中一个成功、另一个决定拦截
					// 成功的发布下游节点的时候会发布此节点
					logger.Infof(logTags,
						"upstream function intercepted. breakout. upstream_function_run_record_id: %s",
						upstreamFunctionRunRecordID.String())
					upstreamFunctionIntercepted = true
					break
				}
			}
		}
		if !upstreamAllSucFinished {
			logger.Infof(logTags, "upstream not all finished. break out")
			event.PubEventAtCertainTime(functionToRunEvent, time.Now().Add(5*time.Second))
			continue
		}
		if upstreamFunctionIntercepted {
			// 上游有节点明确表示拦截了，不能继续往下执行。
			err := flowRunRecordRepo.Intercepted(flowRunRecordIns.ID, "TODO")
			if err != nil {
				logger.Errorf(logTags, "save flow run finished : %v", err)
			}
		}

		functionIns := blocApp.GetFunctionByRepoID(functionRecordIns.FunctionID)
		if functionIns.IsZero() {
			logger.Errorf(logTags, "get function by id match no record")
			continue
		}
		logTags["function_id"] = functionIns.ID.String()

		// 装配输入参数到function_run_record实例【从flowFunction中配置的输入参数的来源（manual/connection）获得】
		// 如果是被覆盖参数方式触发运行的，优先使用传入的覆盖参数
		functionRecordIns.Ipts = make([][]interface{}, len(flowFunction.ParamIpts))
		for paramIndex, paramIpt := range flowFunction.ParamIpts {
			functionRecordIns.Ipts[paramIndex] = make([]interface{}, len(paramIpt))
			for componentIndex, componentIpt := range paramIpt {
				var value interface{} = nil
				customSetted := false
				if _, ok := flowRunRecordIns.OverideIptParams[functionRecordIns.FlowFunctionID]; ok {
					customeParam := flowRunRecordIns.OverideIptParams[functionRecordIns.FlowFunctionID]
					if len(customeParam)-1 >= paramIndex && len(customeParam[paramIndex])-1 >= componentIndex {
						value = customeParam[paramIndex][componentIndex]
						customSetted = true
					}
				}

				if !customSetted {
					if componentIpt.IptWay == value_object.UserIpt {
						value = componentIpt.Value
					} else if componentIpt.IptWay == value_object.Connection {
						// 找到上游对应节点的运行记录并从其opt中取出要的数据
						upstreamFuncRunRecordID := flowFuncIDMapFuncRunRecordID[componentIpt.FlowFunctionID]
						upstreamFuncRunRecordIns, err := funcRunRecordRepo.GetByID(upstreamFuncRunRecordID)
						if err != nil {
							logger.Errorf(logTags,
								"find upstream functionRunRecordIns failed: %v. which id is: %s",
								err, upstreamFuncRunRecordID)
							functionRecordIns.Ipts[paramIndex][componentIndex] = "not valid from find upstream function error"

							err := funcRunRecordRepo.SaveFail(
								functionRecordIns.ID,
								"ipt value get from upstream connection failed")
							if err != nil {
								logger.Errorf(logTags, "funcRunRecord save fail failed: %v", err)
							}
							continue
						}
						if upstreamFuncRunRecordIns.IsZero() {
							logger.Errorf(logTags,
								"find upstream functionRunRecordIns match record. which id is: %s",
								upstreamFuncRunRecordID)
							functionRecordIns.Ipts[paramIndex][componentIndex] = "not valid from not find upstream function"

							err := funcRunRecordRepo.SaveFail(
								functionRecordIns.ID,
								"ipt value get from upstream connection failed")
							if err != nil {
								logger.Errorf(logTags, "funcRunRecord save fail failed: %v", err)
							}
							logger.Errorf(logTags,
								"ipt value get from upstream connection failed. find no valid corresponding functionRunRecordIns.paramIndex: %d, componentIndex: %d",
								paramIndex, componentIndex)
							continue
						}
						optValue, ok := upstreamFuncRunRecordIns.Opt[componentIpt.Key]
						if !ok {
							logger.Errorf(logTags,
								"upstream functionRunRecordIns's opt donnot have this key: %s. which id is: %s",
								componentIpt.Key, upstreamFuncRunRecordID)
							functionRecordIns.Ipts[paramIndex][componentIndex] = "not valid from upstream function opt not have this key"

							err := funcRunRecordRepo.SaveFail(
								functionRecordIns.ID,
								"ipt value get from upstream connection failed")
							if err != nil {
								logger.Errorf(logTags, "funcRunRecord save fail failed: %v", err)
							}
							continue
						}
						isKeyExist, tmp, err := objectStorage.Get(optValue.(string))
						if !isKeyExist || err != nil {
							logger.Errorf(logTags,
								"get upstream functionRunRecordIns's opt from object storage failed: error: %v, is_key_exist: %t",
								optValue.(string), isKeyExist)
							functionRecordIns.Ipts[paramIndex][componentIndex] = "not valid from object storage"

							err := funcRunRecordRepo.SaveFail(functionRecordIns.ID,
								"ipt value get from upstream connection failed")
							if err != nil {
								logger.Errorf(logTags, "funcRunRecord save fail failed: %v", err)
							}
							continue
						}
						json.Unmarshal(tmp, &value)
					}
				}

				if value == nil && componentIpt.Blank {
					// 非必需参数 且 用户没有填写
					continue
				} else {
					// 有效性检查
					dataValid := value_type.CheckValueTypeValueValid(componentIpt.ValueType, value)
					if !dataValid {
						functionRecordIns.Ipts[paramIndex][componentIndex] = "not valid"

						failMsg := fmt.Sprintf(
							"ipt value not valid. ipt_index: %d; component_indxe: %d, value: %v",
							paramIndex, componentIndex, value)
						err := funcRunRecordRepo.SaveFail(functionRecordIns.ID, failMsg)
						if err != nil {
							logger.Errorf(logTags, "funcRunRecord save fail failed: %v", err)
						}
						functionRecordIns.Ipts[paramIndex][componentIndex] = "not valid"
						continue
					}
				}
				functionRecordIns.Ipts[paramIndex][componentIndex] = value
				functionIns.Ipts[paramIndex].Components[componentIndex].Value = value
			}
		}

		err = funcRunRecordRepo.SaveIptBrief(
			funcRunRecordUuid, functionIns.Ipts,
			functionRecordIns.Ipts, objectStorage)
		if err != nil {
			logger.Errorf(logTags, "persist ipt failed: %v", err)
			err := flowRunRecordRepo.Fail(
				flowRunRecordIns.ID,
				fmt.Sprintf(
					"persist function-%s's ipt failed. error: %v",
					functionIns.Name, err),
			)
			if err != nil {
				logger.Errorf(logTags, "persist flow_run_record fail: %v", err)
			}
			continue
		}

		// > ipt装配完成，先保存输入
		// 若装配IPT失败
		if functionRecordIns.ErrorMsg != "" {
			logger.Errorf(logTags,
				"assemble ipt failed, err: %s", functionRecordIns.ErrorMsg)
			funcRunRecordRepo.SaveFail(functionRecordIns.ID, "装配IPT失败")
			continue
		}

		// > 装配IPT已成功，处理flow设置的超时问题
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
				err = funcRunRecordRepo.SetTimeout(funcRunRecordUuid,
					time.Now().Add(time.Duration(leftSeconds)*time.Second))
				if err != nil {
					logger.Errorf(logTags,
						"set timeout for function_run_record failed: %v", err)
				}
			} else { // 已超时
				logger.Infof(logTags,
					"func run record id %s timeout canceled", functionRunRecordIDStr)
				funcRunRecordRepo.SaveCancel(funcRunRecordUuid)
				flowRunRecordRepo.TimeoutCancel(flowRunRecordIns.ID)
				continue
			}
		}
		// 发布运行任务到具体的function provider
		err = event.PubEvent(&event.ClientRunFunction{
			FunctionRunRecordID: funcRunRecordUuid,
			ClientName:          functionIns.ProviderName})
		if err != nil {
			logger.Errorf(logTags, "pub ClientRunFunction event failed: %v", err)
		} else {
			logger.Infof(logTags, "pub function to run event suc")
		}
	}
}
