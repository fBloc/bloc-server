package config

import "time"

const (
	DefaultUserName        = "bloc"
	DefaultUserPassword    = "maytheforcebewithyou"
	DefaultLogKeepDays     = 60
	FlowFunctionStartID    = "0"
	FunctionReportInterval = time.Second * 5
	FunctionReportTimeout  = 4 * FunctionReportInterval
)
