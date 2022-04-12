package mongo

import (
	"context"
	"time"

	"github.com/fBloc/bloc-server/aggregate"
	"github.com/fBloc/bloc-server/internal/conns/mongodb"
	"github.com/fBloc/bloc-server/internal/crontab"
	"github.com/fBloc/bloc-server/internal/filter_options"
	"github.com/fBloc/bloc-server/pkg/add_or_del"
	"github.com/fBloc/bloc-server/pkg/value_type"
	"github.com/fBloc/bloc-server/repository/flow"
	"github.com/fBloc/bloc-server/value_object"
	"go.mongodb.org/mongo-driver/bson/primitive"

	"github.com/pkg/errors"
)

const (
	DefaultCollectionName = "flow"
)

func init() {
	var _ flow.FlowRepository = &MongoRepository{}
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

type mongoIptComponentConfig struct {
	Blank          bool                              `bson:"blank"`
	IptWay         value_object.FunctionParamIptType `bson:"ipt_way,omitempty"`
	ValueType      value_type.ValueType              `bson:"value_type,omitempty"`
	Value          interface{}                       `bson:"value,omitempty"`
	FlowFunctionID string                            `bson:"flow_function_id,omitempty"`
	Key            string                            `bson:"key,omitempty"`
}

type mongoFlowFunction struct {
	FunctionID                value_object.UUID           `bson:"function_id"`
	Note                      string                      `bson:"note"`
	Position                  interface{}                 `bson:"position"`
	UpstreamFlowFunctionIDs   []string                    `bson:"upstream_flowfunction_ids"`
	DownstreamFlowFunctionIDs []string                    `bson:"downstream_flowfunction_ids"`
	ParamIpts                 [][]mongoIptComponentConfig `bson:"param_ipts"` // 第一层对应一个ipt，第二层对应ipt内的component
}

type mongoFlowFunctionIDMapFlowFunction map[string]*mongoFlowFunction

func fromAggFlowFunctionIDMapFlowFunction(
	aggflowFunctionIDMapFlowFunction map[string]*aggregate.FlowFunction,
) *mongoFlowFunctionIDMapFlowFunction {
	funcs := make(mongoFlowFunctionIDMapFlowFunction, len(aggflowFunctionIDMapFlowFunction))

	for flowFuncID, flowFunc := range aggflowFunctionIDMapFlowFunction {
		tmp := mongoFlowFunction{
			FunctionID:                flowFunc.FunctionID,
			Note:                      flowFunc.Note,
			Position:                  flowFunc.Position,
			UpstreamFlowFunctionIDs:   flowFunc.UpstreamFlowFunctionIDs,
			DownstreamFlowFunctionIDs: flowFunc.DownstreamFlowFunctionIDs,
		}
		tmp.ParamIpts = make([][]mongoIptComponentConfig, len(flowFunc.ParamIpts))
		for i, ipt := range flowFunc.ParamIpts {
			tmp.ParamIpts[i] = make([]mongoIptComponentConfig, len(ipt))
			for j, component := range ipt {
				tmp.ParamIpts[i][j] = mongoIptComponentConfig{
					Blank:          component.Blank,
					IptWay:         component.IptWay,
					ValueType:      component.ValueType,
					Value:          component.Value,
					FlowFunctionID: component.FlowFunctionID,
					Key:            component.Key,
				}
			}
		}
		funcs[flowFuncID] = &tmp
	}
	return &funcs
}

func (mF mongoFlowFunctionIDMapFlowFunction) toAggFlowFunctionIDMapFlowFunction() map[string]*aggregate.FlowFunction {
	funcs := make(map[string]*aggregate.FlowFunction, len(mF))
	for flowFuncID, flowFunc := range mF {
		tmp := aggregate.FlowFunction{
			FunctionID:                flowFunc.FunctionID,
			Note:                      flowFunc.Note,
			Position:                  flowFunc.Position,
			UpstreamFlowFunctionIDs:   flowFunc.UpstreamFlowFunctionIDs,
			DownstreamFlowFunctionIDs: flowFunc.DownstreamFlowFunctionIDs,
		}
		tmp.ParamIpts = make([][]aggregate.IptComponentConfig, len(flowFunc.ParamIpts))
		for i, ipt := range flowFunc.ParamIpts {
			tmp.ParamIpts[i] = make([]aggregate.IptComponentConfig, len(ipt))
			for j, component := range ipt {
				iptComponentConfig := aggregate.IptComponentConfig{
					Blank:          component.Blank,
					IptWay:         component.IptWay,
					ValueType:      component.ValueType,
					FlowFunctionID: component.FlowFunctionID,
					Key:            component.Key,
				}
				// for the specific mongo data type primitive.A not influence outside world.
				// do trans it here
				if pa, ok := component.Value.(primitive.A); ok {
					iptComponentConfig.Value = []interface{}(pa)
				} else {
					iptComponentConfig.Value = component.Value
				}
				tmp.ParamIpts[i][j] = iptComponentConfig
			}
		}
		funcs[flowFuncID] = &tmp
	}
	return funcs
}

type mongoFlow struct {
	ID                            value_object.UUID                  `bson:"id"`
	Name                          string                             `bson:"name"`
	IsDraft                       bool                               `bson:"is_draft"`
	Deleted                       bool                               `bson:"deleted"`
	Version                       uint                               `bson:"version"`
	OriginID                      value_object.UUID                  `bson:"origin_id,omitempty"`
	Newest                        bool                               `bson:"newest"`
	CreateUserID                  value_object.UUID                  `bson:"create_user_id"`
	CreateTime                    time.Time                          `bson:"create_time"`
	Position                      interface{}                        `bson:"position"`
	FlowFunctionIDMapFlowFunction mongoFlowFunctionIDMapFlowFunction `bson:"flowFunctionID_map_flowFunction"`
	Crontab                       *crontab.CrontabRepresent          `bson:"crontab,omitempty"`
	TriggerKey                    string                             `bson:"trigger_key"`
	AllowTriggerByKey             bool                               `bson:"allow_trigger_by_key"`
	TimeoutInSeconds              uint32                             `bson:"timeout_in_seconds,omitempty"`
	RetryAmount                   uint16                             `bson:"retry_amount,omitempty"`
	RetryIntervalInSecond         uint16                             `bson:"retry_interval_in_second,omitempty"`
	AllowParallelRun              bool                               `bson:"allow_parallel_run,omitempty"`
	ReadUserIDs                   []value_object.UUID                `bson:"read_user_ids"`
	WriteUserIDs                  []value_object.UUID                `bson:"write_user_ids"`
	ExecuteUserIDs                []value_object.UUID                `bson:"execute_user_ids"`
	DeleteUserIDs                 []value_object.UUID                `bson:"delete_user_ids"`
	AssignPermissionUserIDs       []value_object.UUID                `bson:"assign_permission_user_ids"`
}

func (m mongoFlow) ToAggregate() *aggregate.Flow {
	resp := aggregate.Flow{
		ID:                            m.ID,
		Name:                          m.Name,
		IsDraft:                       m.IsDraft,
		Deleted:                       m.Deleted,
		Version:                       m.Version,
		OriginID:                      m.OriginID,
		Newest:                        m.Newest,
		CreateUserID:                  m.CreateUserID,
		CreateTime:                    m.CreateTime,
		Position:                      m.Position,
		Crontab:                       m.Crontab,
		TriggerKey:                    m.TriggerKey,
		AllowTriggerByKey:             m.AllowTriggerByKey,
		TimeoutInSeconds:              m.TimeoutInSeconds,
		RetryAmount:                   m.RetryAmount,
		RetryIntervalInSecond:         m.RetryIntervalInSecond,
		AllowParallelRun:              m.AllowParallelRun,
		ReadUserIDs:                   m.ReadUserIDs,
		WriteUserIDs:                  m.WriteUserIDs,
		ExecuteUserIDs:                m.ExecuteUserIDs,
		DeleteUserIDs:                 m.DeleteUserIDs,
		AssignPermissionUserIDs:       m.AssignPermissionUserIDs,
		FlowFunctionIDMapFlowFunction: m.FlowFunctionIDMapFlowFunction.toAggFlowFunctionIDMapFlowFunction(),
	}
	return &resp
}

func NewFromFlow(f *aggregate.Flow) *mongoFlow {
	resp := mongoFlow{
		ID:                            f.ID,
		Name:                          f.Name,
		IsDraft:                       f.IsDraft,
		Version:                       f.Version,
		OriginID:                      f.OriginID,
		Newest:                        f.Newest,
		CreateUserID:                  f.CreateUserID,
		CreateTime:                    f.CreateTime,
		Crontab:                       f.Crontab,
		Position:                      f.Position,
		TriggerKey:                    f.TriggerKey,
		AllowTriggerByKey:             f.AllowTriggerByKey,
		TimeoutInSeconds:              f.TimeoutInSeconds,
		RetryAmount:                   f.RetryAmount,
		RetryIntervalInSecond:         f.RetryIntervalInSecond,
		AllowParallelRun:              f.AllowParallelRun,
		ReadUserIDs:                   f.ReadUserIDs,
		WriteUserIDs:                  f.WriteUserIDs,
		ExecuteUserIDs:                f.ExecuteUserIDs,
		DeleteUserIDs:                 f.DeleteUserIDs,
		AssignPermissionUserIDs:       f.AssignPermissionUserIDs,
		FlowFunctionIDMapFlowFunction: *fromAggFlowFunctionIDMapFlowFunction(f.FlowFunctionIDMapFlowFunction),
	}
	return &resp
}

func (mr *MongoRepository) get(filter *mongodb.MongoFilter) (*aggregate.Flow, error) {
	var flow mongoFlow
	err := mr.mongoCollection.Get(filter, nil, &flow)
	if err != nil {
		return nil, err
	}
	return flow.ToAggregate(), err
}

func (mr *MongoRepository) GetByID(id value_object.UUID) (*aggregate.Flow, error) {
	if id.IsNil() {
		return nil, errors.New("must have id")
	}
	return mr.get(mongodb.NewFilter().AddEqual("id", id).AddEqual("deleted", false))
}

func (mr *MongoRepository) GetByIDStr(id string) (*aggregate.Flow, error) {
	if id == "" {
		return nil, errors.New("id cannot be blank")
	}
	uuidFromStr, err := value_object.ParseToUUID(id)
	if err != nil {
		return nil, errors.Wrap(err, "trans id to uuid failed")
	}
	return mr.GetByID(uuidFromStr)
}

func (mr *MongoRepository) GetOnlineByOriginID(originID value_object.UUID) (*aggregate.Flow, error) {
	if originID.IsNil() {
		return nil, errors.New("must have origin_id")
	}
	return mr.get(mongodb.NewFilter().AddEqual("origin_id", originID).AddEqual("is_draft", false).AddEqual("deleted", false))
}

func (mr *MongoRepository) GetLatestByOriginID(originID value_object.UUID) (*aggregate.Flow, error) {
	if originID.IsNil() {
		return nil, errors.New("must have origin_id")
	}
	return mr.get(mongodb.NewFilter().AddEqual("origin_id", originID).AddEqual("deleted", false))
}

func (mr *MongoRepository) GetOnlineByOriginIDStr(originID string) (*aggregate.Flow, error) {
	if originID == "" {
		return nil, errors.New("origin_id cannot be blank")
	}
	uuidFromStr, err := value_object.ParseToUUID(originID)
	if err != nil {
		return nil, errors.Wrap(err, "trans origin_id to uuid failed")
	}
	return mr.GetOnlineByOriginID(uuidFromStr)
}

func (mr *MongoRepository) GetDraftByOriginID(originID value_object.UUID) (*aggregate.Flow, error) {
	if originID.IsNil() {
		return nil, errors.New("must have origin_id")
	}
	return mr.get(mongodb.NewFilter().AddEqual("origin_id", originID).AddEqual("is_draft", true).AddEqual("deleted", false))
}

func (mr *MongoRepository) FilterOnline(user *aggregate.User, nameContains string) ([]aggregate.Flow, error) {
	filter := mongodb.NewFilter().AddEqual("is_draft", false).AddEqual("newest", true).AddEqual("deleted", false)
	if !user.IsZero() && !user.IsSuper {
		filter.AddEqual("read_user_ids", user.ID)
	}
	if nameContains != "" {
		filter.AddContains("name", nameContains)
	}

	var flows []mongoFlow
	err := mr.mongoCollection.Filter(filter, &filter_options.FilterOption{}, &flows)
	if err != nil {
		return nil, err
	}
	ret := make([]aggregate.Flow, 0, len(flows))
	for _, i := range flows {
		ret = append(ret, *i.ToAggregate())
	}
	return ret, err
}

func (mr *MongoRepository) FilterCrontabFlows() ([]aggregate.Flow, error) {
	filter := mongodb.NewFilter().AddEqual("is_draft", false).AddEqual("newest", true).AddExist("crontab").AddEqual("deleted", false)

	var flows []mongoFlow
	err := mr.mongoCollection.Filter(filter, &filter_options.FilterOption{}, &flows)
	if err != nil {
		return nil, err
	}
	ret := make([]aggregate.Flow, 0, len(flows))
	for _, i := range flows {
		if !i.Crontab.IsValid() {
			continue
		}
		ret = append(ret, *i.ToAggregate())
	}
	return ret, err
}

func (mr *MongoRepository) Filter(
	filter *value_object.RepositoryFilter,
) ([]aggregate.Flow, error) {
	var flows []mongoFlow
	filter.AddEqual("is_draft", false).AddEqual("deleted", false)
	err := mr.mongoCollection.CommonFilter(
		*filter, value_object.RepositoryFilterOption{}, &flows)
	if err != nil {
		return nil, err
	}

	ret := make([]aggregate.Flow, 0, len(flows))
	for _, i := range flows {
		ret = append(ret, *i.ToAggregate())
	}
	return ret, err
}

func (mr *MongoRepository) FilterDraft(userID value_object.UUID, nameContains string) ([]aggregate.Flow, error) {
	filter := mongodb.NewFilter().AddEqual("is_draft", true).AddEqual("deleted", false)
	if !userID.IsNil() {
		filter.AddEqual("read_user_ids", userID)
	}
	if nameContains != "" {
		filter.AddContains("name", nameContains)
	}

	var flows []mongoFlow
	err := mr.mongoCollection.Filter(filter, &filter_options.FilterOption{}, &flows)
	if err != nil {
		return nil, err
	}
	ret := make([]aggregate.Flow, 0, len(flows))
	for _, i := range flows {
		ret = append(ret, *i.ToAggregate())
	}
	return ret, err
}

func (mr *MongoRepository) PatchName(id value_object.UUID, name string) error {
	updater := mongodb.NewUpdater().
		AddSet("name", name)
	return mr.mongoCollection.PatchByID(id, updater)
}

// PatchWhetherAllowTriggerByKey 更新crontab配置
func (mr *MongoRepository) PatchPosition(id value_object.UUID, position interface{}) error {
	updater := mongodb.NewUpdater().AddSet("position", position)
	return mr.mongoCollection.PatchByID(id, updater)
}

// OfflineByID 对flow进行下线
func (mr *MongoRepository) OfflineByID(id value_object.UUID) error {
	updater := mongodb.NewUpdater().AddSet("newest", false).AddSet("deleted", true)
	return mr.mongoCollection.PatchByID(id, updater)
}

// PatchTimeout 更新超时配置
// func (mr *MongoRepository) PatchFuncs(id value_object.UUID, funcs map[string]*flow_bloc.FlowBloc) error {
// 	updater := mongodb.NewUpdater().AddSet("funcs", funcs)
// 	return mr.mongoCollection.PatchByID(id, updater)
// }

func (mr *MongoRepository) PatchRetryStrategy(id value_object.UUID, amount, intervalInSecond uint16) error {
	if amount <= 0 || intervalInSecond <= 0 {
		return errors.New("retry_amount & retry_interval_in_second must both > 0")
	}
	updater := mongodb.NewUpdater().
		AddSet("retry_interval_in_second", intervalInSecond).
		AddSet("retry_amount", amount)
	return mr.mongoCollection.PatchByID(id, updater)
}

func (mr *MongoRepository) PatchCrontab(id value_object.UUID, c *crontab.CrontabRepresent) error {
	// 对于非空的crontab设置，需要检查格式是否正确
	if !c.IsZero() && !c.IsValid() {
		return errors.New("crontab expression not valid")
	}
	updater := mongodb.NewUpdater().AddSet("crontab", c)
	return mr.mongoCollection.PatchByID(id, updater)
}

// PatchAllowParallelRun  更新是否在运行的时候有新的发布仍然发布
func (mr *MongoRepository) PatchAllowParallelRun(id value_object.UUID, allowParallel bool) error {
	updater := mongodb.NewUpdater().
		AddSet("allow_parallel_run", allowParallel)
	return mr.mongoCollection.PatchByID(id, updater)
}

// PatchWhetherAllowTriggerByKey 更新触发的key
func (mr *MongoRepository) PatchWhetherAllowTriggerByKey(id value_object.UUID, allowed bool) error {
	updater := mongodb.NewUpdater().AddSet("allow_trigger_by_key", allowed)
	return mr.mongoCollection.PatchByID(id, updater)
}

// PatchFlowFunctionIDMapFlowFunction 更新function构成
func (mr *MongoRepository) PatchFlowFunctionIDMapFlowFunction(
	id value_object.UUID,
	flowFunctionIDMapFlowFunction map[string]*aggregate.FlowFunction,
) error {
	updater := mongodb.NewUpdater().AddSet(
		"flowFunctionID_map_flowFunction",
		fromAggFlowFunctionIDMapFlowFunction(flowFunctionIDMapFlowFunction))
	return mr.mongoCollection.PatchByID(id, updater)
}

// PatchTimeout 更新超时配置
func (mr *MongoRepository) PatchTimeout(id value_object.UUID, tOS uint32) error {
	updater := mongodb.NewUpdater().AddSet("timeout_in_seconds", tOS)
	return mr.mongoCollection.PatchByID(id, updater)
}

func (mr *MongoRepository) ReplaceByID(id value_object.UUID, aggFlow *aggregate.Flow) error {
	if id.IsNil() {
		return errors.New("id cannot be nil")
	}
	mFlow := NewFromFlow(aggFlow)

	mongoID, err := mr.mongoCollection.GetMongoID(id)
	if err != nil {
		return err
	}

	return mr.mongoCollection.ReplaceByID(mongoID, *mFlow)
}

func (mr *MongoRepository) userOperation(
	id, userID value_object.UUID,
	permType value_object.PermissionType, aod add_or_del.AddOrDel,
) error {
	var roleStr string
	if permType == value_object.Read {
		roleStr = "read_user_ids"
	} else if permType == value_object.Write {
		roleStr = "write_user_ids"
	} else if permType == value_object.Execute {
		roleStr = "execute_user_ids"
	} else if permType == value_object.Delete {
		roleStr = "delete_user_ids"
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

func (mr *MongoRepository) AddWriter(id, userID value_object.UUID) error {
	return mr.userOperation(id, userID, value_object.Write, add_or_del.Add)
}

func (mr *MongoRepository) RemoveWriter(id, userID value_object.UUID) error {
	return mr.userOperation(id, userID, value_object.Write, add_or_del.Remove)
}

func (mr *MongoRepository) AddExecuter(id, userID value_object.UUID) error {
	return mr.userOperation(id, userID, value_object.Execute, add_or_del.Add)
}

func (mr *MongoRepository) RemoveExecuter(id, userID value_object.UUID) error {
	return mr.userOperation(id, userID, value_object.Execute, add_or_del.Remove)
}

func (mr *MongoRepository) AddDeleter(id, userID value_object.UUID) error {
	return mr.userOperation(id, userID, value_object.Delete, add_or_del.Add)
}

func (mr *MongoRepository) RemoveDeleter(id, userID value_object.UUID) error {
	return mr.userOperation(id, userID, value_object.Delete, add_or_del.Remove)
}

func (mr *MongoRepository) AddAssigner(id, userID value_object.UUID) error {
	return mr.userOperation(id, userID, value_object.AssignPermission, add_or_del.Add)
}

func (mr *MongoRepository) RemoveAssigner(id, userID value_object.UUID) error {
	return mr.userOperation(id, userID, value_object.AssignPermission, add_or_del.Remove)
}

func (mr *MongoRepository) mongoCreateFromAgg(flow *aggregate.Flow) error {
	m := NewFromFlow(flow)

	_, err := mr.mongoCollection.InsertOne(*m)
	return err
}

func (mr *MongoRepository) CreateDraftFromScratch(
	name string,
	createUserID value_object.UUID,
	position interface{},
	funcs map[string]*aggregate.FlowFunction,
) (*aggregate.Flow, error) {
	if createUserID.IsNil() {
		return nil, errors.New("must have create_user_id")
	}
	aggFlow := aggregate.Flow{
		ID:                            value_object.NewUUID(),
		Name:                          name,
		IsDraft:                       true,
		OriginID:                      value_object.NewUUID(),
		CreateUserID:                  createUserID,
		CreateTime:                    time.Now(),
		Position:                      position,
		FlowFunctionIDMapFlowFunction: funcs,
		TriggerKey:                    value_object.NewUUID().String(),
		ReadUserIDs:                   []value_object.UUID{createUserID},
		WriteUserIDs:                  []value_object.UUID{createUserID},
		ExecuteUserIDs:                []value_object.UUID{createUserID},
		DeleteUserIDs:                 []value_object.UUID{createUserID},
		AssignPermissionUserIDs:       []value_object.UUID{createUserID},
	}

	err := mr.mongoCreateFromAgg(&aggFlow)
	if err != nil {
		return nil, errors.Wrap(err, "create draft flow to repository failed")
	}

	return &aggFlow, nil
}

// CreateDraftFromExistFlow compare to CreateDraftFromScratch, plus field `originID` of the exist flow
func (mr *MongoRepository) CreateDraftFromExistFlow(
	name string,
	createUserID, originID value_object.UUID,
	position interface{},
	funcs map[string]*aggregate.FlowFunction,
) (*aggregate.Flow, error) {
	if createUserID.IsNil() {
		return nil, errors.New("create_user_id cannot be blank")
	}
	if originID.IsNil() {
		return nil, errors.New("origin_id cannot be blank")
	}
	existFlow, err := mr.GetOnlineByOriginID(originID)
	if err != nil {
		return nil, errors.Wrap(err, "origin_id find exist flow error")
	}
	if existFlow.IsZero() {
		return nil, errors.New("origin_id find no exist flow")
	}

	aggFlow := aggregate.Flow{
		ID:                            value_object.NewUUID(),
		Name:                          name,
		IsDraft:                       true,
		OriginID:                      originID,
		CreateUserID:                  createUserID,
		CreateTime:                    time.Now(),
		TriggerKey:                    existFlow.TriggerKey,
		Position:                      position,
		FlowFunctionIDMapFlowFunction: funcs,
		ReadUserIDs:                   existFlow.ReadUserIDs,
		WriteUserIDs:                  existFlow.WriteUserIDs,
		ExecuteUserIDs:                existFlow.ExecuteUserIDs,
		DeleteUserIDs:                 existFlow.DeleteUserIDs,
		AssignPermissionUserIDs:       existFlow.AssignPermissionUserIDs,
	}

	err = mr.mongoCreateFromAgg(&aggFlow)
	if err != nil {
		return nil, errors.Wrap(err, "create draft flow to repository failed")
	}

	return &aggFlow, nil
}

func (mr *MongoRepository) CreateOnlineFromDraft(
	draftF *aggregate.Flow,
) (*aggregate.Flow, error) {
	latestFlow, err := mr.GetOnlineByOriginID(draftF.OriginID)
	if err != nil {
		return nil, errors.Wrap(err, "draft.origin_id find exist flow error")
	}

	aggF := draftF
	aggF.IsDraft = false
	aggF.Newest = true
	aggF.CreateTime = time.Now()
	if latestFlow.IsZero() { // 没有已存在的同origin_id的flow（完全新建的draft提交的情况）
		// 直接创建就是了
		return aggF, mr.mongoCreateFromAgg(aggF)
	}
	// 已有flow
	// 1. 继承一些属性
	aggF.Version = latestFlow.Version + 1
	// 继承权限
	aggF.ReadUserIDs = latestFlow.ReadUserIDs
	aggF.WriteUserIDs = latestFlow.WriteUserIDs
	aggF.ExecuteUserIDs = latestFlow.ExecuteUserIDs
	aggF.DeleteUserIDs = latestFlow.DeleteUserIDs
	aggF.AssignPermissionUserIDs = latestFlow.AssignPermissionUserIDs
	// 运行配置
	aggF.AllowParallelRun = latestFlow.AllowParallelRun
	aggF.Crontab = latestFlow.Crontab
	aggF.TriggerKey = latestFlow.TriggerKey
	aggF.AllowTriggerByKey = latestFlow.AllowTriggerByKey
	aggF.TimeoutInSeconds = latestFlow.TimeoutInSeconds
	aggF.RetryAmount = latestFlow.RetryAmount
	aggF.RetryIntervalInSecond = latestFlow.RetryIntervalInSecond
	// 2. 如果老的在线，需要进行下线
	if latestFlow.Newest {
		err = mr.OfflineByID(latestFlow.ID)
		if err != nil {
			return nil, errors.Wrap(err, "offline old flow error")
		}
	}
	// 3. 创建新的在线flow
	return aggF, mr.mongoCreateFromAgg(aggF)
}

func (mr *MongoRepository) DeleteByID(id value_object.UUID) (int64, error) {
	return mr.mongoCollection.Patch(
		mongodb.NewFilter().AddEqual("id", id).AddEqual("deleted", false),
		mongodb.NewUpdater().AddSet("deleted", true))
}

func (mr *MongoRepository) DeleteByOriginID(originID value_object.UUID) (int64, error) {
	return mr.mongoCollection.Patch(
		mongodb.NewFilter().AddEqual("origin_id", originID).AddEqual("deleted", false),
		mongodb.NewUpdater().AddSet("deleted", true))
}

func (mr *MongoRepository) DeleteDraftByOriginID(originID value_object.UUID) (int64, error) {
	return mr.mongoCollection.Patch(
		mongodb.NewFilter().AddEqual("origin_id", originID).AddEqual("is_draft", true).AddEqual("deleted", false),
		mongodb.NewUpdater().AddSet("deleted", true))
}
