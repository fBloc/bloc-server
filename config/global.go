package config

import "time"

const (
	DefaultUserName        = "bloc"
	DefaultUserPassword    = "maytheforcebewithyou"
	DefaultLogKeepDays     = 60
	FlowFunctionStartID    = "0"
	FunctionReportInterval = time.Second * 30 // this interval is register function interval, not heartbeat interval
	FunctionReportTimeout  = 3 * FunctionReportInterval
)
