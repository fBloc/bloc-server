package function

import (
	"github.com/fBloc/bloc-backend-go/aggregate"

	"github.com/google/uuid"
)

type FunctionRepository interface {
	// create
	Create(f *aggregate.Function) error

	// read
	GetByID(id uuid.UUID) (*aggregate.Function, error)
	GetSameIptOptFunction(iptDigest, optDigest string) (*aggregate.Function, error)
	All() ([]*aggregate.Function, error)
	IDMapFunctionAll() (map[uuid.UUID]*aggregate.Function, error)

	// update
	PatchName(id uuid.UUID, name string) error
	PatchDescription(id uuid.UUID, desc string) error
	PatchGroupName(id uuid.UUID, groupName string) error

	// update user permission
	AddReader(id, userID uuid.UUID) error
	RemoveReader(id, userID uuid.UUID) error
	AddExecuter(id, userID uuid.UUID) error
	RemoveExecuter(id, userID uuid.UUID) error
	AddAssigner(id, userID uuid.UUID) error
	RemoveAssigner(id, userID uuid.UUID) error

	// delete
}
