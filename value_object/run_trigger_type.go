package value_object

type TriggerType int

const (
	UnknownTriggerType TriggerType = iota
	Manual
	Crontab
	Key
)
