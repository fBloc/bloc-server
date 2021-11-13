package flow

import (
	"github.com/fBloc/bloc-backend-go/aggregate"
	"github.com/fBloc/bloc-backend-go/internal/crontab"

	"github.com/google/uuid"
)

type FlowRepository interface {
	// Create
	CreateOnlineFromDraft(
		draftF *aggregate.Flow,
	) (*aggregate.Flow, error)
	CreateDraftFromScratch(
		name string,
		createUserID uuid.UUID,
		position interface{},
		funcs map[string]*aggregate.FlowFunction,
	) (*aggregate.Flow, error)
	CreateDraftFromExistFlow(
		name string,
		createUserID, originID uuid.UUID,
		position interface{},
		funcs map[string]*aggregate.FlowFunction,
	) (*aggregate.Flow, error)

	// Read
	GetByID(id uuid.UUID) (*aggregate.Flow, error)
	GetByIDStr(id string) (*aggregate.Flow, error)

	GetOnlineByOriginID(originID uuid.UUID) (*aggregate.Flow, error)
	GetOnlineByOriginIDStr(originID string) (*aggregate.Flow, error)
	GetDraftByOriginID(originID uuid.UUID) (*aggregate.Flow, error)

	FilterOnline(userID uuid.UUID, nameContains string) (flows []aggregate.Flow, err error)
	FilterCrontabFlows() (flows []aggregate.Flow, err error)
	FilterDraft(userID uuid.UUID, nameContains string) (flows []aggregate.Flow, err error)

	// Update
	PatchName(id uuid.UUID, name string) error
	PatchPosition(id uuid.UUID, position interface{}) error
	// PatchFuncs(id uuid.UUID, funcs map[string]*flow_bloc.) error
	PatchCrontab(id uuid.UUID, c crontab.CrontabRepresent) error
	PatchPubWhileRunning(id uuid.UUID, pub bool) error
	PatchRetryStrategy(id uuid.UUID, amount, intervalInSecond uint16) error
	PatchTriggerKey(id uuid.UUID, key string) error
	PatchTimeout(id uuid.UUID, tOS uint32) error
	// replace
	ReplaceByID(id uuid.UUID, aggFlow *aggregate.Flow) error

	// update user permission
	AddReader(id, userID uuid.UUID) error
	DeleteReader(id, userID uuid.UUID) error
	AddWriter(id, userID uuid.UUID) error
	DeleteWriter(id, userID uuid.UUID) error
	AddExecuter(id, userID uuid.UUID) error
	DeleteExecuter(id, userID uuid.UUID) error
	AddSuper(id, userID uuid.UUID) error
	DeleteSuper(id, userID uuid.UUID) error

	// Delete
	DeleteByID(id uuid.UUID) (int64, error)
	DeleteByOriginID(originID uuid.UUID) (int64, error)
	DeleteDraftByOriginID(originID uuid.UUID) (int64, error)

	// 运行相关

}
