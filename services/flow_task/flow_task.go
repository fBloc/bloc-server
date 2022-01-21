package flow_task

import (
	"context"

	"github.com/fBloc/bloc-server/infrastructure/log"
	"github.com/fBloc/bloc-server/internal/conns/mongodb"
	"github.com/fBloc/bloc-server/repository/flow"
	mongoFlow "github.com/fBloc/bloc-server/repository/flow/mongo"
	"github.com/fBloc/bloc-server/repository/flow_run_record"
	mongoFRR "github.com/fBloc/bloc-server/repository/flow_run_record/mongo"
)

type FlowTaskConfiguration func(fs *FlowTaskService) error

// 聚合静态的flow和flow_run_record
type FlowTaskService struct {
	logger        *log.Logger
	flow          flow.FlowRepository
	flowRunRecord flow_run_record.FlowRunRecordRepository
}

func WithLogger(logger *log.Logger) FlowTaskConfiguration {
	return func(fts *FlowTaskService) error {
		fts.logger = logger
		return nil
	}
}

func WithMongoFlowRepository(
	mC *mongodb.MongoConfig,
) FlowTaskConfiguration {
	return func(fts *FlowTaskService) error {
		ur, err := mongoFlow.New(
			context.Background(),
			mC, mongoFlow.DefaultCollectionName,
		)
		if err != nil {
			return err
		}
		fts.flow = ur
		return nil
	}
}

func WithMongoFunctionRunRecordRepository(
	mC *mongodb.MongoConfig,
) FlowTaskConfiguration {
	return func(fts *FlowTaskService) error {
		ur, err := mongoFRR.New(
			context.Background(),
			mC, mongoFRR.DefaultCollectionName,
		)
		if err != nil {
			return err
		}
		fts.flowRunRecord = ur
		return nil
	}
}
