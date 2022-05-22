package value_object

type RunState int

const (
	UnknownRunState RunState = iota
	Created
	InQueue
	Running
	UserCanceled
	TimeoutCanceled
	Suc
	Fail
	InterceptedCancel
	NotAllowedParallelCancel
	ToSchedule // upper functions not finished. Waiting to schedule
	maxRunStatus
)

func (tS RunState) IsRunFinished() bool {
	return tS > Running && tS < ToSchedule
}

func (tS RunState) IsRunStateValid() bool {
	return tS > 0 && tS < maxRunStatus
}

func NotFinishedRunStatus() []RunState {
	return []RunState{ToSchedule, Created, InQueue, Running}
}
