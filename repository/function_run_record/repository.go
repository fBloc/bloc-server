package function_run_record

import (
	"github.com/fBloc/bloc-backend-go/aggregate"
	"github.com/fBloc/bloc-backend-go/infrastructure/object_storage"
	"github.com/fBloc/bloc-backend-go/internal/filter_options"
	"github.com/fBloc/bloc-backend-go/value_object"
)

type FunctionRunRecordRepository interface {
	// Create
	Create(*aggregate.FunctionRunRecord) error

	// Read
	GetByID(id value_object.UUID) (*aggregate.FunctionRunRecord, error)
	Filter(
		filter value_object.RepositoryFilter,
		filterOption value_object.RepositoryFilterOption,
	) ([]*aggregate.FunctionRunRecord, error)
	FilterByFilterOption(
		kv map[string]interface{},
		filterOptions *filter_options.FilterOption,
	) ([]*aggregate.FunctionRunRecord, error)
	FilterByFlowRunRecordID(
		FlowRunRecordID value_object.UUID,
	) ([]*aggregate.FunctionRunRecord, error)

	// Update
	PatchProgress(id value_object.UUID, progress float32) error
	PatchProgressMsg(id value_object.UUID, progressMsg string) error
	PatchStageIndex(
		id value_object.UUID, progressStageIndex int,
	) error
	PatchProgressStages(
		id value_object.UUID, progressStages []string,
	) error
	SaveIptBrief(
		id value_object.UUID, ipts [][]interface{},
		objectStorageImplement object_storage.ObjectStorage,
	) error

	ClearProgress(id value_object.UUID) error
	SaveSuc(
		id value_object.UUID, desc string,
		opt map[string]string,
		brief map[string]string, intercepted bool,
	) error
	SaveCancel(id value_object.UUID) error
	SaveFail(id value_object.UUID, errMsg string) error

	// Delete
}
