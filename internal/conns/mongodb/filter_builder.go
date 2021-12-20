package mongodb

import (
	"github.com/fBloc/bloc-server/value_object"

	"go.mongodb.org/mongo-driver/bson"
)

type MongoFilter struct {
	filter bson.M
}

func (mf *MongoFilter) FilterExpression() bson.M {
	if mf == nil {
		return bson.M{}
	}
	return mf.filter
}

func NewFilter() *MongoFilter {
	return &MongoFilter{bson.M{}}
}

func newMongoFilterFromCommonFilter(cFilter value_object.RepositoryFilter) *MongoFilter {
	filter := NewFilter()
	eq := cFilter.GetEqual()
	for k, v := range eq {
		filter.AddEqual(k, v)
	}

	notEq := cFilter.GetNotEqual()
	for k, v := range notEq {
		filter.AddNotEqual(k, v)
	}

	strContains := cFilter.GetStrContains()
	for k, v := range strContains {
		filter.AddContains(k, v)
	}

	in := cFilter.GetIn()
	for k, v := range in {
		filter.AddIn(k, v)
	}

	mustExistFields := cFilter.GetFiledExist()
	for _, i := range mustExistFields {
		filter.AddExist(i)
	}

	mustNotExistFields := cFilter.GetFiledNotExist()
	for _, i := range mustNotExistFields {
		filter.AddNotExist(i)
	}

	gte := cFilter.GetGte()
	for k, v := range gte {
		filter.AddGte(k, v)
	}

	gt := cFilter.GetGt()
	for k, v := range gt {
		filter.AddGt(k, v)
	}

	lte := cFilter.GetLte()
	for k, v := range lte {
		filter.AddLte(k, v)
	}

	lt := cFilter.GetLt()
	for k, v := range lt {
		filter.AddLt(k, v)
	}

	return filter
}

func (mf *MongoFilter) AddEqual(key string, val interface{}) *MongoFilter {
	mf.filter[key] = val
	return mf
}

func (mf *MongoFilter) AddNotEqual(key string, val interface{}) *MongoFilter {
	mf.filter[key] = bson.M{"$ne": val}
	return mf
}

func (mf *MongoFilter) AddContains(key, val string) *MongoFilter {
	mf.filter[key] = bson.M{"$regex": ".*" + val + ".*"}
	return mf
}

func (mf *MongoFilter) AddIn(key string, val []interface{}) *MongoFilter {
	mf.filter[key] = bson.M{"$in": val}
	return mf
}

func (mf *MongoFilter) AddExist(key string) *MongoFilter {
	mf.filter[key] = bson.M{"$exists": true, "$ne": ""}
	return mf
}

func (mf *MongoFilter) AddNotExist(key string) *MongoFilter {
	mf.filter[key] = bson.M{"$exists": false}
	return mf
}

func (mf *MongoFilter) AddGt(key string, val interface{}) *MongoFilter {
	mf.filter[key] = bson.M{"$gt": val}
	return mf
}

func (mf *MongoFilter) AddGte(key string, val interface{}) *MongoFilter {
	mf.filter[key] = bson.M{"$gte": val}
	return mf
}

func (mf *MongoFilter) AddLt(key string, val interface{}) *MongoFilter {
	mf.filter[key] = bson.M{"$lt": val}
	return mf
}

func (mf *MongoFilter) AddLte(key string, val interface{}) *MongoFilter {
	mf.filter[key] = bson.M{"$lte": val}
	return mf
}
