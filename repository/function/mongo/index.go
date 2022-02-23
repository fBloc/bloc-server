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
				"name": "text",
			},
			Options: nil,
		},
		{
			Keys: bson.D{
				{Key: "ipt_digest", Value: 1},
				{Key: "opt_digest", Value: 1},
			},
			Options: nil,
		},
		{
			Keys: bson.M{
				"read_user_ids": 1,
			},
			Options: nil,
		},
	}
}
