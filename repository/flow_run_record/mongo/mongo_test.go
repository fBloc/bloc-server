package mongo

import (
	"context"
	"fmt"
	"log"
	"os"
	"testing"
	"time"

	"github.com/brianvoe/gofakeit/v6"
	"github.com/fBloc/bloc-server/aggregate"
	"github.com/fBloc/bloc-server/config"
	"github.com/fBloc/bloc-server/internal/conns/mongodb"
	"github.com/fBloc/bloc-server/pkg/ipt"
	"github.com/fBloc/bloc-server/pkg/opt"
	"github.com/fBloc/bloc-server/pkg/value_type"
	"github.com/fBloc/bloc-server/value_object"
	"github.com/ory/dockertest/v3"
	"github.com/ory/dockertest/v3/docker"
	. "github.com/smartystreets/goconvey/convey"
)

var (
	epo  *MongoRepository
	conf = &mongodb.MongoConfig{
		Db:       "bloc-test-mongo",
		User:     "root",
		Password: "password",
	}
)

var (
	readeUser         = aggregate.User{ID: value_object.NewUUID()}
	executeUser       = aggregate.User{Name: gofakeit.Name(), ID: value_object.NewUUID()}
	allPermissionUser = aggregate.User{ID: value_object.NewUUID()}
)

var functionAdd = aggregate.Function{
	ID:           value_object.NewUUID(),
	Name:         "two add",
	GroupName:    "math operation",
	ProviderName: "test",
	Description:  "test function",
	Ipts: ipt.IptSlice{
		{
			Key:     "to_add_ints",
			Display: "to_add_ints",
			Must:    true,
			Components: []*ipt.IptComponent{
				{
					ValueType:       value_type.IntValueType,
					FormControlType: value_object.InputFormControl,
					Hint:            "加数",
					DefaultValue:    0,
					AllowMulti:      true,
				},
			},
		},
	},
	Opts: []*opt.Opt{
		{
			Key:         "sum",
			Description: "sum of your inputs",
			ValueType:   value_type.IntValueType,
			IsArray:     false,
		},
		{
			Key:         "describe",
			Description: "diff value type opt 4 test",
			ValueType:   value_type.StringValueType,
			IsArray:     false,
		},
	},
	ProcessStages:           []string{"parsing ipt", "finished parse ipt & start do the math", "finished"},
	ReadUserIDs:             []value_object.UUID{readeUser.ID, allPermissionUser.ID},
	ExecuteUserIDs:          []value_object.UUID{executeUser.ID, allPermissionUser.ID},
	AssignPermissionUserIDs: []value_object.UUID{allPermissionUser.ID},
}

var functionMultiply = aggregate.Function{
	ID:           value_object.NewUUID(),
	Name:         "two multiply",
	GroupName:    "math operation",
	ProviderName: "test",
	Description:  "test function",
	Ipts: ipt.IptSlice{
		{
			Key:     "multiplier",
			Display: "multiplier",
			Must:    true,
			Components: []*ipt.IptComponent{
				{
					ValueType:       value_type.IntValueType,
					FormControlType: value_object.InputFormControl,
					Hint:            "乘数",
					DefaultValue:    0,
					AllowMulti:      true,
				},
			},
		},
		{
			Key:     "multiplicand",
			Display: "multiplicand",
			Must:    true,
			Components: []*ipt.IptComponent{
				{
					ValueType:       value_type.IntValueType,
					FormControlType: value_object.InputFormControl,
					Hint:            "被乘数",
					DefaultValue:    0,
					AllowMulti:      true,
				},
			},
		},
	},
	Opts: []*opt.Opt{
		{
			Key:         "result",
			Description: "result of multiply",
			ValueType:   value_type.IntValueType,
			IsArray:     false,
		},
	},
	ProcessStages:           []string{"parsing ipt", "finished parse ipt & start do the math", "finished"},
	ReadUserIDs:             []value_object.UUID{readeUser.ID, allPermissionUser.ID},
	ExecuteUserIDs:          []value_object.UUID{executeUser.ID, allPermissionUser.ID},
	AssignPermissionUserIDs: []value_object.UUID{allPermissionUser.ID},
}

var (
	secondFlowFunctionID               = value_object.NewUUID().String()
	thirdFlowFunctionID                = value_object.NewUUID().String()
	validFlowFunctionIDMapFlowFunction = map[string]*aggregate.FlowFunction{
		config.FlowFunctionStartID: {
			FunctionID:                value_object.NillUUID,
			Note:                      "start node",
			UpstreamFlowFunctionIDs:   []string{},
			DownstreamFlowFunctionIDs: []string{secondFlowFunctionID},
			ParamIpts:                 [][]aggregate.IptComponentConfig{},
		},
		secondFlowFunctionID: {
			FunctionID:                functionAdd.ID,
			Function:                  &functionAdd,
			Note:                      "add",
			UpstreamFlowFunctionIDs:   []string{config.FlowFunctionStartID},
			DownstreamFlowFunctionIDs: []string{thirdFlowFunctionID},
			ParamIpts: [][]aggregate.IptComponentConfig{
				{
					{
						Blank:     false,
						IptWay:    value_object.UserIpt,
						ValueType: value_type.StringValueType,
						Value:     []int{1, 2, 3},
					},
				},
			},
		},
		thirdFlowFunctionID: {
			FunctionID:                functionMultiply.ID,
			Function:                  &functionMultiply,
			Note:                      "multiply",
			UpstreamFlowFunctionIDs:   []string{secondFlowFunctionID},
			DownstreamFlowFunctionIDs: []string{},
			ParamIpts: [][]aggregate.IptComponentConfig{
				{
					{
						Blank:          false,
						IptWay:         value_object.Connection,
						ValueType:      value_type.IntValueType,
						FlowFunctionID: secondFlowFunctionID,
						Key:            "sum",
					},
				},
				{
					{
						Blank:     false,
						IptWay:    value_object.UserIpt,
						ValueType: value_type.IntValueType,
						Value:     10,
					},
				},
			},
		},
	}
	fakeAggregateFlow = aggregate.Flow{
		ID:                            value_object.NewUUID(),
		Name:                          gofakeit.Name(),
		IsDraft:                       true,
		OriginID:                      value_object.NewUUID(),
		CreateUserID:                  value_object.NewUUID(),
		FlowFunctionIDMapFlowFunction: validFlowFunctionIDMapFlowFunction,
		ExecuteUserIDs:                []value_object.UUID{executeUser.ID},
	}
)

func TestCreate(t *testing.T) {
	traceID := value_object.NewTraceID()
	ctx := value_object.SetTraceIDToContext(traceID)
	flowRR, err := aggregate.NewUserTriggeredFlowRunRecord(ctx, &fakeAggregateFlow, &executeUser)
	if err != nil {
		log.Fatal(err)
	}
	if flowRR.IsZero() {
		log.Fatal()
	}

	err = epo.Create(flowRR)
	if err != nil {
		log.Fatal(err)
	}

	Convey("Get", t, func() {
		Convey("GetByID", func() {
			fRR, err := epo.GetByID(flowRR.ID)
			So(err, ShouldBeNil)
			So(fRR.IsZero(), ShouldBeFalse)
			So(fRR.TraceID, ShouldEqual, traceID)
		})

		Convey("GetLatestByFlowOriginID", func() {
			fRR, err := epo.GetLatestByFlowOriginID(flowRR.FlowOriginID)
			So(err, ShouldBeNil)
			So(fRR.IsZero(), ShouldBeFalse)
		})

		Convey("GetLatestByFlowID", func() {
			fRR, err := epo.GetLatestByFlowID(flowRR.FlowID)
			So(err, ShouldBeNil)
			So(fRR.IsZero(), ShouldBeFalse)
		})
	})

	Convey("Filter", t, func() {
		Convey("AllRunRecordOfFlowTriggeredByFlowID", func() {
			fRRs, err := epo.AllRunRecordOfFlowTriggeredByFlowID(flowRR.FlowID)
			So(err, ShouldBeNil)
			So(len(fRRs), ShouldEqual, 1)
		})

		Convey("Filter", func() {
			fRRs, err := epo.Filter(
				*value_object.NewRepositoryFilter(),
				value_object.RepositoryFilterOption{})
			So(err, ShouldBeNil)
			So(len(fRRs), ShouldEqual, 1)
		})
	})

	Convey("Count", t, func() {
		amount, err := epo.Count(
			*value_object.NewRepositoryFilter().AddEqual("id", flowRR.ID))
		So(err, ShouldBeNil)
		So(amount, ShouldEqual, 1)
	})

	Convey("attrs", t, func() {
		Convey("IsHaveRunningTask", func() {
			doHaveRunningTask, err := epo.IsHaveRunningTask(flowRR.FlowID, flowRR.ID)
			So(err, ShouldBeNil)
			So(doHaveRunningTask, ShouldBeFalse)
		})
	})

	Convey("status change", t, func() {
		Convey("start", func() {
			fRR, _ := epo.GetByID(flowRR.ID)
			So(fRR.StartTime.IsZero(), ShouldBeTrue)

			beforeHappendTime := time.Now()
			time.Sleep(time.Second)
			err := epo.Start(fRR.ID)
			So(err, ShouldBeNil)

			fRR, _ = epo.GetByID(flowRR.ID)
			So(fRR.StartTime.Unix(), ShouldBeGreaterThan, beforeHappendTime.Unix())
		})

		Convey("PatchDataForRetry", func() {
			fRR, _ := epo.GetByID(flowRR.ID)
			So(fRR.RetriedAmount, ShouldEqual, 0)

			err := epo.PatchDataForRetry(fRR.ID, 0)
			So(err, ShouldBeNil)

			fRR, _ = epo.GetByID(flowRR.ID)
			So(fRR.RetriedAmount, ShouldEqual, 1)
		})

		Convey("PatchFlowFuncIDMapFuncRunRecordID", func() {
			fRR, _ := epo.GetByID(flowRR.ID)
			So(len(fRR.FlowFuncIDMapFuncRunRecordID), ShouldEqual, 0)

			flowFuncIDMapFuncRunRecordID := map[string]value_object.UUID{
				"key1": value_object.NewUUID()}
			err := epo.PatchFlowFuncIDMapFuncRunRecordID(fRR.ID, flowFuncIDMapFuncRunRecordID)
			So(err, ShouldBeNil)

			fRR, _ = epo.GetByID(flowRR.ID)
			So(fRR.FlowFuncIDMapFuncRunRecordID, ShouldResemble, flowFuncIDMapFuncRunRecordID)

			Convey("AddFlowFuncIDMapFuncRunRecordID", func() {
				secondFlowFunctionRunRecordID := value_object.NewUUID()
				err := epo.AddFlowFuncIDMapFuncRunRecordID(
					fRR.ID,
					secondFlowFunctionID,
					secondFlowFunctionRunRecordID)
				So(err, ShouldBeNil)

				fRR, _ = epo.GetByID(flowRR.ID)
				So(fRR.FlowFuncIDMapFuncRunRecordID[secondFlowFunctionID], ShouldEqual, secondFlowFunctionRunRecordID)
			})
		})

		Convey("Suc", func() {
			fRR, _ := epo.GetByID(flowRR.ID)
			So(fRR.Status, ShouldNotEqual, value_object.Suc)

			err := epo.Suc(fRR.ID)
			So(err, ShouldBeNil)

			fRR, _ = epo.GetByID(flowRR.ID)
			So(fRR.Status, ShouldEqual, value_object.Suc)
		})

		Convey("NotAllowedParallelRun", func() {
			fRR, _ := epo.GetByID(flowRR.ID)
			So(fRR.Status, ShouldNotEqual, value_object.NotAllowedParallelCancel)

			err := epo.NotAllowedParallelRun(fRR.ID)
			So(err, ShouldBeNil)

			fRR, _ = epo.GetByID(flowRR.ID)
			So(fRR.Status, ShouldEqual, value_object.NotAllowedParallelCancel)
		})

		Convey("Fail", func() {
			fRR, _ := epo.GetByID(flowRR.ID)
			So(fRR.Status, ShouldNotEqual, value_object.Fail)

			err := epo.Fail(fRR.ID, "xx")
			So(err, ShouldBeNil)

			fRR, _ = epo.GetByID(flowRR.ID)
			So(fRR.Status, ShouldEqual, value_object.Fail)
			So(fRR.ErrorMsg, ShouldEqual, "xx")
		})

		Convey("Intercepted", func() {
			fRR, _ := epo.GetByID(flowRR.ID)
			So(fRR.Status, ShouldNotEqual, value_object.InterceptedCancel)

			err := epo.Intercepted(fRR.ID, "xx")
			So(err, ShouldBeNil)

			fRR, _ = epo.GetByID(flowRR.ID)
			So(fRR.Status, ShouldEqual, value_object.InterceptedCancel)
			So(fRR.InterceptMsg, ShouldEqual, "xx")
		})

		Convey("UserCancel", func() {
			cancelUserID := value_object.NewUUID()
			err := epo.UserCancel(flowRR.ID, cancelUserID)
			So(err, ShouldBeNil)

			fRR, _ := epo.GetByID(flowRR.ID)
			So(fRR.Canceled, ShouldBeTrue)
			So(fRR.CancelUserID, ShouldEqual, cancelUserID)
		})

		Convey("TimeoutCancel", func() {
			err := epo.TimeoutCancel(flowRR.ID)
			So(err, ShouldBeNil)

			fRR, _ := epo.GetByID(flowRR.ID)
			So(fRR.Canceled, ShouldBeTrue)

			So(epo.ReGetToCheckIsCanceled(flowRR.ID), ShouldBeTrue)
		})
	})

	Convey("CrontabFindOrCreate", t, func() {
		crontabTriggerTime := time.Now()
		created, err := epo.CrontabFindOrCreate(flowRR, crontabTriggerTime)
		So(err, ShouldBeNil)
		So(created, ShouldBeTrue)

		Convey("CrontabFindOrCreate cannot be repub", func() {
			created, err := epo.CrontabFindOrCreate(flowRR, crontabTriggerTime)
			So(err, ShouldBeNil)
			So(created, ShouldBeFalse)
		})
	})
}

func TestMain(m *testing.M) {
	pool, err := dockertest.NewPool("")
	if err != nil {
		log.Fatalf("Could not connect to docker: %s", err)
	}

	// pull mongodb docker image for version 5.0
	resource, err := pool.RunWithOptions(&dockertest.RunOptions{
		Repository: "mongo",
		Tag:        "5.0.5",
		Env: []string{
			// username and password for mongodb superuser
			"MONGO_INITDB_ROOT_USERNAME=" + conf.User,
			"MONGO_INITDB_ROOT_PASSWORD=" + conf.Password,
		},
	}, func(config *docker.HostConfig) {
		// set AutoRemove to true so that stopped container goes away by itself
		config.AutoRemove = true
		config.RestartPolicy = docker.RestartPolicy{
			Name: "no",
		}
	})
	if err != nil {
		log.Fatalf("Could not start resource: %s", err)
	}

	conf.Addresses = []string{
		fmt.Sprintf("localhost:%s", resource.GetPort("27017/tcp"))}

	// exponential backoff-retry, because the application in the container might not be ready to accept connections yet
	err = pool.Retry(func() error {
		var err error
		epo, err = New(context.TODO(), conf, DefaultCollectionName)
		if err != nil {
			return err
		}
		return nil
	})

	if err != nil {
		log.Fatalf("Could not connect to docker: %s", err)
	}

	// run tests
	code := m.Run()

	// When you're done, kill and remove the container
	if err = pool.Purge(resource); err != nil {
		log.Fatalf("Could not purge resource: %s", err)
	}

	os.Exit(code)
}

func TestCreateIndexes(t *testing.T) {
	Convey("create index", t, func() {
		indexes := mongoDBIndexes()
		err := epo.mongoCollection.CreateIndex(indexes)
		So(err, ShouldBeNil)
	})
}
