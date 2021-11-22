package mongo

import (
	"context"
	"time"

	"github.com/fBloc/bloc-backend-go/aggregate"
	"github.com/fBloc/bloc-backend-go/internal/conns/mongodb"
	"github.com/fBloc/bloc-backend-go/internal/filter_options"
	"github.com/fBloc/bloc-backend-go/repository/user"
	"github.com/fBloc/bloc-backend-go/value_object"

	"github.com/pkg/errors"
)

const (
	DefaultCollectionName = "user"
)

func init() {
	var _ user.UserRepository = &MongoRepository{}
}

type MongoRepository struct {
	mongoCollection *mongodb.Collection
}

// Create a new mongodb repository
func New(
	ctx context.Context,
	hosts []string, port int, user, password, db, collectionName string,
) (*MongoRepository, error) {
	collection := mongodb.NewCollection(
		hosts, port, user, password, db, collectionName,
	)
	return &MongoRepository{mongoCollection: collection}, nil
}

type mongoUser struct {
	ID         value_object.UUID `bson:"id"`
	Name       string            `bson:"name"`
	Password   string            `bson:"password"` // 加密的password
	CreateTime time.Time         `bson:"create_time"`
	IsSuper    bool              `bson:"is_super"`
}

func (m mongoUser) ToAggregate() *aggregate.User {
	return &aggregate.User{
		ID:         m.ID,
		Name:       m.Name,
		CreateTime: m.CreateTime,
		Password:   m.Password,
		IsSuper:    m.IsSuper,
	}
}

func NewFromUser(u aggregate.User) *mongoUser {
	mU := mongoUser{
		ID:         u.ID,
		Name:       u.Name,
		Password:   u.Password,
		IsSuper:    u.IsSuper,
		CreateTime: u.CreateTime,
	}
	if mU.CreateTime.IsZero() {
		mU.CreateTime = time.Now()
	}
	return &mU
}

func (mr *MongoRepository) Create(u aggregate.User) error {
	m := NewFromUser(u)

	_, err := mr.mongoCollection.InsertOne(*m)
	if err != nil {
		return errors.Wrap(err, "create flow to mongo failed")
	}

	return nil
}

func (mr *MongoRepository) All() ([]aggregate.User, error) {
	var users []mongoUser
	err := mr.mongoCollection.Filter(nil, nil, &users)
	if err != nil {
		return []aggregate.User{}, err
	}
	resp := make([]aggregate.User, len(users))
	for i, j := range users {
		resp[i] = *j.ToAggregate()
	}
	return resp, nil
}

func (mr *MongoRepository) FilterByNameContains(
	nameContains string,
) (users []aggregate.User, err error) {
	filter := mongodb.NewFilter().AddContains("name", nameContains)
	err = mr.mongoCollection.Filter(filter, &filter_options.FilterOption{}, &users)
	return
}

func (mr *MongoRepository) GetByName(
	name string,
) (*aggregate.User, error) {
	var user mongoUser
	err := mr.mongoCollection.Get(
		mongodb.NewFilter().AddEqual("name", name),
		nil, &user)
	if err != nil {
		return nil, err
	}
	return user.ToAggregate(), nil
}

func (mr *MongoRepository) GetByID(id value_object.UUID) (*aggregate.User, error) {
	var user mongoUser
	err := mr.mongoCollection.GetByID(id, &user)
	if err != nil {
		return nil, err
	}
	return user.ToAggregate(), nil
}

func (mr *MongoRepository) PatchName(id value_object.UUID, name string) error {
	updater := mongodb.NewUpdater().AddSet("name", name)
	return mr.mongoCollection.PatchByID(id, updater)
}

func (mr *MongoRepository) DeleteByID(id value_object.UUID) (int64, error) {
	return mr.mongoCollection.DeleteByID(id)
}
