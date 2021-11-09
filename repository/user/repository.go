package user

import (
	"github.com/fBloc/bloc-backend-go/aggregate"

	"github.com/google/uuid"
)

type UserRepository interface {
	// create
	Create(aggregate.User) error

	// read
	GetByName(name string) (*aggregate.User, error)
	GetByID(id uuid.UUID) (*aggregate.User, error)
	All() (users []aggregate.User, err error)
	FilterByNameContains(nameContains string) (users []aggregate.User, err error)

	// update
	PatchName(id uuid.UUID, name string) error

	// delete
	DeleteByID(id uuid.UUID) (int64, error)
}
