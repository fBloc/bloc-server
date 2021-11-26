package function_developer_implement

import (
	"context"

	"github.com/fBloc/bloc-backend-go/infrastructure/log"
	"github.com/fBloc/bloc-backend-go/pkg/ipt"
	"github.com/fBloc/bloc-backend-go/pkg/opt"
	"github.com/fBloc/bloc-backend-go/value_object"
)

type FunctionDeveloperImplementInterface interface {
	Run(
		context.Context,
		ipt.IptSlice,
		chan value_object.FunctionRunStatus,
		chan *value_object.FunctionRunOpt,
		*log.Logger,
	)
	IptConfig() ipt.IptSlice
	OptConfig() []*opt.Opt
	AllProcessStages() []string
}
