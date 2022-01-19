package value_object

type LogType int

const (
	UnknownLogType LogType = iota
	HttpServerLog
	ScheduleLog
	FuncRunRecordLog
	maxLogType
)

func (l LogType) String() string {
	switch l {
	case HttpServerLog:
		return "http-server"
	case ScheduleLog:
		return "schedule"
	case FuncRunRecordLog:
		return "func-run-record"
	default:
		return "unknown"
	}
}

func (l LogType) IsValid() bool {
	return int(l) < int(maxLogType) && l != UnknownLogType
}

func AllLogTypes() (resp []LogType) {
	for i := 1; i < int(maxLogType); i++ {
		resp = append(resp, LogType(i))
	}
	return
}
