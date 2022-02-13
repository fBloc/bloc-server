package mongodb

import (
	"context"
	"fmt"
	"log"
	"os"
	"testing"

	"github.com/brianvoe/gofakeit/v6"
	"github.com/fBloc/bloc-server/value_object"
	"github.com/ory/dockertest/v3"
	"github.com/ory/dockertest/v3/docker"
	. "github.com/smartystreets/goconvey/convey"
	"go.mongodb.org/mongo-driver/mongo"
)

var dbClient *mongo.Client
var collec *Collection

var conf = &MongoConfig{
	Db:       "bloc-test-mongo",
	User:     "root",
	Password: "password",
}

type testData struct {
	ID   value_object.UUID `bson:"id"`
	Name string            `bson:"name"`
	Age  int               `bson:"age"`
}

var test = testData{ID: value_object.NewUUID(), Name: gofakeit.Name()}

func TestMongo(t *testing.T) {
	Convey("mongo", t, func() {
		Convey("test count. initial no record", func() {
			amount, err := collec.Count(NewFilter().AddEqual("id", test.ID))
			So(err, ShouldBeNil)
			So(amount, ShouldEqual, 0)
		})

		Convey("insert one doc", func() {
			_, err := collec.InsertOne(test)
			So(err, ShouldBeNil)
		})

		Convey("insert one doc and count", func() {
			amount, err := collec.Count(NewFilter())
			So(err, ShouldBeNil)
			So(amount, ShouldEqual, 0)

			collec.InsertOne(test)

			amount, err = collec.Count(NewFilter())
			So(err, ShouldBeNil)
			So(amount, ShouldEqual, 1)
		})

		Convey("insert one doc and clear collection", func() {
			collec.InsertOne(test)

			err := collec.ClearCollection()
			So(err, ShouldBeNil)

			amount, err := collec.Count(NewFilter())
			So(err, ShouldBeNil)
			So(amount, ShouldEqual, 0)
		})

		Convey("insert one doc and delete by id", func() {
			collec.InsertOne(test)

			deleteAmount, err := collec.DeleteByID(test.ID)
			So(err, ShouldBeNil)
			So(deleteAmount, ShouldEqual, 1)

			amount, _ := collec.Count(NewFilter().AddEqual("id", test.ID))
			So(err, ShouldBeNil)
			So(amount, ShouldEqual, 0)
		})

		Convey("insert one doc and delete by filter", func() {
			collec.InsertOne(test)

			deleteAmount, err := collec.Delete(NewFilter().AddEqual("name", test.Name))
			So(err, ShouldBeNil)
			So(deleteAmount, ShouldEqual, 1)

			amount, _ := collec.Count(NewFilter().AddEqual("id", test.ID))
			So(amount, ShouldEqual, 0)
		})

		Convey("insert one doc and Get by id", func() {
			collec.InsertOne(test)

			var resp testData
			err := collec.GetByID(test.ID, &resp)
			So(err, ShouldBeNil)
			So(resp, ShouldNotBeNil)
			So(resp.Name, ShouldEqual, test.Name)
		})

		Convey("insert one doc and Get by filter", func() {
			collec.InsertOne(test)

			var byNameGetResp testData
			err := collec.Get(
				NewFilter().AddEqual("name", test.Name),
				nil, &byNameGetResp)
			So(err, ShouldBeNil)
			So(byNameGetResp, ShouldNotBeNil)
			So(byNameGetResp.Name, ShouldEqual, test.Name)
		})

		Convey("insert one doc and filter", func() {
			collec.InsertOne(test)

			var byNameFilterResp []testData
			err := collec.Filter(
				NewFilter().AddEqual("name", test.Name),
				nil, &byNameFilterResp)
			So(err, ShouldBeNil)
			So(len(byNameFilterResp), ShouldEqual, 1)
			var existTheNameRecord bool
			for _, i := range byNameFilterResp {
				if i.Name == test.Name {
					existTheNameRecord = true
				}
			}
			So(existTheNameRecord, ShouldBeTrue)
		})

		Convey("insert one and patch field", func() {
			collec.InsertOne(test)

			newName := gofakeit.Name()
			for newName == test.Name {
				newName = gofakeit.Name()
			}

			err := collec.PatchByID(test.ID, NewUpdater().AddSet("name", newName))
			So(err, ShouldBeNil)

			var resp testData
			collec.GetByID(test.ID, &resp)
			So(resp.Name, ShouldNotEqual, test.Name)
			So(resp.Name, ShouldEqual, newName)
		})

		Convey("insert one doc and replace", func() {
			var resp testData

			_, err := collec.InsertOne(test)
			So(err, ShouldBeNil)

			collec.GetByID(test.ID, &resp)
			So(resp.Name, ShouldEqual, test.Name)
			So(resp.Age, ShouldEqual, 0)

			// do replace
			newTest := testData{
				ID:   test.ID,
				Name: gofakeit.Name(),
				Age:  24,
			}
			mID, err := collec.GetMongoID(newTest.ID)
			So(err, ShouldBeNil)

			collec.ReplaceByID(mID, newTest)

			// donnot create new doc
			amount, _ := collec.Count(NewFilter())
			So(amount, ShouldEqual, 1)

			// age do changed
			collec.GetByID(test.ID, &resp)
			So(resp.Name, ShouldEqual, newTest.Name)
			So(resp.Age, ShouldEqual, newTest.Age)
		})

		Convey("FindOneOrInsert", func() {
			var resp testData
			exist, err := collec.FindOneOrInsert(
				NewFilter().AddEqual("name", test.Name),
				test, &resp)
			So(err, ShouldBeNil)
			So(exist, ShouldBeFalse)
			if !resp.ID.IsNil() {
				t.Fatal("should not resp data!")
			}
			insertedAmount, _ := collec.Count(NewFilter())
			So(insertedAmount, ShouldEqual, 1)

			exist, err = collec.FindOneOrInsert(
				NewFilter().AddEqual("name", test.Name),
				test, &resp)
			So(err, ShouldBeNil)
			So(exist, ShouldBeTrue)
			if resp.ID.IsNil() {
				t.Fatal("should resp data!")
			}

			insertedAmount, _ = collec.Count(NewFilter())
			So(insertedAmount, ShouldEqual, 1)
		})

		Convey("insert multi same name docs and test filter & count", func() {
			theName := gofakeit.Name()
			insertedDocs := make([]testData, 3)
			for i := 0; i < len(insertedDocs); i++ {
				doc := testData{
					ID:   value_object.NewUUID(),
					Name: theName,
				}
				insertedDocs[i] = doc
				_, err := collec.InsertOne(doc)
				So(err, ShouldBeNil)
			}

			amount, err := collec.Count(NewFilter().AddEqual("name", theName))
			So(err, ShouldBeNil)
			So(amount, ShouldEqual, len(insertedDocs))

			var byNameFilterResp []testData
			collec.Filter(
				NewFilter().AddEqual("name", theName),
				nil, &byNameFilterResp)
			So(len(byNameFilterResp), ShouldEqual, len(insertedDocs))
			everyNameTheSame := true
			for _, i := range byNameFilterResp {
				if i.Name != theName {
					everyNameTheSame = false
				}
			}
			So(everyNameTheSame, ShouldBeTrue)

			collec.ClearCollection()

			amount, _ = collec.Count(NewFilter().AddEqual("name", theName))
			So(amount, ShouldEqual, 0)
		})

		Reset(func() {
			collec.ClearCollection()
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

		dbClient, err = InitClient(conf)
		if err != nil {
			return err
		}
		return nil
	})

	if err != nil {
		log.Fatalf("Could not connect to docker: %s", err)
	}

	collec, err = NewCollection(conf, "test")
	if err != nil {
		log.Fatalf("get connection error: %s", err)
	}

	// run tests
	code := m.Run()

	// When you're done, kill and remove the container
	if err = pool.Purge(resource); err != nil {
		log.Fatalf("Could not purge resource: %s", err)
	}

	// disconnect mongodb client
	if err = dbClient.Disconnect(context.TODO()); err != nil {
		panic(err)
	}

	os.Exit(code)
}
