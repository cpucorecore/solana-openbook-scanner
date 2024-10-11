package main

import (
	"context"
	"fmt"
	"github.com/blocto/solana-go-sdk/rpc"
	"time"
)

type BlockGetter struct {
	cli rpc.RpcClient
}

const (
	rpcEndpoint = "https://api.mainnet-beta.solana.com"
	Delay       = time.Millisecond * 1000
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

func NewBlockGetter() *BlockGetter {
	cli := rpc.NewRpcClient(rpcEndpoint)
	return &BlockGetter{cli: cli}
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

func (bg *BlockGetter) run(ctx context.Context, taskCh chan uint64, blockCh chan *rpc.GetBlock) {
	for {
		select {
		case <-ctx.Done():
			return

		default:
			for height := range taskCh {
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
					blockCh <- b
					break
				}
				Logger.Info(fmt.Sprintf("GetBlock:%d succeed", height))
			}
		}
	}
}

func (bg *BlockGetter) keepGenerateTaskMock(startHeight uint64, count uint64, taskCh chan uint64) {
	start := startHeight
	end := startHeight + count
	for start < end {
		taskCh <- start
		start += 1
	}
	close(taskCh)
}
