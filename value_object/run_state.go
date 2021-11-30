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
	maxRunStatus
)

func (tS RunState) IsRunFinished() bool {
	return tS > Running
}

func (tS RunState) IsRunStateValid() bool {
	return tS > 0 && tS < maxRunStatus
}
