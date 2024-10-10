package main

import (
	"context"
	"encoding/binary"
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"time"

	"github.com/gagliardetto/solana-go"
	"github.com/gagliardetto/solana-go/rpc"
	"github.com/gagliardetto/solana-go/rpc/jsonrpc"

	"solana-openbook-scanner/common"
	"solana-openbook-scanner/finished_block_manager"
	"solana-openbook-scanner/log"
)

const (
	NativeDenom               = "SOL"
	SystemInstructionTransfer = 2
)

var (
	memoProgramId           = solana.MustPublicKeyFromBase58("opnb2LAfJYbRMAHHvqjCwQxanZn7ReEHp1k81EohpZb")
	systemTransferProgramId = solana.MustPublicKeyFromBase58("11111111111111111111111111111111")
)

func SOLDispatchTasksMock(startHeight uint64, taskCh chan uint64) {
	taskCh <- startHeight
}

func SOLDispatchTasks(startHeight uint64, taskCh chan uint64) {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	endpoint := rpc.LocalNet_RPC
	cli := rpc.New(endpoint)

	taskCh <- startHeight
	cursor := startHeight

	var errCnt int
	var start time.Time
	var duration time.Duration
	for {
		errCnt = 0
		start = time.Now()
		latestBlockHeight, err := cli.GetBlockHeight(ctx, rpc.CommitmentFinalized)
		duration = time.Now().Sub(start)
		if err != nil {
			errCnt++
			log.Logger.Warn(fmt.Sprintf("sol GetBlockHeight failed %d times with err %s, time elapse ms %d", errCnt, err.Error(), duration.Milliseconds()))
			time.Sleep(time.Second * 3)
			continue
		}
		log.Logger.Info(fmt.Sprintf("sol GetBlockHeight success with retry count %d, time elapse ms %d", errCnt, duration.Milliseconds()))

		if latestBlockHeight <= cursor {
			log.Logger.Warn(fmt.Sprintf("sol GetBlockHeight remote height %d <= local height %d", latestBlockHeight, cursor))
			time.Sleep(time.Second * 1)
			continue
		}
		log.Logger.Info(fmt.Sprintf("sol GetBlockHeight: cursor %d, remote height %d", cursor, latestBlockHeight))

		for height := cursor + 1; height <= latestBlockHeight; height++ {
			taskCh <- height
		}

		cursor = latestBlockHeight
	}
}

func SOLSyncBlocks(workerId int, taskCh chan uint64, blockCh chan *rpc.GetBlockResult) {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	endpoint := rpc.MainNetBeta_RPC
	cli := rpc.New(endpoint)

	workerBufferCh := make(chan *rpc.GetBlockResult, 50)
	go func() {
		for b := range workerBufferCh {
			taskCoordinate := fmt.Sprintf("(workerId%d, task%d) workerBuffer length:%d", workerId, *b.BlockHeight, len(workerBufferCh))
			var waitCnt int64 = 0
			for { // wait the worker's turn to commit finished work
				fh := finished_block_manager.Get()
				if b.ParentSlot == fh {
					log.Logger.Info(fmt.Sprintf("task %s committed with wait time ms:%v", taskCoordinate, waitCnt*time.Millisecond.Milliseconds()))
					break
				}
				time.Sleep(time.Millisecond * 5)
				waitCnt++
				if waitCnt%1000 == 0 {
					log.Logger.Info(fmt.Sprintf("task %s wait with time ms:%v", taskCoordinate, waitCnt*time.Millisecond.Milliseconds()))
				}
			}
			blockCh <- b
		}
	}()

	var start, end time.Time
	for task := range taskCh {
		taskCoordinate := fmt.Sprintf("(workerId%d, task%d)", workerId, task)
		log.Logger.Info(fmt.Sprintf("task %s begin", taskCoordinate))

		includeRewards := false

		getBlockFailedCnt := 0
		getBlockNilCnt := 0
		MaxSupportedTransactionVersion := uint64(0)
		for {
			start = time.Now()
			b, err := cli.GetBlockWithOpts(ctx, task, &rpc.GetBlockOpts{
				Encoding:                       solana.EncodingJSONParsed,
				Commitment:                     rpc.CommitmentConfirmed,
				TransactionDetails:             rpc.TransactionDetailsFull,
				Rewards:                        &includeRewards,
				MaxSupportedTransactionVersion: &MaxSupportedTransactionVersion,
			})
			end = time.Now()
			durationMs := end.Sub(start).Milliseconds()
			if err != nil {
				var rpcError *jsonrpc.RPCError
				if errors.As(err, &rpcError) {
					if rpcError.Code == -32007 {
						log.Logger.Warn(fmt.Sprintf("slot %d skipped, we skipped also, err: %v", task, rpcError))
						break
					}
				}
				getBlockFailedCnt++
				log.Logger.Warn(fmt.Sprintf("task %s do 'GetBlock' failed %d times with err: [%v], elapse ms:%v", taskCoordinate, getBlockFailedCnt, err, durationMs))
				time.Sleep(time.Second * 5)
				continue
			}
			if b == nil {
				getBlockNilCnt++
				log.Logger.Warn(fmt.Sprintf("task %s do 'GetBlock' returns nil %d times ,elapse ms:%v", taskCoordinate, getBlockNilCnt, durationMs))
				time.Sleep(time.Second * 5)
				continue
			}
			log.Logger.Info(fmt.Sprintf("task %s do 'GetBlock' succeed with failed count %d with returns nil count %d, elapse ms:%v", taskCoordinate, getBlockFailedCnt, getBlockNilCnt, durationMs))

			workerBufferCh <- b
			log.Logger.Info(fmt.Sprintf("task %s commit to worker buffer", taskCoordinate))
			break
		}
	}
}

func isTheProgramId(expected, actual solana.PublicKey) bool {
	return expected.Equals(actual)
}

func isSystemTransferProgramId(id solana.PublicKey) bool {
	return isTheProgramId(systemTransferProgramId, id)
}

func isMemoProgramId(id solana.PublicKey) bool {
	log.Logger.Info(fmt.Sprintf("target:%s, actual:%s\n", memoProgramId.String(), id.String()))
	return isTheProgramId(memoProgramId, id)
}

func findInstructionIndexes(publicKeys []solana.PublicKey, instructions []solana.CompiledInstruction, filter func(solana.PublicKey) bool) (indexes []int) {
	for index, inst := range instructions {
		if filter(publicKeys[inst.ProgramIDIndex]) {
			indexes = append(indexes, index)
		}
	}
	return
}

func parseSystemInstructionCallData(callData []byte) (sysInst uint32, value uint64, err error) {
	switch len(callData) {
	case 12:
		sysInst = binary.LittleEndian.Uint32(callData[:4])
		value = binary.LittleEndian.Uint64(callData[4:])
	case 9:
		sysInst = uint32(callData[0])
		value = binary.LittleEndian.Uint64(callData[1:])
	default:
		err = errors.New(fmt.Sprintf("wrong system transfer call data len:%d", len(callData)))
	}
	return
}

func ParseTx(blockHeight uint64, txIdx int, txWithMeta *rpc.TransactionWithMeta) {
	tx, err := txWithMeta.GetTransaction()
	if err != nil {
		// TODO FIXME
		// TODO release resources
		os.Exit(-1)
	}

	//if len(tx.Signatures) != 1 {
	//	// TODO FIXME: until now we don't known how to process a transaction with multiple signatures
	//	err = errors.New(fmt.Sprintf("tx signatures len %d != 1, ignore", len(tx.Signatures)))
	//	return
	//}

	txCoordinate := common.TxCoordinate(blockHeight, txIdx, tx.Signatures[0].String())
	log.Logger.Info(fmt.Sprintf("--%s begin", txCoordinate))

	if txWithMeta.Meta.Err != nil {
		err = errors.New(fmt.Sprintf("tx not success on chain with err: %v", txWithMeta.Meta.Err))
		return
	}

	memoProgramInstructionIndexes := findInstructionIndexes(tx.Message.AccountKeys, tx.Message.Instructions, isMemoProgramId)
	if len(memoProgramInstructionIndexes) == 0 {
		err = errors.New(fmt.Sprintf("no memo instruction"))
		return
	}

	log.Logger.Info(fmt.Sprintf("%v", tx.String()))
	txJson, _ := json.Marshal(tx)
	log.Logger.Info(fmt.Sprintf("tx=%v", string(txJson)))

	//memo := string(tx.Message.Instructions[memoProgramInstructionIndexes[0]].Data)
	//
	//log.Logger.Info(memo)
}
