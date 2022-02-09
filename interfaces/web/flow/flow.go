package flow

import (
	"sync"
	"time"

	"github.com/fBloc/bloc-server/aggregate"
	"github.com/fBloc/bloc-server/internal/crontab"
	"github.com/fBloc/bloc-server/pkg/value_type"
	"github.com/fBloc/bloc-server/services/flow"
	"github.com/fBloc/bloc-server/value_object"
)

var fService *flow.FlowService

func InjectFlowService(
	f *flow.FlowService,
) {
	fService = f
}

type LatestRun struct {
	ID                           value_object.UUID                    `json:"id"`
	ArrangementID                value_object.UUID                    `json:"arrangement_id,omitempty"`
	ArrangementFlowID            string                               `json:"arrangement_flow_id,omitempty"`
	ArrangementRunRecordID       string                               `json:"arrangement_run_record_id,omitempty"`
	FlowID                       value_object.UUID                    `json:"flow_id"`
	FlowOriginID                 value_object.UUID                    `json:"flow_origin_id"`
	FlowFuncIDMapFuncRunRecordID map[string]value_object.UUID         `json:"flowFunctionID_map_functionRunRecordID"`
	TriggerType                  value_object.TriggerType             `json:"trigger_type"`
	TriggerKey                   string                               `json:"trigger_key"`
	TriggerSource                value_object.FlowTriggeredSourceType `json:"trigger_source"`
	TriggerUserName              string                               `json:"trigger_user_name"`
	TriggerTime                  value_object.JsonDate                `json:"trigger_time"`
	StartTime                    value_object.JsonDate                `json:"start_time"`
	EndTime                      value_object.JsonDate                `json:"end_time"`
	Status                       value_object.RunState                `json:"status"`
	ErrorMsg                     string                               `json:"error_msg"`
	RetriedAmount                uint16                               `json:"retried_amount"`
	TimeoutCanceled              bool                                 `json:"timeout_canceled"`
	Canceled                     bool                                 `json:"canceled"`
	CancelUserName               string                               `json:"cancel_user_name"`
}

func newLatestRunFromAgg(flowRunRecord *aggregate.FlowRunRecord) *LatestRun {
	if flowRunRecord.IsZero() {
		return nil
	}
	resp := &LatestRun{
		ID:                           flowRunRecord.ID,
		FlowID:                       flowRunRecord.FlowID,
		FlowOriginID:                 flowRunRecord.FlowOriginID,
		FlowFuncIDMapFuncRunRecordID: flowRunRecord.FlowFuncIDMapFuncRunRecordID,
		TriggerType:                  flowRunRecord.TriggerType,
		TriggerKey:                   flowRunRecord.TriggerKey,
		TriggerSource:                flowRunRecord.TriggerSource,
		TriggerTime:                  value_object.NewJsonDate(flowRunRecord.TriggerTime),
		StartTime:                    value_object.NewJsonDate(flowRunRecord.StartTime),
		EndTime:                      value_object.NewJsonDate(flowRunRecord.EndTime),
		Status:                       flowRunRecord.Status,
		ErrorMsg:                     flowRunRecord.ErrorMsg,
		RetriedAmount:                flowRunRecord.RetriedAmount,
		TimeoutCanceled:              flowRunRecord.TimeoutCanceled,
		Canceled:                     flowRunRecord.Canceled,
	}
	if flowRunRecord.CancelUserID.IsNil() && flowRunRecord.TriggerUserID.IsNil() {
		return resp
	}
	if !flowRunRecord.CancelUserID.IsNil() {
		user, _ := fService.UserCacheService.GetUserByID(flowRunRecord.CancelUserID)
		if !user.IsZero() {
			resp.CancelUserName = user.Name
		}
	}
	if !flowRunRecord.TriggerUserID.IsNil() {
		user, _ := fService.UserCacheService.GetUserByID(flowRunRecord.TriggerUserID)
		if !user.IsZero() {
			resp.TriggerUserName = user.Name
		}
	}

	return resp
}

type IptComponentConfig struct {
	Blank     bool                              `json:"blank"`
	IptWay    value_object.FunctionParamIptType `json:"ipt_way"`
	ValueType value_type.ValueType              `json:"value_type"`
	// 当且仅当为user_ipt时才会有此
	Value interface{} `json:"value"`
	// 当且仅当为connection时才会有此
	FlowFunctionID string `json:"flow_function_id"`
	Key            string `json:"key"`
}

type FlowFunction struct {
	FunctionID                value_object.UUID      `json:"function_id"`
	Note                      string                 `json:"note"`
	Position                  interface{}            `json:"position"`
	UpstreamFlowFunctionIDs   []string               `json:"upstream_flowfunction_ids"`
	DownstreamFlowFunctionIDs []string               `json:"downstream_flowfunction_ids"`
	ParamIpts                 [][]IptComponentConfig `json:"param_ipts"`
}

func (flowFunc FlowFunction) formatToAggFlowFunction() *aggregate.FlowFunction {
	paramIpts := make([][]aggregate.IptComponentConfig, len(flowFunc.ParamIpts))
	for i, j := range flowFunc.ParamIpts {
		paramIpts[i] = make([]aggregate.IptComponentConfig, len(j))
		for z, k := range j {
			paramIpts[i][z] = aggregate.IptComponentConfig{
				Blank:          k.Blank,
				IptWay:         k.IptWay,
				ValueType:      k.ValueType,
				Value:          k.Value,
				FlowFunctionID: k.FlowFunctionID,
				Key:            k.Key,
			}
		}
	}
	return &aggregate.FlowFunction{
		FunctionID:                flowFunc.FunctionID,
		Note:                      flowFunc.Note,
		Position:                  flowFunc.Position,
		UpstreamFlowFunctionIDs:   flowFunc.UpstreamFlowFunctionIDs,
		DownstreamFlowFunctionIDs: flowFunc.DownstreamFlowFunctionIDs,
		ParamIpts:                 paramIpts,
	}
}

type FunctionRunInfo struct {
	Status              value_object.RunState `json:"status"`
	FunctionRunRecordID value_object.UUID     `json:"function_run_record_id"`
}

type Flow struct {
	ID                            value_object.UUID        `json:"id"`
	Name                          string                   `json:"name"`
	IsDraft                       bool                     `json:"is_draft"`
	Version                       uint                     `json:"version"`
	OriginID                      value_object.UUID        `json:"origin_id"`
	Newest                        bool                     `json:"newest"`
	CreateUserID                  value_object.UUID        `json:"create_user_id,omitempty"`
	CreateUserName                string                   `json:"create_user_name"`
	CreateTime                    time.Time                `json:"create_time"`
	Position                      interface{}              `json:"position"`
	FlowFunctionIDMapFlowFunction map[string]*FlowFunction `json:"flowFunctionID_map_flowFunction"`
	// 运行控制相关
	Crontab               *crontab.CrontabRepresent `json:"crontab"`
	TriggerKey            string                    `json:"trigger_key"`
	TimeoutInSeconds      uint32                    `json:"timeout_in_seconds"`
	RetryAmount           uint16                    `json:"retry_amount"`
	RetryIntervalInSecond uint16                    `json:"retry_interval_in_second"`
	AllowParallelRun      bool                      `json:"allow_parallel_run"`
	// permission
	Read             bool `json:"read"`
	Write            bool `json:"write"`
	Execute          bool `json:"execute"`
	Delete           bool `json:"delete"`
	AssignPermission bool `json:"assign_permission"`
	// latest run status
	LatestRun                                 *LatestRun                 `json:"latest_run,omitempty"`
	LatestRunFlowFunctionIDMapFunctionRunInfo map[string]FunctionRunInfo `json:"latestRun_flowFunctionID_map_functionRunInfo"`
}

func (f *Flow) getAggregateFlowFunctionIDMapFlowFunction() map[string]*aggregate.FlowFunction {
	resp := make(map[string]*aggregate.FlowFunction, len(f.FlowFunctionIDMapFlowFunction))
	for k, v := range f.FlowFunctionIDMapFlowFunction {
		resp[k] = v.formatToAggFlowFunction()
	}
	return resp
}

func (f *Flow) IsZero() bool {
	return f == nil
}

func fromAggWithoutUserPermission(aggF *aggregate.Flow) *Flow {
	if aggF.IsZero() {
		return nil
	}
	httpFuncs := make(map[string]*FlowFunction, len(aggF.FlowFunctionIDMapFlowFunction))
	for k, v := range aggF.FlowFunctionIDMapFlowFunction {
		paramIpts := make([][]IptComponentConfig, len(v.ParamIpts))
		for i, j := range v.ParamIpts {
			paramIpts[i] = make([]IptComponentConfig, len(j))
			for z, k := range j {
				paramIpts[i][z] = IptComponentConfig{
					Blank:          k.Blank,
					IptWay:         k.IptWay,
					ValueType:      k.ValueType,
					Value:          k.Value,
					FlowFunctionID: k.FlowFunctionID,
					Key:            k.Key,
				}
			}
		}

		httpFuncs[k] = &FlowFunction{
			FunctionID:                v.FunctionID,
			Note:                      v.Note,
			Position:                  v.Position,
			UpstreamFlowFunctionIDs:   v.UpstreamFlowFunctionIDs,
			DownstreamFlowFunctionIDs: v.DownstreamFlowFunctionIDs,
			ParamIpts:                 paramIpts,
		}
	}
	retFlow := &Flow{
		ID:                            aggF.ID,
		Name:                          aggF.Name,
		IsDraft:                       aggF.IsDraft,
		Version:                       aggF.Version,
		OriginID:                      aggF.OriginID,
		Newest:                        aggF.Newest,
		CreateTime:                    aggF.CreateTime,
		Position:                      aggF.Position,
		FlowFunctionIDMapFlowFunction: httpFuncs,
		Crontab:                       aggF.Crontab,
		TriggerKey:                    aggF.TriggerKey,
		TimeoutInSeconds:              aggF.TimeoutInSeconds,
		RetryAmount:                   aggF.RetryAmount,
		RetryIntervalInSecond:         aggF.RetryIntervalInSecond,
		AllowParallelRun:              aggF.AllowParallelRun,
	}
	creator, err := fService.UserCacheService.GetUserByID(aggF.CreateUserID)
	if err == nil && !creator.IsZero() {
		retFlow.CreateUserName = creator.Name
	} else {
		retFlow.CreateUserName = "unknown"
	}
	return retFlow
}

func fromAgg(aggF *aggregate.Flow, reqUser *aggregate.User) *Flow {
	bareFlow := fromAggWithoutUserPermission(aggF)
	if bareFlow.IsZero() {
		return nil
	}
	bareFlow.Read = aggF.UserCanRead(reqUser)
	bareFlow.Write = aggF.UserCanWrite(reqUser)
	bareFlow.Execute = aggF.UserCanExecute(reqUser)
	bareFlow.Delete = aggF.UserCanDelete(reqUser)
	bareFlow.AssignPermission = aggF.UserCanAssignPermission(reqUser)
	return bareFlow
}

// fromAggWithLatestRunFlowView 附带此flow最近一次的运行记录，注意是flow_run_record的
// 当前的使用场景是在获取flow列表的时候会返回其中每个flow的最近一次运行状态
func fromAggWithLatestRunFlowView(aggF *aggregate.Flow, reqUser *aggregate.User) *Flow {
	retFlow := fromAgg(aggF, reqUser)
	if retFlow.IsZero() {
		return nil
	}

	latestFlowRunRecord, err := fService.GetLatestRunRecordByFlowID(aggF.ID)
	if err != nil {
		retFlow.LatestRun = &LatestRun{ErrorMsg: "visit latest run failed: " + err.Error()}
	} else {
		retFlow.LatestRun = newLatestRunFromAgg(latestFlowRunRecord)
	}
	return retFlow
}

// fromAggWithCertainRunFunctionView 附带此flow及其特定运行记录下的「各个function node的运行状态」
// 当前使用场景是点击了单个flow的时候，此时会进行渲染其下的functions，需要渲染其最近那次运行下的各个function状态
func fromAggWithCertainRunFunctionView(
	aggF *aggregate.Flow,
	theFlowRunRecord *aggregate.FlowRunRecord,
	reqUser *aggregate.User,
) *Flow {
	retFlow := fromAgg(aggF, reqUser)
	if retFlow.IsZero() {
		return nil
	}

	wg := sync.WaitGroup{}
	wg.Add(len(theFlowRunRecord.FlowFuncIDMapFuncRunRecordID))

	type funcRunState struct {
		flowfunctionID      string
		functionRunRecordID value_object.UUID
		functionRunState    value_object.RunState
	}
	resp := make(chan funcRunState, len(theFlowRunRecord.FlowFuncIDMapFuncRunRecordID))

	for flowFuncID, functionRunRecordID := range theFlowRunRecord.FlowFuncIDMapFuncRunRecordID {
		go func(
			flowFuncID string, functionRunRecordID value_object.UUID,
			retChan chan funcRunState, wg *sync.WaitGroup,
		) {
			defer wg.Done()
			funcRecord, err := fService.FunctionRunRecord.GetByID(functionRunRecordID)
			if err != nil {
				resp <- funcRunState{
					flowfunctionID:      flowFuncID,
					functionRunRecordID: functionRunRecordID,
					functionRunState:    value_object.UnknownRunState,
				}
				return
			}
			var thisFuncRunState value_object.RunState
			if funcRecord.Finished() {
				if funcRecord.Suc {
					thisFuncRunState = value_object.Suc
					goto RET
				}
				if funcRecord.Canceled {
					if theFlowRunRecord.TimeoutCanceled {
						thisFuncRunState = value_object.TimeoutCanceled
						goto RET
					}
					if theFlowRunRecord.Canceled {
						thisFuncRunState = value_object.UserCanceled
						goto RET
					}
				}
				thisFuncRunState = value_object.Fail
			} else {
				thisFuncRunState = value_object.Running
			}
		RET:
			resp <- funcRunState{
				flowfunctionID:      flowFuncID,
				functionRunRecordID: functionRunRecordID,
				functionRunState:    thisFuncRunState,
			}
		}(flowFuncID, functionRunRecordID, resp, &wg)
	}

	wg.Wait()
	close(resp)

	if retFlow.LatestRunFlowFunctionIDMapFunctionRunInfo == nil {
		retFlow.LatestRunFlowFunctionIDMapFunctionRunInfo = make(
			map[string]FunctionRunInfo,
			len(theFlowRunRecord.FlowFuncIDMapFuncRunRecordID),
		)
	}
	for i := range resp {
		retFlow.LatestRunFlowFunctionIDMapFunctionRunInfo[i.flowfunctionID] = FunctionRunInfo{
			Status:              i.functionRunState,
			FunctionRunRecordID: i.functionRunRecordID,
		}
	}

	retFlow.LatestRun = newLatestRunFromAgg(theFlowRunRecord)
	return retFlow
}

// fromAggWithLatestRunFunctionView 附带此flow最近一次的运行记录下的「各个function node的运行状态」
// 当前使用场景是点击了单个flow的时候，此时会进行渲染其下的functions，需要渲染其最近那次运行下的各个function状态
func fromAggWithLatestRunFunctionView(
	aggF *aggregate.Flow,
	reqUser *aggregate.User,
) *Flow {
	retFlow := fromAgg(aggF, reqUser)
	if retFlow.IsZero() {
		return nil
	}

	latestFlowRunRecord, err := fService.GetLatestRunRecordByFlowID(aggF.ID)
	if err != nil {
		retFlow.LatestRun = &LatestRun{ErrorMsg: "visit latest run failed: " + err.Error()}
	} else {
		wg := sync.WaitGroup{}
		wg.Add(len(latestFlowRunRecord.FlowFuncIDMapFuncRunRecordID))

		type funcRunState struct {
			flowfunctionID      string
			functionRunRecordID value_object.UUID
			functionRunState    value_object.RunState
		}
		resp := make(chan funcRunState, len(latestFlowRunRecord.FlowFuncIDMapFuncRunRecordID))

		for flowFuncID, functionRunRecordID := range latestFlowRunRecord.FlowFuncIDMapFuncRunRecordID {
			go func(
				flowFuncID string, functionRunRecordID value_object.UUID,
				retChan chan funcRunState, wg *sync.WaitGroup,
			) {
				defer wg.Done()
				funcRecord, err := fService.FunctionRunRecord.GetByID(functionRunRecordID)
				if err != nil {
					resp <- funcRunState{
						flowfunctionID:      flowFuncID,
						functionRunRecordID: functionRunRecordID,
						functionRunState:    value_object.UnknownRunState,
					}
					return
				}
				var thisFuncRunState value_object.RunState
				if funcRecord.Finished() {
					if funcRecord.Suc {
						thisFuncRunState = value_object.Suc
						goto RET
					}
					if funcRecord.Canceled {
						if latestFlowRunRecord.TimeoutCanceled {
							thisFuncRunState = value_object.TimeoutCanceled
							goto RET
						}
						if latestFlowRunRecord.Canceled {
							thisFuncRunState = value_object.UserCanceled
							goto RET
						}
					}
					thisFuncRunState = value_object.Fail
				} else {
					thisFuncRunState = value_object.Running
				}
			RET:
				resp <- funcRunState{
					flowfunctionID:      flowFuncID,
					functionRunRecordID: functionRunRecordID,
					functionRunState:    thisFuncRunState,
				}
			}(flowFuncID, functionRunRecordID, resp, &wg)
		}

		wg.Wait()
		close(resp)

		if retFlow.LatestRunFlowFunctionIDMapFunctionRunInfo == nil {
			retFlow.LatestRunFlowFunctionIDMapFunctionRunInfo = make(
				map[string]FunctionRunInfo,
				len(latestFlowRunRecord.FlowFuncIDMapFuncRunRecordID),
			)
		}
		for i := range resp {
			retFlow.LatestRunFlowFunctionIDMapFunctionRunInfo[i.flowfunctionID] = FunctionRunInfo{
				Status:              i.functionRunState,
				FunctionRunRecordID: i.functionRunRecordID,
			}
		}

		retFlow.LatestRun = newLatestRunFromAgg(latestFlowRunRecord)
	}
	return retFlow
}

// 附带特定arrangement下的此flow最近一次运行记录
func fromAggWithLatestRunOfCertainArrangement(
	aggF *aggregate.Flow, reqUser *aggregate.User, arrangementFlowID string,
) *Flow {
	retFlow := fromAgg(aggF, reqUser)
	if retFlow.IsZero() {
		return nil
	}

	latestFlowRunRecord, err := fService.GetLatestRunRecordByFlowID(aggF.ID)
	if err != nil {
		retFlow.LatestRun = &LatestRun{ErrorMsg: "visit latest run failed: " + err.Error()}
	} else {
		// TODO 需要返回其下各个function的运行状态吗？
		retFlow.LatestRun = newLatestRunFromAgg(latestFlowRunRecord)
	}
	return retFlow
}

// 列表返回
func fromAggSliceWithLatestRun(aggFs []aggregate.Flow, reqUser *aggregate.User) []*Flow {
	if len(aggFs) <= 0 {
		return []*Flow{}
	}
	type orderedFlow struct {
		*Flow
		flowIndex int
	}

	flowChan := make(chan *orderedFlow, len(aggFs))
	wGroup := sync.WaitGroup{}
	wGroup.Add(len(aggFs))

	for index, aggF := range aggFs {
		go func(wg *sync.WaitGroup, index int, aF aggregate.Flow, flowChan chan *orderedFlow) {
			defer wg.Done()
			retFlow := fromAggWithLatestRunFlowView(&aF, reqUser)
			if retFlow.IsZero() {
				return
			}
			flowChan <- &orderedFlow{flowIndex: index, Flow: retFlow}
		}(&wGroup, index, aggF, flowChan)
	}

	wGroup.Wait()
	close(flowChan)

	ret := make([]*Flow, len(aggFs))
	for orderedFlow := range flowChan {
		ret[orderedFlow.flowIndex] = orderedFlow.Flow
	}

	return ret
}

// 列表返回
func fromAggSlice(aggFs []aggregate.Flow, reqUser *aggregate.User) []*Flow {
	if len(aggFs) <= 0 {
		return []*Flow{}
	}

	ret := make([]*Flow, len(aggFs))
	for index, aggF := range aggFs {
		ret[index] = fromAgg(&aggF, reqUser)
	}

	return ret
}
