package function

import (
	"github.com/fBloc/bloc-backend-go/aggregate"
	"github.com/fBloc/bloc-backend-go/value_object"
)

type FunctionRepository interface {
	// create
	Create(f *aggregate.Function) error

	// read
	GetByID(id value_object.UUID) (*aggregate.Function, error)
	GetSameIptOptFunction(iptDigest, optDigest string) (*aggregate.Function, error)
	All() ([]*aggregate.Function, error)
	IDMapFunctionAll() (map[value_object.UUID]*aggregate.Function, error)

	// update
	PatchName(id value_object.UUID, name string) error
	PatchDescription(id value_object.UUID, desc string) error
	PatchGroupName(id value_object.UUID, groupName string) error
	PatchProviderName(id value_object.UUID, provider string) error
	AliveReport(id value_object.UUID) error

	// update user permission
	AddReader(id, userID value_object.UUID) error
	RemoveReader(id, userID value_object.UUID) error
	AddExecuter(id, userID value_object.UUID) error
	RemoveExecuter(id, userID value_object.UUID) error
	AddAssigner(id, userID value_object.UUID) error
	RemoveAssigner(id, userID value_object.UUID) error

	// delete
}
