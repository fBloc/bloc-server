package mongo

import (
	"context"
	"time"

	"github.com/fBloc/bloc-backend-go/aggregate"
	"github.com/fBloc/bloc-backend-go/internal/conns/mongodb"
	"github.com/fBloc/bloc-backend-go/internal/filter_options"
	"github.com/fBloc/bloc-backend-go/repository/function_execute_heartbeat"
	"github.com/fBloc/bloc-backend-go/value_object"
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
	hosts []string, port int, user, password, db, collectionName string,
) (*MongoRepository, error) {
	collection := mongodb.NewCollection(
		hosts, port, user, password, db, collectionName,
	)
	return &MongoRepository{mongoCollection: collection}, nil
}

type mongoFunctionExecuteHeartBeat struct {
	ID                  value_object.UUID `bson:"id"`
	FunctionRunRecordID value_object.UUID `bson:"function_run_record_id"`
	StartTime           time.Time         `bson:"start_time"`
	LatestHeartbeatTime time.Time         `bson:"latest_heartbeat_time"`
}

func (m *mongoFunctionExecuteHeartBeat) ToAggregate() *aggregate.FunctionExecuteHeartBeat {
	return &aggregate.FunctionExecuteHeartBeat{
		ID:                  m.ID,
		FunctionRunRecordID: m.FunctionRunRecordID,
		StartTime:           m.StartTime,
		LatestHeartbeatTime: m.LatestHeartbeatTime}
}

func NewFromAggregate(f *aggregate.FunctionExecuteHeartBeat) *mongoFunctionExecuteHeartBeat {
	resp := mongoFunctionExecuteHeartBeat{
		ID:                  f.ID,
		FunctionRunRecordID: f.FunctionRunRecordID,
		StartTime:           f.StartTime,
		LatestHeartbeatTime: f.LatestHeartbeatTime,
	}
	return &resp
}

func (mr *MongoRepository) Create(
	f *aggregate.FunctionExecuteHeartBeat,
) error {
	_, err := mr.mongoCollection.InsertOne(*f)
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

func (mr *MongoRepository) AllDeads() ([]*aggregate.FunctionExecuteHeartBeat, error) {
	var mSlice []mongoFunctionExecuteHeartBeat
	err := mr.mongoCollection.Filter(
		mongodb.NewFilter().AddLt("latest_heartbeat_time", time.Now().Add(aggregate.HeartBeatDeadThreshold)),
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

func (mr *MongoRepository) Delete(
	id value_object.UUID,
) (int64, error) {
	return mr.mongoCollection.Delete(
		mongodb.NewFilter().AddEqual("id", id))
}
