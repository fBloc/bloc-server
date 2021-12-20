package flow_run_record

import (
	"time"

	"github.com/fBloc/bloc-backend-go/aggregate"
	"github.com/fBloc/bloc-backend-go/value_object"
)

type FlowRunRecordRepository interface {
	// Create
	Create(*aggregate.FlowRunRecord) error
	CrontabFindOrCreate(
		fRR *aggregate.FlowRunRecord, crontabTime time.Time,
	) (created bool, err error)

	// Read
	GetByID(id value_object.UUID) (*aggregate.FlowRunRecord, error)
	ReGetToCheckIsCanceled(id value_object.UUID) bool

	GetLatestByFlowOriginID(flowOriginID value_object.UUID) (*aggregate.FlowRunRecord, error)
	GetLatestByFlowID(flowID value_object.UUID) (*aggregate.FlowRunRecord, error)
	GetLatestByArrangementFlowID(arrangementFlowID string) (*aggregate.FlowRunRecord, error)

	Filter(
		filter value_object.RepositoryFilter,
		filterOption value_object.RepositoryFilterOption,
	) ([]*aggregate.FlowRunRecord, error)
	Count(
		filter value_object.RepositoryFilter,
	) (int64, error)
	IsHaveRunningTask(
		flowID, thisFlowRunRecordID value_object.UUID,
	) (bool, error)
	AllRunRecordOfFlowTriggeredByFlowID(
		flowID value_object.UUID,
	) ([]*aggregate.FlowRunRecord, error)

	// Update
	PatchDataForRetry(id value_object.UUID, retriedAmount uint16) error
	PatchFlowFuncIDMapFuncRunRecordID(
		id value_object.UUID,
		FlowFuncIDMapFuncRunRecordID map[string]value_object.UUID,
	) error
	AddFlowFuncIDMapFuncRunRecordID(
		id value_object.UUID,
		flowFuncID string,
		funcRunRecordID value_object.UUID,
	) error

	Start(id value_object.UUID) error
	Suc(id value_object.UUID) error
	Fail(id value_object.UUID, errorMsg string) error
	Intercepted(id value_object.UUID, msg string) error
	TimeoutCancel(id value_object.UUID) error
	UserCancel(id, userID value_object.UUID) error
	NotAllowedParallelRun(id value_object.UUID) error

	// Delete
}
