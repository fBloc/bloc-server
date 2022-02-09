package flow

import (
	"github.com/fBloc/bloc-server/aggregate"
	"github.com/fBloc/bloc-server/internal/crontab"
	"github.com/fBloc/bloc-server/value_object"
)

type FlowRepository interface {
	// Create
	CreateOnlineFromDraft(
		draftF *aggregate.Flow,
	) (*aggregate.Flow, error)
	CreateDraftFromScratch(
		name string,
		createUserID value_object.UUID,
		position interface{},
		funcs map[string]*aggregate.FlowFunction,
	) (*aggregate.Flow, error)
	CreateDraftFromExistFlow(
		name string,
		createUserID, originID value_object.UUID,
		position interface{},
		funcs map[string]*aggregate.FlowFunction,
	) (*aggregate.Flow, error)

	// Read
	GetByID(id value_object.UUID) (*aggregate.Flow, error)
	GetByIDStr(id string) (*aggregate.Flow, error)

	GetOnlineByOriginID(originID value_object.UUID) (*aggregate.Flow, error)
	GetOnlineByOriginIDStr(originID string) (*aggregate.Flow, error)
	GetDraftByOriginID(originID value_object.UUID) (*aggregate.Flow, error)

	FilterOnline(user *aggregate.User, nameContains string) (flows []aggregate.Flow, err error)
	FilterCrontabFlows() (flows []aggregate.Flow, err error)
	FilterDraft(userID value_object.UUID, nameContains string) (flows []aggregate.Flow, err error)

	// Update
	PatchName(id value_object.UUID, name string) error
	PatchPosition(id value_object.UUID, position interface{}) error
	// PatchFuncs(id value_object.UUID, funcs map[string]*flow_bloc.) error
	PatchCrontab(id value_object.UUID, c crontab.CrontabRepresent) error
	PatchAllowParallelRun(id value_object.UUID, pub bool) error
	PatchRetryStrategy(id value_object.UUID, amount, intervalInSecond uint16) error
	PatchTriggerKey(id value_object.UUID, key string) error
	PatchTimeout(id value_object.UUID, tOS uint32) error
	// replace
	ReplaceByID(id value_object.UUID, aggFlow *aggregate.Flow) error

	// update user permission
	AddReader(id, userID value_object.UUID) error
	RemoveReader(id, userID value_object.UUID) error
	AddWriter(id, userID value_object.UUID) error
	RemoveWriter(id, userID value_object.UUID) error
	AddExecuter(id, userID value_object.UUID) error
	RemoveExecuter(id, userID value_object.UUID) error
	AddDeleter(id, userID value_object.UUID) error
	RemoveDeleter(id, userID value_object.UUID) error
	AddAssigner(id, userID value_object.UUID) error
	RemoveAssigner(id, userID value_object.UUID) error

	// Delete
	DeleteByID(id value_object.UUID) (int64, error)
	DeleteByOriginID(originID value_object.UUID) (int64, error)
	DeleteDraftByOriginID(originID value_object.UUID) (int64, error)

	// 运行相关

}
