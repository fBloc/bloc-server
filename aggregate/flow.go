package aggregate

import (
	"fmt"
	"sync"
	"time"

	"github.com/fBloc/bloc-server/config"
	"github.com/fBloc/bloc-server/internal/crontab"
	"github.com/fBloc/bloc-server/pkg/value_type"
	"github.com/fBloc/bloc-server/value_object"
)

type IptComponentConfig struct {
	Blank     bool
	IptWay    value_object.FunctionParamIptType
	ValueType value_type.ValueType
	// 当且仅当为user_ipt时才会有此
	Value interface{}
	// 当且仅当为connection时才会有此
	FlowFunctionID string
	Key            string
}

type FlowFunction struct {
	FunctionID                value_object.UUID
	Function                  *Function
	Note                      string
	Position                  interface{}
	UpstreamFlowFunctionIDs   []string
	DownstreamFlowFunctionIDs []string
	ParamIpts                 [][]IptComponentConfig // 第一层对应一个ipt，第二层对应ipt内的component
	findAllUpstreamIDOnce     sync.Once
	allUpstreamIDsMap         map[string]struct{}
}

func (flowFunc *FlowFunction) Name() string {
	if flowFunc.Note != "" {
		return flowFunc.Note
	}
	if flowFunc.Function == nil {
		return ""
	}
	return flowFunc.Function.Name
}

// AllUpstreamFlowFunctionIDsMap including father、grandfather、...'s flow_function_ids
func (flowFunc *FlowFunction) AllUpstreamFlowFunctionIDsMap(flowFuncIDMapFlowFunction map[string]*FlowFunction) map[string]struct{} {
	if flowFunc.allUpstreamIDsMap != nil {
		return flowFunc.allUpstreamIDsMap
	}
	var upIDs []string
	var iterFunc func(thisFlowFunc *FlowFunction, upIDs *[]string)
	iterFunc = func(thisFlowFunc *FlowFunction, upIDs *[]string) {
		for _, upstreamFlowFuncID := range thisFlowFunc.UpstreamFlowFunctionIDs {
			if upstreamFlowFuncID == config.FlowFunctionStartID {
				return
			}
			*upIDs = append(*upIDs, upstreamFlowFuncID)
			iterFunc(flowFuncIDMapFlowFunction[upstreamFlowFuncID], upIDs)
		}
	}

	flowFunc.findAllUpstreamIDOnce.Do(
		func() {
			iterFunc(flowFunc, &upIDs)
			upIDsMap := make(map[string]struct{}, len(upIDs))
			for _, i := range upIDs {
				upIDsMap[i] = struct{}{}
			}
			flowFunc.allUpstreamIDsMap = upIDsMap
		})
	return flowFunc.allUpstreamIDsMap
}

func (flowFunc *FlowFunction) CheckWhetherParamValidIsValid(
	userIptValue [][]interface{},
) (bool, error) {
	if flowFunc.Function.IsZero() {
		return false, fmt.Errorf(
			"not set function to flow_function: %s, cannot check whether param is valid",
			flowFunc.Name())
	}

	// 重要，这里的参数检测仅仅采用半检测 - 即只检测用户输入的参数类型是有效的。不考虑整体的有效性
	// 不考虑整体的是为了兼容一种参数半动态输入驱动运行的情况
	for iptIndex, iptParamVals := range userIptValue {
		if len(iptParamVals) <= 0 { // 只是占位
			continue
		}
		if len(iptParamVals) > len(flowFunc.ParamIpts[iptIndex]) {
			return false, fmt.Errorf("ipt index:%d's component value exceed", iptIndex)
		}
		for componentIndex, componentVal := range iptParamVals {
			componentParamConfig := flowFunc.ParamIpts[iptIndex][componentIndex]
			valueValid := value_type.CheckValueTypeValueValid(
				componentParamConfig.ValueType, componentVal)
			if !valueValid {
				return false, fmt.Errorf(
					"user_ipt value type wrong",
				)
			}
		}
	}
	return true, nil
}

// CheckValid 检测此function node的配置是否正确且有效
func (flowFunc *FlowFunction) CheckValid(
	thisFlowFuncID string,
	flowFuncIDMapFlowFunction map[string]*FlowFunction,
) (bool, error) {
	// 节点上游id必须是在数据里的（防止前端传过来的数据不对）
	for _, funcID := range flowFunc.UpstreamFlowFunctionIDs {
		if _, ok := flowFuncIDMapFlowFunction[funcID]; !ok {
			return false, fmt.Errorf(
				`「%s」节点填写的上游节点flow_function_id(%s)无效
				-没有找到对应的flow_function_id`,
				flowFunc.Name(), funcID)
		}
	}

	// 节点下游id必须是在数据里的（防止前端传过来的数据不对）
	for _, funcID := range flowFunc.DownstreamFlowFunctionIDs {
		if _, ok := flowFuncIDMapFlowFunction[funcID]; !ok {
			return false, fmt.Errorf(
				`「%s」节点填写的下游节点flow_function_id(%s)无效
				-没有找到对应的flow_function_id`,
				flowFunc.Name(), funcID)
		}
	}

	// 开始节点的特殊情况
	if thisFlowFuncID == config.FlowFunctionStartID {
		if len(flowFunc.DownstreamFlowFunctionIDs) == 0 && len(flowFuncIDMapFlowFunction) > 1 {
			return false, fmt.Errorf(
				"start function says no downstreams, but get all %d function",
				len(flowFuncIDMapFlowFunction))
		}
		return true, nil
	}

	// 下面开始都是非开始节点才需要做的检测

	// 所有节点都应该要有输入节点
	if len(flowFunc.UpstreamFlowFunctionIDs) <= 0 {
		return false, fmt.Errorf(
			"「%s」节点没有上游节点 - 不允许",
			flowFunc.Name())
	}

	// 检测输入参数的类型是否正确
	for iptIndex, iptParamConfig := range flowFunc.ParamIpts { // 对应到ipt
		for componentIndex, componentParamConfig := range iptParamConfig { // 对应到component
			if flowFunc.Function.IsZero() {
				return false, fmt.Errorf(
					"not set function to flow_function: %s, cannot check whether ipt param is valid",
					thisFlowFuncID)
			}

			// 若是必填参数、但没有配置参数
			if componentParamConfig.Blank && flowFunc.Function.Ipts[iptIndex].Must {
				return false, fmt.Errorf(
					"「%s」节点第%d个ipt下的第%d个component要求必填，但是没填",
					flowFunc.Name(),
					iptIndex, componentIndex)
			}

			if componentParamConfig.IptWay == value_object.Connection { // 配置的参数输入方式是链接
				// connection写的节点ID是否存在
				if _, ok := flowFuncIDMapFlowFunction[componentParamConfig.FlowFunctionID]; !ok {
					return false, fmt.Errorf(
						`「%s」节点第%d个ipt下的第%d个component输入的上游flow_function节点id(%s)无效
						-没有此flow_function_id的上游节点`,
						flowFunc.Name(),
						iptIndex, componentIndex, componentParamConfig.FlowFunctionID,
					)
				}

				// 2. connection写的节点ID需要是此节点的上游
				// 注意：可以不是直接上游！
				allUpStreamsMap := flowFunc.AllUpstreamFlowFunctionIDsMap(flowFuncIDMapFlowFunction)
				if _, ok := allUpStreamsMap[componentParamConfig.FlowFunctionID]; !ok {
					return false, fmt.Errorf(
						`「%s」节点第%d个ipt下的第%d个component输入的上游flow_function节点id(%s)无效
						-此flow_function_id对应的节点不是直接上游节点、不能作为输入`,
						flowFunc.Name(),
						iptIndex, componentIndex, componentParamConfig.FlowFunctionID,
					)
				}

				// 3. 检查对应出参的类型是不是和此参数的输入要求类型一致
				iptNode := flowFuncIDMapFlowFunction[componentParamConfig.FlowFunctionID]
				for _, optItem := range iptNode.Function.Opts {
					if optItem.Key != componentParamConfig.Key {
						continue
					}
					if optItem.ValueType != componentParamConfig.ValueType {
						return false, fmt.Errorf(
							`「%s」节点第%d个ipt下的第%d个component输入的上游flow_function节点id(%s)无效
							-此flow_function_id对应的节点不是直接上游节点、不能作为输入`,
							flowFunc.Name(),
							iptIndex, componentIndex, componentParamConfig.FlowFunctionID)
					}
				}
			} else if componentParamConfig.IptWay == value_object.UserIpt { // 配置的参数输入方式是用户输入的值
				valueValid := value_type.CheckValueTypeValueValid(
					componentParamConfig.ValueType, componentParamConfig.Value)
				if !valueValid {
					return false, fmt.Errorf(
						"user_ipt value type wrong. ipt_index: %d, component_idex: %d, needed_value_type: %s, get_value: %v",
						iptIndex, componentIndex, componentParamConfig.ValueType, componentParamConfig.Value,
					)
				}
			}
		}
	}
	return true, nil
}

type Flow struct {
	ID                            value_object.UUID
	Name                          string
	IsDraft                       bool
	Deleted                       bool
	Version                       uint
	OriginID                      value_object.UUID
	Newest                        bool
	CreateUserID                  value_object.UUID
	CreateUserName                string
	CreateTime                    time.Time
	Position                      interface{}
	FlowFunctionIDMapFlowFunction map[string]*FlowFunction
	// 是否允许正在运行时再次运行
	AllowParallelRun bool
	// 运行触发
	Crontab           *crontab.CrontabRepresent
	TriggerKey        string
	AllowTriggerByKey bool
	// 重试策略
	TimeoutInSeconds      uint32
	RetryAmount           uint16
	RetryIntervalInSecond uint16
	// 用于权限
	ReadUserIDs             []value_object.UUID
	WriteUserIDs            []value_object.UUID
	ExecuteUserIDs          []value_object.UUID
	DeleteUserIDs           []value_object.UUID
	AssignPermissionUserIDs []value_object.UUID
}

func (flow *Flow) IsZero() bool {
	if flow == nil {
		return true
	}
	return flow.ID.IsNil()
}

func (flow *Flow) LinedFlowIDs() []string {
	if flow == nil {
		return []string{}
	}
	if len(flow.FlowFunctionIDMapFlowFunction[config.FlowFunctionStartID].DownstreamFlowFunctionIDs) == 0 {
		return []string{}
	}

	ids := make([]string, 0, len(flow.FlowFunctionIDMapFlowFunction)-1)
	var getNodeDownstreamFlowFunctionIDs func(flowFuncID string, flowFuncIDSlice *[]string)
	getNodeDownstreamFlowFunctionIDs = func(
		flowFuncID string, flowFuncIDSlice *[]string,
	) {
		flowFunc, ok := flow.FlowFunctionIDMapFlowFunction[flowFuncID]
		if !ok {
			return
		}
		if flowFuncID != config.FlowFunctionStartID {
			(*flowFuncIDSlice) = append((*flowFuncIDSlice), flowFuncID)
		}
		for _, downstreamFlowFuncID := range flowFunc.DownstreamFlowFunctionIDs {
			getNodeDownstreamFlowFunctionIDs(downstreamFlowFuncID, flowFuncIDSlice)
		}
	}
	getNodeDownstreamFlowFunctionIDs(config.FlowFunctionStartID, &ids)
	return ids
}

func (flow *Flow) HaveRetryStrategy() bool {
	if flow.IsZero() {
		return false
	}
	return flow.RetryAmount > 0 && flow.RetryIntervalInSecond > 0
}

func (flow *Flow) UserCanRead(user *User) bool {
	if user.IsSuper {
		return true
	}
	userID := user.ID
	for _, readAbleUserID := range flow.ReadUserIDs {
		if readAbleUserID == userID {
			return true
		}
	}
	return false
}

func (flow *Flow) UserCanWrite(user *User) bool {
	if user.IsSuper {
		return true
	}
	userID := user.ID
	for _, writeAbleUserID := range flow.WriteUserIDs {
		if writeAbleUserID == userID {
			return true
		}
	}
	return false
}

func (flow *Flow) UserCanExecute(user *User) bool {
	if user.IsSuper {
		return true
	}
	userID := user.ID
	for _, exeAbleUserID := range flow.ExecuteUserIDs {
		if exeAbleUserID == userID {
			return true
		}
	}
	return false
}

func (flow *Flow) UserCanDelete(user *User) bool {
	if user.IsSuper {
		return true
	}
	userID := user.ID
	for _, uID := range flow.DeleteUserIDs {
		if uID == userID {
			return true
		}
	}
	return false
}

func (flow *Flow) UserCanAssignPermission(user *User) bool {
	if user.IsSuper {
		return true
	}
	userID := user.ID
	for _, uID := range flow.AssignPermissionUserIDs {
		if uID == userID {
			return true
		}
	}
	return false
}
