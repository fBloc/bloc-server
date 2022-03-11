package function_execute_heartbeat

import (
	"context"

	"github.com/fBloc/bloc-server/infrastructure/log"
	"github.com/fBloc/bloc-server/internal/conns/mongodb"
	"github.com/fBloc/bloc-server/repository/function_execute_heartbeat"
	"github.com/fBloc/bloc-server/repository/function_run_record"
	mongoFRRC "github.com/fBloc/bloc-server/repository/function_run_record/mongo"
)

type FunctionConfiguration func(fs *FunctionExecuteHeartbeatService) error

type FunctionExecuteHeartbeatService struct {
	Logger            *log.Logger
	HeartBeatRepo     function_execute_heartbeat.FunctionExecuteHeartbeatRepository
	FunctionRunRecord function_run_record.FunctionRunRecordRepository
}

func NewFunctionExecuteHeartbeatService(
	cfgs ...FunctionConfiguration,
) (*FunctionExecuteHeartbeatService, error) {
	fs := &FunctionExecuteHeartbeatService{}
	for _, cfg := range cfgs {
		err := cfg(fs)
		if err != nil {
			return nil, err
		}
	}
	return fs, nil
}

func WithLogger(logger *log.Logger) FunctionConfiguration {
	return func(us *FunctionExecuteHeartbeatService) error {
		us.Logger = logger
		return nil
	}
}

func WithFunctionHeartbeatRepository(
	fHR function_execute_heartbeat.FunctionExecuteHeartbeatRepository,
) FunctionConfiguration {
	return func(fs *FunctionExecuteHeartbeatService) error {
		fs.HeartBeatRepo = fHR
		return nil
	}
}

func WithFunctionRunRecordRepository(
	fRR function_run_record.FunctionRunRecordRepository,
) FunctionConfiguration {
	return func(fs *FunctionExecuteHeartbeatService) error {
		fs.FunctionRunRecord = fRR
		return nil
	}
}

func WithMongoFunctionRepository(
	mC *mongodb.MongoConfig,
) FunctionConfiguration {
	return func(fs *FunctionExecuteHeartbeatService) error {
		ur, err := mongoFRRC.New(
			context.TODO(),
			mC, mongoFRRC.DefaultCollectionName)
		if err != nil {
			return err
		}
		fs.FunctionRunRecord = ur
		return nil
	}
}
