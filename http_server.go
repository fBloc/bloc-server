package bloc

import (
	"log"
	"net/http"

	"github.com/fBloc/bloc-backend-go/event"
	"github.com/fBloc/bloc-backend-go/interfaces/web/client"
	"github.com/fBloc/bloc-backend-go/interfaces/web/flow"
	"github.com/fBloc/bloc-backend-go/interfaces/web/flow_run_record"
	"github.com/fBloc/bloc-backend-go/interfaces/web/function"
	"github.com/fBloc/bloc-backend-go/interfaces/web/function_run_record"
	"github.com/fBloc/bloc-backend-go/interfaces/web/log_data"
	"github.com/fBloc/bloc-backend-go/interfaces/web/middleware"
	"github.com/fBloc/bloc-backend-go/interfaces/web/object_storage"
	"github.com/fBloc/bloc-backend-go/interfaces/web/user"
	flow_service "github.com/fBloc/bloc-backend-go/services/flow"
	flowRunRecord_service "github.com/fBloc/bloc-backend-go/services/flow_run_record"
	function_service "github.com/fBloc/bloc-backend-go/services/function"
	functionRunRecord_service "github.com/fBloc/bloc-backend-go/services/function_run_record"
	user_service "github.com/fBloc/bloc-backend-go/services/user"
	user_cache "github.com/fBloc/bloc-backend-go/services/userid_cache"

	"github.com/julienschmidt/httprouter"
	"github.com/rs/cors"
)

func (blocApp *BlocApp) RunHttpServer() {
	router := httprouter.New()

	httpConsumer := blocApp.GetOrCreateHttpLogger()
	// TODO 放这里合适吗？
	event.InjectMq(blocApp.GetOrCreateEventMQ())

	uCacheService, err := user_cache.NewUserCacheService(
		user_cache.WithLogger(httpConsumer),
		user_cache.WithUser(blocApp.GetOrCreateUserRepository()),
	)
	if err != nil {
		panic(err)
	}

	// middleware 依赖资源注入
	middleware.InjectUserIDCacheService(uCacheService)

	// user
	{
		// initial relied services
		userService, err := user_service.NewUserService(
			user_service.WithLogger(httpConsumer),
			user_service.WithUserRepository(blocApp.GetOrCreateUserRepository()))
		if err != nil {
			panic(err)
		}
		user.InjectUserService(userService)

		// 确保默认用户存在（否则没法登录前端、查看功能）
		initialUserName, initialUserPasswd := blocApp.InitialUserInfo()
		user.InitialUserExistOrCreate(initialUserName, initialUserPasswd)

		// router
		router.POST("/api/v1/login", user.LoginHandler)

		basicPath := "/api/v1/user"
		router.GET(basicPath, middleware.LoginAuth(user.FilterByName))
		router.POST(basicPath, middleware.SuperuserAuth(user.AddUser))
		router.DELETE(basicPath+"/delete_by_id/:id", middleware.SuperuserAuth(user.DeleteUser))
	}

	// function
	{
		// initial relied services
		funcService, err := function_service.NewFunctionService(
			function_service.WithLogger(httpConsumer),
			function_service.WithFunctionRepository(
				blocApp.GetOrCreateFunctionRepository()),
			function_service.WithUserCacheService(uCacheService),
		)
		if err != nil {
			panic(err)
		}
		function.InjectFunctionService(funcService)

		// 确保用户实现的函数已经持久化到存储层了
		function.MakeSureAllUserImplementFunctionsInRepository(
			blocApp.AllFunctions(),
		)

		// router
		{
			// function本身
			basicPath := "/api/v1/function"
			router.GET(basicPath, middleware.LoginAuth(function.All))
		}
		{
			// function权限
			basicPath := "/api/v1/function_permission"
			router.GET(basicPath, middleware.LoginAuth(function.GetPermissionByFunctionID))
			router.POST(basicPath+"/add_permission", middleware.LoginAuth(function.AddUserPermission))
			router.DELETE(basicPath+"/remove_permission", middleware.LoginAuth(function.DeleteUserPermission))
		}
	}

	// function_run_record
	{
		fRRS, err := functionRunRecord_service.NewService(
			functionRunRecord_service.WithLogger(httpConsumer),
			functionRunRecord_service.WithFunctionRunRecordRepository(
				blocApp.GetOrCreateFunctionRunRecordRepository()),
			functionRunRecord_service.WithUserCacheService(uCacheService),
		)
		if err != nil {
			panic(err)
		}
		function_run_record.InjectFunctionRunRecordService(fRRS)
		function_run_record.InjectLogCollectBackend(blocApp.GetOrCreateLogBackEnd())

		// router
		basicPath := "/api/v1/function_run_record"
		router.GET(basicPath, middleware.LoginAuth(function_run_record.Filter))
		router.GET(basicPath+"/get_by_id/:id", middleware.LoginAuth(function_run_record.Get))
		router.GET(basicPath+"/list_logkeys_by_id/:id", middleware.LoginAuth(function_run_record.ListLogKeys))
		router.GET(basicPath+"/get_log_by_logkey/:log_key", middleware.LoginAuth(function_run_record.GetLogByKey))
	}

	// flow相关
	{
		// initial relied services
		flowService, err := flow_service.NewFlowService(
			flow_service.WithLogger(httpConsumer),
			flow_service.WithFlowRepository(
				blocApp.GetOrCreateFlowRepository(),
			),
			flow_service.WithFunctionRepository(
				blocApp.GetOrCreateFunctionRepository(),
			),
			flow_service.WithFunctionRunRecordRepository(
				blocApp.GetOrCreateFunctionRunRecordRepository(),
			),
			flow_service.WithFlowRunRecordRepository(
				blocApp.GetOrCreateFlowRunRecordRepository(),
			),
			flow_service.WithUserCacheService(uCacheService),
		)
		if err != nil {
			panic(err)
		}
		flow.InjectFlowService(flowService)

		// config
		{
			// 约定的一些字段
			basicPath := "/api/v1/configs"
			router.GET(basicPath, middleware.LoginAuth(flow.FilterDraftByName))
		}

		// router
		{
			// draft flow
			basicPath := "/api/v1/draft_flow"
			router.GET(basicPath, middleware.LoginAuth(flow.FilterDraftByName))
			router.GET(basicPath+"/get_by_origin_id/:origin_id", middleware.LoginAuth(flow.GetDraftByOriginID))
			router.GET(basicPath+"/commit_by_id/:id", middleware.LoginAuth(flow.PubDraft))
			router.POST(basicPath, middleware.LoginAuth(flow.CreateDraft))
			router.PATCH(basicPath, middleware.LoginAuth(flow.UpdateDraft))
			router.DELETE(
				basicPath+"/delete_by_origin_id/:origin_id",
				middleware.LoginAuth(flow.DeleteDraftByOriginID))
		}

		{
			// online flow
			basicPath := "/api/v1/flow"
			router.GET(basicPath, middleware.LoginAuth(flow.FilterFlow))
			router.GET(basicPath+"/get_by_id/:id", middleware.LoginAuth(flow.GetFlowByID))
			router.GET(basicPath+"/get_latestonline_by_origin_id/:origin_id", middleware.LoginAuth(flow.GetFlowByOriginID))
			router.PATCH(basicPath+"/set_execute_control_attributes", middleware.LoginAuth(flow.SetExecuteControlAttributes))
			router.DELETE(basicPath+"delete_by_origin_id/:origin_id", middleware.LoginAuth(flow.DeleteFlowByOriginID))
		}

		{
			// 运行相关
			basicPath := "/api/v1/flow"
			router.GET(basicPath+"/run/by_origin_id/:origin_id", middleware.LoginAuth(flow.Run))
			router.GET(basicPath+"/cancel_run/by_origin_id/:origin_id", middleware.LoginAuth(flow.CancelRun))
		}

		{
			// 权限
			basicPath := "/api/v1/flow_permission"
			router.GET(basicPath, middleware.LoginAuth(flow.GetPermission))
			router.POST(basicPath+"/add_permission", middleware.LoginAuth(flow.AddUserPermission))
			router.DELETE(basicPath+"/remove_permission", middleware.LoginAuth(flow.DeleteUserPermission))
		}
	}

	// flow_run_record
	{
		// initial relied services
		flowRunRecordService, err := flowRunRecord_service.NewService(
			flowRunRecord_service.WithLogger(httpConsumer),
			flowRunRecord_service.WithFlowRunRecordRepository(
				blocApp.GetOrCreateFlowRunRecordRepository()),
			flowRunRecord_service.WithUserCacheService(uCacheService),
		)
		if err != nil {
			panic(err)
		}
		flow_run_record.InjectFlowRunRecordService(flowRunRecordService)

		// router
		basicPath := "/api/v1/flow_run_record"
		router.GET(basicPath, middleware.LoginAuth(flow_run_record.Filter))
	}

	// object storage
	{
		object_storage.InjectObjectStorageImplement(
			blocApp.GetOrCreateConsumerObjectStorage(),
		)
		{
			basicPath := "/api/v1/object_storage"
			router.GET(basicPath+"/get_string_value_by_key/:key", object_storage.GetValueByKeyReturnString)
		}
	}

	// log
	{
		log_data.InjectLogCollectBackend(
			blocApp.GetOrCreateLogBackEnd(),
		)
		{
			basicPath := "/api/v1/log"
			router.POST(basicPath+"/pull_log_between_time", log_data.PullLog)
		}
	}

	// function provider client
	{

		funcService, err := function_service.NewFunctionService(
			function_service.WithLogger(httpConsumer),
			function_service.WithFunctionRepository(
				blocApp.GetOrCreateFunctionRepository()),
			function_service.WithUserCacheService(uCacheService),
		)
		if err != nil {
			panic(err)
		}
		client.InjectFunctionService(funcService)
		client.InjectLogBackend(blocApp.GetOrCreateLogBackEnd())
		fRRS, err := functionRunRecord_service.NewService(
			functionRunRecord_service.WithLogger(httpConsumer),
			functionRunRecord_service.WithFunctionRunRecordRepository(
				blocApp.GetOrCreateFunctionRunRecordRepository()),
			functionRunRecord_service.WithUserCacheService(uCacheService),
		)
		if err != nil {
			panic(err)
		}
		client.InjectFunctionRunRecordService(fRRS)

		flowService, err := flow_service.NewFlowService(
			flow_service.WithLogger(httpConsumer),
			flow_service.WithFlowRepository(
				blocApp.GetOrCreateFlowRepository(),
			),
			flow_service.WithFunctionRepository(
				blocApp.GetOrCreateFunctionRepository(),
			),
			flow_service.WithFunctionRunRecordRepository(
				blocApp.GetOrCreateFunctionRunRecordRepository(),
			),
			flow_service.WithFlowRunRecordRepository(
				blocApp.GetOrCreateFlowRunRecordRepository(),
			),
			flow_service.WithUserCacheService(uCacheService),
		)
		if err != nil {
			panic(err)
		}
		client.InjectFlowService(flowService)

		client.InjectConsumerLogger(blocApp.GetOrCreateConsumerLogger())

		flowRunRecordService, err := flowRunRecord_service.NewService(
			flowRunRecord_service.WithLogger(httpConsumer),
			flowRunRecord_service.WithFlowRunRecordRepository(
				blocApp.GetOrCreateFlowRunRecordRepository()),
			flowRunRecord_service.WithUserCacheService(uCacheService),
		)
		if err != nil {
			panic(err)
		}
		client.InjectFlowRunRecordService(flowRunRecordService)

		basicPath := "/api/v1/client"
		{
			router.POST(basicPath+"/register_functions", client.RegisterFunctions)
			router.POST(basicPath+"/report_log", client.ReportLog)
			router.POST(basicPath+"/report_progress", client.ReportProgress)
			router.POST(basicPath+"/function_run_finished", client.FunctionRunFinished)
			router.GET(basicPath+"/get_function_run_record_by_id/:id", function_run_record.Get)
			router.GET(basicPath+"/check_flowRun_is_canceled_by_flowRunID/:id", client.FlowRunRecordIsCanceled)
		}
	}

	// start http server
	log.Printf("start http server at http://%s", blocApp.HttpAddress())
	handler := cors.AllowAll().Handler(router)
	log.Fatal(http.Serve(blocApp.HttpListener(), handler))
}
