package mongo

import (
	"context"
	"errors"
	"time"

	"github.com/fBloc/bloc-server/internal/conns/mongodb"
	"github.com/fBloc/bloc-server/internal/filter_options"
	"github.com/fBloc/bloc-server/pkg/add_or_del"
	"github.com/fBloc/bloc-server/pkg/ipt"
	"github.com/fBloc/bloc-server/pkg/opt"
	"github.com/fBloc/bloc-server/repository/function"
	"github.com/fBloc/bloc-server/value_object"

	"github.com/fBloc/bloc-server/aggregate"
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
	mC *mongodb.MongoConfig, collectionName string,
) (*MongoRepository, error) {
	collection, err := mongodb.NewCollection(mC, collectionName)
	if err != nil {
		return nil, err
	}
	return &MongoRepository{mongoCollection: collection}, nil
}

type mongoFunction struct {
	ID                      value_object.UUID   `bson:"id"`
	Name                    string              `bson:"name"`
	GroupName               string              `bson:"group_name"`
	ProviderName            string              `bson:"provider_name"`
	RegisterTime            time.Time           `bson:"register_time"`
	LastAliveTime           time.Time           `bson:"last_alive_time"`
	Description             string              `bson:"description"`
	Ipts                    ipt.IptSlice        `bson:"ipts"`
	Opts                    []*opt.Opt          `bson:"opts"`
	IptDigest               string              `bson:"ipt_digest"`
	OptDigest               string              `bson:"opt_digest"`
	ProgressMilestones      []string            `bson:"progress_milestones"`
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
		RegisterTime:            m.RegisterTime,
		LastAliveTime:           m.LastAliveTime,
		Description:             m.Description,
		Ipts:                    m.Ipts,
		Opts:                    m.Opts,
		IptDigest:               m.IptDigest,
		OptDigest:               m.OptDigest,
		ProgressMilestones:      m.ProgressMilestones,
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
		RegisterTime:            f.RegisterTime,
		LastAliveTime:           f.LastAliveTime,
		Description:             f.Description,
		Ipts:                    f.Ipts,
		Opts:                    f.Opts,
		IptDigest:               f.IptDigest,
		OptDigest:               f.OptDigest,
		ProgressMilestones:      f.ProgressMilestones,
		ReadUserIDs:             f.ReadUserIDs,
		ExecuteUserIDs:          f.ExecuteUserIDs,
		AssignPermissionUserIDs: f.AssignPermissionUserIDs,
	}

	// below set to []value_object.UUID{} is because mongo's $push not support push to nil
	if f.ReadUserIDs == nil {
		resp.ReadUserIDs = []value_object.UUID{}
	}
	if f.ExecuteUserIDs == nil {
		resp.ExecuteUserIDs = []value_object.UUID{}
	}
	if f.AssignPermissionUserIDs == nil {
		resp.AssignPermissionUserIDs = []value_object.UUID{}
	}
	return &resp
}

func (mr *MongoRepository) Create(
	f *aggregate.Function,
) error {
	mF := NewFromFunction(f)
	if mF.RegisterTime.IsZero() {
		mF.RegisterTime = time.Now()
	}
	_, err := mr.mongoCollection.InsertOne(mF)
	return err
}

func (mr *MongoRepository) All(withoutFields []string) ([]*aggregate.Function, error) {
	var m []mongoFunction
	filterOption := filter_options.NewFilterOption().SetSortByNaturalAsc()
	filterOption.AddWithoutFields(withoutFields...)

	err := mr.mongoCollection.Filter(nil, filterOption, &m)
	if err != nil {
		return nil, err
	}
	ret := make([]*aggregate.Function, len(m))
	for i, j := range m {
		ret[i] = j.ToAggregate()
	}
	return ret, nil
}

func (mr *MongoRepository) UserReadAbleAll(
	user *aggregate.User, withoutFields []string,
) ([]*aggregate.Function, error) {
	if user.IsZero() {
		return nil, errors.New("ipt user is nil")
	}
	var m []mongoFunction

	filterOption := filter_options.NewFilterOption().SetSortByNaturalAsc()
	filterOption.AddWithoutFields(withoutFields...)

	err := mr.mongoCollection.Filter(
		mongodb.NewFilter().AddEqual("read_user_ids", user.ID),
		filterOption, &m)
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

func (mr *MongoRepository) GetByIDForCheckAliveTime(
	id value_object.UUID,
) (*aggregate.Function, error) {
	var m mongoFunction
	err := mr.mongoCollection.GetByIDWithFieldCtrl(
		id,
		[]string{"id", "last_alive_time", "name", "provider_name"},
		[]string{},
		&m)
	if err != nil {
		return nil, err
	}
	return m.ToAggregate(), nil
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

func (mr *MongoRepository) PatchProgressMilestones(
	id value_object.UUID,
	progressMilestones []string,
) error {
	updater := mongodb.NewUpdater().AddSet("progress_milestones", progressMilestones)
	return mr.mongoCollection.PatchByID(id, updater)
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

func (mr *MongoRepository) AliveReport(
	id value_object.UUID,
) error {
	updater := mongodb.NewUpdater().AddSet("last_alive_time", time.Now())
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
