package mongodb

import (
	"go.mongodb.org/mongo-driver/bson"
)

type MongoUpdater struct {
	setter bson.M
	pusher bson.M
	puller bson.M
}

func NewUpdater() *MongoUpdater {
	return &MongoUpdater{
		bson.M{},
		bson.M{},
		bson.M{}}
}

func (ms *MongoUpdater) AddSet(key string, val interface{}) *MongoUpdater {
	ms.setter[key] = val
	return ms
}

func (ms *MongoUpdater) AddPush(key string, val interface{}) *MongoUpdater {
	ms.pusher[key] = val
	return ms
}

func (ms *MongoUpdater) AddPull(key string, val interface{}) *MongoUpdater {
	ms.puller[key] = val
	return ms
}

func (ms *MongoUpdater) IsZero() bool {
	return len(ms.setter) == 0 && len(ms.puller) == 0 && len(ms.pusher) == 0
}

func (ms *MongoUpdater) finalStatement() bson.M {
	resp := bson.M{}

	if len(ms.setter) > 0 {
		resp["$set"] = ms.setter
	}
	if len(ms.puller) > 0 {
		resp["$pull"] = ms.puller
	}
	if len(ms.pusher) > 0 {
		resp["$push"] = ms.pusher
	}
	return resp
}
