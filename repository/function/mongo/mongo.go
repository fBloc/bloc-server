package mongo

import (
	"context"
	"errors"

	"github.com/fBloc/bloc-backend-go/internal/conns/mongodb"
	"github.com/fBloc/bloc-backend-go/pkg/add_or_del"
	"github.com/fBloc/bloc-backend-go/pkg/ipt"
	"github.com/fBloc/bloc-backend-go/pkg/opt"
	"github.com/fBloc/bloc-backend-go/repository/function"
	"github.com/fBloc/bloc-backend-go/value_object"

	"github.com/fBloc/bloc-backend-go/aggregate"
)

const (
	DefaultCollectionName = "function"
)

func init() {
	var _ function.FunctionRepository = &MongoRepository{}
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

type mongoFunction struct {
	ID                      value_object.UUID   `bson:"id"`
	Name                    string              `bson:"name"`
	GroupName               string              `bson:"group_name"`
	ProviderName            string              `bson:"provider_name"`
	Description             string              `bson:"description"`
	Ipts                    ipt.IptSlice        `bson:"ipts"`
	Opts                    []*opt.Opt          `bson:"opts"`
	IptDigest               string              `bson:"ipt_digest"`
	OptDigest               string              `bson:"opt_digest"`
	ProcessStages           []string            `bson:"process_stages"`
	ReadUserIDs             []value_object.UUID `bson:"read_user_ids"`
	ExecuteUserIDs          []value_object.UUID `bson:"execute_user_ids"`
	AssignPermissionUserIDs []value_object.UUID `bson:"assign_permission_user_ids"`
}

func (m *mongoFunction) ToAggregate() *aggregate.Function {
	return &aggregate.Function{
		ID:                      m.ID,
		Name:                    m.Name,
		GroupName:               m.GroupName,
		ProviderName:            m.ProviderName,
		Description:             m.Description,
		Ipts:                    m.Ipts,
		Opts:                    m.Opts,
		IptDigest:               m.IptDigest,
		OptDigest:               m.OptDigest,
		ProcessStages:           m.ProcessStages,
		ReadUserIDs:             m.ReadUserIDs,
		ExecuteUserIDs:          m.ExecuteUserIDs,
		AssignPermissionUserIDs: m.AssignPermissionUserIDs,
	}
}

func NewFromFunction(f *aggregate.Function) *mongoFunction {
	resp := mongoFunction{
		ID:                      f.ID,
		Name:                    f.Name,
		GroupName:               f.GroupName,
		ProviderName:            f.ProviderName,
		Description:             f.Description,
		Ipts:                    f.Ipts,
		Opts:                    f.Opts,
		IptDigest:               f.IptDigest,
		OptDigest:               f.OptDigest,
		ProcessStages:           f.ProcessStages,
		ReadUserIDs:             f.ReadUserIDs,
		ExecuteUserIDs:          f.ExecuteUserIDs,
		AssignPermissionUserIDs: f.AssignPermissionUserIDs,
	}
	return &resp
}

func (mr *MongoRepository) Create(
	f *aggregate.Function,
) error {
	_, err := mr.mongoCollection.InsertOne(NewFromFunction(f))
	return err
}

func (mr *MongoRepository) All() ([]*aggregate.Function, error) {
	var m []mongoFunction
	err := mr.mongoCollection.Filter(nil, nil, &m)
	if err != nil {
		return nil, err
	}
	ret := make([]*aggregate.Function, len(m))
	for i, j := range m {
		ret[i] = j.ToAggregate()
	}
	return ret, nil
}

func (mr *MongoRepository) IDMapFunctionAll() (map[value_object.UUID]*aggregate.Function, error) {
	var m []mongoFunction
	err := mr.mongoCollection.Filter(nil, nil, &m)
	if err != nil {
		return nil, err
	}
	ret := make(map[value_object.UUID]*aggregate.Function, len(m))
	for _, i := range m {
		ret[i.ID] = i.ToAggregate()
	}
	return ret, nil
}

func (mr *MongoRepository) GetByID(
	id value_object.UUID,
) (*aggregate.Function, error) {
	var m mongoFunction
	err := mr.mongoCollection.GetByID(id, &m)
	if err != nil {
		return nil, err
	}
	return m.ToAggregate(), nil
}

func (mr *MongoRepository) GetSameIptOptFunction(
	iptDigest, optDigest string,
) (*aggregate.Function, error) {
	var m mongoFunction
	err := mr.mongoCollection.Get(
		mongodb.NewFilter().
			AddEqual("ipt_digest", iptDigest).
			AddEqual("opt_digest", optDigest),
		nil, &m)
	if err != nil {
		return nil, err
	}
	return m.ToAggregate(), nil
}

func (mr *MongoRepository) PatchName(id value_object.UUID, name string) error {
	updater := mongodb.NewUpdater().AddSet("name", name)
	return mr.mongoCollection.PatchByID(id, updater)
}

func (mr *MongoRepository) PatchDescription(id value_object.UUID, desc string) error {
	updater := mongodb.NewUpdater().AddSet("description", desc)
	return mr.mongoCollection.PatchByID(id, updater)
}

func (mr *MongoRepository) PatchGroupName(id value_object.UUID, groupName string) error {
	updater := mongodb.NewUpdater().AddSet("group_name", groupName)
	return mr.mongoCollection.PatchByID(id, updater)
}

func (mr *MongoRepository) PatchProviderName(
	id value_object.UUID, providerName string,
) error {
	updater := mongodb.NewUpdater().AddSet("provider_name", providerName)
	return mr.mongoCollection.PatchByID(id, updater)
}

func (mr *MongoRepository) userOperation(
	id, userID value_object.UUID, permType value_object.PermissionType, aod add_or_del.AddOrDel,
) error {
	var roleStr string
	if permType == value_object.Read {
		roleStr = "read_user_ids"
	} else if permType == value_object.Execute {
		roleStr = "execute_user_ids"
	} else if permType == value_object.AssignPermission {
		roleStr = "assign_permission_user_ids"
	} else {
		return errors.New("permission type wrong")
	}

	updater := mongodb.NewUpdater()
	if aod == add_or_del.Remove {
		updater.AddPull(roleStr, userID)
	} else {
		updater.AddPush(roleStr, userID)
	}
	return mr.mongoCollection.PatchByID(id, updater)
}

func (mr *MongoRepository) AddReader(id, userID value_object.UUID) error {
	return mr.userOperation(id, userID, value_object.Read, add_or_del.Add)
}
func (mr *MongoRepository) RemoveReader(id, userID value_object.UUID) error {
	return mr.userOperation(id, userID, value_object.Read, add_or_del.Remove)
}

func (mr *MongoRepository) AddExecuter(id, userID value_object.UUID) error {
	return mr.userOperation(id, userID, value_object.Execute, add_or_del.Add)
}

func (mr *MongoRepository) RemoveExecuter(id, userID value_object.UUID) error {
	return mr.userOperation(id, userID, value_object.Execute, add_or_del.Remove)
}

func (mr *MongoRepository) AddAssigner(id, userID value_object.UUID) error {
	return mr.userOperation(id, userID, value_object.AssignPermission, add_or_del.Add)
}

func (mr *MongoRepository) RemoveAssigner(id, userID value_object.UUID) error {
	return mr.userOperation(id, userID, value_object.AssignPermission, add_or_del.Remove)
}
