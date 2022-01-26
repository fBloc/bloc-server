package mongodb

import (
	"context"
	"errors"

	"github.com/fBloc/bloc-server/internal/filter_options"
	"github.com/fBloc/bloc-server/value_object"

	"go.mongodb.org/mongo-driver/bson"

	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

type Collection struct {
	Name       string
	collection *mongo.Collection
}

// NewCollection return mongo collection(if u not familiar with it, think as sql table)
func NewCollection(
	mC *MongoConfig, collectionName string,
) (*Collection, error) {
	client, err := InitClient(mC)
	if err != nil {
		return nil, err
	}

	collection := client.Database(mC.Db).Collection(collectionName)
	return &Collection{Name: collectionName, collection: collection}, nil
}

// GetByID get by id
func (c *Collection) GetByID(id value_object.UUID, resultPointer interface{}) error {
	if id.IsNil() {
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

func (c *Collection) CommonCount(
	filter value_object.RepositoryFilter,
) (int64, error) {
	mongoFilter := newMongoFilterFromCommonFilter(filter)
	return c.collection.CountDocuments(context.TODO(), mongoFilter.filter)
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

// FindOneOrInsert first try to find the doc by filter:
// if exist, do nothing & put the find record to oldDocResultPointer
// if not exist, insert the insertData & oldDocResultPointer keep point to blank content
func (c *Collection) FindOneOrInsert(
	mFilter *MongoFilter,
	insertData interface{},
	oldDocResultPointer interface{},
) (alreadyExist bool, err error) {
	err = c.collection.FindOneAndUpdate(
		context.TODO(),
		mFilter.filter,
		bson.M{"$setOnInsert": insertData},
		options.FindOneAndUpdate().SetUpsert(true).SetReturnDocument(options.Before),
	).Decode(oldDocResultPointer)
	if err != nil && err == mongo.ErrNoDocuments {
		err = nil
	} else {
		alreadyExist = true
	}
	return
}

// PatchByID partially update a doc, only update ipt fields
func (c *Collection) PatchByID(id value_object.UUID, mSetter *MongoUpdater) error {
	return c.Patch(NewFilter().AddEqual("id", id), mSetter)
}

func (c *Collection) Patch(mFilter *MongoFilter, mSetter *MongoUpdater) error {
	_, err := c.collection.UpdateMany(context.TODO(), mFilter.filter, mSetter.finalStatement())
	return err
}

// UpdateByID require full doc, replace all except id
func (c *Collection) ReplaceByID(id value_object.UUID, insertData interface{}) error {
	_, err := c.collection.ReplaceOne(
		context.TODO(),
		NewFilter().AddEqual("id", id).filter,
		insertData)
	return err
}

// DeleteByID delete
func (c *Collection) DeleteByID(id value_object.UUID) (int64, error) {
	if id.IsNil() {
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
