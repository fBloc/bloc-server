package mq_msg

type MqMsg interface {
	String() string
	Ack() error
	Nack() error
}
