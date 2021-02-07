package modules

import (
	"context"
	"log"
	"os"
	"time"

	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
	"go.mongodb.org/mongo-driver/mongo/readpref"
)

func MongoDBConnect(ctx context.Context) *mongo.Client {
	client, err := mongo.NewClient(options.Client().ApplyURI(os.Getenv("MONGO_URL")))
	if err != nil {
		log.Fatal(err)
	}

	ctxConnectTimeout, cancelConnectTimeout := context.WithTimeout(ctx, 30*time.Second)
	defer cancelConnectTimeout()

	err = client.Connect(ctxConnectTimeout)
	if err != nil {
		log.Fatal("Cannot Connect To MongoDB Server")
	}

	ctxPingTimeout, cancelPingTimeout := context.WithTimeout(ctx, 30*time.Second)
	defer cancelPingTimeout()

	err = client.Ping(ctxPingTimeout, readpref.Primary())
	if err != nil {
		log.Fatal("Cannot Ping To MongoDB Server")
	}

	return client
}
