package minio

import (
	"encoding/json"
	"fmt"
	"log"
	"os"
	"testing"

	"github.com/ory/dockertest/v3"
	. "github.com/smartystreets/goconvey/convey"
)

var conn *MinioCon
var err error

var conf = &MinioConfig{
	BucketName:     "test",
	AccessKey:      "blocMinio",
	AccessPassword: "blocMinioPasswd",
}

func TestMinioConf(t *testing.T) {
	Convey("conf nil test", t, func() {
		var conf *MinioConfig
		So(conf.IsNil(), ShouldBeTrue)

		conf = &MinioConfig{
			BucketName:     "test",
			AccessKey:      "blocMinio",
			AccessPassword: "blocMinioPasswd",
		}
		So(conf.IsNil(), ShouldBeTrue)

		Convey("nil cannot gen signature", func() {
			So(func() { conf.signature() }, ShouldPanic)
		})
	})
}

func TestMinio(t *testing.T) {
	Convey("minio set & get", t, func() {
		key := "key"
		value := "value"

		valByte, err := json.Marshal(value)
		So(err, ShouldBeNil)

		err = conn.Set(key, valByte)
		So(err, ShouldBeNil)

		keyExist, getedByte, err := conn.Get(key)
		So(err, ShouldBeNil)
		So(keyExist, ShouldBeTrue)
		var getedData string
		err = json.Unmarshal(getedByte, &getedData)
		So(err, ShouldBeNil)
		So(getedData, ShouldEqual, value)

		keyExist, _, err = conn.Get(key + "miss")
		So(err, ShouldBeNil)
		So(keyExist, ShouldBeFalse)
	})
}

func TestMain(m *testing.M) {
	pool, err := dockertest.NewPool("")
	if err != nil {
		log.Fatalf("Could not connect to docker: %s", err)
	}

	options := &dockertest.RunOptions{
		Repository: "minio/minio",
		Tag:        "RELEASE.2021-11-24T23-19-33Z",
		Cmd:        []string{"server", "/data"},
		Env: []string{
			"MINIO_ROOT_USER=" + conf.AccessKey,
			"MINIO_ROOT_PASSWORD=" + conf.AccessPassword,
		},
	}

	resource, err := pool.RunWithOptions(options)
	if err != nil {
		log.Fatalf("Could not start resource: %s", err)
	}

	conf.Addresses = []string{
		fmt.Sprintf("localhost:%s", resource.GetPort("9000/tcp"))}

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
