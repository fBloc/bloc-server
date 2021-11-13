package flow_run_record

import (
	"time"

	"github.com/fBloc/bloc-backend-go/aggregate"
	"github.com/fBloc/bloc-backend-go/value_object"

	"github.com/google/uuid"
)

type FlowRunRecordRepository interface {
	// Create
	Create(*aggregate.FlowRunRecord) error
	CrontabFindOrCreate(
		fRR *aggregate.FlowRunRecord, crontabTime time.Time,
	) (created bool, err error)

	// Read
	GetByID(id uuid.UUID) (*aggregate.FlowRunRecord, error)
	ReGetToCheckIsCanceled(id uuid.UUID) bool

	GetLatestByFlowOriginID(flowOriginID uuid.UUID) (*aggregate.FlowRunRecord, error)
	GetLatestByFlowID(flowID uuid.UUID) (*aggregate.FlowRunRecord, error)
	GetLatestByArrangementFlowID(arrangementFlowID string) (*aggregate.FlowRunRecord, error)

	Filter(
		filter value_object.RepositoryFilter,
		filterOption value_object.RepositoryFilterOption,
	) ([]*aggregate.FlowRunRecord, error)
	AllRunRecordOfFlowTriggeredByFlowID(
		flowID uuid.UUID,
	) ([]*aggregate.FlowRunRecord, error)

	// Update
	PatchDataForRetry(id uuid.UUID, retriedAmount uint16) error
	PatchFlowFuncIDMapFuncRunRecordID(
		id uuid.UUID,
		FlowFuncIDMapFuncRunRecordID map[string]uuid.UUID,
	) error
	AddFlowFuncIDMapFuncRunRecordID(
		id uuid.UUID,
		flowFuncID string,
		funcRunRecordID uuid.UUID,
	) error
	Start(id uuid.UUID) error
	Suc(id uuid.UUID) error
	Fail(id uuid.UUID, errorMsg string) error
	TimeoutCancel(id uuid.UUID) error
	UserCancel(id, userID uuid.UUID) error

	// Delete
}
