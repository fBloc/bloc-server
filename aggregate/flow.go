package aggregate

import (
	"fmt"
	"time"

	"github.com/fBloc/bloc-backend-go/internal/crontab"
	"github.com/fBloc/bloc-backend-go/pkg/value_type"
	"github.com/fBloc/bloc-backend-go/value_object"
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
}

// CheckValid 检测此function node的配置是否正确且有效
func (flowFunc *FlowFunction) CheckValid(
	flowFuncIDMapFlowFunction map[string]*FlowFunction,
) (bool, string) {
	// TODO 优化错误输出的提示信息，要能够提供精确的信息
	// 所有节点都应该要有输入节点
	if len(flowFunc.UpstreamFlowFunctionIDs) <= 0 {
		return false, "flowFunc must have input flowFunc"
	}

	// 节点上游id必须是在数据里的（防止前端传过来的数据不对）
	for _, funcID := range flowFunc.UpstreamFlowFunctionIDs {
		if _, ok := flowFuncIDMapFlowFunction[funcID]; !ok {
			return false, fmt.Sprintf("节点上游id(%s)无效", funcID)
		}
	}

	// 节点下游id必须是在数据里的（防止前端传过来的数据不对）
	for _, funcID := range flowFunc.DownstreamFlowFunctionIDs {
		if _, ok := flowFuncIDMapFlowFunction[funcID]; !ok {
			return false, fmt.Sprintf("节点下游id(%s)无效", funcID)
		}
	}

	// 检测输入参数的类型是否正确
	for iptIndex, iptParamConfig := range flowFunc.ParamIpts { // 对应到ipt
		for componentIndex, componentParamConfig := range iptParamConfig { // 对应到component
			// 若是必填参数、但没有配置参数
			if componentParamConfig.Blank {
				thisParamIpt := flowFunc.Function.Ipts[iptIndex]
				if thisParamIpt.Must {
					return false, fmt.Sprintf(
						"节点参数序列为%d - %d必填",
						iptIndex, componentIndex)
				}
				continue
			}

			if componentParamConfig.IptWay == value_object.Connection { // 配置的参数输入方式是链接
				// connection写的节点ID是否存在
				if _, ok := flowFuncIDMapFlowFunction[componentParamConfig.FlowFunctionID]; !ok {
					return false, fmt.Sprintf(
						"节点参数序列为%d - %d设置的输入链接节点id(%s)无效-没有此id的节点",
						iptIndex, componentIndex, componentParamConfig.FlowFunctionID,
					)
				}

				// 2.connection写的节点ID需要是此节点的直接上游
				var isUpstreamFuncNodeID bool
				for _, upFuncNodeID := range flowFunc.UpstreamFlowFunctionIDs {
					if upFuncNodeID == componentParamConfig.FlowFunctionID {
						isUpstreamFuncNodeID = true
						break
					}
				}
				if !isUpstreamFuncNodeID {
					return false, "节点写的参数输入链接并不是节点的上游节点，无效"
				}

				// 3. 检查对应出参的类型是不是和此参数的输入要求类型一致
				// TODO 这里的iptNode.Function为nil，需要想办法处理附上function实例才能检查
				// iptNode := flowFuncIDMapFlowFunction[componentParamConfig.FlowFunctionID]
				// for _, optItem := range iptNode.Function.Opts {
				// 	if optItem.Key != componentParamConfig.Key {
				// 		continue
				// 	}
				// 	if optItem.ValueType != componentParamConfig.ValueType {
				// 		return false, "connection value type wrong"
				// 	}
				// 	return true, ""
				// }
			} else if componentParamConfig.IptWay == value_object.UserIpt { // 配置的参数输入方式是用户输入的值
				valueValid := value_type.CheckValueTypeValueValid(
					componentParamConfig.ValueType, componentParamConfig.Value)
				if !valueValid {
					return false, "user_ipt value type wrong"
				}
			} else {
				// TODO 这里的值user_ipt/connection不应该写死
				return false, "ipt config type not right(should be in user_ipt/connection)"
			}
		}
	}
	return true, ""
}

type Flow struct {
	ID                            value_object.UUID
	Name                          string
	IsDraft                       bool
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
	Crontab    crontab.CrontabRepresent
	TriggerKey string
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

func (flow *Flow) HaveRetryStrategy() bool {
	if flow == nil {
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
