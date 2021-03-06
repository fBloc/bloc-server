package function

import (
	"github.com/fBloc/bloc-server/aggregate"
	"github.com/fBloc/bloc-server/value_object"
)

type FunctionRepository interface {
	// create
	Create(f *aggregate.Function) error
	FindOrCreate(f *aggregate.Function) (alreadyExistFunction *aggregate.Function, err error)

	// read
	GetByIDForCheckAliveTime(id value_object.UUID) (*aggregate.Function, error)
	GetByID(id value_object.UUID) (*aggregate.Function, error)
	GetSameIptOptFunction(iptDigest, optDigest string) (*aggregate.Function, error)
	All(withoutFields []string) ([]*aggregate.Function, error)
	UserReadAbleAll(user *aggregate.User, withoutFields []string) ([]*aggregate.Function, error)
	IDMapFunctionAll() (map[value_object.UUID]*aggregate.Function, error)

	// update
	PatchName(id value_object.UUID, name string) error
	PatchProgressMilestones(id value_object.UUID, progressMilestones []string) error
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
