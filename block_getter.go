package main

import (
	"context"
	"fmt"
	"sync"
	"time"

	"github.com/blocto/solana-go-sdk/rpc"

	"solana-openbook-scanner/block_slot_manager"
)

type BlockGetter struct {
	workerCnt int
	cli       rpc.RpcClient
	bhm       block_slot_manager.BlockSlotManager
}

const (
	SolanaRpcEndpoint = "https://api.mainnet-beta.solana.com"
	Delay             = time.Millisecond * 1000
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

func NewBlockGetter(workerCnt int, bhm block_slot_manager.BlockSlotManager) *BlockGetter {
	cli := rpc.NewRpcClient(SolanaRpcEndpoint)
	return &BlockGetter{workerCnt: workerCnt, cli: cli, bhm: bhm}
}

func (bg *BlockGetter) GetBlock(slot uint64) (b *rpc.GetBlock, err error) {
	ctx := context.Background()
	resp, err := bg.cli.GetBlockWithConfig(ctx, slot, getBlockConfig)
	if err != nil {
		Logger.Error(fmt.Sprintf("get block slot:%d err:%s", slot, err.Error()))
		return nil, err
	}

	if resp.Error != nil {
		Logger.Error(fmt.Sprintf("get block slot:%d JsonRpc err:%s", slot, resp.Error.Error()))
		return nil, resp.Error
	}

	return resp.Result, nil
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

			case height := <-taskCh:
				if height == 0 {
					Logger.Info(fmt.Sprintf("worker:%d all task finish", id))
					return
				}

				Logger.Info(fmt.Sprintf("GetBlock:%d start", height))
				failCnt := 0
				for {
					time.Sleep(Delay)
					b, err := bg.GetBlock(height)
					if err != nil {
						failCnt += 1
						Logger.Info(fmt.Sprintf("GetBlock:%d failed:%d", height, failCnt))
						continue
					}

					for {
						Logger.Debug(fmt.Sprintf("ParentSlot:%d, height:%d", b.ParentSlot, *b.BlockHeight))
						if bg.bhm.CanCommit(b.ParentSlot) {
							blockCh <- b
							bg.bhm.Commit(uint64(b.ParentSlot + 1))
							break
						}
						time.Sleep(time.Millisecond * 100)
					}
					break
				}
				Logger.Info(fmt.Sprintf("GetBlock:%d succeed", height))
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
