package mongodb

import (
	"context"
	"sync"
	"time"

	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
	"go.mongodb.org/mongo-driver/mongo/readpref"
)

var (
	confSigMapClient     = make(map[confSignature]*mongo.Client)
	confSigMapClientLock sync.Mutex
)

// InitClient same config can only create single one client.
// this can also used to check the server is valid
func InitClient(conf *MongoConfig) (*mongo.Client, error) {
	confSigMapClientLock.Lock()
	defer confSigMapClientLock.Unlock()

	confSig := conf.signature()

	client, ok := confSigMapClient[confSig]
	if ok && client != nil {
		return client, nil
	}

	copiedConf := conf
	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	var err error
	client, err = mongo.Connect(
		ctx, options.Client().ApplyURI(copiedConf.ConnectionUrl()),
		options.Client().SetReadPreference(readpref.SecondaryPreferred()))
	if err != nil {
		return nil, err
	}

	pingCtx, pingCancel := context.WithTimeout(context.Background(), 2*time.Second)
	defer pingCancel()
	err = client.Ping(pingCtx, readpref.Primary())
	if err != nil {
		return nil, err
	}

	confSigMapClient[confSig] = client
	return client, err
}
