package rabbit

import (
	"fmt"

	"github.com/sirius1024/go-amqp-reconnect/rabbitmq"

	"github.com/streadway/amqp"
)

type RabbitChannel struct {
	conf    *RabbitConfig
	channel *rabbitmq.Channel
}

func InitChannel(conf *RabbitConfig) (*RabbitChannel, error) {
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
	if err != nil {
		return nil, err
	}

	channel, err := connection.Channel()
	if err != nil {
		return nil, err
	}

	channel.Qos(1, 0, false)
	return &RabbitChannel{conf: conf, channel: channel}, nil
}

func (rC *RabbitChannel) IniExchange(exchange string, exchangeType string) error {
	return rC.channel.ExchangeDeclare(exchange, exchangeType, true, false, false, false, nil)
}

func (rC *RabbitChannel) initQueAndBindToExchange(queue, exchange, routingKey string) error {
	var err error

	q, err := rC.channel.QueueDeclare(
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

	err = rC.channel.QueueBind(
		q.Name,     // queue name
		routingKey, // routing key
		exchange,   // exchange
		false,
		nil)
	return err
}

func (rC *RabbitChannel) Pub(
	exchange, routingKey string, value []byte,
) (err error) {
	err = rC.IniExchange(exchange, "topic")
	if err != nil {
		return err
	}

	err = rC.channel.Publish(
		exchange,   // exchange
		routingKey, // routing key
		false,      // mandatory
		false,      // immediate
		amqp.Publishing{
			DeliveryMode: amqp.Persistent,
			ContentType:  "text/plain",
			Body:         value,
		})
	return
}

func (rC *RabbitChannel) Pull(
	exchange, routingKey, queue string,
	autoAck bool,
	respMsgByteChan chan []byte,
) (err error) {
	err = rC.IniExchange(exchange, "topic")
	if err != nil {
		return
	}

	err = rC.initQueAndBindToExchange(queue, exchange, routingKey)
	if err != nil {
		return
	}

	msgs, err := rC.channel.Consume(
		queue,   // queue
		"",      // consumer
		autoAck, // auto-ack
		false,   // exclusive
		false,   // no-local
		false,   // no-wait
		nil,     // args
	)
	if err != nil {
		return
	}

	go func() {
		for d := range msgs {
			respMsgByteChan <- d.Body
		}
	}()
	return
}
