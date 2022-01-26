package mongodb

import "go.mongodb.org/mongo-driver/mongo"

func Connect(conf *MongoConfig) (client *mongo.Client, err error) {
	client, err = InitClient(conf)
	return
}
