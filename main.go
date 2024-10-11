package main

import (
	"go.mongodb.org/mongo-driver/v2/mongo"
	"go.mongodb.org/mongo-driver/v2/mongo/options"
	"log"
)

func main() {
	startSlot := uint64(294577906)

	client, err := mongo.Connect(options.Client().ApplyURI("mongodb://localhost:27017"))
	if err != nil {
		log.Fatal(err)
	}
	rpcEndpoint := "https://api.mainnet-beta.solana.com"
	worker := NewWorker(rpcEndpoint, client)
	worker.getBlock(startSlot)
}
