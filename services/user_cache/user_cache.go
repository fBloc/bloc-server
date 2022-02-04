package user_cache

import (
	"context"
	"errors"
	"sync"

	"github.com/fBloc/bloc-server/aggregate"
	"github.com/fBloc/bloc-server/infrastructure/log"
	"github.com/fBloc/bloc-server/internal/conns/mongodb"
	"github.com/fBloc/bloc-server/repository/user"
	mongoUser "github.com/fBloc/bloc-server/repository/user/mongo"
	"github.com/fBloc/bloc-server/value_object"
)

type UserConfiguration func(us *UserCacheService) error

type UserCacheService struct {
	logger *log.Logger
	user   user.UserRepository
}

func NewUserCacheService(
	cfgs ...UserConfiguration,
) (*UserCacheService, error) {
	us := &UserCacheService{}
	for _, cfg := range cfgs {
		err := cfg(us)
		if err != nil {
			return nil, err
		}
	}
	us.initialCache()
	return us, nil
}

func WithMongoUserRepository(
	mC *mongodb.MongoConfig,
) UserConfiguration {
	return func(us *UserCacheService) error {
		ur, err := mongoUser.New(
			context.Background(),
			mC, mongoUser.DefaultCollectionName,
		)
		if err != nil {
			return err
		}
		us.user = ur
		return nil
	}
}

func WithUser(uR user.UserRepository) UserConfiguration {
	return func(us *UserCacheService) error {
		us.user = uR
		return nil
	}
}

func WithLogger(logger *log.Logger) UserConfiguration {
	return func(us *UserCacheService) error {
		us.logger = logger
		return nil
	}
}

type localCache struct {
	userTokenMapUser map[value_object.UUID]aggregate.User
	userIDMapUser    map[value_object.UUID]aggregate.User
	sync.Mutex
}

var cache = &localCache{
	userIDMapUser:    make(map[value_object.UUID]aggregate.User),
	userTokenMapUser: make(map[value_object.UUID]aggregate.User),
}

func (us *UserCacheService) initialCache() {
	cache.Lock()
	defer cache.Unlock()

	allUsers, err := us.user.All()
	if err != nil {
		panic(err)
	}
	idMapUser := make(map[value_object.UUID]aggregate.User, len(allUsers))
	tokenMapUser := make(map[value_object.UUID]aggregate.User, len(allUsers))
	for _, i := range allUsers {
		idMapUser[i.ID] = i
		tokenMapUser[i.Token] = i
	}
	cache.userIDMapUser = idMapUser
	cache.userTokenMapUser = tokenMapUser
}

func (us *UserCacheService) visitRepositoryByToken(token value_object.UUID) (aggregate.User, error) {
	resp, err := us.user.GetByToken(token)
	if err != nil {
		return aggregate.User{}, err
	}
	if resp.IsZero() {
		us.logger.Warningf(
			map[string]string{},
			"get user by token missed:%s", token.String())
		return aggregate.User{}, nil
	}
	cache.Lock()
	defer cache.Unlock()
	cache.userIDMapUser[resp.ID] = *resp
	cache.userTokenMapUser[resp.Token] = *resp
	return *resp, nil
}

func (us *UserCacheService) GetUserByToken(token value_object.UUID) (aggregate.User, error) {
	if userIns, ok := cache.userTokenMapUser[token]; ok {
		return userIns, nil
	}
	return us.visitRepositoryByToken(token)
}

func (us *UserCacheService) GetUserByTokenString(token string) (aggregate.User, error) {
	if token == "" {
		return aggregate.User{}, errors.New("token cannot be blank string")
	}
	tokenUID, err := value_object.ParseToUUID(token)
	if err != nil {
		return aggregate.User{}, err
	}
	return us.GetUserByToken(tokenUID)
}

func (us *UserCacheService) visitRepositoryByID(
	id value_object.UUID,
) (aggregate.User, error) {
	resp, err := us.user.GetByID(id)
	if err != nil {
		return aggregate.User{}, err
	}
	if resp.IsZero() {
		us.logger.Warningf(
			map[string]string{},
			"get user by ID missed:%s", id.String())
		return aggregate.User{}, nil
	}
	cache.Lock()
	defer cache.Unlock()
	cache.userIDMapUser[resp.ID] = *resp
	cache.userTokenMapUser[resp.Token] = *resp
	return *resp, nil
}

func (us *UserCacheService) GetUserByID(id value_object.UUID) (aggregate.User, error) {
	if userIns, ok := cache.userIDMapUser[id]; ok {
		return userIns, nil
	}
	return us.visitRepositoryByID(id)
}

func (us *UserCacheService) GetUserByIDString(id string) (aggregate.User, error) {
	if id == "" {
		return aggregate.User{}, errors.New("token cannot be blank string")
	}
	uUID, err := value_object.ParseToUUID(id)
	if err != nil {
		return aggregate.User{}, err
	}
	return us.visitRepositoryByID(uUID)
}
