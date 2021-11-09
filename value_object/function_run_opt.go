package value_object

import "sync"

type FunctionRunOpt struct {
	Suc         bool
	Canceled    bool
	Pass        bool
	ErrorMsg    string
	Description string
	Detail      map[string]interface{}
	Brief       map[string]string
	sync.Mutex
}

func CanceldBlocOpt() *FunctionRunOpt {
	return &FunctionRunOpt{Canceled: true}
}
