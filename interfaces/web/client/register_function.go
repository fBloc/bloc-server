package client

import (
	"sync"
	"time"

	"github.com/fBloc/bloc-server/aggregate"
	"github.com/fBloc/bloc-server/pkg/ipt"
	"github.com/fBloc/bloc-server/pkg/opt"
	"github.com/fBloc/bloc-server/services/function"
	"github.com/fBloc/bloc-server/value_object"
)

var fService *function.FunctionService

func InjectFunctionService(fS *function.FunctionService) {
	fService = fS
}

type reportFunction struct {
	ID                 value_object.UUID
	Name               string
	GroupName          string
	ProviderName       string
	ProgressMilestones []string
	LastReportTime     time.Time
}

type reportedGroupNameMapFuncNameMapFunc struct {
	groupNameMapFuncNameMapFunc map[string]map[string]*reportFunction
	idMapFunc                   map[value_object.UUID]aggregate.Function
	sync.Mutex
}

var reported = reportedGroupNameMapFuncNameMapFunc{
	groupNameMapFuncNameMapFunc: make(map[string]map[string]*reportFunction),
	idMapFunc:                   make(map[value_object.UUID]aggregate.Function),
}

type RegisterFuncReq struct {
	Who                   string                     `json:"who"`
	GroupNameMapFunctions map[string][]*HttpFunction `json:"groupName_map_functions"`
}
type GroupNameMapFunctions map[string][]*HttpFunction

// HttpFunction
// follow fields should not be provided in request but will be present in return data:
// 	ID: the server will gen the ID and return it when first register.
// 	ErrorMsg: like when function register failed. will put error msg in this field to return.
type HttpFunction struct {
	ID                 value_object.UUID `json:"id"`
	Name               string            `json:"name"`
	GroupName          string            `json:"group_name"`
	Description        string            `json:"description"`
	Ipts               []*ipt.Ipt        `json:"ipts"`
	Opts               []*opt.Opt        `json:"opts"`
	ProgressMilestones []string          `json:"progress_milestones"`
	ErrorMsg           string            `json:"error_msg"`
}
