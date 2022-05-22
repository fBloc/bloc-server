package bloc

func (blocApp *BlocApp) RunScheduler() {
	// 监听发布flow运行任务消息的consumer
	go blocApp.FlowTaskStartConsumer()

	// 监听发布flow运行完成消息的consumer
	go blocApp.FlowTaskFinishedConsumer()

	//监听bloc task完成消息的consumer
	go blocApp.FunctionRunConsumer()

	//监听是否有运行中途因为各项原因退出了的function，触发其重新运行
	go blocApp.RePubDeadRuns()

	// crontab watcher
	go blocApp.CrontabWatcher()
}
