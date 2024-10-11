package main

import (
	"context"
	"github.com/blocto/solana-go-sdk/rpc"
	"go.mongodb.org/mongo-driver/v2/bson"
)

func main() {
	startHeight := uint64(294577906) //294577923
	//bhm := block_height_manager.NewBlockHeightManager(startHeight)

	taskCh := make(chan uint64, 1000)
	blockCh := make(chan *rpc.GetBlock, 100)
	txRawCh := make(chan string, 100)
	ixRawCh := make(chan string, 100)
	ixIndexCh := make(chan bson.M, 100)
	ixCh := make(chan bson.M, 100)

	ctx := context.Background()

	attendant := NewMongoAttendant(txRawCh, ixRawCh, ixIndexCh, ixCh)
	go attendant.serve(ctx)

	blockProcessor := NewBlockProcessorAdmin(blockCh, txRawCh, ixRawCh, ixIndexCh, ixCh)
	go blockProcessor.run(ctx)

	blockGetter := NewBlockGetter()
	go blockGetter.keepGenerateTaskMock(startHeight, 10, taskCh)
	blockGetter.run(ctx, taskCh, blockCh)
}
