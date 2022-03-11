package mongo

import (
	"context"
	"fmt"
	"log"
	"os"
	"testing"
	"time"

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
)

func TestFunctionExecuteHeartBeat(t *testing.T) {
	Convey("function execute heartbeat", t, func() {
		funcRunRecordID := value_object.NewUUID()
		aggFunctionExecuteHeartBeat := aggregate.NewFunctionExecuteHeartBeat(funcRunRecordID)
		So(aggFunctionExecuteHeartBeat.IsZero(), ShouldBeFalse)
		So(aggFunctionExecuteHeartBeat.ID, ShouldNotEqual, value_object.NillUUID)

		err := epo.Create(aggFunctionExecuteHeartBeat)
		So(err, ShouldBeNil)

		Convey("GetByFunctionRunRecordID", func() {
			aggFEHB, err := epo.GetByFunctionRunRecordID(funcRunRecordID)
			So(err, ShouldBeNil)
			So(aggFEHB.IsZero(), ShouldBeFalse)
			So(aggFEHB.FunctionRunRecordID, ShouldEqual, funcRunRecordID)
		})

		Convey("GetByID", func() {
			aggFEHB, _ := epo.GetByFunctionRunRecordID(funcRunRecordID)
			aggFEHB, err := epo.GetByID(aggFEHB.ID)
			So(err, ShouldBeNil)
			So(aggFEHB.IsZero(), ShouldBeFalse)
			So(aggFEHB.FunctionRunRecordID, ShouldEqual, funcRunRecordID)
		})

		Convey("AliveReport", func() {
			aggFEHB, _ := epo.GetByFunctionRunRecordID(funcRunRecordID)

			beforeTime := time.Now()
			time.Sleep(1 * time.Second)
			err := epo.AliveReport(aggFEHB.ID)
			So(err, ShouldBeNil)

			aggFEHB, _ = epo.GetByID(aggFEHB.ID)
			So(aggFEHB.LatestHeartbeatTime, ShouldHappenAfter, beforeTime)
		})

		Convey("AllDeads filter", func() {
			var err error
			var deads []*aggregate.FunctionExecuteHeartBeat

			deads, err = epo.AllDeads(time.Hour)
			So(err, ShouldBeNil)
			So(len(deads), ShouldEqual, 0)

			time.Sleep(2 * time.Second)

			deads, err = epo.AllDeads(time.Second)
			So(err, ShouldBeNil)
			So(len(deads), ShouldEqual, 1)
		})

		Convey("Delete by id", func() {
			deleteAmount, err := epo.Delete(aggFunctionExecuteHeartBeat.ID)
			So(err, ShouldBeNil)
			So(deleteAmount, ShouldEqual, 1)
		})

		Convey("Delete by function_run_record_id", func() {
			deleteAmount, err := epo.DeleteByFunctionRunRecordID(funcRunRecordID)
			So(err, ShouldBeNil)
			So(deleteAmount, ShouldEqual, 1)
		})

		Reset(func() {
			epo.Delete(aggFunctionExecuteHeartBeat.ID)
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
