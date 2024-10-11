package main

import (
	"context"
	"fmt"
	"log"
	"sync"

	"go.mongodb.org/mongo-driver/v2/bson"
	"go.mongodb.org/mongo-driver/v2/mongo"
	"go.mongodb.org/mongo-driver/v2/mongo/options"
)

type MongoAttendant struct {
	cli       *mongo.Client
	txRawCh   chan string
	ixRawCh   chan string
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
		cli:       client,
		txRawCh:   txRawChan,
		ixRawCh:   ixRawChan,
		ixIndexCh: ixIndexCh,
		ixCh:      ixCh,
	}
}

func (ma *MongoAttendant) startServe(ctx context.Context, wg *sync.WaitGroup) {
	wg.Add(1)
	go ma.serveTxRaw(ctx, wg)

	wg.Add(1)
	go ma.serveIxRaw(ctx, wg)

	wg.Add(1)
	go ma.serveIxIndex(ctx, wg)

	wg.Add(1)
	go ma.serveIx(ctx, wg)
}

func (ma *MongoAttendant) serveTxRaw(ctx context.Context, wg *sync.WaitGroup) {
	defer wg.Done()

	for {
		select {
		case <-ctx.Done():
			return

		case txRaw := <-ma.txRawCh:
			if txRaw == "" {
				Logger.Info("txRawCh @ done")
				return
			}

			_, err := ma.cli.Database(Database).Collection(CollectionTxRaw).InsertOne(ctx, bson.M{"d": txRaw})
			if err != nil {
				Logger.Error(fmt.Sprintf("insert tx raw err:%s", err.Error()))
			}
		}
	}
}

func (ma *MongoAttendant) serveIxRaw(ctx context.Context, wg *sync.WaitGroup) {
	defer wg.Done()

	for {
		select {
		case <-ctx.Done():
			return

		case ixRaw := <-ma.ixRawCh:
			if ixRaw == "" {
				Logger.Info("ixRawCh @ done")
				return
			}

			_, err := ma.cli.Database(Database).Collection(CollectionIxRaw).InsertOne(ctx, bson.M{"d": ixRaw})
			if err != nil {
				Logger.Error(fmt.Sprintf("insert ix raw err:%s", err.Error()))
			}
		}
	}
}

func (ma *MongoAttendant) serveIxIndex(ctx context.Context, wg *sync.WaitGroup) {
	defer wg.Done()

	for {
		select {
		case <-ctx.Done():
			return

		case ixIndex := <-ma.ixIndexCh:
			if ixIndex == nil {
				Logger.Info("ixIndexCh @ done")
				return
			}

			_, err := ma.cli.Database(Database).Collection(CollectionIxIndex).InsertOne(ctx, ixIndex)
			if err != nil {
				Logger.Error(fmt.Sprintf("insert ix index err:%s", err.Error()))
			}
		}
	}
}

func (ma *MongoAttendant) serveIx(ctx context.Context, wg *sync.WaitGroup) {
	defer wg.Done()

	for {
		select {
		case <-ctx.Done():
			return

		case ix := <-ma.ixCh:
			if ix == nil {
				Logger.Info("ixCh @ done")
				return
			}

			_, err := ma.cli.Database(Database).Collection(CollectionIx).InsertOne(ctx, ix)
			if err != nil {
				Logger.Error(fmt.Sprintf("insert ix err:%s", err.Error()))
			}
		}
	}
}
