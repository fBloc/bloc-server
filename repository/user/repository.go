package user

import (
	"github.com/fBloc/bloc-server/aggregate"
	"github.com/fBloc/bloc-server/value_object"
)

type UserRepository interface {
	// create
	Create(aggregate.User) error

	// read
	GetByName(name string) (*aggregate.User, error)
	GetByID(id value_object.UUID) (*aggregate.User, error)
	All() (users []aggregate.User, err error)
	FilterByNameContains(nameContains string) (users []aggregate.User, err error)

	// update
	PatchName(id value_object.UUID, name string) error

	// delete
	DeleteByID(id value_object.UUID) (int64, error)
}
