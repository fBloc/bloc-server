package rabbit

import (
	"sync"

	rabbit_con "github.com/fBloc/bloc-server/internal/conns/rabbit"
)

const topicExchangeName = "bloc_topic_exchange"

var (
	rabbitChannel *RabbitChannel = nil
	initialMutex  sync.Mutex
)

type RabbitChannel struct {
	conRabbitChannel *rabbit_con.RabbitChannel
}

func Connect(conf *rabbit_con.RabbitConfig) (*RabbitChannel, error) {
	initialMutex.Lock()
	defer initialMutex.Unlock()

	if rabbitChannel != nil {
		return rabbitChannel, nil
	}

	channel, err := rabbit_con.InitChannel(conf)
	if err != nil {
		return nil, err
	}

	err = channel.IniExchange(topicExchangeName, "topic")
	if err != nil {
		return nil, err
	}
	rabbitChannel = &RabbitChannel{conRabbitChannel: channel}

	return rabbitChannel, err
}

func (rmq *RabbitChannel) Pub(topic string, data []byte) error {
	return rmq.conRabbitChannel.Pub(topicExchangeName, topic, data)
}

func (rmq *RabbitChannel) Pull(
	topic, pullerTag string,
	respMsgByteChan chan []byte,
) error {
	return rmq.conRabbitChannel.Pull(topicExchangeName, topic, pullerTag, true, respMsgByteChan)
}
