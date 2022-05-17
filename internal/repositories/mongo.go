package repositories

import (
	"context"
	"time"

	"github.com/bervimo/events/internal/core/domain"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
	"golang.org/x/exp/slices"
)

type Document struct {
	Id primitive.ObjectID `bson:"_id" json:"id"`
}

type ChangeStream struct {
	FullDocument domain.Event `bson:"fullDocument" json:"fullDocument"`
}

type Mongo struct {
	Collection *mongo.Collection
}

type MongoOptions struct {
	URI        string
	Database   string
	Collection string
}

// NewMongo
func NewMongo(mopts MongoOptions) (*Mongo, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)

	defer cancel()

	opts := options.Client().
		ApplyURI(mopts.URI).
		SetConnectTimeout(5 * time.Second).
		SetMaxConnIdleTime(5 * time.Minute).
		SetMaxPoolSize(100)

	client, err := mongo.Connect(ctx, opts)

	if err != nil {
		return nil, err
	}

	coll := client.Database(mopts.Database).Collection(mopts.Collection)

	return &Mongo{Collection: coll}, nil
}

func (mll *Mongo) getHead(ctx context.Context) (*Document, error) {
	filter := bson.D{primitive.E{Key: "head", Value: true}}

	var head *Document

	err := mll.Collection.FindOne(ctx, filter).Decode(&head)

	if err != nil {
		return nil, err
	}

	return head, nil
}

func (mll *Mongo) unsetHead(ctx context.Context) (bool, error) {
	head, err := mll.getHead(ctx)

	if head == nil || err != nil {
		return false, err
	}

	ops := []mongo.WriteModel{
		&mongo.UpdateOneModel{
			Filter: bson.D{primitive.E{Key: "_id", Value: head.Id}},
			Update: bson.D{primitive.E{Key: "$unset", Value: bson.D{primitive.E{Key: "head", Value: ""}}}},
		},
	}

	res, err := mll.Collection.BulkWrite(ctx, ops)

	return res.ModifiedCount == 1, err
}

func (mll *Mongo) insert(ctx context.Context, events []domain.Event) func(mongo.SessionContext) (any, error) {
	return func(sc mongo.SessionContext) (any, error) {
		_, err := mll.unsetHead(ctx)

		if err != nil {
			return false, err
		}

		// Insert nodes
		ops := make([]mongo.WriteModel, len(events))

		upsert := true

		for i, event := range events {
			now := time.Now().UTC().Format("2006-01-02T15:04:05.000+00:00")

			event.UpdatedAt = &now

			ops[i] = &mongo.UpdateOneModel{
				Filter: bson.D{primitive.E{Key: "id", Value: event.Id}},
				Update: bson.D{
					primitive.E{Key: "$set", Value: event},
					primitive.E{Key: "$setOnInsert", Value: bson.D{primitive.E{Key: "createdAt", Value: now}}},
				},
				Upsert: &upsert,
			}
		}

		res, err := mll.Collection.BulkWrite(ctx, ops)

		if err != nil {
			return false, err
		}

		return res.UpsertedCount > 0, err
	}

}

// Insert
func (mll *Mongo) Insert(events []domain.Event) (bool, error) {
	session, err := mll.Collection.
		Database().
		Client().
		StartSession()

	if err != nil {
		return false, err
	}

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)

	defer cancel()
	defer session.EndSession(ctx)

	result, err := session.WithTransaction(ctx, mll.insert(ctx, events))

	return result.(bool), err
}

func sort(events []domain.Event) {
	slices.SortFunc(events, func(a, b domain.Event) bool {
		next := b.Next

		if next == nil {
			return true
		}

		return *next < a.Id
	})
}

// Get
func (mll *Mongo) Get(matchId string) ([]domain.Event, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)

	defer cancel()

	filter := bson.D{primitive.E{Key: "matchId", Value: matchId}}

	cur, err := mll.Collection.Find(ctx, filter)

	if err != nil {
		return nil, err
	}

	var events []domain.Event

	err = cur.All(ctx, &events)

	if err != nil {
		return nil, err
	}

	sort(events)

	return events, nil
}

// OnEvents
func (mll *Mongo) OnEvents(callback func(*domain.Event, error)) {
	ctx, cancel := context.WithCancel(context.Background())

	defer cancel()

	opts := options.ChangeStream().SetFullDocument(options.UpdateLookup)

	pipe := bson.D{bson.E{
		Key: "$match",
		Value: bson.D{bson.E{
			Key: "operationType",
			Value: bson.D{bson.E{
				Key: "$in",
				Value: bson.A{
					"insert",
					"update",
					"replace",
				},
			}},
		}},
	}}

	mll.Collection.Database().Client()

	stream, err := mll.Collection.Watch(ctx, mongo.Pipeline{pipe}, opts)

	if err != nil {
		callback(nil, err)

		return
	}

	defer stream.Close(ctx)

	for stream.Next(ctx) {
		var cs ChangeStream

		err := stream.Decode(&cs)

		callback(&cs.FullDocument, err)
	}
}

// Close
func (mll *Mongo) Close() error {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)

	defer cancel()

	return mll.Collection.Database().Client().Disconnect(ctx)
}
