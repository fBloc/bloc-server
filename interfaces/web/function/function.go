package function

import (
	"github.com/fBloc/bloc-backend-go/aggregate"
	"github.com/fBloc/bloc-backend-go/pkg/ipt"
	"github.com/fBloc/bloc-backend-go/pkg/opt"
	"github.com/fBloc/bloc-backend-go/services/function"
	"github.com/fBloc/bloc-backend-go/value_object"

	"github.com/pkg/errors"
)

var fService *function.FunctionService

func InjectFunctionService(fS *function.FunctionService) {
	fService = fS
}

// MakeSureAllUserImplementFunctionsInRepository 确保用户注册进来的函数已经在存储层了
func MakeSureAllUserImplementFunctionsInRepository(
	functions []*aggregate.Function,
) error {
	if fService == nil {
		return errors.New(
			"must inject functionService in web.function first")
	}
	for index, aggFunction := range functions {
		sameCoreIns, err := fService.Function.GetSameIptOptFunction(
			aggFunction.IptDigest, aggFunction.OptDigest)
		if err != nil {
			return err
		}
		if !sameCoreIns.IsZero() { // 已经在存储层了，修改对应的ID
			functions[index].ID = sameCoreIns.ID
		} else { // 不在存储层，持久化之
			functions[index].ID = value_object.NewUUID()
			err := fService.Function.Create(aggFunction)
			if err != nil {
				return errors.Wrap(err,
					"persistent user impolement function to repository layer failed")
			}
		}
	}
	return nil
}

type Function struct {
	ID          value_object.UUID `json:"id"`
	Name        string            `json:"name"`
	GroupName   string            `json:"group_name"`
	Description string            `json:"description"`
	Ipt         ipt.IptSlice      `json:"ipt"`
	Opt         []*opt.Opt        `json:"opt"`
}

func newFunctionFromAgg(aggF *aggregate.Function) *Function {
	if aggF.IsZero() {
		return nil
	}
	return &Function{
		ID:          aggF.ID,
		Name:        aggF.Name,
		GroupName:   aggF.GroupName,
		Description: aggF.Description,
		Ipt:         aggF.Ipts,
		Opt:         aggF.Opts}
}

type GroupFunctions struct {
	GroupName string     `json:"group_name"`
	Functions []Function `json:"functions"`
}

func newGroupedFunctionsFromAggFunctions(
	aggFuncs []*aggregate.Function,
) ([]GroupFunctions, error) {
	groupNameMapGroup := make(map[string]*GroupFunctions)
	for _, aggF := range aggFuncs {
		f := newFunctionFromAgg(aggF)
		if f == nil {
			continue
		}
		if _, ok := groupNameMapGroup[f.GroupName]; !ok {
			groupNameMapGroup[f.GroupName] = &GroupFunctions{
				GroupName: f.GroupName,
			}
		}
		groupNameMapGroup[f.GroupName].Functions = append(
			groupNameMapGroup[f.GroupName].Functions, *f)
	}
	ret := make([]GroupFunctions, 0, len(groupNameMapGroup))
	for _, groupFuncs := range groupNameMapGroup {
		ret = append(ret, *groupFuncs)
	}
	return ret, nil
}
