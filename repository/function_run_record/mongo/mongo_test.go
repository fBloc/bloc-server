package mongo

import (
	"context"
	"fmt"
	"log"
	"os"
	"sync"
	"testing"
	"time"

	"github.com/brianvoe/gofakeit/v6"
	"github.com/fBloc/bloc-server/aggregate"
	"github.com/fBloc/bloc-server/config"
	"github.com/fBloc/bloc-server/infrastructure/object_storage"
	"github.com/fBloc/bloc-server/internal/conns/mongodb"
	"github.com/fBloc/bloc-server/pkg/ipt"
	"github.com/fBloc/bloc-server/pkg/opt"
	"github.com/fBloc/bloc-server/pkg/value_type"
	"github.com/fBloc/bloc-server/value_object"
	"github.com/ory/dockertest/v3"
	"github.com/ory/dockertest/v3/docker"
	. "github.com/smartystreets/goconvey/convey"
)

type mockObjectStorageImplement struct {
	keyMapData map[string][]byte
	sync.Mutex
}

func (mOSI *mockObjectStorageImplement) Set(key string, data []byte) error {
	mOSI.Lock()
	defer mOSI.Unlock()
	mOSI.keyMapData[key] = data
	return nil
}

func (mOSI *mockObjectStorageImplement) Get(key string) (bool, []byte, error) {
	data, ok := mOSI.keyMapData[key]
	return ok, data, nil
}

var _ object_storage.ObjectStorage = &mockObjectStorageImplement{}

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

func TestFunctionRunRecord(t *testing.T) {
	aggFlowRunRecord := aggregate.NewCrontabTriggeredRunRecord(context.TODO(), &fakeAggregateFlow)
	aggFunctionRunRecord := aggregate.NewFunctionRunRecordFromFlowDriven(
		context.TODO(), functionAdd, *aggFlowRunRecord, secondFlowFunctionID)
	err := epo.Create(aggFunctionRunRecord)
	if err != nil {
		log.Fatal(err)
	}

	Convey("query", t, func() {
		Convey("GetByID", func() {
			fRR, err := epo.GetByID(aggFunctionRunRecord.ID)
			So(err, ShouldBeNil)
			So(fRR.IsZero(), ShouldBeFalse)
		})

		Convey("Filter", func() {
			fRR, err := epo.Filter(
				*value_object.NewRepositoryFilter().AddEqual("id", aggFunctionRunRecord.ID),
				value_object.RepositoryFilterOption{})
			So(err, ShouldBeNil)
			So(len(fRR), ShouldEqual, 1)
		})

		Convey("FilterByFlowRunRecordID", func() {
			fRR, err := epo.FilterByFlowRunRecordID(aggFlowRunRecord.ID)
			So(err, ShouldBeNil)
			So(len(fRR), ShouldEqual, 1)
		})
	})

	Convey("Count", t, func() {
		amount, err := epo.Count(
			*value_object.NewRepositoryFilter().AddEqual("id", aggFunctionRunRecord.ID),
		)
		So(err, ShouldBeNil)
		So(amount, ShouldEqual, 1)
	})

	Convey("update", t, func() {
		Convey("PatchProgress", func() {
			fRR, err := epo.GetByID(aggFunctionRunRecord.ID)
			So(err, ShouldBeNil)
			So(fRR.Progress, ShouldEqual, 0)

			progress := 60
			err = epo.PatchProgress(aggFunctionRunRecord.ID, float32(progress))
			So(err, ShouldBeNil)

			fRR, _ = epo.GetByID(aggFunctionRunRecord.ID)
			So(fRR.Progress, ShouldEqual, progress)

			Convey("clearProgress", func() {
				err := epo.ClearProgress(aggFunctionRunRecord.ID)
				So(err, ShouldBeNil)

				fRR, _ = epo.GetByID(aggFunctionRunRecord.ID)
				So(fRR.Progress, ShouldEqual, 0)
			})
		})

		Convey("PatchProgressMsg", func() {
			fRR, _ := epo.GetByID(aggFunctionRunRecord.ID)
			So(len(fRR.ProgressMsg), ShouldEqual, 0)

			progressMsg := "wuhu"
			err = epo.PatchProgressMsg(aggFunctionRunRecord.ID, progressMsg)
			So(err, ShouldBeNil)

			fRR, _ = epo.GetByID(aggFunctionRunRecord.ID)
			So(fRR.ProgressMsg, ShouldResemble, []string{progressMsg})
		})

		Convey("PatchStageIndex", func() {
			fRR, _ := epo.GetByID(aggFunctionRunRecord.ID)
			So(fRR.ProcessStageIndex, ShouldEqual, 0)

			stageIndex := 1
			err = epo.PatchStageIndex(aggFunctionRunRecord.ID, stageIndex)
			So(err, ShouldBeNil)

			fRR, _ = epo.GetByID(aggFunctionRunRecord.ID)
			So(fRR.ProcessStageIndex, ShouldResemble, stageIndex)
		})

		Convey("SetTimeout", func() {
			fRR, _ := epo.GetByID(aggFunctionRunRecord.ID)
			So(fRR.ShouldBeCanceledAt.IsZero(), ShouldBeTrue)

			theTimeOutTime := time.Now().Add(10 * time.Second)
			err = epo.SetTimeout(aggFunctionRunRecord.ID, theTimeOutTime)
			So(err, ShouldBeNil)

			fRR, _ = epo.GetByID(aggFunctionRunRecord.ID)
			So(fRR.ShouldBeCanceledAt.Unix(), ShouldEqual, theTimeOutTime.Unix())
		})

		Convey("SaveIptBrief", func() {
			objectStorage := &mockObjectStorageImplement{
				keyMapData: make(map[string][]byte)}
			err := epo.SaveIptBrief(
				aggFunctionRunRecord.ID,
				nil, nil, objectStorage)
			So(err, ShouldBeNil)
		})

		Convey("SaveStart", func() {
			fRR, _ := epo.GetByID(aggFunctionRunRecord.ID)
			So(fRR.Start.IsZero(), ShouldBeTrue)

			err := epo.SaveStart(aggFunctionRunRecord.ID)
			So(err, ShouldBeNil)

			fRR, _ = epo.GetByID(aggFunctionRunRecord.ID)
			So(fRR.Start.IsZero(), ShouldBeFalse)
		})

		Convey("SaveSuc", func() {
			err := epo.SaveSuc(
				aggFunctionRunRecord.ID, "test",
				nil, nil, nil, nil, false)
			So(err, ShouldBeNil)

			fRR, _ := epo.GetByID(aggFunctionRunRecord.ID)
			So(fRR.Suc, ShouldBeTrue)
		})

		Convey("SaveFail", func() {
			failMsg := "xxx"
			err := epo.SaveFail(aggFunctionRunRecord.ID, failMsg)
			So(err, ShouldBeNil)

			fRR, _ := epo.GetByID(aggFunctionRunRecord.ID)
			So(fRR.ErrorMsg, ShouldEqual, failMsg)
		})

		Convey("SaveCancel", func() {
			err := epo.SaveCancel(aggFunctionRunRecord.ID)
			So(err, ShouldBeNil)

			fRR, _ := epo.GetByID(aggFunctionRunRecord.ID)
			So(fRR.Canceled, ShouldBeTrue)
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
