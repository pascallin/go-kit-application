package db

import (
	"context"
	"fmt"
	"go.mongodb.org/mongo-driver/mongo/readpref"
	"os"
	"time"

	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

const (
	// Timeout operations after N seconds
	connectTimeout           = 5
	mongoConnStringTemplate = "mongodb://%s:%s@%s"
)

func getMongoConnURL() string {
	username := os.Getenv("MONGODB_USERNAME")
	password := os.Getenv("MONGODB_PASSWORD")
	clusterEndpoint := os.Getenv("MONGODB_ENDPOINT")

	connectionURI := fmt.Sprintf(mongoConnStringTemplate, username, password, clusterEndpoint)
	return connectionURI
}

type MongoDatabase struct {
	DB      *mongo.Database
	Client  *mongo.Client
	Context context.Context
}

var MongoDB *MongoDatabase

func NewMongoDatabase() (*MongoDatabase, error) {
	connection := getMongoConnURL()
	dbname := os.Getenv("MONGODB_DATABASE")
	ctx, cancel := context.WithTimeout(context.Background(), connectTimeout*time.Second)
	defer cancel()
	client, err := mongo.Connect(ctx, options.Client().ApplyURI(connection))
	if err != nil {
		return nil, err
	}
	ctxping, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	err = client.Ping(ctxping, readpref.Primary())
	if err != nil {
		return nil, err
	}
	db := client.Database(dbname)
	MongoDB = &MongoDatabase{DB: db, Client: client, Context: ctx}
	fmt.Println("mongodb connected")
	return MongoDB, nil
}

func (d *MongoDatabase) Close() {
	d.Client.Disconnect(d.Context)
}