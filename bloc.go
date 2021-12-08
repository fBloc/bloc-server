package bloc

import (
	"context"

	"github.com/fBloc/bloc-backend-go/aggregate"
	"github.com/fBloc/bloc-backend-go/config"
	"github.com/fBloc/bloc-backend-go/event"
	mongo_futureEventStorage "github.com/fBloc/bloc-backend-go/event/mongo_event_storage"
	"github.com/fBloc/bloc-backend-go/infrastructure/log"
	"github.com/fBloc/bloc-backend-go/infrastructure/log_collect_backend"
	minio_logBackend "github.com/fBloc/bloc-backend-go/infrastructure/log_collect_backend/minio"
	"github.com/fBloc/bloc-backend-go/infrastructure/mq"
	"github.com/fBloc/bloc-backend-go/infrastructure/object_storage"
	minioInf "github.com/fBloc/bloc-backend-go/infrastructure/object_storage/minio"
	"github.com/fBloc/bloc-backend-go/internal/conns/minio"
	"github.com/fBloc/bloc-backend-go/internal/conns/mongodb"
	"github.com/fBloc/bloc-backend-go/internal/util"
	"github.com/fBloc/bloc-backend-go/pkg/function_developer_implement"
	flow_repository "github.com/fBloc/bloc-backend-go/repository/flow"
	mongo_flow "github.com/fBloc/bloc-backend-go/repository/flow/mongo"
	flowRunRecord_repository "github.com/fBloc/bloc-backend-go/repository/flow_run_record"
	mongo_flowRunRecord "github.com/fBloc/bloc-backend-go/repository/flow_run_record/mongo"
	function_repository "github.com/fBloc/bloc-backend-go/repository/function"
	mongo_func "github.com/fBloc/bloc-backend-go/repository/function/mongo"
	function_execute_heartbeat_repository "github.com/fBloc/bloc-backend-go/repository/function_execute_heartbeat"
	mongo_funcRunHBeat "github.com/fBloc/bloc-backend-go/repository/function_execute_heartbeat/mongo"
	funcRunRec_repository "github.com/fBloc/bloc-backend-go/repository/function_run_record"
	mongo_funcRunRecord "github.com/fBloc/bloc-backend-go/repository/function_run_record/mongo"
	mongo_user "github.com/fBloc/bloc-backend-go/repository/user/mongo"
	"github.com/fBloc/bloc-backend-go/value_object"

	"fmt"
	"net"
	"strings"
	"sync"

	"github.com/fBloc/bloc-backend-go/infrastructure/mq/rabbit"
	user_repository "github.com/fBloc/bloc-backend-go/repository/user"
)

type DefaultUserConfig struct {
	Name     string
	Password string
}

func (dUC *DefaultUserConfig) IsNil() bool {
	if dUC == nil {
		return true
	}
	return dUC.Name != "" && dUC.Password != ""
}

type HttpServerConfig struct {
	IP   string
	Port int
}

func (hSC *HttpServerConfig) HttpAddress() string {
	return fmt.Sprintf("%s:%d", hSC.IP, hSC.Port)
}

type LogConfig struct {
	MaxKeepDays int
}

func (lF *LogConfig) IsNil() bool {
	if lF == nil {
		return true
	}
	return lF.MaxKeepDays == 0
}

type ConfigBuilder struct {
	DefaultUserConf *DefaultUserConfig
	HttpServerConf  *HttpServerConfig
	RabbitConf      *rabbit.RabbitConfig
	mongoConf       *mongodb.MongoConfig
	minioConf       *minio.MinioConfig
	LogConf         *LogConfig
}

func (confbder *ConfigBuilder) SetDefaultUser(name, password string) *ConfigBuilder {
	confbder.DefaultUserConf.Name = name
	confbder.DefaultUserConf.Password = password
	return confbder
}

func (confbder *ConfigBuilder) SetHttpServer(ip string, port int) *ConfigBuilder {
	confbder.HttpServerConf = &HttpServerConfig{IP: ip, Port: port}
	return confbder
}

func (confbder *ConfigBuilder) SetRabbitConfig(
	user, password, host string, port int, vHost string,
) *ConfigBuilder {
	confbder.RabbitConf = &rabbit.RabbitConfig{
		User:     user,
		Password: password,
		Host:     host,
		Port:     port,
		Vhost:    vHost}
	return confbder
}

func (confbder *ConfigBuilder) SetMongoConfig(hosts []string, port int, db, user, password string) *ConfigBuilder {
	confbder.mongoConf = &mongodb.MongoConfig{
		Hosts:    hosts,
		Port:     port,
		Db:       db,
		User:     user,
		Password: password,
	}
	return confbder
}

func (confbder *ConfigBuilder) SetMinioConfig(
	bucketName string, addresses []string, key, password string) *ConfigBuilder {
	// minio名称不允许有下划线
	bucketName = strings.Replace(bucketName, "_", "", -1)
	confbder.minioConf = &minio.MinioConfig{
		BucketName:     bucketName,
		Addresses:      addresses,
		AccessKey:      key,
		AccessPassword: password}
	return confbder
}

func (confbder *ConfigBuilder) SetLogConfig(maxKeepDays int) *ConfigBuilder {
	confbder.LogConf = &LogConfig{
		MaxKeepDays: maxKeepDays}
	return confbder
}

// BuildUp 对于必须要输入的做输入检查 & 有效性检查
func (congbder *ConfigBuilder) BuildUp() {
	// DefaultUserConf 默认超级用户设置检测
	if congbder.DefaultUserConf.IsNil() {
		congbder.DefaultUserConf = &DefaultUserConfig{
			Name:     config.DefaultUserName,
			Password: config.DefaultUserPassword,
		}
	}

	// HttpServerConf http server 地址配置。
	// 不用检查，没有设置就自行分配就是了

	// RabbitConf。需要检查输入的配置能够建立有效的链接
	if congbder.RabbitConf.IsNil() {
		panic("must set rabbit config")
	}
	rabbit.InitChannel(congbder.RabbitConf)

	// mongoConf 查看mongo是否能够有效链接
	if congbder.mongoConf.IsNil() {
		panic("must set mongo config")
	}
	mongodb.CheckConfValid(congbder.mongoConf)

	// minioConf 查看minIO是否能够有效工作
	if congbder.minioConf.IsNil() {
		panic("must set minio config")
	}
	minio.Init(congbder.minioConf)

	// LogConf
	if congbder.LogConf.IsNil() {
		congbder.LogConf = &LogConfig{
			MaxKeepDays: config.DefaultLogKeepDays,
		}
	}
}

type BlocApp struct {
	Name                             string // 构建的项目名称
	FunctionGroups                   []FunctionGroup
	functionRepoIDMapFunction        map[value_object.UUID]aggregate.Function
	functionRepoIDMapExecuteFunction map[value_object.UUID]function_developer_implement.FunctionDeveloperImplementInterface
	configBuilder                    *ConfigBuilder
	httpServerLogger                 *log.Logger
	consumerLogger                   *log.Logger
	logBackEnd                       log_collect_backend.LogBackEnd
	userRepository                   user_repository.UserRepository
	flowRepository                   flow_repository.FlowRepository
	functionRepository               function_repository.FunctionRepository
	functionExecuteHBeatRepository   function_execute_heartbeat_repository.FunctionExecuteHeartbeatRepository
	functionRunRecordRepository      funcRunRec_repository.FunctionRunRecordRepository
	flowRunRecordRepository          flowRunRecord_repository.FlowRunRecordRepository
	eventMQ                          mq.MsgQueue
	futureEventStorage               event.FuturePubEventStorage
	consumerObjectStorage            object_storage.ObjectStorage
	sync.Mutex
}

// GetConfigBuilder
func (bloc *BlocApp) GetConfigBuilder() *ConfigBuilder {
	bloc.configBuilder = &ConfigBuilder{}
	return bloc.configBuilder
}

func NewBlocApp(appName string) *BlocApp {
	return &BlocApp{Name: appName}
}

func (bA *BlocApp) HttpAddress() string {
	return bA.configBuilder.HttpServerConf.HttpAddress()
}

func (bA *BlocApp) InitialUserInfo() (name, rawPassword string) {
	return bA.configBuilder.DefaultUserConf.Name, bA.configBuilder.DefaultUserConf.Password
}

func (bA *BlocApp) AllFunctions() []*aggregate.Function {
	var ret []*aggregate.Function
	for _, funcGroup := range bA.FunctionGroups {
		ret = append(ret, funcGroup.Functions...)
	}
	return ret
}

func (bA *BlocApp) HttpListener() net.Listener {
	if bA.configBuilder.HttpServerConf == nil {
		ip, port, listener := util.NewAutoAddressNetListener()
		bA.configBuilder.HttpServerConf.IP = ip
		bA.configBuilder.HttpServerConf.Port = port
		return listener
	}
	return util.NewNetListener(
		bA.configBuilder.HttpServerConf.IP,
		bA.configBuilder.HttpServerConf.Port)
}

func (bA *BlocApp) RegisterFunctionGroup(name string) *FunctionGroup {
	for _, i := range bA.FunctionGroups {
		if i.Name == name {
			panic("should not register same name group")
		}
	}
	functionGroup := FunctionGroup{
		Name:    name,
		blocApp: bA}
	bA.FunctionGroups = append(bA.FunctionGroups, functionGroup)
	return &functionGroup
}

func (bA *BlocApp) GetOrCreateLogBackEnd() log_collect_backend.LogBackEnd {
	if bA.logBackEnd != nil {
		return bA.logBackEnd
	}

	bA.logBackEnd = minio_logBackend.New(
		bA.configBuilder.minioConf.BucketName,
		bA.configBuilder.minioConf.Addresses,
		bA.configBuilder.minioConf.AccessKey,
		bA.configBuilder.minioConf.AccessPassword)
	return bA.logBackEnd
}

func (bA *BlocApp) GetOrCreateHttpLogger() *log.Logger {
	bA.Lock()
	defer bA.Unlock()
	if !bA.httpServerLogger.IsZero() {
		return bA.httpServerLogger
	}

	bA.httpServerLogger = log.NewWithPeriodicUpload(
		value_object.HttpServerLog.String(),
		bA.GetOrCreateLogBackEnd())
	return bA.httpServerLogger
}

func (bA *BlocApp) GetOrCreateConsumerLogger() *log.Logger {
	bA.Lock()
	defer bA.Unlock()
	if !bA.consumerLogger.IsZero() {
		return bA.consumerLogger
	}

	logBackEnd := bA.GetOrCreateLogBackEnd()
	logger := log.NewWithPeriodicUpload(
		value_object.ConsumerLog.String(),
		logBackEnd)
	bA.consumerLogger = logger
	return bA.consumerLogger
}

func (bA *BlocApp) CreateFunctionRunLogger(funcRunRecordID value_object.UUID) *log.Logger {
	logBackEnd := bA.GetOrCreateLogBackEnd()
	return log.NewWithPeriodicUpload(
		value_object.FuncRunRecordLog.String()+"-"+funcRunRecordID.String(),
		logBackEnd)
}

func (bA *BlocApp) GetOrCreateUserRepository() user_repository.UserRepository {
	bA.Lock()
	defer bA.Unlock()
	if bA.userRepository != nil {
		return bA.userRepository
	}

	ur, err := mongo_user.New(
		context.Background(),
		bA.configBuilder.mongoConf.Hosts,
		bA.configBuilder.mongoConf.Port,
		bA.configBuilder.mongoConf.User,
		bA.configBuilder.mongoConf.Password,
		bA.configBuilder.mongoConf.Db,
		mongo_user.DefaultCollectionName,
	)
	if err != nil {
		panic(err)
	}

	bA.userRepository = ur
	return bA.userRepository
}

func (bA *BlocApp) GetOrCreateFlowRepository() flow_repository.FlowRepository {
	bA.Lock()
	defer bA.Unlock()
	if bA.flowRepository != nil {
		return bA.flowRepository
	}

	fr, err := mongo_flow.New(
		context.Background(),
		bA.configBuilder.mongoConf.Hosts,
		bA.configBuilder.mongoConf.Port,
		bA.configBuilder.mongoConf.User,
		bA.configBuilder.mongoConf.Password,
		bA.configBuilder.mongoConf.Db,
		mongo_flow.DefaultCollectionName,
	)
	if err != nil {
		panic(err)
	}

	bA.flowRepository = fr
	return fr
}

func (bA *BlocApp) GetOrCreateFunctionRepository() function_repository.FunctionRepository {
	bA.Lock()
	defer bA.Unlock()
	if bA.functionRepository != nil {
		return bA.functionRepository
	}

	fR, err := mongo_func.New(
		context.Background(),
		bA.configBuilder.mongoConf.Hosts,
		bA.configBuilder.mongoConf.Port,
		bA.configBuilder.mongoConf.User,
		bA.configBuilder.mongoConf.Password,
		bA.configBuilder.mongoConf.Db,
		mongo_func.DefaultCollectionName,
	)
	if err != nil {
		panic(err)
	}

	bA.functionRepository = fR
	return bA.functionRepository
}

func (bA *BlocApp) GetOrCreateFunctionRunRecordRepository() funcRunRec_repository.FunctionRunRecordRepository {
	bA.Lock()
	defer bA.Unlock()
	if bA.functionRunRecordRepository != nil {
		return bA.functionRunRecordRepository
	}

	fR, err := mongo_funcRunRecord.New(
		context.Background(),
		bA.configBuilder.mongoConf.Hosts,
		bA.configBuilder.mongoConf.Port,
		bA.configBuilder.mongoConf.User,
		bA.configBuilder.mongoConf.Password,
		bA.configBuilder.mongoConf.Db,
		mongo_funcRunRecord.DefaultCollectionName,
	)
	if err != nil {
		panic(err)
	}

	bA.functionRunRecordRepository = fR
	return bA.functionRunRecordRepository
}

func (bA *BlocApp) GetOrCreateFlowRunRecordRepository() flowRunRecord_repository.FlowRunRecordRepository {
	bA.Lock()
	defer bA.Unlock()
	if bA.flowRunRecordRepository != nil {
		return bA.flowRunRecordRepository
	}

	fR, err := mongo_flowRunRecord.New(
		context.Background(),
		bA.configBuilder.mongoConf.Hosts,
		bA.configBuilder.mongoConf.Port,
		bA.configBuilder.mongoConf.User,
		bA.configBuilder.mongoConf.Password,
		bA.configBuilder.mongoConf.Db,
		mongo_flowRunRecord.DefaultCollectionName,
	)
	if err != nil {
		panic(err)
	}

	bA.flowRunRecordRepository = fR
	return bA.flowRunRecordRepository
}

func (bA *BlocApp) GetOrCreateFuncRunHBeatRepository() function_execute_heartbeat_repository.FunctionExecuteHeartbeatRepository {
	bA.Lock()
	defer bA.Unlock()
	if bA.functionExecuteHBeatRepository != nil {
		return bA.functionExecuteHBeatRepository
	}

	fR, err := mongo_funcRunHBeat.New(
		context.Background(),
		bA.configBuilder.mongoConf.Hosts,
		bA.configBuilder.mongoConf.Port,
		bA.configBuilder.mongoConf.User,
		bA.configBuilder.mongoConf.Password,
		bA.configBuilder.mongoConf.Db,
		mongo_funcRunHBeat.DefaultCollectionName,
	)
	if err != nil {
		panic(err)
	}

	bA.functionExecuteHBeatRepository = fR
	return bA.functionExecuteHBeatRepository
}

func (bA *BlocApp) GetOrCreateEventMQ() mq.MsgQueue {
	bA.Lock()
	defer bA.Unlock()
	if bA.eventMQ != nil {
		return bA.eventMQ
	}

	rabbitMQ := rabbit.InitChannel(bA.configBuilder.RabbitConf)
	bA.eventMQ = rabbitMQ

	return bA.eventMQ
}

func (bA *BlocApp) GetOrCreateFutureEventStorage() event.FuturePubEventStorage {
	bA.Lock()
	defer bA.Unlock()
	if bA.futureEventStorage != nil {
		return bA.futureEventStorage
	}

	mongoStorage, err := mongo_futureEventStorage.New(
		context.Background(),
		bA.configBuilder.mongoConf.Hosts,
		bA.configBuilder.mongoConf.Port,
		bA.configBuilder.mongoConf.User,
		bA.configBuilder.mongoConf.Password,
		bA.configBuilder.mongoConf.Db,
		mongo_futureEventStorage.DefaultCollectionName,
	)
	if err != nil {
		panic(err)
	}
	bA.futureEventStorage = mongoStorage

	return bA.futureEventStorage
}

func (bA *BlocApp) GetOrCreateConsumerObjectStorage() object_storage.ObjectStorage {
	bA.Lock()
	defer bA.Unlock()
	if bA.consumerObjectStorage != nil {
		return bA.consumerObjectStorage
	}

	minioOS := minioInf.New(
		bA.configBuilder.minioConf.Addresses,
		bA.configBuilder.minioConf.AccessKey,
		bA.configBuilder.minioConf.AccessPassword,
		bA.configBuilder.minioConf.BucketName,
	)
	bA.consumerObjectStorage = minioOS

	return bA.consumerObjectStorage
}

func (bA *BlocApp) GetFunctionByRepoID(functionRepoID value_object.UUID) aggregate.Function {
	if bA.functionRepoIDMapFunction == nil {
		bA.functionRepoIDMapFunction = make(map[value_object.UUID]aggregate.Function)
	}
	if ins, ok := bA.functionRepoIDMapFunction[functionRepoID]; ok {
		return ins
	}
	functionRepo := bA.GetOrCreateFunctionRepository()
	allFunctions, err := functionRepo.All()
	if err != nil {
		panic(err)
	}

	tmp := make(map[value_object.UUID]aggregate.Function, len(allFunctions))
	for _, i := range allFunctions {
		tmp[i.ID] = *i
	}

	bA.Lock()
	defer bA.Unlock()
	bA.functionRepoIDMapFunction = tmp

	return bA.functionRepoIDMapFunction[functionRepoID]
}

func (bA *BlocApp) GetExecuteFunctionByRepoID(
	functionRepoID value_object.UUID,
) function_developer_implement.FunctionDeveloperImplementInterface {
	return bA.functionRepoIDMapExecuteFunction[functionRepoID]
}
