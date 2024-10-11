package main

import (
	"context"
	"sync"

	"github.com/blocto/solana-go-sdk/rpc"
	"go.mongodb.org/mongo-driver/v2/bson"

	"solana-openbook-scanner/block_height_manager"
)

const (
	ConfigBlockGetterWorkerCnt = 3
	ConfigStartSlot            = uint64(294577906) //294577906
)

func main() {
	taskCh := make(chan uint64, 1000)

	blockCh := make(chan *rpc.GetBlock, 100)

	txRawCh := make(chan string, 100)
	ixRawCh := make(chan string, 100)
	ixIndexCh := make(chan bson.M, 100)
	ixCh := make(chan bson.M, 100)

	ctx := context.Background()
	var wg sync.WaitGroup

	mongo := NewMongoAttendant(txRawCh, ixRawCh, ixIndexCh, ixCh)
	mongo.startServe(ctx, &wg)

	blockProcessor := NewBlockProcessorAdmin(blockCh, txRawCh, ixRawCh, ixIndexCh, ixCh)
	wg.Add(1)
	go blockProcessor.run(ctx, &wg)

	bhm := block_height_manager.NewBlockHeightManager()
	blockGetter := NewBlockGetter(ConfigBlockGetterWorkerCnt, bhm)
	startBlockHeight := blockGetter.getBlockHeightBySlot(ConfigStartSlot)
	bhm.Init(startBlockHeight - 1)

	wg.Add(1)
	go blockGetter.run(ctx, &wg, taskCh, blockCh)

	taskDispatch := NewBlockTaskDispatch()
	wg.Add(1)
	go taskDispatch.keepDispatchTaskMock(&wg, ConfigStartSlot, 2000, taskCh)

	Logger.Info("wait all goroutine done")
	wg.Wait()
	Logger.Info("all goroutine done")
}
