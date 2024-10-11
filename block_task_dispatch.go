package main

import (
	"sync"

	"github.com/blocto/solana-go-sdk/rpc"
)

type BlockTaskDispatch struct {
	cli rpc.RpcClient
}

func NewBlockTaskDispatch() *BlockTaskDispatch {
	cli := rpc.NewRpcClient(SolanaRpcEndpoint)
	return &BlockTaskDispatch{cli: cli}
}

func (btd *BlockTaskDispatch) keepDispatchTaskMock(wg *sync.WaitGroup, startSlot uint64, count uint64, taskCh chan uint64) {
	defer wg.Done()

	start := startSlot
	end := startSlot + count
	for start < end {
		taskCh <- start
		start += 1
	}

	close(taskCh)
}
