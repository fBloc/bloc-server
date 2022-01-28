package rabbit

import (
	"fmt"
	"log"
	"os"
	"testing"

	"github.com/brianvoe/gofakeit/v6"
	"github.com/ory/dockertest/v3"
	"github.com/ory/dockertest/v3/docker"
)

var (
	channel      *RabbitChannel
	msg          = gofakeit.Name()
	exchangeName = "test-exhange"
	routingKey   = "xx"
	queueNAme    = "queue"
	conf         = &RabbitConfig{
		User:     "blocRabbit",
		Password: "blocRabbitPasswd",
	}
)

func TestRabbit(t *testing.T) {
	var err error
	exist := make(chan struct{})

	respChan := make(chan []byte)
	err = channel.Pull(
		exchangeName, routingKey, queueNAme, true, respChan,
	)
	if err != nil {
		t.Fatal(err)
	}

	go func() {
		for i := range respChan {
			if string(i) != msg {
				t.Fatalf("received msg not right, expect: %s, received: %s", msg, string(i))
			}
			exist <- struct{}{}
			return
		}
	}()

	err = channel.Pub(exchangeName, routingKey, []byte(msg))
	if err != nil {
		t.Fatal(err)
	}

	<-exist
}

func TestMain(m *testing.M) {
	pool, err := dockertest.NewPool("")
	if err != nil {
		log.Fatalf("Could not connect to docker: %s", err)
	}

	resource, err := pool.RunWithOptions(&dockertest.RunOptions{
		Repository: "rabbitmq",
		Tag:        "3.9.11-management",
		Env: []string{
			// username and password for mongodb superuser
			"RABBITMQ_DEFAULT_USER=" + conf.User,
			"RABBITMQ_DEFAULT_PASS=" + conf.Password,
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

	conf.Host = []string{
		fmt.Sprintf("localhost:%s", resource.GetPort("5672/tcp"))}

	// exponential backoff-retry, because the application in the container might not be ready to accept connections yet
	err = pool.Retry(func() error {
		var err error
		channel, err = InitChannel(conf)
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
