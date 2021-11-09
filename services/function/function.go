package function

import (
	"context"

	"github.com/fBloc/bloc-backend-go/infrastructure/log"
	minioLog "github.com/fBloc/bloc-backend-go/infrastructure/log/minio"
	"github.com/fBloc/bloc-backend-go/repository/function"
	mongoFunc "github.com/fBloc/bloc-backend-go/repository/function/mongo"
	user_cache "github.com/fBloc/bloc-backend-go/services/userid_cache"
)

type FunctionConfiguration func(fs *FunctionService) error

type FunctionService struct {
	logger           log.Logger
	Function         function.FunctionRepository
	UserCacheService *user_cache.UserCacheService
}

func NewFunctionService(
	cfgs ...FunctionConfiguration,
) (*FunctionService, error) {
	fs := &FunctionService{}
	for _, cfg := range cfgs {
		err := cfg(fs)
		if err != nil {
			return nil, err
		}
	}
	return fs, nil
}

func WithLogger(logger log.Logger) FunctionConfiguration {
	return func(us *FunctionService) error {
		us.logger = logger
		return nil
	}
}

func WithMinioLogger(
	name string, bucketName string, addresses []string, key, password string,
) FunctionConfiguration {
	return func(fs *FunctionService) error {
		fs.logger = minioLog.New(name, bucketName, addresses, key, password)
		return nil
	}
}

func WithFunctionRepository(
	fR function.FunctionRepository,
) FunctionConfiguration {
	return func(fs *FunctionService) error {
		fs.Function = fR
		return nil
	}
}

func WithMongoFunctionRepository(
	hosts []string, port int, db, user, password string,
) FunctionConfiguration {
	return func(fs *FunctionService) error {
		ur, err := mongoFunc.New(
			context.Background(),
			hosts, port, db, user, password, mongoFunc.DefaultCollectionName,
		)
		if err != nil {
			return err
		}
		fs.Function = ur
		return nil
	}
}

func WithUserCacheService(
	userCacheService *user_cache.UserCacheService,
) FunctionConfiguration {
	return func(t *FunctionService) error {
		t.UserCacheService = userCacheService
		return nil
	}
}
