package main

import (
	"context"
	"log"

	"go.mongodb.org/mongo-driver/v2/bson"
	"go.mongodb.org/mongo-driver/v2/mongo"
	"go.mongodb.org/mongo-driver/v2/mongo/options"
)

type MongoAttendant struct {
	client    *mongo.Client
	txRawChan chan string
	ixRawChan chan string
	ixIndexCh chan bson.M
	ixCh      chan bson.M
}

const (
	MongoDataSource   = "mongodb://localhost:27017"
	Database          = "openbook_v2"
	CollectionTxRaw   = "tx_raw"
	CollectionIxRaw   = "ix_raw"
	CollectionIxIndex = "ix_index"
	CollectionIx      = "ix"
)

func NewMongoAttendant(txRawChan chan string, ixRawChan chan string, ixIndexCh chan bson.M, ixCh chan bson.M) *MongoAttendant {
	client, err := mongo.Connect(options.Client().ApplyURI(MongoDataSource))
	if err != nil {
		log.Fatal(err)
	}

	return &MongoAttendant{
		client:    client,
		txRawChan: txRawChan,
		ixRawChan: ixRawChan,
		ixIndexCh: ixIndexCh,
		ixCh:      ixCh,
	}
}

func (ma *MongoAttendant) serve(ctx context.Context) {
	go ma.serveTxRaw(ctx)
	go ma.serveIxRaw(ctx)
	go ma.serveIxIndex(ctx)
	go ma.serveIx(ctx)
}

func (ma *MongoAttendant) serveTxRaw(ctx context.Context) {
	for txRaw := range ma.txRawChan {
		ma.client.Database(Database).Collection(CollectionTxRaw).InsertOne(ctx, txRaw)
	}
}

func (ma *MongoAttendant) serveIxRaw(ctx context.Context) {
	for ixRaw := range ma.ixRawChan {
		ma.client.Database(Database).Collection(CollectionIxRaw).InsertOne(ctx, ixRaw)
	}
}

func (ma *MongoAttendant) serveIxIndex(ctx context.Context) {
	for ixIndex := range ma.ixIndexCh {
		ma.client.Database(Database).Collection(CollectionIxIndex).InsertOne(ctx, ixIndex)
	}
}

func (ma *MongoAttendant) serveIx(ctx context.Context) {
	for ix := range ma.ixCh {
		ma.client.Database(Database).Collection(CollectionIx).InsertOne(ctx, ix)
	}
}
