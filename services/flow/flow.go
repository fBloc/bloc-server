package flow

import (
	"context"

	"github.com/fBloc/bloc-backend-go/aggregate"
	"github.com/fBloc/bloc-backend-go/infrastructure/log"
	flow_repo "github.com/fBloc/bloc-backend-go/repository/flow"
	mongo_flow "github.com/fBloc/bloc-backend-go/repository/flow/mongo"
	"github.com/fBloc/bloc-backend-go/repository/flow_run_record"
	mongoFlowRRecord "github.com/fBloc/bloc-backend-go/repository/flow_run_record/mongo"
	"github.com/fBloc/bloc-backend-go/repository/function"
	mongoFunction "github.com/fBloc/bloc-backend-go/repository/function/mongo"
	"github.com/fBloc/bloc-backend-go/repository/function_run_record"
	user_cache "github.com/fBloc/bloc-backend-go/services/userid_cache"
	"github.com/fBloc/bloc-backend-go/value_object"
)

type FlowConfiguration func(fs *FlowService) error

type FlowService struct {
	Logger            *log.Logger
	Flow              flow_repo.FlowRepository
	FlowRunRecord     flow_run_record.FlowRunRecordRepository
	Function          function.FunctionRepository
	FunctionRunRecord function_run_record.FunctionRunRecordRepository
	UserCacheService  *user_cache.UserCacheService
}

func NewFlowService(cfgs ...FlowConfiguration) (*FlowService, error) {
	fs := &FlowService{}
	for _, cfg := range cfgs {
		err := cfg(fs)
		if err != nil {
			return nil, err
		}
	}
	return fs, nil
}

func WithLogger(logger *log.Logger) FlowConfiguration {
	return func(fs *FlowService) error {
		fs.Logger = logger
		return nil
	}
}

func WithFlowRepository(
	fR flow_repo.FlowRepository,
) FlowConfiguration {
	return func(fs *FlowService) error {
		fs.Flow = fR
		return nil
	}
}

func WithFunctionRunRecordRepository(
	f function_run_record.FunctionRunRecordRepository,
) FlowConfiguration {
	return func(fs *FlowService) error {
		fs.FunctionRunRecord = f
		return nil
	}
}

func WithMongoFlowRepository(
	hosts []string, port int, db, user, password string,
) FlowConfiguration {
	return func(fs *FlowService) error {
		mF, err := mongo_flow.New(
			context.Background(),
			hosts, port, db, user, password, mongo_flow.DefaultCollectionName,
		)
		if err != nil {
			return err
		}
		fs.Flow = mF
		return nil
	}
}

func WithMongoFlowRunRecordRepository(
	hosts []string, port int, db, user, password string,
) FlowConfiguration {
	return func(fs *FlowService) error {
		mF, err := mongoFlowRRecord.New(
			context.Background(),
			hosts, port, db, user, password, mongo_flow.DefaultCollectionName,
		)
		if err != nil {
			return err
		}
		fs.FlowRunRecord = mF
		return nil
	}
}

func WithFlowRunRecordRepository(
	f flow_run_record.FlowRunRecordRepository,
) FlowConfiguration {
	return func(fs *FlowService) error {
		fs.FlowRunRecord = f
		return nil
	}
}

func WithMongoFunctionRepository(
	hosts []string, port int, db, user, password string,
) FlowConfiguration {
	return func(fs *FlowService) error {
		mF, err := mongoFunction.New(
			context.Background(),
			hosts, port, db, user, password, mongo_flow.DefaultCollectionName,
		)
		if err != nil {
			return err
		}
		fs.Function = mF
		return nil
	}
}

func WithFunctionRepository(
	f function.FunctionRepository,
) FlowConfiguration {
	return func(fs *FlowService) error {
		fs.Function = f
		return nil
	}
}

func WithUserCacheService(
	userCacheService *user_cache.UserCacheService,
) FlowConfiguration {
	return func(t *FlowService) error {
		t.UserCacheService = userCacheService
		return nil
	}
}

func (u *FlowService) GetLatestRunRecordByFlowID(
	flowID value_object.UUID,
) (*aggregate.FlowRunRecord, error) {
	if flowID.IsNil() {
		return nil, nil
	}
	return u.FlowRunRecord.GetLatestByFlowID(flowID)
}

func (u *FlowService) GetLatestRunRecordByArrangementFlowID(
	arrFlowID string,
) (*aggregate.FlowRunRecord, error) {
	if arrFlowID == "" {
		return nil, nil
	}
	return u.FlowRunRecord.GetLatestByArrangementFlowID(arrFlowID)
}
