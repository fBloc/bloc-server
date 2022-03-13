package mongo

import (
	"context"
	"fmt"
	"time"

	"github.com/fBloc/bloc-server/aggregate"
	"github.com/fBloc/bloc-server/internal/conns/mongodb"
	"github.com/fBloc/bloc-server/internal/crontab"
	"github.com/fBloc/bloc-server/repository/flow_run_record"
	"github.com/fBloc/bloc-server/value_object"
)

const (
	DefaultCollectionName = "flow_run_record"
)

func init() {
	var _ flow_run_record.FlowRunRecordRepository = &MongoRepository{}
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

type mongoFlowRunRecord struct {
	ID                           value_object.UUID                    `bson:"id"`
	ArrangementID                value_object.UUID                    `bson:"arrangement_id,omitempty"`
	ArrangementFlowID            string                               `bson:"arrangement_flow_id,omitempty"`
	ArrangementRunRecordID       string                               `bson:"arrangement_task_id,omitempty"`
	FlowID                       value_object.UUID                    `bson:"flow_id"`
	FlowOriginID                 value_object.UUID                    `bson:"flow_origin_id"`
	FlowFuncIDMapFuncRunRecordID map[string]value_object.UUID         `bson:"flowFuncID_map_funcRunRecordID"`
	TriggerTime                  time.Time                            `bson:"trigger_time"`
	TriggerKey                   string                               `bson:"trigger_key"`
	TriggerSource                value_object.FlowTriggeredSourceType `bson:"source_type"`
	TriggerType                  value_object.TriggerType             `bson:"trigger_type"`
	TriggerUserID                value_object.UUID                    `bson:"trigger_user_id"`
	CrontabTriggerflag           string                               `bson:"crontab_trigger_flag,omitempty"` // same crontab not repub
	StartTime                    time.Time                            `bson:"start_time,omitempty"`
	EndTime                      time.Time                            `bson:"end_time,omitempty"`
	Status                       value_object.RunState                `bson:"status"`
	ErrorMsg                     string                               `bson:"error_msg,omitempty"`
	InterceptMsg                 string                               `bson:"intercept_msg,omitempty"`
	RetriedAmount                uint16                               `bson:"retried_amount"`
	TimeoutCanceled              bool                                 `bson:"timeout_canceled,omitempty"`
	Canceled                     bool                                 `bson:"canceled"`
	CancelUserID                 value_object.UUID                    `bson:"cancel_user_id"`
	TraceID                      string                               `bson:"trace_id"`
	OverideIptParams             map[string][][]interface{}           `bson:"overide_ipt_params,omitempty"`
}

func (m *mongoFlowRunRecord) IsZero() bool {
	if m == nil {
		return true
	}
	if m.ID.IsNil() {
		return true
	}
	return false
}

func NewFromAggregate(
	fRR *aggregate.FlowRunRecord,
) *mongoFlowRunRecord {
	resp := mongoFlowRunRecord{
		ID:                           fRR.ID,
		ArrangementID:                fRR.ArrangementID,
		ArrangementFlowID:            fRR.ArrangementFlowID,
		ArrangementRunRecordID:       fRR.ArrangementRunRecordID,
		FlowID:                       fRR.FlowID,
		FlowOriginID:                 fRR.FlowOriginID,
		FlowFuncIDMapFuncRunRecordID: fRR.FlowFuncIDMapFuncRunRecordID,
		TriggerTime:                  fRR.TriggerTime,
		TriggerKey:                   fRR.TriggerKey,
		TriggerSource:                fRR.TriggerSource,
		TriggerType:                  fRR.TriggerType,
		TriggerUserID:                fRR.TriggerUserID,
		StartTime:                    fRR.StartTime,
		EndTime:                      fRR.EndTime,
		Status:                       fRR.Status,
		ErrorMsg:                     fRR.ErrorMsg,
		InterceptMsg:                 fRR.InterceptMsg,
		RetriedAmount:                fRR.RetriedAmount,
		TimeoutCanceled:              fRR.TimeoutCanceled,
		Canceled:                     fRR.Canceled,
		CancelUserID:                 fRR.CancelUserID,
		TraceID:                      fRR.TraceID,
		OverideIptParams:             fRR.OverideIptParams,
	}
	return &resp
}

func (m mongoFlowRunRecord) ToAggregate() *aggregate.FlowRunRecord {
	resp := aggregate.FlowRunRecord{
		ID:                           m.ID,
		ArrangementID:                m.ArrangementID,
		ArrangementFlowID:            m.ArrangementFlowID,
		ArrangementRunRecordID:       m.ArrangementRunRecordID,
		FlowID:                       m.FlowID,
		FlowOriginID:                 m.FlowOriginID,
		FlowFuncIDMapFuncRunRecordID: m.FlowFuncIDMapFuncRunRecordID,
		TriggerTime:                  m.TriggerTime,
		TriggerKey:                   m.TriggerKey,
		TriggerSource:                m.TriggerSource,
		TriggerType:                  m.TriggerType,
		TriggerUserID:                m.TriggerUserID,
		StartTime:                    m.StartTime,
		EndTime:                      m.EndTime,
		Status:                       m.Status,
		ErrorMsg:                     m.ErrorMsg,
		InterceptMsg:                 m.InterceptMsg,
		RetriedAmount:                m.RetriedAmount,
		TimeoutCanceled:              m.TimeoutCanceled,
		Canceled:                     m.Canceled,
		CancelUserID:                 m.CancelUserID,
		TraceID:                      m.TraceID,
		OverideIptParams:             m.OverideIptParams,
	}
	return &resp
}

// create
func (mr *MongoRepository) Create(fRR *aggregate.FlowRunRecord) error {
	m := NewFromAggregate(fRR)
	_, err := mr.mongoCollection.InsertOne(*m)
	return err
}

// CrontabFindOrCreate 创建来源是crontab触发的
func (mr *MongoRepository) CrontabFindOrCreate(
	fRR *aggregate.FlowRunRecord,
	crontabTime time.Time,
) (created bool, err error) {
	m := NewFromAggregate(fRR)
	var old mongoFlowRunRecord

	crontabRep := fmt.Sprintf("%s_%s",
		fRR.FlowID.String(), crontab.TriggeredTimeFlag(crontabTime))
	_, err = mr.mongoCollection.FindOneOrInsert(
		mongodb.NewFilter().AddEqual("crontab_trigger_flag", crontabRep),
		*m,
		&old)
	if err != nil {
		return
	}
	if old.IsZero() { // 老文档不存在，表示此次为新建
		created = true
	}
	return
}

// Read
func (mr *MongoRepository) get(
	filter *mongodb.MongoFilter,
) (*aggregate.FlowRunRecord, error) {
	var mFRR mongoFlowRunRecord
	err := mr.mongoCollection.Get(filter, nil, &mFRR)
	if err != nil {
		return nil, err
	}
	return mFRR.ToAggregate(), err
}

func (mr *MongoRepository) GetByID(
	id value_object.UUID,
) (*aggregate.FlowRunRecord, error) {
	return mr.get(mongodb.NewFilter().AddEqual("id", id))
}

func (mr *MongoRepository) ReGetToCheckIsCanceled(
	id value_object.UUID,
) bool {
	aggFRR, err := mr.get(mongodb.NewFilter().AddEqual("id", id))
	if err != nil {
		return false // 访问失败的，保守处理为没有取消
	}
	return aggFRR.Canceled
}

func (mr *MongoRepository) GetLatestByFlowOriginID(
	flowOriginID value_object.UUID,
) (*aggregate.FlowRunRecord, error) {
	return mr.get(mongodb.NewFilter().AddEqual("flow_origin_id", flowOriginID))
}

func (mr *MongoRepository) GetLatestByFlowID(
	flowID value_object.UUID,
) (*aggregate.FlowRunRecord, error) {
	return mr.get(mongodb.NewFilter().AddEqual("flow_id", flowID))
}

func (mr *MongoRepository) IsHaveRunningTask(
	flowID, thisFlowRunRecordID value_object.UUID,
) (bool, error) {
	filter := value_object.NewRepositoryFilter()
	filter.AddEqual("flow_id", flowID).AddNotEqual("id", thisFlowRunRecordID)

	filterOption := value_object.NewRepositoryFilterOption()
	filterOption.SetDesc()
	filterOption.SetLimit(1)

	var mFRRs []mongoFlowRunRecord
	err := mr.mongoCollection.CommonFilter(
		*filter, *filterOption, &mFRRs)
	if err != nil {
		return false, err
	}

	if len(mFRRs) < 1 {
		return false, nil
	}
	if mFRRs[0].EndTime.IsZero() {
		return true, nil
	}

	return false, nil
}

func (mr *MongoRepository) Filter(
	filter value_object.RepositoryFilter,
	filterOption value_object.RepositoryFilterOption,
) ([]*aggregate.FlowRunRecord, error) {
	var mFRRs []mongoFlowRunRecord
	err := mr.mongoCollection.CommonFilter(filter, filterOption, &mFRRs)
	if err != nil {
		return nil, err
	}

	resp := make([]*aggregate.FlowRunRecord, 0, len(mFRRs))
	for _, i := range mFRRs {
		resp = append(resp, i.ToAggregate())
	}

	return resp, nil
}

func (mr *MongoRepository) Count(
	filter value_object.RepositoryFilter,
) (int64, error) {
	return mr.mongoCollection.CommonCount(filter)
}

// 返回某个flow作为运行源的其全部`运行中`记录
func (mr *MongoRepository) AllRunRecordOfFlowTriggeredByFlowID(
	flowID value_object.UUID,
) ([]*aggregate.FlowRunRecord, error) {
	var mFRRs []mongoFlowRunRecord
	filter := value_object.NewRepositoryFilter()
	filter.AddEqual("flow_id", flowID)

	err := mr.mongoCollection.CommonFilter(
		*filter, *value_object.NewRepositoryFilterOption(), &mFRRs,
	)
	if err != nil {
		return nil, err
	}

	resp := make([]*aggregate.FlowRunRecord, 0, len(mFRRs))
	for _, i := range mFRRs {
		aggRR := i.ToAggregate()
		resp = append(resp, aggRR)
	}

	return resp, nil
}

// update
func (mr *MongoRepository) PatchDataForRetry(
	id value_object.UUID, retriedAmount uint16,
) error {
	return mr.mongoCollection.PatchByID(
		id,
		mongodb.NewUpdater().
			AddSet("retried_amount", retriedAmount+1))
}

func (mr *MongoRepository) PatchFlowFuncIDMapFuncRunRecordID(
	id value_object.UUID,
	FlowFuncIDMapFuncRunRecordID map[string]value_object.UUID,
) error {
	return mr.mongoCollection.PatchByID(
		id,
		mongodb.NewUpdater().
			AddSet("flowFuncID_map_funcRunRecordID", FlowFuncIDMapFuncRunRecordID).
			AddSet("status", value_object.Running))
}

func (mr *MongoRepository) AddFlowFuncIDMapFuncRunRecordID(
	id value_object.UUID,
	flowFuncID string,
	funcRunRecordID value_object.UUID,
) error {
	return mr.mongoCollection.PatchByID(
		id,
		mongodb.NewUpdater().AddSet(
			"flowFuncID_map_funcRunRecordID."+flowFuncID,
			funcRunRecordID),
	)
}

func (mr *MongoRepository) Start(id value_object.UUID) error {
	return mr.mongoCollection.PatchByID(
		id,
		mongodb.NewUpdater().
			AddSet("status", value_object.Running).
			AddSet("start_time", time.Now()))
}

func (mr *MongoRepository) Suc(id value_object.UUID) error {
	return mr.mongoCollection.PatchByID(
		id,
		mongodb.NewUpdater().
			AddSet("status", value_object.Suc).
			AddSet("end_time", time.Now()))
}

func (mr *MongoRepository) NotAllowedParallelRun(
	id value_object.UUID) error {
	return mr.mongoCollection.PatchByID(
		id,
		mongodb.NewUpdater().
			AddSet("status", value_object.NotAllowedParallelCancel).
			AddSet("end_time", time.Now()))
}

func (mr *MongoRepository) Fail(id value_object.UUID, errorMsg string) error {
	return mr.mongoCollection.PatchByID(
		id,
		mongodb.NewUpdater().
			AddSet("status", value_object.Fail).
			AddSet("error_msg", errorMsg).
			AddSet("end_time", time.Now()))
}

func (mr *MongoRepository) FunctionDead(id value_object.UUID, errorMsg string) error {
	return mr.mongoCollection.PatchByID(
		id,
		mongodb.NewUpdater().
			AddSet("status", value_object.Fail).
			AddSet("error_msg", errorMsg).
			AddSet("end_time", time.Now()))
}

func (mr *MongoRepository) Intercepted(id value_object.UUID, msg string) error {
	return mr.mongoCollection.PatchByID(
		id,
		mongodb.NewUpdater().
			AddSet("status", value_object.InterceptedCancel).
			AddSet("intercept_msg", msg).
			AddSet("end_time", time.Now()))
}

func (mr *MongoRepository) TimeoutCancel(id value_object.UUID) error {
	return mr.mongoCollection.PatchByID(
		id,
		mongodb.NewUpdater().
			AddSet("status", value_object.TimeoutCanceled).
			AddSet("canceled", true).
			AddSet("end_time", time.Now()),
	)
}

func (mr *MongoRepository) UserCancel(id, userID value_object.UUID) error {
	return mr.mongoCollection.PatchByID(
		id,
		mongodb.NewUpdater().
			AddSet("status", value_object.UserCanceled).
			AddSet("canceled", true).
			AddSet("cancel_user_id", userID).
			AddSet("end_time", time.Now()),
	)
}

// Delete
