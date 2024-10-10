package main

import (
	"fmt"
	"github.com/gagliardetto/solana-go/rpc"
	"solana-openbook-scanner/finished_block_manager"
	"solana-openbook-scanner/log"
)

func main() {
	startBlock := uint64(294577906)
	finished_block_manager.Setup(startBlock)

	taskCh := make(chan uint64, 10000)
	go SOLDispatchTasksMock(startBlock, taskCh)

	blockCh := make(chan *rpc.GetBlockResult, 1000)
	for workerId := 0; workerId < 1; workerId++ {
		go SOLSyncBlocks(workerId, taskCh, blockCh)
	}

	for b := range blockCh {
		curSlot := b.ParentSlot + 1
		log.Logger.Info(fmt.Sprintf("slot:%d with %d txs begin", curSlot, len(b.Transactions)))

		for txIdx, txWithMeta := range b.Transactions {
			ParseTx(*b.BlockHeight, txIdx, &txWithMeta)
		}

		log.Logger.Info(fmt.Sprintf("block:%d all operations commit to queue", curSlot))
		finished_block_manager.Update(curSlot)
	}
}
