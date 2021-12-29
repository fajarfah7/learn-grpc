package mongodb

// tutorial install mongodb on ubuntu 18.04
// https://linuxize.com/post/how-to-install-mongodb-on-ubuntu-18-04/

// get golang package for mongodb
// go get go.mongodb.org/mongo-driver/mongo

// list comparison of sql command and mongo command
// https://docs.mongodb.com/manual/reference/sql-comparison/

import (
	"context"
	"log"
	"time"

	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

type MongoDB struct {
	URI   string
	DB    string
	Table string
}

func (m *MongoDB) Connect() (*mongo.Client, *mongo.Collection, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 20*time.Second)
	defer cancel()

	client, err := mongo.Connect(ctx, options.Client().ApplyURI(m.URI))
	if err != nil {
		log.Fatalf("error while setup mongo client\n%v", err)
		return nil, nil, err
	}
	return client, client.Database(m.DB).Collection(m.Table), nil
}

func (m *MongoDB) Disconnect(client *mongo.Client) {
	client.Disconnect(context.Background())
}
