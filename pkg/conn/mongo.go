package conn

import (
	"context"
	"time"

	"github.com/sirupsen/logrus"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
	"go.mongodb.org/mongo-driver/mongo/readpref"

	"github.com/pascallin/go-kit-application/config"
)

const (
	// Timeout operations after N seconds
	connectTimeout = 5
)

type Mongo struct {
	DB     *mongo.Database
	Client *mongo.Client
}

func GetMongo(ctx context.Context) (*Mongo, error) {
	connectionURI := config.GetMongoConfig().URI

	// defer cancel()
	client, err := mongo.Connect(ctx, options.Client().ApplyURI(connectionURI))
	if err != nil {
		return nil, err
	}

	db := client.Database(config.GetMongoConfig().DATABASE)

	return &Mongo{DB: db, Client: client}, nil
}

func Ping() string {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	mongo, err := GetMongo(ctx)
	if err != nil {
		logrus.Error(err)
		return "fail"
	}
	err = mongo.Client.Ping(ctx, readpref.Primary())
	if err != nil {
		logrus.Error(err)
		return "fail"
	}

	return "ok"
}
