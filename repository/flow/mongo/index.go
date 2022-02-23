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
				"origin_id": "hashed",
			},
			Options: &options.IndexOptions{
				Sparse: &truePoint,
			},
		},
		{
			Keys: bson.M{
				"is_draft": 1,
			},
			Options: nil,
		},
		{
			Keys: bson.M{
				"read_user_ids": 1,
			},
			Options: nil,
		},
		{
			Keys: bson.M{
				"name": "text",
			},
			Options: nil,
		},
	}
}
