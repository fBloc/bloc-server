package value_object

// 表示函数运行进度
type FunctionRunningProgress struct {
	Progress          float32
	Msg               string
	ProcessStageIndex int
}
