package main

import (
	"context"
	"sync"

	"github.com/blocto/solana-go-sdk/rpc"
	"go.mongodb.org/mongo-driver/v2/bson"

	"solana-openbook-scanner/block_slot_manager"
)

const (
	ConfigBlockGetterWorkerCnt = 3
	ConfigStartHeight          = uint64(294577906)
)

func main() {
	bhm := block_slot_manager.NewBlockSlotManager(ConfigStartHeight)

	taskCh := make(chan uint64, 1000)
	blockCh := make(chan *rpc.GetBlock, 100)
	txRawCh := make(chan string, 100)
	ixRawCh := make(chan string, 100)
	ixIndexCh := make(chan bson.M, 100)
	ixCh := make(chan bson.M, 100)

	ctx := context.Background()

	var wg sync.WaitGroup

	mga := NewMongoAttendant(txRawCh, ixRawCh, ixIndexCh, ixCh)
	mga.startServe(ctx, &wg)

	blockProcessor := NewBlockProcessorAdmin(blockCh, txRawCh, ixRawCh, ixIndexCh, ixCh)
	wg.Add(1)
	go blockProcessor.run(ctx, &wg)

	blockGetter := NewBlockGetter(ConfigBlockGetterWorkerCnt, bhm)
	wg.Add(1)
	go blockGetter.run(ctx, &wg, taskCh, blockCh)

	taskDispatch := NewBlockTaskDispatch()
	wg.Add(1)
	go taskDispatch.keepDispatchTaskMock(&wg, ConfigStartHeight, 5, taskCh)

	Logger.Info("wait all goroutine done")
	wg.Wait()
	Logger.Info("all goroutine done")
}
