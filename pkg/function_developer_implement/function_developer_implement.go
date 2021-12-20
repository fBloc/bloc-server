package function_developer_implement

import (
	"context"

	"github.com/fBloc/bloc-server/infrastructure/log"
	"github.com/fBloc/bloc-server/pkg/ipt"
	"github.com/fBloc/bloc-server/pkg/opt"
	"github.com/fBloc/bloc-server/value_object"
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
