package value_object

type FlowTriggeredSourceType int

const (
	UnknownSourceType FlowTriggeredSourceType = iota
	ArrangementTriggerSource
	FlowTriggerSource
)
