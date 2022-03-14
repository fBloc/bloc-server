package mongo

import (
	"context"
	"time"

	"github.com/fBloc/bloc-server/aggregate"
	"github.com/fBloc/bloc-server/internal/conns/mongodb"
	"github.com/fBloc/bloc-server/internal/filter_options"
	"github.com/fBloc/bloc-server/repository/function_execute_heartbeat"
	"github.com/fBloc/bloc-server/value_object"
)

const (
	DefaultCollectionName = "function_execute_heartbeat"
)

func init() {
	var _ function_execute_heartbeat.FunctionExecuteHeartbeatRepository = &MongoRepository{}
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

	indexes := mongoDBIndexes()
	collection.CreateIndex(indexes)

	return &MongoRepository{mongoCollection: collection}, nil
}

type mongoFunctionExecuteHeartBeat struct {
	FunctionRunRecordID value_object.UUID `bson:"function_run_record_id"`
	LatestHeartbeatTime time.Time         `bson:"latest_heartbeat_time"`
}

func (m *mongoFunctionExecuteHeartBeat) ToAggregate() *aggregate.FunctionExecuteHeartBeat {
	return &aggregate.FunctionExecuteHeartBeat{
		FunctionRunRecordID: m.FunctionRunRecordID,
		LatestHeartbeatTime: m.LatestHeartbeatTime}
}

func NewFromAggregate(f *aggregate.FunctionExecuteHeartBeat) *mongoFunctionExecuteHeartBeat {
	resp := mongoFunctionExecuteHeartBeat{
		FunctionRunRecordID: f.FunctionRunRecordID,
	}
	if f.LatestHeartbeatTime.IsZero() {
		resp.LatestHeartbeatTime = f.LatestHeartbeatTime
	} else {
		resp.LatestHeartbeatTime = time.Now()
	}
	return &resp
}

func (mr *MongoRepository) Create(
	f *aggregate.FunctionExecuteHeartBeat,
) error {
	mFunctionExecuteHeartBeat := NewFromAggregate(f)
	_, err := mr.mongoCollection.InsertOne(mFunctionExecuteHeartBeat)
	return err
}

func (mr *MongoRepository) GetByID(
	id value_object.UUID,
) (*aggregate.FunctionExecuteHeartBeat, error) {
	var m mongoFunctionExecuteHeartBeat
	err := mr.mongoCollection.GetByID(id, &m)
	if err != nil {
		return nil, err
	}
	return m.ToAggregate(), nil
}

func (mr *MongoRepository) GetByFunctionRunRecordID(
	funcRunRecordID value_object.UUID,
) (*aggregate.FunctionExecuteHeartBeat, error) {
	var m mongoFunctionExecuteHeartBeat
	err := mr.mongoCollection.Get(
		mongodb.NewFilter().AddEqual("function_run_record_id", funcRunRecordID),
		nil, &m)
	if err != nil {
		return nil, err
	}
	return m.ToAggregate(), nil
}

func (mr *MongoRepository) AllDeads(
	timeoutThreshold time.Duration,
) ([]*aggregate.FunctionExecuteHeartBeat, error) {
	var mSlice []mongoFunctionExecuteHeartBeat
	err := mr.mongoCollection.Filter(
		mongodb.NewFilter().AddLt("latest_heartbeat_time", time.Now().Add(-timeoutThreshold)),
		&filter_options.FilterOption{}, &mSlice,
	)
	if err != nil {
		return nil, err
	}

	ret := make([]*aggregate.FunctionExecuteHeartBeat, 0, len(mSlice))
	for _, m := range mSlice {
		ret = append(ret, m.ToAggregate())
	}

	return ret, nil
}

func (mr *MongoRepository) AliveReport(
	id value_object.UUID,
) error {
	updater := mongodb.NewUpdater().AddSet("latest_heartbeat_time", time.Now())
	return mr.mongoCollection.PatchByID(id, updater)
}

func (mr *MongoRepository) AliveReportByFuncRunRecordID(
	funcRunRecordID value_object.UUID,
) error {
	err := mr.mongoCollection.UpdateOneOrInsert(
		mongodb.NewFilter().AddEqual("function_run_record_id", funcRunRecordID),
		mongodb.NewUpdater().AddSet("latest_heartbeat_time", time.Now()).AddSet("function_run_record_id", funcRunRecordID))
	return err
}

func (mr *MongoRepository) Delete(
	id value_object.UUID,
) (int64, error) {
	return mr.mongoCollection.Delete(
		mongodb.NewFilter().AddEqual("id", id))
}

func (mr *MongoRepository) DeleteByFunctionRunRecordID(
	functionRunRecordID value_object.UUID,
) (int64, error) {
	return mr.mongoCollection.Delete(
		mongodb.NewFilter().AddEqual("function_run_record_id", functionRunRecordID))
}
