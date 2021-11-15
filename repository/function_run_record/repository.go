package function_run_record

import (
	"github.com/fBloc/bloc-backend-go/aggregate"
	"github.com/fBloc/bloc-backend-go/infrastructure/object_storage"
	"github.com/fBloc/bloc-backend-go/internal/filter_options"
	"github.com/fBloc/bloc-backend-go/value_object"

	"github.com/google/uuid"
)

type FunctionRunRecordRepository interface {
	// Create
	Create(*aggregate.FunctionRunRecord) error

	// Read
	GetByID(id uuid.UUID) (*aggregate.FunctionRunRecord, error)
	Filter(
		filter value_object.RepositoryFilter,
		filterOption value_object.RepositoryFilterOption,
	) ([]*aggregate.FunctionRunRecord, error)
	FilterByFilterOption(
		kv map[string]interface{},
		filterOptions *filter_options.FilterOption,
	) ([]*aggregate.FunctionRunRecord, error)
	FilterByFlowRunRecordID(
		FlowRunRecordID uuid.UUID,
	) ([]*aggregate.FunctionRunRecord, error)

	// Update
	PatchProgress(id uuid.UUID, progress float32) error
	PatchProgressMsg(id uuid.UUID, progressMsg string) error
	PatchStageIndex(
		id uuid.UUID, progressStageIndex int,
	) error
	PatchProgressStages(
		id uuid.UUID, progressStages []string,
	) error
	SaveIptBrief(
		id uuid.UUID, ipts [][]interface{},
		objectStorageImplement object_storage.ObjectStorage,
	) error

	ClearProgress(id uuid.UUID) error
	SaveSuc(
		id uuid.UUID, desc string, opt map[string]interface{},
		brief map[string]string, pass bool,
	) error
	SaveCancel(id uuid.UUID) error
	SaveFail(id uuid.UUID, errMsg string) error

	// Delete
}
