package flow_task

import (
	"context"

	"github.com/fBloc/bloc-backend-go/infrastructure/log"
	"github.com/fBloc/bloc-backend-go/repository/flow"
	mongoFlow "github.com/fBloc/bloc-backend-go/repository/flow/mongo"
	"github.com/fBloc/bloc-backend-go/repository/flow_run_record"
	mongoFRR "github.com/fBloc/bloc-backend-go/repository/flow_run_record/mongo"
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

func WithMongoFlowRepository(hosts []string, port int, db, user, password string) FlowTaskConfiguration {
	return func(fts *FlowTaskService) error {
		ur, err := mongoFlow.New(
			context.Background(),
			hosts, port, db, user, password, mongoFlow.DefaultCollectionName,
		)
		if err != nil {
			return err
		}
		fts.flow = ur
		return nil
	}
}

func WithMongoFunctionRunRecordRepository(
	hosts []string, port int, db, user, password string,
) FlowTaskConfiguration {
	return func(fts *FlowTaskService) error {
		ur, err := mongoFRR.New(
			context.Background(),
			hosts, port, db, user, password, mongoFRR.DefaultCollectionName,
		)
		if err != nil {
			return err
		}
		fts.flowRunRecord = ur
		return nil
	}
}
