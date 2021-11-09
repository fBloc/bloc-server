package mongo

import (
	"context"
	"encoding/json"
	"fmt"
	"strconv"
	"strings"
	"time"

	"github.com/fBloc/bloc-backend-go/aggregate"
	"github.com/fBloc/bloc-backend-go/infrastructure/object_storage"
	"github.com/fBloc/bloc-backend-go/internal/conns/mongodb"
	"github.com/fBloc/bloc-backend-go/internal/filter_options"
	"github.com/fBloc/bloc-backend-go/repository/function_run_record"
	"github.com/fBloc/bloc-backend-go/value_object"

	"github.com/google/uuid"
)

const (
	DefaultCollectionName = "function_run_record"
)

func init() {
	var _ function_run_record.FunctionRunRecordRepository = &MongoRepository{}
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

type mongoIptBriefAndKey struct {
	Brief   interface{} `bson:"brief"`
	FullKey string      `bson:"full_key"`
}

type mongoFunctionRunRecord struct {
	ID                uuid.UUID               `bson:"id"`
	FlowID            uuid.UUID               `bson:"flow_id"`
	FlowOriginID      uuid.UUID               `bson:"flow_origin_id"`
	ArrangementFlowID string                  `bson:"arrangement_flow_id"`
	FunctionID        uuid.UUID               `bson:"function_id"`
	FlowFunctionID    string                  `bson:"flow_function_id"`
	FlowRunRecordID   uuid.UUID               `bson:"flow_run_record_id"`
	Start             time.Time               `bson:"start"`
	End               time.Time               `bson:"end,omitempty"`
	Suc               bool                    `bson:"suc"`
	Pass              bool                    `bson:"pass"`
	Canceled          bool                    `bson:"canceled,omitempty"`
	Description       string                  `bson:"description,omitempty"`
	ErrorMsg          string                  `bson:"error_msg,omitempty"`
	IptBriefAndObskey [][]mongoIptBriefAndKey `bson:"ipt,omitempty"`
	Opt               map[string]interface{}  `bson:"opt,omitempty"`
	OptBrief          map[string]string       `bson:"opt_brief,omitempty"`
	Progress          float32                 `bson:"progress"`
	ProgressMsg       []string                `bson:"progress_msg"`
	ProcessStages     []string                `bson:"process_stages"`
	ProcessStageIndex int                     `bson:"process_stage_index"`
}

func NewFromAggregate(fRR *aggregate.FunctionRunRecord) *mongoFunctionRunRecord {
	resp := mongoFunctionRunRecord{
		ID:                fRR.ID,
		FlowID:            fRR.FlowID,
		FlowOriginID:      fRR.FlowOriginID,
		ArrangementFlowID: fRR.ArrangementFlowID,
		FunctionID:        fRR.FunctionID,
		FlowFunctionID:    fRR.FlowFunctionID,
		FlowRunRecordID:   fRR.FlowRunRecordID,
		Start:             fRR.Start,
		End:               fRR.End,
		Suc:               fRR.Suc,
		Pass:              fRR.Pass,
		Canceled:          fRR.Canceled,
		Description:       fRR.Description,
		ErrorMsg:          fRR.ErrorMsg,
		Opt:               fRR.Opt,
		OptBrief:          fRR.OptBrief,
		Progress:          fRR.Progress,
		ProgressMsg:       fRR.ProgressMsg,
		ProcessStages:     fRR.ProcessStages,
		ProcessStageIndex: fRR.ProcessStageIndex,
	}
	resp.IptBriefAndObskey = make([][]mongoIptBriefAndKey, len(fRR.IptBriefAndObskey))
	for i, param := range fRR.IptBriefAndObskey {
		resp.IptBriefAndObskey[i] = make([]mongoIptBriefAndKey, len(param))
		for j, component := range param {
			resp.IptBriefAndObskey[i][j] = mongoIptBriefAndKey{
				Brief:   component.Brief,
				FullKey: component.FullKey,
			}
		}
	}
	return &resp
}

func (m mongoFunctionRunRecord) ToAggregate() *aggregate.FunctionRunRecord {
	resp := aggregate.FunctionRunRecord{
		ID:                m.ID,
		FlowID:            m.FlowID,
		FlowOriginID:      m.FlowOriginID,
		ArrangementFlowID: m.ArrangementFlowID,
		FunctionID:        m.FunctionID,
		FlowFunctionID:    m.FlowFunctionID,
		FlowRunRecordID:   m.FlowRunRecordID,
		Start:             m.Start,
		End:               m.End,
		Suc:               m.Suc,
		Pass:              m.Pass,
		Canceled:          m.Canceled,
		Description:       m.Description,
		ErrorMsg:          m.ErrorMsg,
		Opt:               m.Opt,
		OptBrief:          m.OptBrief,
		Progress:          m.Progress,
		ProgressMsg:       m.ProgressMsg,
		ProcessStages:     m.ProcessStages,
		ProcessStageIndex: m.ProcessStageIndex,
	}
	resp.IptBriefAndObskey = make([][]aggregate.IptBriefAndKey, len(m.IptBriefAndObskey))
	for i, param := range m.IptBriefAndObskey {
		resp.IptBriefAndObskey[i] = make([]aggregate.IptBriefAndKey, len(param))
		for j, component := range param {
			resp.IptBriefAndObskey[i][j] = aggregate.IptBriefAndKey{
				Brief:   component.Brief,
				FullKey: component.FullKey,
			}
		}
	}
	return &resp
}

// create
func (mr *MongoRepository) Create(fRR *aggregate.FunctionRunRecord) error {
	m := NewFromAggregate(fRR)
	_, err := mr.mongoCollection.InsertOne(*m)
	return err
}

// Read
func (mr *MongoRepository) get(filter *mongodb.MongoFilter) (*aggregate.FunctionRunRecord, error) {
	var mFRR mongoFunctionRunRecord
	err := mr.mongoCollection.Get(filter, nil, &mFRR)
	if err != nil {
		return nil, err
	}
	return mFRR.ToAggregate(), err
}

func (mr *MongoRepository) GetByID(
	id uuid.UUID,
) (*aggregate.FunctionRunRecord, error) {
	return mr.get(mongodb.NewFilter().AddEqual("id", id))
}

func (mr *MongoRepository) FilterByFilterOption(
	kv map[string]interface{}, filterOptions *filter_options.FilterOption,
) ([]*aggregate.FunctionRunRecord, error) {
	filter := mongodb.NewFilter()
	for k, v := range kv {
		if strings.HasSuffix(k, "__contains") {
			realKey := strings.Split(k, "__")[0]
			filter.AddContains(realKey, fmt.Sprintf("%v", v))
			continue
		}
		if strings.HasSuffix(k, "__in") {
			realKey := strings.Split(k, "__")[0]
			valueSlice := strings.Split(fmt.Sprintf("%v", v), ",")
			interfaceVals := make([]interface{}, 0, len(valueSlice))
			for _, v := range valueSlice {
				intVar, err := strconv.Atoi(v)
				if err != nil {
					interfaceVals = append(interfaceVals, v)
				} else {
					interfaceVals = append(interfaceVals, intVar)
				}
			}
			filter.AddIn(realKey, interfaceVals)
			continue
		}
		filter.AddEqual(k, v)
	}

	var mRRRs []mongoFunctionRunRecord
	err := mr.mongoCollection.Filter(filter, filterOptions, &mRRRs)
	if err != nil {
		return nil, err
	}

	ret := make([]*aggregate.FunctionRunRecord, len(mRRRs))
	for i, j := range mRRRs {
		ret[i] = j.ToAggregate()
	}
	return ret, nil
}

func (mr *MongoRepository) Filter(
	filter value_object.RepositoryFilter,
	filterOption value_object.RepositoryFilterOption,
) ([]*aggregate.FunctionRunRecord, error) {
	var mRRRs []mongoFunctionRunRecord
	err := mr.mongoCollection.CommonFilter(filter, filterOption, &mRRRs)
	if err != nil {
		return nil, err
	}

	resp := make([]*aggregate.FunctionRunRecord, 0, len(mRRRs))
	for _, i := range mRRRs {
		resp = append(resp, i.ToAggregate())
	}

	return resp, nil
}

func (mr *MongoRepository) FilterByFlowRunRecordID(
	FlowRunRecordID uuid.UUID,
) ([]*aggregate.FunctionRunRecord, error) {
	var mRRRs []mongoFunctionRunRecord
	err := mr.mongoCollection.Filter(
		mongodb.NewFilter().AddEqual("flow_Run_record_id", FlowRunRecordID),
		&filter_options.FilterOption{}, &mRRRs)
	if err != nil {
		return nil, err
	}

	resp := make([]*aggregate.FunctionRunRecord, 0, len(mRRRs))
	for _, i := range mRRRs {
		resp = append(resp, i.ToAggregate())
	}

	return resp, nil
}

// Update
func (mr *MongoRepository) PatchProgress(id uuid.UUID, progress float32) error {
	return mr.mongoCollection.PatchByID(
		id,
		mongodb.NewUpdater().AddSet("progress", progress))
}

func (mr *MongoRepository) PatchProgressMsg(id uuid.UUID, progressMsg string) error {
	return mr.mongoCollection.PatchByID(
		id,
		mongodb.NewUpdater().AddSet("progress_msg", progressMsg))
}

func (mr *MongoRepository) PatchStageIndex(id uuid.UUID, progressStageIndex int) error {
	return mr.mongoCollection.PatchByID(
		id,
		mongodb.NewUpdater().AddSet("process_stage_index", progressStageIndex))
}

func (mr *MongoRepository) PatchProgressStages(id uuid.UUID, progressStages []string) error {
	return mr.mongoCollection.PatchByID(
		id,
		mongodb.NewUpdater().AddSet("process_stages", progressStages))
}

func (mr *MongoRepository) SaveIptBrief(
	id uuid.UUID, ipts [][]interface{},
	objectStorageImplement object_storage.ObjectStorage,
) error {
	iptBAOk := make([][]mongoIptBriefAndKey, len(ipts), len(ipts))
	for paramIndex, param := range ipts {
		iptBAOk[paramIndex] = make([]mongoIptBriefAndKey, 0, len(param))
		for componentIndex, componentVal := range param {
			uploadByte, _ := json.Marshal(componentVal)
			byteInrune := []rune(string(uploadByte))
			minLength := 51
			if len(byteInrune) < minLength {
				minLength = len(byteInrune)
			}

			key := fmt.Sprintf("%s_%d_%d", id, paramIndex, componentIndex)
			iptBAOk[paramIndex] = append(iptBAOk[paramIndex], mongoIptBriefAndKey{
				Brief:   string(byteInrune[:minLength-1]),
				FullKey: key})
			// TODO oss set error check
			objectStorageImplement.Set(key, uploadByte)
		}
	}
	return mr.mongoCollection.PatchByID(
		id,
		mongodb.NewUpdater().AddSet("ipt", iptBAOk))
}

func (mr *MongoRepository) ClearProgress(id uuid.UUID, progressStages []string) error {
	return mr.mongoCollection.PatchByID(
		id,
		mongodb.NewUpdater().AddSet("process", 0).AddSet("progress_msg", []string{}),
	)
}

func (mr *MongoRepository) SaveSuc(
	id uuid.UUID,
	desc string, opt map[string]interface{}, brief map[string]string, pass bool,
) error {
	return mr.mongoCollection.PatchByID(
		id,
		mongodb.NewUpdater().
			AddSet("end", time.Now()).
			AddSet("suc", true).
			AddSet("pass", pass).
			AddSet("opt", opt).
			AddSet("opt_brief", brief).
			AddSet("description", desc))
}

func (mr *MongoRepository) SaveFail(
	id uuid.UUID, errMsg string,
) error {
	return mr.mongoCollection.PatchByID(
		id,
		mongodb.NewUpdater().
			AddSet("end", time.Now()).
			AddSet("error_msg", errMsg))
}

func (mr *MongoRepository) SaveCancel(
	id uuid.UUID,
) error {
	return mr.mongoCollection.PatchByID(
		id,
		mongodb.NewUpdater().
			AddSet("end", time.Now()).
			AddSet("canceled", true))
}