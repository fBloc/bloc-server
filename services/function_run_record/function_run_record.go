package function_run_record

import (
	"context"

	"github.com/fBloc/bloc-server/infrastructure/log"
	"github.com/fBloc/bloc-server/repository/function_run_record"
	mongoFRRC "github.com/fBloc/bloc-server/repository/function_run_record/mongo"
	user_cache "github.com/fBloc/bloc-server/services/userid_cache"
)

type FunctionRunRecordConfiguration func(
	fRRS *FunctionRunRecordService,
) error

type FunctionRunRecordService struct {
	logger             *log.Logger
	UserCacheService   *user_cache.UserCacheService
	FunctionRunRecords function_run_record.FunctionRunRecordRepository
}

func NewService(
	cfgs ...FunctionRunRecordConfiguration,
) (*FunctionRunRecordService, error) {
	fRRS := &FunctionRunRecordService{}
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
) FunctionRunRecordConfiguration {
	return func(t *FunctionRunRecordService) error {
		t.UserCacheService = userCacheService
		return nil
	}
}

func WithLogger(logger *log.Logger) FunctionRunRecordConfiguration {
	return func(us *FunctionRunRecordService) error {
		us.logger = logger
		return nil
	}
}

func WithMongoFunctionRunRecordRepository(
	hosts []string, port int, db, user, password string,
) FunctionRunRecordConfiguration {
	return func(fts *FunctionRunRecordService) error {
		ur, err := mongoFRRC.New(
			context.Background(),
			hosts, port, db, user, password, mongoFRRC.DefaultCollectionName,
		)
		if err != nil {
			return err
		}
		fts.FunctionRunRecords = ur
		return nil
	}
}

func WithFunctionRunRecordRepository(
	fRR function_run_record.FunctionRunRecordRepository,
) FunctionRunRecordConfiguration {
	return func(fts *FunctionRunRecordService) error {
		fts.FunctionRunRecords = fRR
		return nil
	}
}
