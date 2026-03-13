package database

import (
	"context"
	"log"
	"time"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

var Client *mongo.Client
var DB *mongo.Database

func Connect(uri, dbName string) error {
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	clientOptions := options.Client().ApplyURI(uri)
	client, err := mongo.Connect(ctx, clientOptions)
	if err != nil {
		return err
	}

	if err := client.Ping(ctx, nil); err != nil {
		return err
	}

	Client = client
	DB = client.Database(dbName)

	log.Printf("Connected to MongoDB: %s", dbName)
	return nil
}

func Disconnect() {
	if Client != nil {
		ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
		defer cancel()
		Client.Disconnect(ctx)
	}
}

// Collections

func Products() *mongo.Collection {
	return DB.Collection("products")
}

func Categories() *mongo.Collection {
	return DB.Collection("categories")
}

func Orders() *mongo.Collection {
	return DB.Collection("orders")
}

func Images() *mongo.Collection {
	return DB.Collection("images")
}

// EnsureIndexes creates required indexes for the 3D store collections
func EnsureIndexes() error {
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	// products: unique slug
	_, err := Products().Indexes().CreateOne(ctx, mongo.IndexModel{
		Keys:    bson.D{{Key: "slug", Value: 1}},
		Options: options.Index().SetUnique(true),
	})
	if err != nil {
		return err
	}

	// products: index on category_id
	_, err = Products().Indexes().CreateOne(ctx, mongo.IndexModel{
		Keys: bson.D{{Key: "category_id", Value: 1}},
	})
	if err != nil {
		return err
	}

	// products: compound index on active + featured
	_, err = Products().Indexes().CreateOne(ctx, mongo.IndexModel{
		Keys: bson.D{{Key: "active", Value: 1}, {Key: "featured", Value: 1}},
	})
	if err != nil {
		return err
	}

	// products: text index on name and description
	_, err = Products().Indexes().CreateOne(ctx, mongo.IndexModel{
		Keys: bson.D{{Key: "name", Value: "text"}, {Key: "description", Value: "text"}},
	})
	if err != nil {
		return err
	}

	// categories: unique slug
	_, err = Categories().Indexes().CreateOne(ctx, mongo.IndexModel{
		Keys:    bson.D{{Key: "slug", Value: 1}},
		Options: options.Index().SetUnique(true),
	})
	if err != nil {
		return err
	}

	// categories: compound index on active + sort_order
	_, err = Categories().Indexes().CreateOne(ctx, mongo.IndexModel{
		Keys: bson.D{{Key: "active", Value: 1}, {Key: "sort_order", Value: 1}},
	})
	if err != nil {
		return err
	}

	// orders: compound index on user_id + created_at
	_, err = Orders().Indexes().CreateOne(ctx, mongo.IndexModel{
		Keys: bson.D{{Key: "user_id", Value: 1}, {Key: "created_at", Value: -1}},
	})
	if err != nil {
		return err
	}

	// orders: compound index on status + created_at
	_, err = Orders().Indexes().CreateOne(ctx, mongo.IndexModel{
		Keys: bson.D{{Key: "status", Value: 1}, {Key: "created_at", Value: -1}},
	})
	if err != nil {
		return err
	}

	// images: compound index on group_id + size_label
	_, err = Images().Indexes().CreateOne(ctx, mongo.IndexModel{
		Keys: bson.D{{Key: "group_id", Value: 1}, {Key: "size_label", Value: 1}},
	})
	if err != nil {
		return err
	}

	log.Println("3D Store indexes ensured")
	return nil
}
