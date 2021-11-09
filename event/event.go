package event

import (
	"time"

	"github.com/fBloc/bloc-backend-go/infrastructure/mq"

	"github.com/pkg/errors"
)

type DomainEvent interface {
	Topic() string
	Marshal() ([]byte, error)
	Unmarshal(data []byte) error
	Identity() string
}

var needInitialMqInsAsEventChannelError = errors.New("lack init event mq rely")

// EventChannel 存取event的的通道，
// 由于是分布式的，故肯定需要引入消息队列中间件
type eventDriver struct {
	mqIns                       mq.MsgQueue
	futureEventStorageImplement FuturePubEventStorage
}

var (
	driver = eventDriver{}
)

func InjectMq(eventChannel mq.MsgQueue) {
	driver.mqIns = eventChannel
}

func InjectFutureEventStorageImplement(fES FuturePubEventStorage) {
	driver.futureEventStorageImplement = fES
}

func PubEvent(event DomainEvent) error {
	if driver.mqIns == nil {
		panic(needInitialMqInsAsEventChannelError)
	}

	eventByteData, err := event.Marshal()
	if err != nil {
		return errors.Wrap(err, "event marshall failed")
	}

	err = driver.mqIns.Pub(event.Topic(), eventByteData)
	if err != nil {
		panic(err)
		return errors.Wrap(err, "pub event failed")
	}
	return nil
}

// PubEventAtCertainTime 在未来某个时间发布事件
func PubEventAtCertainTime(event DomainEvent, pubTime time.Time) error {
	if driver.mqIns == nil {
		panic(needInitialMqInsAsEventChannelError)
	}

	// 只能发布未来的
	if pubTime.Before(time.Now()) {
		return errors.New("cannot pub event in the past")
	}

	return driver.futureEventStorageImplement.Add(event, pubTime)
}

/*
ListenEvent 监听某项事件

对比PubEvent，为什么多了listenerTag参数呢？
因为发布是发布一种类型的事件，其不需要也不应该知道有哪些地方需要订阅此事件
也就是说对于同一个事件的发布，可能有多个订阅者，所以需要传入订阅者的标识
*/
func ListenEvent(
	event DomainEvent, listenerTag string,
	respEventChan chan DomainEvent,
) error {
	if driver.mqIns == nil {
		panic(needInitialMqInsAsEventChannelError)
	}

	msgByteChan := make(chan []byte)
	err := driver.mqIns.Pull(event.Topic(), listenerTag, msgByteChan)
	if err != nil {
		return errors.Wrap(err, "pull event failed")
	}

	go func() {
		for msgByte := range msgByteChan {
			err = event.Unmarshal(msgByte)
			if err != nil {
				panic(err)
			}
			respEventChan <- event
		}
	}()

	return nil
}
