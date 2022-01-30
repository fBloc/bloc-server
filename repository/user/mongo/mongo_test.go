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
	fakeUserName   = gofakeit.Name()
	fakeUserPasswd = gofakeit.Password(false, false, false, false, false, 16)
)

func TestNewFromUser(t *testing.T) {
	Convey("build from aggregate user", t, func() {
		aggregateUser := aggregate.NewUser(
			fakeUserName, fakeUserPasswd, false)
		mUser := NewFromUser(aggregateUser)
		So(mUser, ShouldNotBeNil)
		So(mUser, ShouldHaveSameTypeAs, &mongoUser{})
		So(mUser.Name, ShouldEqual, aggregateUser.Name)
	})
}

func TestToAggregate(t *testing.T) {
	Convey("build from aggregate user", t, func() {
		aggregateUser := aggregate.NewUser(
			fakeUserName, fakeUserPasswd, false)
		mUser := NewFromUser(aggregateUser)
		mToAggUser := mUser.ToAggregate()
		So(mToAggUser, ShouldNotBeNil)
		So(mToAggUser, ShouldHaveSameTypeAs, &aggregate.User{})
		So(mToAggUser.Name, ShouldEqual, fakeUserName)
		So(mToAggUser.RawPassword, ShouldEqual, "") // raw password should lost
		So(mToAggUser.Password, ShouldNotEqual, fakeUserPasswd)
	})
}

func TestCreate(t *testing.T) {
	Convey("create", t, func() {
		aggregateUser := aggregate.NewUser(
			fakeUserName, fakeUserPasswd, false)
		err := epo.Create(aggregateUser)
		So(err, ShouldBeNil)

		Reset(func() {
			epo.DeleteByID(aggregateUser.ID)
		})
	})
}

func TestDeleteByID(t *testing.T) {
	Convey("DeleteByID", t, func() {
		aggregateUser := aggregate.NewUser(
			fakeUserName, fakeUserPasswd, false)
		err := epo.Create(aggregateUser)
		So(err, ShouldBeNil)

		Reset(func() {
			amount, err := epo.DeleteByID(aggregateUser.ID)
			So(err, ShouldBeNil)
			So(amount, ShouldEqual, 1)
		})
	})
}

func TestQuery(t *testing.T) {
	aggregateUser := aggregate.NewUser(
		fakeUserName, fakeUserPasswd, false)
	epo.Create(aggregateUser)

	Convey("GetByID miss", t, func() {
		aggUser, err := epo.GetByID(value_object.NewUUID())
		So(err, ShouldBeNil)
		So(aggUser, ShouldBeNil)
	})

	Convey("GetByID hit", t, func() {
		aggUser, err := epo.GetByID(aggregateUser.ID)
		So(err, ShouldBeNil)
		So(aggUser, ShouldNotBeNil)
		So(aggUser.Name, ShouldEqual, fakeUserName)
	})

	Convey("GetByName miss", t, func() {
		aggUser, err := epo.GetByName(fakeUserName + "miss")
		So(err, ShouldBeNil)
		So(aggUser, ShouldBeNil)
	})

	Convey("GetByName hit", t, func() {
		aggUser, err := epo.GetByName(fakeUserName)
		So(err, ShouldBeNil)
		So(aggUser, ShouldNotBeNil)
		So(aggUser, ShouldHaveSameTypeAs, &aggregate.User{})
		So(aggUser.Name, ShouldEqual, fakeUserName)
	})

	Convey("FilterByNameContains miss", t, func() {
		nameContains := fakeUserName[1:len(fakeUserName)-1] + "miss"
		aggUsers, err := epo.FilterByNameContains(nameContains)
		So(err, ShouldBeNil)
		So(len(aggUsers), ShouldEqual, 0)
	})

	Convey("FilterByNameContains hit", t, func() {
		nameContains := fakeUserName[1 : len(fakeUserName)-1]
		aggUsers, err := epo.FilterByNameContains(nameContains)
		So(err, ShouldBeNil)
		So(len(aggUsers), ShouldEqual, 1)
		So(aggUsers[0].Name, ShouldEqual, fakeUserName)
	})

	Convey("All", t, func() {
		aggUsers, err := epo.All()
		So(err, ShouldBeNil)
		So(len(aggUsers), ShouldEqual, 1)
		So(aggUsers[0].Name, ShouldEqual, fakeUserName)
	})

	Convey("PatchName", t, func() {
		newName := fakeUserName + "new"
		err := epo.PatchName(aggregateUser.ID, newName)
		So(err, ShouldBeNil)

		aggUser, _ := epo.GetByID(aggregateUser.ID)
		So(aggUser.Name, ShouldEqual, newName)
	})

	epo.DeleteByID(aggregateUser.ID)
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
