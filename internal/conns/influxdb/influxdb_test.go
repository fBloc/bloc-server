package influxdb

import (
	"log"
	"os"
	"testing"
	"time"

	"github.com/ory/dockertest/v3"
	. "github.com/smartystreets/goconvey/convey"
)

var (
	conf = &InfluxDBConfig{
		UserName:     "testUser",
		Password:     "password",
		Token:        "1326143eba0c8e1a408a014e9a63d767",
		Organization: "bloc-test",
	}
	conn *Connection
)

const (
	bucketName  = "test-log"
	measurement = "unit-test"
)

func TestBucketOperations(t *testing.T) {
	bucketClient, err := conn.NewBucketClient(
		bucketName, time.Hour)
	if err != nil {
		t.Fatalf("new bucket client error: %s", err)
	}

	Convey("influxdb set/query", t, func() {
		bucketClient.Write(
			measurement, map[string]string{},
			map[string]interface{}{"data": "valOne"},
			time.Now())

		bucketClient.Flush()
		time.Sleep(2 * time.Second)

		items, err := bucketClient.QueryAll(
			measurement,
			map[string]string{})
		So(err, ShouldBeNil)
		So(len(items), ShouldEqual, 1)
		item := items[0]
		value, ok := item["data"]
		So(ok, ShouldBeTrue)
		So(value, ShouldEqual, "valOne")

		items, err = bucketClient.Query(
			measurement,
			map[string]string{},
			time.Now().Add(-3*time.Minute),
			time.Now(),
		)
		So(err, ShouldBeNil)
		So(len(items), ShouldEqual, 1)
		item = items[0]
		value, ok = item["data"]
		So(ok, ShouldBeTrue)
		So(value, ShouldEqual, "valOne")

		items, err = bucketClient.Query(
			measurement,
			map[string]string{},
			time.Time{},
			time.Now().Add(-3*time.Minute),
		)
		So(err, ShouldBeNil)
		So(len(items), ShouldEqual, 0)
	})

	Convey("influxdb query", t, func() {
		bucketClient.Write(
			measurement,
			map[string]string{"tagK": "tagV"},
			map[string]interface{}{"data": "valTwo"},
			time.Now())

		bucketClient.Flush()
		time.Sleep(2 * time.Second)

		items, err := bucketClient.QueryAll(
			measurement,
			map[string]string{"tagK": "tagWrongV"})
		So(err, ShouldBeNil)
		So(len(items), ShouldEqual, 0)

		items, err = bucketClient.QueryAll(
			measurement,
			map[string]string{"tagK": "tagV"})
		So(err, ShouldBeNil)
		So(len(items), ShouldEqual, 1)
		item := items[0]
		value, ok := item["data"]
		So(ok, ShouldBeTrue)
		So(value, ShouldEqual, "valTwo")

		items, err = bucketClient.Query(
			measurement,
			map[string]string{"tagK": "tagWrongV"},
			time.Now().Add(-3*time.Minute),
			time.Now(),
		)
		So(err, ShouldBeNil)
		So(len(items), ShouldEqual, 0)

		items, err = bucketClient.Query(
			measurement,
			map[string]string{"tagK": "tagV"},
			time.Now().Add(-3*time.Minute),
			time.Now(),
		)
		So(err, ShouldBeNil)
		So(len(items), ShouldEqual, 1)
		item = items[0]
		value, ok = item["data"]
		So(ok, ShouldBeTrue)
		So(value, ShouldEqual, "valTwo")
	})
}

func TestMain(m *testing.M) {
	pool, err := dockertest.NewPool("")
	if err != nil {
		log.Fatalf("Could not connect to docker: %s", err)
	}

	options := &dockertest.RunOptions{
		Repository: "influxdb",
		Tag:        "2.1.1",
	}

	resource, err := pool.RunWithOptions(options)
	if err != nil {
		log.Fatalf("Could not start resource: %s", err)
	}

	conf.Address = "localhost:" + resource.GetPort("8086/tcp")

	// exponential backoff-retry
	// because the application in the container might not be ready to accept connections yet
	if err := pool.Retry(func() error {
		conn, err = Connect(conf)
		if err != nil {
			return err
		}
		return nil
	}); err != nil {
		log.Fatalf("Could not connect to docker: %s", err)
	}

	// run tests
	code := m.Run()

	// remove container
	if err = pool.Purge(resource); err != nil {
		log.Fatalf("Could not purge resource: %s", err)
	}

	os.Exit(code)
}
