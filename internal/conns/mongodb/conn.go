package mongodb

import (
	"context"
	"errors"
	"strconv"
	"strings"
	"time"

	"github.com/fBloc/bloc-backend-go/internal/filter_options"
	"github.com/fBloc/bloc-backend-go/value_object"

	"github.com/google/uuid"
	"go.mongodb.org/mongo-driver/bson"

	"github.com/fBloc/bloc-backend-go/internal/util"

	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
	"go.mongodb.org/mongo-driver/mongo/readpref"
)

type MongoConfig struct {
	Hosts    []string
	Port     int
	Db       string
	User     string
	Password string
}

func (mC *MongoConfig) IsNil() bool {
	if mC == nil {
		return true
	}
	return len(mC.Hosts) == 0 || mC.Port == 0 ||
		mC.Db == "" || mC.User == "" ||
		mC.Password == ""
}

func (mC MongoConfig) Equal(anotherMC MongoConfig) bool {
	if mC.Port != anotherMC.Port || mC.Db != anotherMC.Db ||
		mC.User != anotherMC.User || mC.Password != anotherMC.Password {
		return false
	}
	if len(mC.Hosts) != len(anotherMC.Hosts) {
		return false
	}
	hostMapAmount := make(map[string]uint, len(mC.Hosts))
	for _, v := range mC.Hosts {
		hostMapAmount[v]++
	}
	for _, v := range anotherMC.Hosts {
		hostMapAmount[v]--
	}
	for _, v := range hostMapAmount {
		if v != 0 {
			return false
		}
	}
	return true
}

var (
	client *mongo.Client
	config *MongoConfig
)

func CheckConfValid(conf *MongoConfig) {
	InitClient(conf)
}

func InitClient(conf *MongoConfig) *mongo.Client {
	config = conf

	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()
	client, _ = mongo.Connect(ctx, options.Client().ApplyURI(
		strings.Join([]string{
			"mongodb://",
			util.EncodeString(conf.User), ":",
			util.EncodeString(conf.Password), "@",
			conf.Hosts[0], ":",
			strconv.Itoa(conf.Port)},
			"")))

	pingCtx, pingCancel := context.WithTimeout(context.Background(), 2*time.Second)
	defer pingCancel()
	err := client.Ping(pingCtx, readpref.Primary())
	if err != nil {
		panic(err)
	}
	return client
}

func GetCollection(name string) *mongo.Collection {
	collection := client.Database(config.Db).Collection(name)
	return collection
}

type Collection struct {
	Name       string
	collection *mongo.Collection
}

func NewCollection(
	hosts []string, port int, user, password, db, collectionName string,
) *Collection {
	client := InitClient(&MongoConfig{
		Hosts:    hosts,
		Port:     port,
		User:     user,
		Password: password,
		Db:       db,
	})
	collection := client.Database(db).Collection(collectionName)
	return &Collection{
		Name:       collectionName,
		collection: collection,
	}
}

// GetByID get by id
func (c *Collection) GetByID(id uuid.UUID, resultPointer interface{}) error {
	if id == uuid.Nil {
		return errors.New("id cannot be blank string")
	}
	return c.Get(NewFilter().AddEqual("id", id), &filter_options.FilterOption{}, resultPointer)
}

func (c *Collection) Get(
	mFilter *MongoFilter,
	filterOptions *filter_options.FilterOption,
	resultPointer interface{},
) error {
	findOptions := options.FindOneOptions{}
	if filterOptions != nil {
		if len(filterOptions.SortAscFields) > 0 || len(filterOptions.SortDescFields) > 0 {
			sortOptions := bson.D{}
			for _, i := range filterOptions.SortAscFields {
				sortOptions = append(sortOptions, bson.E{Key: i, Value: 1})
			}
			for _, i := range filterOptions.SortDescFields {
				sortOptions = append(sortOptions, bson.E{Key: i, Value: -1})
			}
			findOptions.SetSort(sortOptions)
		}
	} else {
		findOptions.SetSort(bson.M{"$natural": -1})
	}
	err := c.collection.FindOne(context.TODO(), mFilter.filter, &findOptions).Decode(resultPointer)
	if err != nil && err == mongo.ErrNoDocuments {
		return nil
	}
	return err
}

func (c *Collection) FindOneAndDelete(
	mFilter *MongoFilter,
	filterOptions *filter_options.FilterOption,
	resultPointer interface{},
) error {
	findOptions := options.FindOneAndDeleteOptions{}
	if filterOptions != nil {
		if len(filterOptions.SortAscFields) > 0 || len(filterOptions.SortDescFields) > 0 {
			sortOptions := bson.D{}
			for _, i := range filterOptions.SortAscFields {
				sortOptions = append(sortOptions, bson.E{Key: i, Value: 1})
			}
			for _, i := range filterOptions.SortDescFields {
				sortOptions = append(sortOptions, bson.E{Key: i, Value: -1})
			}
			findOptions.SetSort(sortOptions)
		}
	} else {
		findOptions.SetSort(bson.M{"$natural": -1})
	}
	err := c.collection.FindOneAndDelete(
		context.TODO(),
		mFilter.filter,
		&findOptions).Decode(resultPointer)
	if err != nil && err == mongo.ErrNoDocuments {
		return nil
	}
	return err
}

// TODO 用CommonFilter替换Filter？？？
func (c *Collection) CommonFilter(
	filter value_object.RepositoryFilter,
	filterOptions value_object.RepositoryFilterOption,
	resultSlicePointer interface{},
) error {
	mongoFilter := newMongoFilterFromCommonFilter(filter)
	mongoFitlerOptions := options.FindOptions{}
	if filterOptions.Limit > 0 {
		mongoFitlerOptions.SetLimit(filterOptions.Limit)
	}
	if filterOptions.OffSet > 0 {
		mongoFitlerOptions.SetSkip(filterOptions.OffSet)
	}
	if filterOptions.Asc {
		mongoFitlerOptions.SetSort(bson.M{"$natural": 1})
	} else { // 默认使用倒序
		mongoFitlerOptions.SetSort(bson.M{"$natural": -1})
	}

	cursor, _ := c.collection.Find(context.TODO(), mongoFilter.FilterExpression(), &mongoFitlerOptions)
	return cursor.All(context.TODO(), resultSlicePointer)
}

// Filter all
func (c *Collection) Filter(
	mFilter *MongoFilter,
	filterOptions *filter_options.FilterOption,
	resultSlicePointer interface{},
) error {
	findOptions := options.FindOptions{}
	if filterOptions != nil {
		if len(filterOptions.SortAscFields) == 0 && len(filterOptions.SortDescFields) == 0 {
			findOptions.SetSort(bson.M{"$natural": -1})
		} else {
			sortOptions := bson.D{}
			for _, i := range filterOptions.SortAscFields {
				sortOptions = append(sortOptions, bson.E{Key: i, Value: 1})
			}
			for _, i := range filterOptions.SortDescFields {
				sortOptions = append(sortOptions, bson.E{Key: i, Value: -1})
			}
			findOptions.SetSort(sortOptions)
		}
		if filterOptions.Limit > 0 {
			findOptions.SetLimit(filterOptions.Limit)
		}
		if filterOptions.OffSet > 0 {
			findOptions.SetSkip(filterOptions.OffSet)
		}
	}

	cursor, _ := c.collection.Find(context.TODO(), mFilter.FilterExpression(), &findOptions)
	return cursor.All(context.TODO(), resultSlicePointer)
}

// Count count of document
func (c *Collection) Count(mFilter *MongoFilter) (int64, error) {
	return c.collection.CountDocuments(context.TODO(), mFilter.filter)
}

// InsertOne insert document
func (c *Collection) InsertOne(insertData interface{}) (string, error) {
	insertResult, err := c.collection.InsertOne(context.TODO(), insertData)
	if err != nil {
		return "", err
	}
	if oid, ok := insertResult.InsertedID.(primitive.ObjectID); ok {
		return oid.Hex(), nil
	}
	return "", errors.New("insert ok. gen ID failed")
}

// PatchByID partially update a doc, only update ipt fields
func (c *Collection) PatchByID(id uuid.UUID, mSetter *MongoUpdater) error {
	return c.Patch(NewFilter().AddEqual("id", id), mSetter)
}

func (c *Collection) Patch(mFilter *MongoFilter, mSetter *MongoUpdater) error {
	_, err := c.collection.UpdateMany(context.TODO(), mFilter.filter, mSetter.finalStatement())
	return err
}

// UpdateByID require full doc, replace all except id
func (c *Collection) ReplaceByID(id uuid.UUID, insertData interface{}) error {
	_, err := c.collection.ReplaceOne(
		context.TODO(),
		NewFilter().AddEqual("id", id).filter,
		insertData)
	return err
}

// DeleteByID delete
func (c *Collection) DeleteByID(id uuid.UUID) (int64, error) {
	if id == uuid.Nil {
		return 0, nil
	}

	return c.Delete(NewFilter().AddEqual("id", id))
}

func (c *Collection) Delete(mFilter *MongoFilter) (int64, error) {
	deleteResult, err := c.collection.DeleteMany(context.TODO(), mFilter.filter)
	return deleteResult.DeletedCount, err
}

// ClearCollection purge collection
func (c *Collection) ClearCollection() error {
	_, err := c.collection.DeleteMany(context.TODO(), map[string]interface{}{})
	return err
}
