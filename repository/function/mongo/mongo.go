package mongo

import (
	"context"

	"github.com/fBloc/bloc-backend-go/internal/conns/mongodb"
	"github.com/fBloc/bloc-backend-go/pkg/add_or_del"
	"github.com/fBloc/bloc-backend-go/pkg/ipt"
	"github.com/fBloc/bloc-backend-go/pkg/op_role"
	"github.com/fBloc/bloc-backend-go/pkg/opt"
	"github.com/fBloc/bloc-backend-go/repository/function"

	"github.com/fBloc/bloc-backend-go/aggregate"

	"github.com/google/uuid"
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
	ID            uuid.UUID    `bson:"id"`
	Name          string       `bson:"name"`
	GroupName     string       `bson:"group_name"`
	Description   string       `bson:"description"`
	Ipts          ipt.IptSlice `bson:"ipts"`
	Opts          []*opt.Opt   `bson:"opts"`
	IptDigest     string       `bson:"ipt_digest"`
	OptDigest     string       `bson:"opt_digest"`
	ProcessStages []string     `bson:"process_stages"`
}

func (m *mongoFunction) ToAggregate() *aggregate.Function {
	return &aggregate.Function{
		ID:            m.ID,
		Name:          m.Name,
		GroupName:     m.GroupName,
		Description:   m.Description,
		Ipts:          m.Ipts,
		Opts:          m.Opts,
		IptDigest:     m.IptDigest,
		OptDigest:     m.OptDigest,
		ProcessStages: m.ProcessStages}
}

func NewFromFunction(f *aggregate.Function) *mongoFunction {
	resp := mongoFunction{
		ID:            f.ID,
		Name:          f.Name,
		GroupName:     f.GroupName,
		Description:   f.Description,
		Ipts:          f.Ipts,
		Opts:          f.Opts,
		IptDigest:     f.IptDigest,
		OptDigest:     f.OptDigest,
		ProcessStages: f.ProcessStages,
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

func (mr *MongoRepository) IDMapFunctionAll() (map[uuid.UUID]*aggregate.Function, error) {
	var m []mongoFunction
	err := mr.mongoCollection.Filter(nil, nil, &m)
	if err != nil {
		return nil, err
	}
	ret := make(map[uuid.UUID]*aggregate.Function, len(m))
	for _, i := range m {
		ret[i.ID] = i.ToAggregate()
	}
	return ret, nil
}

func (mr *MongoRepository) GetByID(
	id uuid.UUID,
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

func (mr *MongoRepository) PatchName(id uuid.UUID, name string) error {
	updater := mongodb.NewUpdater().AddSet("name", name)
	return mr.mongoCollection.PatchByID(id, updater)
}

func (mr *MongoRepository) PatchDescription(id uuid.UUID, desc string) error {
	updater := mongodb.NewUpdater().AddSet("description", desc)
	return mr.mongoCollection.PatchByID(id, updater)
}

func (mr *MongoRepository) PatchGroupName(id uuid.UUID, groupName string) error {
	updater := mongodb.NewUpdater().AddSet("group_name", groupName)
	return mr.mongoCollection.PatchByID(id, updater)
}

func (mr *MongoRepository) userOperation(
	id, userID uuid.UUID, role op_role.OpRole, aod add_or_del.AddOrDel,
) error {
	roleStr := "read_user_ids"
	if role == op_role.Writer {
		roleStr = "write_user_ids"
	} else if role == op_role.Executer {
		roleStr = "execute_user_ids"
	} else if role == op_role.Super {
		roleStr = "super_user_ids"
	}

	updater := mongodb.NewUpdater()
	if aod == add_or_del.Remove {
		updater.AddPull(roleStr, userID)
	} else {
		updater.AddPush(roleStr, userID)
	}
	return mr.mongoCollection.PatchByID(id, updater)
}

func (mr *MongoRepository) AddReader(id, userID uuid.UUID) error {
	return mr.userOperation(id, userID, op_role.Reader, add_or_del.Add)
}
func (mr *MongoRepository) DeleteReader(id, userID uuid.UUID) error {
	return mr.userOperation(id, userID, op_role.Reader, add_or_del.Remove)
}

func (mr *MongoRepository) AddWriter(id, userID uuid.UUID) error {
	return mr.userOperation(id, userID, op_role.Writer, add_or_del.Add)
}

func (mr *MongoRepository) AddSuper(id, userID uuid.UUID) error {
	return mr.userOperation(id, userID, op_role.Super, add_or_del.Add)
}

func (mr *MongoRepository) DeleteWriter(id, userID uuid.UUID) error {
	return mr.userOperation(id, userID, op_role.Writer, add_or_del.Remove)
}

func (mr *MongoRepository) AddExecuter(id, userID uuid.UUID) error {
	return mr.userOperation(id, userID, op_role.Executer, add_or_del.Add)
}

func (mr *MongoRepository) DeleteExecuter(id, userID uuid.UUID) error {
	return mr.userOperation(id, userID, op_role.Executer, add_or_del.Remove)
}

func (mr *MongoRepository) DeleteSuper(id, userID uuid.UUID) error {
	return mr.userOperation(id, userID, op_role.Super, add_or_del.Remove)
}
