package bloc

func (blocApp *BlocApp) RunConsumer() {
	// 监听发布flow运行任务消息的consumer
	go blocApp.FlowTaskStartConsumer()

	// 监听发布flow运行完成消息的consumer
	go blocApp.FlowTaskFinishedConsumer()

	//监听bloc task完成消息的consumer
	go blocApp.FunctionRunConsumer()

	//监听是否有运行bloc中途因为各项原因退出了的bloc，触发重新运行
	// go blocApp.LookDeadBlocAndRepub()

	// crontab watcher
	go blocApp.CrontabWatcher()
}
