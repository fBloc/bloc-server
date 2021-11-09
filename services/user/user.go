package user

import (
	"context"
	"errors"

	"github.com/fBloc/bloc-backend-go/aggregate"
	"github.com/fBloc/bloc-backend-go/infrastructure/log"
	minioLog "github.com/fBloc/bloc-backend-go/infrastructure/log/minio"
	"github.com/fBloc/bloc-backend-go/repository/user"
	mongoUser "github.com/fBloc/bloc-backend-go/repository/user/mongo"

	"github.com/google/uuid"
)

type UserConfiguration func(us *UserService) error

type UserService struct {
	logger log.Logger
	user   user.UserRepository
}

func NewUserService(cfgs ...UserConfiguration) (*UserService, error) {
	us := &UserService{}
	for _, cfg := range cfgs {
		err := cfg(us)
		if err != nil {
			return nil, err
		}
	}
	return us, nil
}

func WithUserRepository(uR user.UserRepository) UserConfiguration {
	return func(us *UserService) error {
		us.user = uR
		return nil
	}
}

func WithMongoUserRepository(hosts []string, port int, db, user, password string) UserConfiguration {
	return func(us *UserService) error {
		ur, err := mongoUser.New(
			context.Background(),
			hosts, port, user, password,
			db, mongoUser.DefaultCollectionName,
		)
		if err != nil {
			return err
		}
		us.user = ur
		return nil
	}
}

func WithLogger(logger log.Logger) UserConfiguration {
	return func(us *UserService) error {
		us.logger = logger
		return nil
	}
}

func WithMinioLogger(
	name string, bucketName string,
	addresses []string, key, password string,
) UserConfiguration {
	return func(us *UserService) error {
		us.logger = minioLog.New(name, bucketName, addresses, key, password)
		return nil
	}
}

func (u *UserService) Login(
	name, rawPassword string,
) (suc bool, sameNameUser *aggregate.User, err error) {
	sameNameUser, err = u.user.GetByName(name)
	if err != nil {
		return false, nil, err
	}
	suc, err = sameNameUser.IsRawPasswordMatch(rawPassword)
	return
}

func (u *UserService) GetByName(
	name string,
) (*aggregate.User, error) {
	if name == "" {
		return nil, nil
	}
	return u.user.GetByName(name)
}

func (u *UserService) FilterByNameContains(
	nameContains string,
) ([]aggregate.User, error) {
	if nameContains == "" {
		return u.user.All()
	}
	return u.user.FilterByNameContains(nameContains)
}

func (u *UserService) AddUser(name, rawPassword string, isSuper bool) error {
	if name == "" || rawPassword == "" {
		return errors.New("password & name all must exit")
	}
	sameNameIns, err := u.user.GetByName(name)
	if err != nil {
		return err
	}
	if !sameNameIns.IsZero() {
		return errors.New("cannot create same name user")
	}
	return u.user.Create(aggregate.NewUser(name, rawPassword, isSuper))
}

func (u *UserService) DeleteUserByID(id uuid.UUID) (int64, error) {
	if id == uuid.Nil {
		return 0, nil
	}
	return u.user.DeleteByID(id)
}

func (u *UserService) DeleteUserByIDString(id string) (int64, error) {
	if id == "" {
		return 0, nil
	}
	uuidFromStr, err := uuid.Parse(id)
	if err != nil {
		return 0, errors.New("id not valid")
	}
	return u.DeleteUserByID(uuidFromStr)
}
