package rabbit

import (
	"fmt"
	"log"

	"github.com/fBloc/bloc-server/internal/mq_msg"
	"github.com/sirius1024/go-amqp-reconnect/rabbitmq"

	"github.com/streadway/amqp"
)

type RabbitConfig struct {
	User     string
	Password string
	Host     []string
	Vhost    string
}

func (rC *RabbitConfig) IsNil() bool {
	if rC == nil {
		return true
	}
	return rC.User == "" || rC.Password == "" || len(rC.Host) <= 0
}

func failOnError(err error, msg string) {
	if err != nil {
		log.Fatalf("%s: %s", msg, err)
	}
}

var channel *amqp.Channel

func InitChannel(conf *RabbitConfig) {
	var connection *rabbitmq.Connection
	var err error
	if len(conf.Host) > 1 { // cluster
		conStrs := make([]string, 0, len(conf.Host))
		for _, i := range conf.Host {
			conStrs = append(
				conStrs,
				fmt.Sprintf("amqp://%s:%s@%s/%s", conf.User, conf.Password, i, conf.Vhost))
		}
		connection, err = rabbitmq.DialCluster(conStrs)
	} else {
		connection, err = rabbitmq.Dial(
			fmt.Sprintf(
				"amqp://%s:%s@%s/%s",
				conf.User, conf.Password,
				conf.Host[0], conf.Vhost))
	}
	failOnError(err, "Failed to connect to RabbitMQ")

	channel, err := connection.Channel()
	failOnError(err, "Failed to open a channel")

	channel.Qos(1, 0, false)
}
func GetChannel() *amqp.Channel {
	return channel
}

func iniExchange(exchange string, ch *amqp.Channel) error {
	return ch.ExchangeDeclare(exchange, "direct", true, false, false, false, nil)
}

func initQueAndBindToExchange(queue, exchange, routingKey string) error {
	var err error
	ch := GetChannel()

	q, err := ch.QueueDeclare(
		queue, // name
		true,  // durable
		false, // delete when unused
		false, // exclusive
		false, // no-wait
		nil,   // arguments
	)
	if err != nil {
		return err
	}

	err = ch.QueueBind(
		q.Name,     // queue name
		routingKey, // routing key
		exchange,   // exchange
		false,
		nil)
	return err
}

func Pub(exchange string, routingKey string, value string) error {
	var err error
	ch := GetChannel()
	err = iniExchange(exchange, ch)
	if err != nil {
		return err
	}

	err = ch.Publish(
		exchange,   // exchange
		routingKey, // routing key
		false,      // mandatory
		false,      // immediate
		amqp.Publishing{
			DeliveryMode: amqp.Persistent,
			ContentType:  "text/plain",
			Body:         []byte(value),
		})
	return err
}

type mqMsg struct {
	d *amqp.Delivery
}

func (msg mqMsg) String() string {
	return string(msg.d.Body)
}

func (msg mqMsg) Ack() error {
	return msg.d.Ack(false)
}

func (msg mqMsg) Nack() error {
	return msg.d.Nack(false, false)
}

func Pull(exchange, queue string, autoAck bool, msg chan mq_msg.MqMsg) {
	ch := GetChannel()
	iniExchange(exchange, ch)
	initQueAndBindToExchange(queue, exchange, queue)

	msgs, err := ch.Consume(
		queue,   // queue
		"",      // consumer
		autoAck, // auto-ack
		false,   // exclusive
		false,   // no-local
		false,   // no-wait
		nil,     // args
	)
	if err != nil {
		panic(err.Error())
	}

	go func() {
		for d := range msgs {
			m := mqMsg{&d}
			msg <- m
		}
	}()
}
