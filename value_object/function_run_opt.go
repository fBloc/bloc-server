package value_object

import (
	"sync"
)

type FunctionRunOpt struct {
	Suc                       bool
	Canceled                  bool
	InterceptBelowFunctionRun bool // 拦截后续的运行
	ErrorMsg                  string
	Description               string
	Detail                    map[string]interface{}
	Brief                     map[string]string
	sync.Mutex
}

func CanceldBlocOpt() *FunctionRunOpt {
	return &FunctionRunOpt{Canceled: true}
}
