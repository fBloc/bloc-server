package mongo

import (
	"context"
	"fmt"
	"log"
	"os"
	"testing"

	"github.com/brianvoe/gofakeit/v6"
	"github.com/fBloc/bloc-server/aggregate"
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
	funcID            = value_object.NewUUID()
	ReadeUser         = aggregate.User{ID: value_object.NewUUID()}
	ExecuteUser       = aggregate.User{ID: value_object.NewUUID()}
	AllPermissionUser = aggregate.User{ID: value_object.NewUUID()}
	SuperUser         = aggregate.User{ID: value_object.NewUUID(), IsSuper: true}
	NewPermissionUser = aggregate.User{ID: value_object.NewUUID()}
	fakeFunction      = aggregate.Function{
		ID:           funcID,
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
		},
		ProcessStages:           []string{"parsing ipt", "finished parse ipt & start do the math", "finished"},
		ReadUserIDs:             []value_object.UUID{ReadeUser.ID, AllPermissionUser.ID},
		ExecuteUserIDs:          []value_object.UUID{ExecuteUser.ID, AllPermissionUser.ID},
		AssignPermissionUserIDs: []value_object.UUID{AllPermissionUser.ID},
	}
)

func TestCreate(t *testing.T) {
	Convey("create function", t, func() {
		err := epo.Create(&fakeFunction)
		So(err, ShouldBeNil)
	})
}

func TestQuery(t *testing.T) {
	Convey("All", t, func() {
		funcs, err := epo.All()
		So(err, ShouldBeNil)
		So(len(funcs), ShouldEqual, 1)
		So(funcs[0].Name, ShouldEqual, fakeFunction.Name)
	})

	Convey("GetByID miss", t, func() {
		aggFunc, err := epo.GetByID(value_object.NewUUID())
		So(err, ShouldBeNil)
		So(aggFunc.IsZero(), ShouldBeTrue)
	})

	Convey("GetByID hit", t, func() {
		aggFunc, err := epo.GetByID(fakeFunction.ID)
		So(err, ShouldBeNil)
		So(aggFunc.IsZero(), ShouldBeFalse)
		So(aggFunc.Name, ShouldEqual, fakeFunction.Name)
	})

	Convey("GetSameIptOptFunction hit", t, func() {
		aggFunc, err := epo.GetSameIptOptFunction(
			fakeFunction.IptDigest, fakeFunction.OptDigest)
		So(err, ShouldBeNil)
		So(aggFunc.IsZero(), ShouldBeFalse)
		So(aggFunc.Name, ShouldEqual, fakeFunction.Name)
	})

	Convey("GetSameIptOptFunction miss", t, func() {
		aggFunc, err := epo.GetSameIptOptFunction(
			fakeFunction.IptDigest, fakeFunction.OptDigest+"miss")
		So(err, ShouldBeNil)
		So(aggFunc.IsZero(), ShouldBeTrue)
	})

	Convey("IDMapFunctionAll", t, func() {
		idMapFunctionAll, err := epo.IDMapFunctionAll()
		So(err, ShouldBeNil)
		aggFunc, ok := idMapFunctionAll[fakeFunction.ID]
		So(ok, ShouldBeTrue)
		So(aggFunc.IsZero(), ShouldBeFalse)
		So(aggFunc.Name, ShouldEqual, fakeFunction.Name)
	})
}

func TestPatch(t *testing.T) {
	Convey("PatchName", t, func() {
		newName := gofakeit.Name()
		err := epo.PatchName(fakeFunction.ID, newName)
		So(err, ShouldBeNil)

		aggFunc, _ := epo.GetByID(fakeFunction.ID)
		So(aggFunc.Name, ShouldEqual, newName)
	})

	Convey("PatchDescription", t, func() {
		newDesc := gofakeit.Name()
		err := epo.PatchDescription(fakeFunction.ID, newDesc)
		So(err, ShouldBeNil)

		aggFunc, _ := epo.GetByID(fakeFunction.ID)
		So(aggFunc.Description, ShouldEqual, newDesc)
	})

	Convey("PatchGroupName", t, func() {
		groupName := gofakeit.Name()
		err := epo.PatchGroupName(fakeFunction.ID, groupName)
		So(err, ShouldBeNil)

		aggFunc, _ := epo.GetByID(fakeFunction.ID)
		So(aggFunc.GroupName, ShouldEqual, groupName)
	})

	Convey("PatchProviderName", t, func() {
		providerName := gofakeit.Name()
		err := epo.PatchProviderName(fakeFunction.ID, providerName)
		So(err, ShouldBeNil)

		aggFunc, _ := epo.GetByID(fakeFunction.ID)
		So(aggFunc.ProviderName, ShouldEqual, providerName)
	})
}

func TestPermission(t *testing.T) {
	Convey("read", t, func() {
		canRead := fakeFunction.UserCanRead(&ReadeUser)
		So(canRead, ShouldBeTrue)

		err := epo.RemoveReader(fakeFunction.ID, ReadeUser.ID)
		So(err, ShouldBeNil)

		aggFunc, _ := epo.GetByID(fakeFunction.ID)
		canRead = aggFunc.UserCanRead(&ReadeUser)
		So(canRead, ShouldBeFalse)

		err = epo.AddReader(fakeFunction.ID, ReadeUser.ID)
		So(err, ShouldBeNil)

		aggFunc, _ = epo.GetByID(fakeFunction.ID)
		canRead = aggFunc.UserCanRead(&ReadeUser)
		So(canRead, ShouldBeTrue)
	})

	Convey("execute", t, func() {
		canExecute := fakeFunction.UserCanExecute(&ExecuteUser)
		So(canExecute, ShouldBeTrue)

		err := epo.RemoveExecuter(fakeFunction.ID, ExecuteUser.ID)
		So(err, ShouldBeNil)

		aggFunc, _ := epo.GetByID(fakeFunction.ID)
		canExecute = aggFunc.UserCanExecute(&ExecuteUser)
		So(canExecute, ShouldBeFalse)

		err = epo.AddExecuter(fakeFunction.ID, ExecuteUser.ID)
		So(err, ShouldBeNil)

		aggFunc, _ = epo.GetByID(fakeFunction.ID)
		canExecute = aggFunc.UserCanExecute(&ExecuteUser)
		So(canExecute, ShouldBeTrue)
	})

	Convey("assign", t, func() {
		canAssign := fakeFunction.UserCanAssignPermission(&AllPermissionUser)
		So(canAssign, ShouldBeTrue)

		err := epo.RemoveAssigner(fakeFunction.ID, AllPermissionUser.ID)
		So(err, ShouldBeNil)

		aggFunc, _ := epo.GetByID(fakeFunction.ID)
		canAssign = aggFunc.UserCanAssignPermission(&AllPermissionUser)
		So(canAssign, ShouldBeFalse)

		err = epo.AddAssigner(fakeFunction.ID, AllPermissionUser.ID)
		So(err, ShouldBeNil)

		aggFunc, _ = epo.GetByID(fakeFunction.ID)
		canAssign = aggFunc.UserCanAssignPermission(&AllPermissionUser)
		So(canAssign, ShouldBeTrue)
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
