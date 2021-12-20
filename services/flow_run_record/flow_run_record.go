package flow_run_record

import (
	"context"

	"github.com/fBloc/bloc-server/infrastructure/log"
	"github.com/fBloc/bloc-server/repository/flow_run_record"
	mongoFRRC "github.com/fBloc/bloc-server/repository/flow_run_record/mongo"
	user_cache "github.com/fBloc/bloc-server/services/userid_cache"
)

type FlowRunRecordConfiguration func(fRRS *FlowRunRecordService) error

type FlowRunRecordService struct {
	logger           *log.Logger
	UserCacheService *user_cache.UserCacheService
	FlowRunRecord    flow_run_record.FlowRunRecordRepository
}

func NewService(
	cfgs ...FlowRunRecordConfiguration,
) (*FlowRunRecordService, error) {
	fRRS := &FlowRunRecordService{}
	for _, cfg := range cfgs {
		err := cfg(fRRS)
		if err != nil {
			return nil, err
		}
	}
	return fRRS, nil
}

func WithUserCacheService(
	userCacheService *user_cache.UserCacheService,
) FlowRunRecordConfiguration {
	return func(t *FlowRunRecordService) error {
		t.UserCacheService = userCacheService
		return nil
	}
}

func WithLogger(logger *log.Logger) FlowRunRecordConfiguration {
	return func(us *FlowRunRecordService) error {
		us.logger = logger
		return nil
	}
}

func WithMongoFlowRunRecordRepository(
	hosts []string, port int, db, user, password string,
) FlowRunRecordConfiguration {
	return func(fts *FlowRunRecordService) error {
		ur, err := mongoFRRC.New(
			context.Background(),
			hosts, port, db, user, password, mongoFRRC.DefaultCollectionName,
		)
		if err != nil {
			return err
		}
		fts.FlowRunRecord = ur
		return nil
	}
}

func WithFlowRunRecordRepository(
	flowRunRecordRepository flow_run_record.FlowRunRecordRepository,
) FlowRunRecordConfiguration {
	return func(fts *FlowRunRecordService) error {
		fts.FlowRunRecord = flowRunRecordRepository
		return nil
	}
}
