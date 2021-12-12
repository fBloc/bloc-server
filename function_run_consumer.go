package bloc

import (
	"encoding/json"
	"fmt"
	"time"

	"github.com/fBloc/bloc-backend-go/event"
	"github.com/fBloc/bloc-backend-go/pkg/value_type"
	"github.com/fBloc/bloc-backend-go/value_object"
)

// FunctionRunConsumer 接收到要运行的function，主要有以下预操作：
// 1. 装配ipt具体值
// 2. 检测是否已超时
// 3. 都没问题发布client能识别的的运行消息
func (blocApp *BlocApp) FunctionRunConsumer() {
	event.InjectMq(blocApp.GetOrCreateEventMQ())
	event.InjectFutureEventStorageImplement(blocApp.GetOrCreateFutureEventStorage())
	logger := blocApp.GetOrCreateConsumerLogger()
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

		// 装配function_run_record对应function的具体输入参数值
		flowFunction := flowIns.FlowFunctionIDMapFlowFunction[functionRecordIns.FlowFunctionID]
		// 需确保所有上游节点都已经运行完成了
		upstreamAllSucFinished := true
		upstreamFunctionIntercepted := false
		if len(flowFunction.UpstreamFlowFunctionIDs) > 1 { // 在只有一个上游节点的情况下，不需要检测
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
				if upstreamFunctionRunRecordIns.InterceptBelowFunctionRun {
					// 为什么会出现这种情况的说明：可能有两个上游节点，其中一个成功、另一个决定拦截
					// 成功的发布下游节点的时候会发布此节点
					logger.Infof(
						"upstream function intercepted. breakout. upstream_function_run_record_id: %s",
						upstreamFunctionRunRecordID.String())
					upstreamFunctionIntercepted = true
					break
				}
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

		// 装配输入参数到function_run_record实例【从flowFunction中配置的输入参数的来源（manual/connection）获得】
		functionRecordIns.Ipts = make([][]interface{}, len(flowFunction.ParamIpts))
		for paramIndex, paramIpt := range flowFunction.ParamIpts {
			functionRecordIns.Ipts[paramIndex] = make([]interface{}, len(paramIpt))
			for componentIndex, componentIpt := range paramIpt {
				var value interface{}
				if componentIpt.IptWay == value_object.UserIpt {
					value = componentIpt.Value
				} else if componentIpt.IptWay == value_object.Connection {
					// 找到上游对应节点的运行记录并从其opt中取出要的数据
					upstreamFuncRunRecordID := flowFuncIDMapFuncRunRecordID[componentIpt.FlowFunctionID]
					upstreamFuncRunRecordIns, err := funcRunRecordRepo.GetByID(upstreamFuncRunRecordID)
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
			funcRunRecordUuid, functionIns.Ipts,
			functionRecordIns.Ipts, objectStorage)
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
					logger.Errorf("set timeout for function_run_record failed: %v", err)
				}
			} else { // 已超时
				logger.Infof(
					"|----> func run record id %s timeout canceled", functionRunRecordIDStr)
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
			logger.Errorf("pub ClientRunFunction event failed: %v", err)
		}
	}
}
