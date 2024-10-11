package main

import (
	"context"
	"fmt"
	"sync"
	"time"

	"github.com/blocto/solana-go-sdk/rpc"

	"solana-openbook-scanner/block_height_manager"
)

type BlockGetter struct {
	workerCnt int
	cli       rpc.RpcClient
	bhm       block_height_manager.BlockHeightManager
}

const (
	SolanaRpcEndpoint = "https://api.mainnet-beta.solana.com"
	Delay             = time.Millisecond * 10000
)

var (
	transactionVersion = uint8(0)
	rewards            = false
	getBlockConfig     = rpc.GetBlockConfig{
		Encoding:                       rpc.GetBlockConfigEncodingJsonParsed,
		TransactionDetails:             rpc.GetBlockConfigTransactionDetailsFull,
		Rewards:                        &rewards,
		Commitment:                     rpc.CommitmentFinalized,
		MaxSupportedTransactionVersion: &transactionVersion,
	}
)

func NewBlockGetter(workerCnt int, bhm block_height_manager.BlockHeightManager) *BlockGetter {
	cli := rpc.NewRpcClient(SolanaRpcEndpoint)
	return &BlockGetter{workerCnt: workerCnt, cli: cli, bhm: bhm}
}

func (bg *BlockGetter) getBlockHeightBySlot(slot uint64) int64 {
	ctx := context.Background()
	resp, err := bg.cli.GetBlockWithConfig(ctx, slot, getBlockConfig)
	if err != nil {
		Logger.Fatal(err.Error())
	}

	if resp.Error != nil {
		Logger.Fatal(resp.Error.Error())
	}

	return *resp.Result.BlockHeight
}

func (bg *BlockGetter) run(ctx context.Context, wg *sync.WaitGroup, taskCh chan uint64, blockCh chan *rpc.GetBlock) {
	defer wg.Done()
	defer close(blockCh)

	var wgWorker sync.WaitGroup
	wgWorker.Add(bg.workerCnt)

	worker := func(id int) {
		defer wgWorker.Done()

		for {
			select {
			case <-ctx.Done():
				return

			case slot := <-taskCh:
				if slot == 0 {
					Logger.Info(fmt.Sprintf("worker:%d all task finish", id))
					return
				}

				Logger.Info(fmt.Sprintf("GetBlock:%d start", slot))
				failCnt := 0
				for {
					time.Sleep(Delay)

					resp, err := bg.cli.GetBlockWithConfig(ctx, slot, getBlockConfig)
					if err != nil {
						failCnt += 1
						Logger.Info(fmt.Sprintf("get slot:%d failed:%d", slot, failCnt))
						continue
					}

					if resp.Error != nil {
						Logger.Error(fmt.Sprintf("TODO check! get slot:%d JsonRpc err:%s", slot, resp.Error.Error()))
						// ignore this slot: TODO check
						break
					}

					Logger.Info(fmt.Sprintf("get slot:%d succeed", slot))

					for {
						Logger.Debug(fmt.Sprintf("current height:%d, ParentSlot:%d, height:%d", bg.bhm.Get(), resp.Result.ParentSlot, *resp.Result.BlockHeight))
						if bg.bhm.CanCommit(*resp.Result.BlockHeight) {
							blockCh <- resp.Result
							bg.bhm.Commit(*resp.Result.BlockHeight)
							break
						}
						time.Sleep(time.Millisecond * 100)
					}

					break
				}
			}
		}
	}

	for i := 0; i < bg.workerCnt; i++ {
		go worker(i)
	}

	Logger.Info("wait all workers done")
	wgWorker.Wait()
	Logger.Info("all workers done")
}
