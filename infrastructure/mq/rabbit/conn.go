package rabbit

import (
	"log"
	"strings"

	"github.com/fBloc/bloc-server/infrastructure/mq"

	"github.com/pkg/errors"
	"github.com/streadway/amqp"
)

func init() {
	var _ mq.MsgQueue = &RabbitMQ{}
}

const topicExchangeName = "bloc_topic_exchange"

type RabbitMQ struct {
	channel *amqp.Channel
}

type RabbitConfig struct {
	User     string
	Password string
	Host     string
	Vhost    string
}

func (rC *RabbitConfig) IsNil() bool {
	if rC == nil {
		return true
	}
	return rC.Host == "" ||
		rC.User == "" ||
		rC.Password == ""
}

func failOnError(err error, msg string) {
	if err != nil {
		log.Fatalf("%s: %s", msg, err)
	}
}

func InitChannel(conf *RabbitConfig) *RabbitMQ {
	conStr := strings.Join([]string{
		"amqp://",
		conf.User, ":",
		conf.Password, "@",
		conf.Host, "/",
		conf.Vhost}, "")
	connection, err := amqp.Dial(conStr)
	failOnError(err, "Failed to connect to RabbitMQ")

	channel, err := connection.Channel()
	failOnError(err, "Failed to open a channel")

	channel.Qos(1, 0, false)
	channel.ExchangeDeclare(
		topicExchangeName,
		"topic",
		true, false, false, false, nil)
	return &RabbitMQ{channel: channel}
}

func (rmq *RabbitMQ) initQueueAndBindToExchange(
	topic, queueName string,
) (amqp.Queue, error) {
	var err error
	q, err := rmq.channel.QueueDeclare(
		queueName, // name
		true,      // durable
		false,     // delete when unused
		false,     // exclusive
		false,     // no-wait
		nil,       // arguments
	)
	if err != nil {
		return amqp.Queue{}, err
	}

	err = rmq.channel.QueueBind(
		q.Name,            // queue name
		topic,             // routing key
		topicExchangeName, // exchange
		false,
		nil)
	return q, err
}

func (rmq *RabbitMQ) Pub(topic string, data []byte) error {
	err := rmq.channel.Publish(
		topicExchangeName, // exchange
		topic,             // routing key
		false,             // mandatory
		false,             // immediate
		amqp.Publishing{
			DeliveryMode: amqp.Persistent,
			ContentType:  "text/plain",
			Body:         data,
		})
	return err
}

func (rmq *RabbitMQ) Pull(
	topic, pullerTag string,
	respMsgByteChan chan []byte,
) error {
	queue, err := rmq.initQueueAndBindToExchange(topic, pullerTag)
	if err != nil {
		return errors.Wrap(err, "initial queue & bind to exchange failed")
	}

	msgs, err := rmq.channel.Consume(
		queue.Name, // queue
		"",         // consumer
		true,       // auto-ack, 不依赖底层具体实现的consume保证，应用层去保证
		false,      // exclusive
		false,      // no-local
		false,      // no-wait
		nil,        // args
	)
	if err != nil {
		return errors.Wrap(err, "failed to register a consumer")
	}

	go func() {
		for d := range msgs {
			respMsgByteChan <- d.Body
		}
	}()

	return nil
}
