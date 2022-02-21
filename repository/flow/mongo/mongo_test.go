package mongo

import (
	"context"
	"fmt"
	"log"
	"os"
	"testing"

	"github.com/brianvoe/gofakeit/v6"
	"github.com/fBloc/bloc-server/aggregate"
	"github.com/fBloc/bloc-server/config"
	"github.com/fBloc/bloc-server/internal/conns/mongodb"
	"github.com/fBloc/bloc-server/internal/crontab"
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
	executeUser       = aggregate.User{ID: value_object.NewUUID()}
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
	}
)

func TestNewFromFlow(t *testing.T) {
	Convey("NewFromFlow", t, func() {
		mFlow := NewFromFlow(&fakeAggregateFlow)
		So(mFlow.ID, ShouldNotEqual, value_object.NillUUID)
		So(mFlow.ID, ShouldEqual, fakeAggregateFlow.ID)
	})
}

func TestToAggregate(t *testing.T) {
	Convey("NewFromFlow", t, func() {
		mFlow := NewFromFlow(&fakeAggregateFlow)
		aggFlow := mFlow.ToAggregate()
		So(aggFlow.IsZero(), ShouldBeFalse)
	})
}

func TestDraft(t *testing.T) {
	fakeName := gofakeit.Name()
	Convey("CreateDraftFromScratch", t, func() {
		draftFlow, err := epo.CreateDraftFromScratch(
			fakeName, value_object.NewUUID(),
			nil, validFlowFunctionIDMapFlowFunction)
		So(err, ShouldBeNil)
		So(draftFlow.IsZero(), ShouldBeFalse)
		So(draftFlow.IsDraft, ShouldBeTrue)
		So(draftFlow.ID, ShouldNotEqual, value_object.NillUUID)
		So(draftFlow.OriginID, ShouldNotEqual, value_object.NillUUID)
		epo.DeleteDraftByOriginID(draftFlow.OriginID)
	})

	Convey("CreateDraftFromExistFlow", t, func() {
		draftFlow, err := epo.CreateDraftFromExistFlow(
			fakeName, value_object.NewUUID(),
			value_object.NewUUID(), nil,
			validFlowFunctionIDMapFlowFunction,
		)
		So(err, ShouldNotBeNil)
		So(draftFlow.IsZero(), ShouldBeTrue)
	})

	Convey("Query", t, func() {
		draftFlow, _ := epo.CreateDraftFromScratch(
			fakeName, readeUser.ID,
			nil, validFlowFunctionIDMapFlowFunction)

		Convey("GetByIDStr", func() {
			f, err := epo.GetByIDStr(draftFlow.ID.String())
			So(err, ShouldBeNil)
			So(f.IsZero(), ShouldBeFalse)
		})

		Convey("GetDraftByOriginID", func() {
			f, err := epo.GetDraftByOriginID(draftFlow.OriginID)
			So(err, ShouldBeNil)
			So(f.IsZero(), ShouldBeFalse)
		})

		Convey("FilterDraft", func() {
			flows, err := epo.FilterDraft(readeUser.ID, fakeName)
			So(err, ShouldBeNil)
			So(len(flows), ShouldEqual, 1)
			So(flows[0].Name, ShouldEqual, fakeName)
		})

		Reset(func() {
			epo.DeleteByID(draftFlow.ID)
		})
	})

	Convey("delete", t, func() {
		draftFlow, _ := epo.CreateDraftFromScratch(
			fakeName, value_object.NewUUID(),
			nil, validFlowFunctionIDMapFlowFunction)

		Convey("DeleteByID", func() {
			deleteAmount, err := epo.DeleteByID(draftFlow.ID)
			So(err, ShouldBeNil)
			So(deleteAmount, ShouldEqual, 1)
		})

		Convey("DeleteByOriginID", func() {
			deleteAmount, err := epo.DeleteByOriginID(draftFlow.OriginID)
			So(err, ShouldBeNil)
			So(deleteAmount, ShouldEqual, 1)
		})

		Convey("DeleteDraftByOriginID", func() {
			deleteAmount, err := epo.DeleteDraftByOriginID(draftFlow.OriginID)
			So(err, ShouldBeNil)
			So(deleteAmount, ShouldEqual, 1)
		})

		Reset(func() {
			epo.DeleteByID(draftFlow.ID)
		})
	})
}

func TestOnline(t *testing.T) {
	fakeName := gofakeit.Name()
	Convey("CreateOnlineFromDraft", t, func() {
		draftFlow, _ := epo.CreateDraftFromScratch(
			fakeName, value_object.NewUUID(),
			nil, validFlowFunctionIDMapFlowFunction)

		onlineFlow, err := epo.CreateOnlineFromDraft(draftFlow)
		So(err, ShouldBeNil)
		So(onlineFlow.IsZero(), ShouldBeFalse)
		So(onlineFlow.IsDraft, ShouldBeFalse)
		So(onlineFlow.Newest, ShouldBeTrue)
		epo.DeleteDraftByOriginID(draftFlow.OriginID)
		epo.DeleteByOriginID(onlineFlow.OriginID)
	})

	Convey("Query", t, func() {
		draftFlow, _ := epo.CreateDraftFromScratch(
			fakeName, readeUser.ID,
			nil, validFlowFunctionIDMapFlowFunction)
		onlineFlow, _ := epo.CreateOnlineFromDraft(draftFlow)
		epo.DeleteDraftByOriginID(draftFlow.OriginID)

		Convey("FilterOnline", func() {
			flows, err := epo.FilterOnline(&readeUser, fakeName)
			So(err, ShouldBeNil)
			So(len(flows), ShouldEqual, 1)
			So(flows[0].Name, ShouldEqual, fakeName)
		})

		Convey("GetOnlineByOriginID", func() {
			flow, err := epo.GetOnlineByOriginID(onlineFlow.OriginID)
			So(err, ShouldBeNil)
			So(flow.IsZero(), ShouldBeFalse)
			So(onlineFlow.Name, ShouldEqual, fakeName)
		})

		Convey("GetOnlineByOriginIDStr", func() {
			flow, err := epo.GetOnlineByOriginIDStr(onlineFlow.OriginID.String())
			So(err, ShouldBeNil)
			So(flow.IsZero(), ShouldBeFalse)
			So(onlineFlow.Name, ShouldEqual, fakeName)
		})

		Convey("GetLatestByOriginID", func() {
			flow, err := epo.GetLatestByOriginID(onlineFlow.OriginID)
			So(err, ShouldBeNil)
			So(flow.IsZero(), ShouldBeFalse)
			So(flow.ID, ShouldEqual, onlineFlow.ID)
		})

		Reset(func() {
			epo.DeleteByID(onlineFlow.ID)
		})
	})

	Convey("Patch", t, func() {
		draftFlow, _ := epo.CreateDraftFromScratch(
			fakeName, readeUser.ID,
			nil, validFlowFunctionIDMapFlowFunction)
		onlineFlow, _ := epo.CreateOnlineFromDraft(draftFlow)
		epo.DeleteDraftByOriginID(draftFlow.OriginID)

		Convey("PatchRetryStrategy", func() {
			So(onlineFlow.HaveRetryStrategy(), ShouldBeFalse)

			err := epo.PatchRetryStrategy(onlineFlow.ID, 3, 60)
			So(err, ShouldBeNil)

			f, err := epo.GetByID(onlineFlow.ID)
			So(err, ShouldBeNil)
			So(f.IsZero(), ShouldBeFalse)
			So(f.HaveRetryStrategy(), ShouldBeTrue)
		})

		Convey("PatchCrontab", func() {
			flows, err := epo.FilterCrontabFlows()
			So(err, ShouldBeNil)
			So(len(flows), ShouldEqual, 0)

			err = epo.PatchCrontab(onlineFlow.ID, *crontab.BuildCrontab("* * * * *"))
			So(err, ShouldBeNil)

			flows, err = epo.FilterCrontabFlows()
			So(err, ShouldBeNil)
			So(len(flows), ShouldEqual, 1)
			So(flows[0].Crontab.IsValid(), ShouldBeTrue)
		})

		Convey("PatchAllowParallelRun", func() {
			f, _ := epo.GetByID(onlineFlow.ID)
			So(f.AllowParallelRun, ShouldBeFalse)

			err := epo.PatchAllowParallelRun(onlineFlow.ID, true)
			So(err, ShouldBeNil)

			f, _ = epo.GetByID(onlineFlow.ID)
			So(f.AllowParallelRun, ShouldBeTrue)
		})

		Convey("PatchTriggerKey", func() {
			f, _ := epo.GetByID(onlineFlow.ID)
			So(f.TriggerKey, ShouldEqual, "")

			triggerKey := "xxxx"
			err := epo.PatchTriggerKey(onlineFlow.ID, triggerKey)
			So(err, ShouldBeNil)

			f, _ = epo.GetByID(onlineFlow.ID)
			So(f.TriggerKey, ShouldEqual, triggerKey)
		})

		Convey("PatchTimeout", func() {
			f, _ := epo.GetByID(onlineFlow.ID)
			So(f.TimeoutInSeconds, ShouldEqual, 0)

			err := epo.PatchTimeout(onlineFlow.ID, 1000)
			So(err, ShouldBeNil)

			f, _ = epo.GetByID(onlineFlow.ID)
			So(f.TimeoutInSeconds, ShouldEqual, 1000)
		})

		Convey("PatchName", func() {
			f, _ := epo.GetByID(onlineFlow.ID)
			So(f.Name, ShouldEqual, fakeName)

			newName := gofakeit.Name()
			err := epo.PatchName(onlineFlow.ID, newName)
			So(err, ShouldBeNil)

			f, _ = epo.GetByID(onlineFlow.ID)
			So(f.Name, ShouldEqual, newName)
		})

		Convey("OfflineByID", func() {
			f, _ := epo.GetByID(onlineFlow.ID)
			So(f.IsDraft, ShouldBeFalse)

			err := epo.OfflineByID(onlineFlow.ID)
			So(err, ShouldBeNil)

			f, _ = epo.GetByID(onlineFlow.ID)
			So(f.IsDraft, ShouldBeTrue)
		})

		Convey("CreateDraftFromExistFlow", func() {
			draftFlow, err := epo.CreateDraftFromExistFlow(
				fakeName, value_object.NewUUID(),
				onlineFlow.OriginID, nil,
				validFlowFunctionIDMapFlowFunction,
			)
			So(err, ShouldBeNil)
			So(draftFlow.IsZero(), ShouldBeFalse)
		})

		Reset(func() {
			epo.DeleteByOriginID(onlineFlow.OriginID)
		})
	})

	Convey("Delete", t, func() {
		draftFlow, _ := epo.CreateDraftFromScratch(
			fakeName, value_object.NewUUID(),
			nil, validFlowFunctionIDMapFlowFunction)
		onlineFlow, _ := epo.CreateOnlineFromDraft(draftFlow)
		epo.DeleteDraftByOriginID(draftFlow.OriginID)

		Convey("DeleteDraftByOriginID should delete nothing", func() {
			deleteAmount, err := epo.DeleteDraftByOriginID(onlineFlow.ID)
			So(err, ShouldBeNil)
			So(deleteAmount, ShouldEqual, 0)
		})

		Convey("DeleteByOriginID", func() {
			deleteAmount, err := epo.DeleteByOriginID(onlineFlow.OriginID)
			So(err, ShouldBeNil)
			So(deleteAmount, ShouldEqual, 1)
		})

		Convey("DeleteByID", func() {
			deleteAmount, err := epo.DeleteByID(onlineFlow.ID)
			So(err, ShouldBeNil)
			So(deleteAmount, ShouldEqual, 1)

			// after delete. there should no such flow
			f, err := epo.GetByID(onlineFlow.ID)
			So(err, ShouldBeNil)
			So(f.IsZero(), ShouldBeTrue)
		})

		Reset(func() {
			epo.DeleteByID(draftFlow.ID)
		})
	})

	Convey("user permission", t, func() {
		draftFlow, _ := epo.CreateDraftFromScratch(
			fakeName, value_object.NewUUID(),
			nil, validFlowFunctionIDMapFlowFunction)
		onlineFlow, _ := epo.CreateOnlineFromDraft(draftFlow)
		epo.DeleteDraftByOriginID(draftFlow.OriginID)

		Convey("read", func() {
			newUser := aggregate.User{ID: value_object.NewUUID()}
			So(onlineFlow.UserCanRead(&newUser), ShouldBeFalse)

			err := epo.AddReader(onlineFlow.ID, newUser.ID)
			So(err, ShouldBeNil)

			f, _ := epo.GetByID(onlineFlow.ID)
			So(f.UserCanRead(&newUser), ShouldBeTrue)

			err = epo.RemoveReader(onlineFlow.ID, newUser.ID)
			So(err, ShouldBeNil)

			f, _ = epo.GetByID(onlineFlow.ID)
			So(f.UserCanRead(&newUser), ShouldBeFalse)
		})

		Convey("execute", func() {
			newUser := aggregate.User{ID: value_object.NewUUID()}
			So(onlineFlow.UserCanExecute(&newUser), ShouldBeFalse)

			err := epo.AddExecuter(onlineFlow.ID, newUser.ID)
			So(err, ShouldBeNil)

			f, _ := epo.GetByID(onlineFlow.ID)
			So(f.UserCanExecute(&newUser), ShouldBeTrue)

			err = epo.RemoveExecuter(onlineFlow.ID, newUser.ID)
			So(err, ShouldBeNil)

			f, _ = epo.GetByID(onlineFlow.ID)
			So(f.UserCanExecute(&newUser), ShouldBeFalse)
		})

		Convey("writer", func() {
			newUser := aggregate.User{ID: value_object.NewUUID()}
			So(onlineFlow.UserCanWrite(&newUser), ShouldBeFalse)

			err := epo.AddWriter(onlineFlow.ID, newUser.ID)
			So(err, ShouldBeNil)

			f, _ := epo.GetByID(onlineFlow.ID)
			So(f.UserCanWrite(&newUser), ShouldBeTrue)

			err = epo.RemoveWriter(onlineFlow.ID, newUser.ID)
			So(err, ShouldBeNil)

			f, _ = epo.GetByID(onlineFlow.ID)
			So(f.UserCanWrite(&newUser), ShouldBeFalse)
		})

		Convey("deleter", func() {
			newUser := aggregate.User{ID: value_object.NewUUID()}
			So(onlineFlow.UserCanDelete(&newUser), ShouldBeFalse)

			err := epo.AddDeleter(onlineFlow.ID, newUser.ID)
			So(err, ShouldBeNil)

			f, _ := epo.GetByID(onlineFlow.ID)
			So(f.UserCanDelete(&newUser), ShouldBeTrue)

			err = epo.RemoveDeleter(onlineFlow.ID, newUser.ID)
			So(err, ShouldBeNil)

			f, _ = epo.GetByID(onlineFlow.ID)
			So(f.UserCanDelete(&newUser), ShouldBeFalse)
		})

		Convey("assigner", func() {
			newUser := aggregate.User{ID: value_object.NewUUID()}
			So(onlineFlow.UserCanAssignPermission(&newUser), ShouldBeFalse)

			err := epo.AddAssigner(onlineFlow.ID, newUser.ID)
			So(err, ShouldBeNil)

			f, _ := epo.GetByID(onlineFlow.ID)
			So(f.UserCanAssignPermission(&newUser), ShouldBeTrue)

			err = epo.RemoveAssigner(onlineFlow.ID, newUser.ID)
			So(err, ShouldBeNil)

			f, _ = epo.GetByID(onlineFlow.ID)
			So(f.UserCanAssignPermission(&newUser), ShouldBeFalse)
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
