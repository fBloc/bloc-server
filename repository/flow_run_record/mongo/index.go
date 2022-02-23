package mongo

import (
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

func mongoDBIndexes() []mongo.IndexModel {
	truePoint := true
	return []mongo.IndexModel{
		{
			Keys: bson.M{
				"id": "hashed",
			},
			Options: &options.IndexOptions{
				Sparse: &truePoint,
			},
		},
		{
			Keys: bson.M{
				"flow_id": "hashed",
			},
			Options: &options.IndexOptions{
				Sparse: &truePoint,
			},
		},
		{
			Keys: bson.M{
				"flow_origin_id": "hashed",
			},
			Options: &options.IndexOptions{
				Sparse: &truePoint,
			},
		},
	}
}
