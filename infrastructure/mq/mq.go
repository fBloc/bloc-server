package mq

type MsgQueue interface {
	Pub(topic string, data []byte) error
	Pull(topic, pullerTag string, respMsgByteChan chan []byte) error
}
