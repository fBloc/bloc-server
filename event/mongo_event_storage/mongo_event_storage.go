package mongo_event_storage

import (
	"context"
	"time"

	"github.com/fBloc/bloc-backend-go/event"
	"github.com/fBloc/bloc-backend-go/internal/conns/mongodb"
	"github.com/fBloc/bloc-backend-go/internal/filter_options"
)

func init() {
	var _ event.FuturePubEventStorage = &MongoEventStorage{}
}

const (
	DefaultCollectionName = "event_storage"
)

type MongoEventStorage struct {
	mongoCollection *mongodb.Collection
}

func New(
	ctx context.Context,
	hosts []string, port int, user, password, db, collectionName string,
) (*MongoEventStorage, error) {
	collection := mongodb.NewCollection(
		hosts, port, user, password, db, collectionName,
	)
	return &MongoEventStorage{mongoCollection: collection}, nil
}

type mongoFutureEvent struct {
	Tag        string    `bson:"event_tag"`
	EventData  []byte    `bson:"event_data"`
	PubTime    time.Time `bson:"pub_time"`
	recordTime time.Time `bson:"record_time"`
}

func NewFromEvent(
	e event.DomainEvent,
	pubTime time.Time,
) (*mongoFutureEvent, error) {
	eventData, err := e.Marshal()
	if err != nil {
		return nil, err
	}

	resp := mongoFutureEvent{
		Tag:        e.Topic(),
		EventData:  eventData,
		PubTime:    pubTime,
		recordTime: time.Now(),
	}
	return &resp, nil
}

func (m *MongoEventStorage) Add(
	event event.DomainEvent, pubTime time.Time,
) error {
	mIns, err := NewFromEvent(event, pubTime)
	if err != nil {
		return err
	}

	_, err = m.mongoCollection.InsertOne(*mIns)
	return err
}

// PopLatestBeforeATime 发布特定时间之前的最晚的一条记录
func (m *MongoEventStorage) PopLatestBeforeATime(
	theTime time.Time,
) (tag string, data []byte, err error) {
	var resp mongoFutureEvent
	err = m.mongoCollection.FindOneAndDelete(
		mongodb.NewFilter().AddGte("pub_time", theTime),
		&filter_options.FilterOption{SortDescFields: []string{"record_time"}},
		&resp)
	if err != nil {
		return
	}
	tag = resp.Tag
	data = resp.EventData
	return
}

// PopEarliestAfterATime 发布特定时间之后的最早的一条记录
func (m *MongoEventStorage) PopEarliestAfterATime(
	theTime time.Time,
) (tag string, data []byte, err error) {
	var resp mongoFutureEvent
	err = m.mongoCollection.FindOneAndDelete(
		mongodb.NewFilter().AddLte("pub_time", theTime),
		&filter_options.FilterOption{SortAscFields: []string{"record_time"}},
		&resp)
	if err != nil {
		return
	}
	tag = resp.Tag
	data = resp.EventData
	return
}
