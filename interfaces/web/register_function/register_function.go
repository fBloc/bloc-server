package register_function

import (
	"time"

	"github.com/fBloc/bloc-backend-go/pkg/ipt"
	"github.com/fBloc/bloc-backend-go/pkg/opt"
	"github.com/fBloc/bloc-backend-go/services/function"
	"github.com/fBloc/bloc-backend-go/value_object"
)

var fService *function.FunctionService

func InjectFunctionService(fS *function.FunctionService) {
	fService = fS
}

type reportFunction struct {
	ConsumerFlag   string
	GroupName      string
	Name           string
	ID             value_object.UUID
	LastReportTime time.Time
}

var groupNameMapFuncNameMapFunc = make(map[string]map[string]*reportFunction)

type RegisterFuncReq struct {
	Flag                        string                     `json:"flag"`
	GroupNameMapFuncNameMapFunc map[string][]*HttpFunction `json:"groupName_map_functionName_map_function"`
}
type GroupNameMapFunctions map[string][]*HttpFunction

type HttpFunction struct {
	ID            value_object.UUID `json:"id"`
	Name          string            `json:"name"`
	GroupName     string            `json:"group_name"`
	Description   string            `json:"description"`
	Ipts          []*ipt.Ipt        `json:"ipts"`
	Opts          []*opt.Opt        `json:"opts"`
	ProcessStages []string          `json:"process_stages"`
	ErrorMsg      string            `json:"error_msg"`
}
