package bloc

import (
	"fmt"

	"github.com/fBloc/bloc-backend-go/aggregate"
	"github.com/fBloc/bloc-backend-go/pkg/function_developer_implement"
	"github.com/fBloc/bloc-backend-go/pkg/ipt"
	"github.com/fBloc/bloc-backend-go/pkg/opt"
	"github.com/fBloc/bloc-backend-go/value_object"
)

type FunctionGroup struct {
	blocApp   *BlocApp
	Name      string
	Functions []*aggregate.Function
}

func (functionGroup *FunctionGroup) AddFunction(
	name string,
	description string,
	userImplementedFunc function_developer_implement.FunctionDeveloperImplementInterface) {
	for _, function := range functionGroup.Functions {
		if function.Name == name {
			errorInfo := fmt.Sprintf(
				"should not have same function name(%s) under same group(%s)",
				name, functionGroup.Name)
			panic(errorInfo)
		}
	}

	aggFunction := aggregate.Function{
		Name:          name,
		GroupName:     functionGroup.Name,
		Description:   description,
		Ipts:          userImplementedFunc.IptConfig(),
		IptDigest:     ipt.GenIptDigest(userImplementedFunc.IptConfig()),
		Opts:          userImplementedFunc.OptConfig(),
		OptDigest:     opt.GenOptDigest(userImplementedFunc.OptConfig()),
		ProcessStages: userImplementedFunc.AllProcessStages(),
		ExeFunc:       userImplementedFunc}

	funcRepo := functionGroup.blocApp.GetOrCreateFunctionRepository()
	sameIns, err := funcRepo.GetSameIptOptFunction(
		aggFunction.IptDigest, aggFunction.OptDigest,
	)
	if err != nil {
		panic(err)
	}
	if sameIns.IsZero() {
		aggFunction.ID = value_object.NewUUID()
		err = funcRepo.Create(&aggFunction)
		if err != nil {
			panic(err)
		}
	} else {
		aggFunction.ID = sameIns.ID
	}

	if functionGroup.blocApp.functionRepoIDMapExecuteFunction == nil {
		functionGroup.blocApp.functionRepoIDMapExecuteFunction = make(map[value_object.UUID]function_developer_implement.FunctionDeveloperImplementInterface)
	}
	functionGroup.blocApp.functionRepoIDMapExecuteFunction[aggFunction.ID] = userImplementedFunc
	functionGroup.Functions = append(functionGroup.Functions, &aggFunction)
}
