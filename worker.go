package main

import (
	"context"
	"encoding/json"
	"fmt"
	"github.com/blocto/solana-go-sdk/rpc"
	"github.com/mr-tron/base58"
	"os"
	"solana-openbook-scanner/openbook_v2"

	ag_solanago "github.com/gagliardetto/solana-go"
)

type Worker struct {
	cli rpc.RpcClient
}

func NewWorker(address string) *Worker {
	return &Worker{
		cli: rpc.NewRpcClient(address),
	}
}

func (worker *Worker) getBlock(slot uint64) {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	var getBlockConfig rpc.GetBlockConfig
	version := uint8(0)
	getBlockConfig.Encoding = rpc.GetBlockConfigEncodingJsonParsed
	getBlockConfig.Commitment = rpc.CommitmentFinalized
	getBlockConfig.MaxSupportedTransactionVersion = &version
	getBlockConfig.TransactionDetails = rpc.GetBlockConfigTransactionDetailsFull
	resp, err := worker.cli.GetBlockWithConfig(ctx, slot, getBlockConfig)
	if err != nil {
		Logger.Error(fmt.Sprintf("get block with slot:%d err:%s", slot, err.Error()))
	}

	if resp.Error != nil {
		Logger.Error(fmt.Sprintf("get block with slot:%d err:%s", slot, resp.Error.Error()))
	}

	blockData, err := json.Marshal(resp.Result)
	if err != nil {
		Logger.Error(fmt.Sprintf("json marshal err:%s", err.Error()))
	}
	f, err := os.Create(fmt.Sprintf("block%d.json", slot))
	if err != nil {
		Logger.Error(fmt.Sprintf("create file err:%s", err.Error()))
	}
	f.Write(blockData)
	f.Close()

	var ixF rpc.InstructionFull
	for _, tx := range resp.Result.Transactions {
		if tx.Meta.Err != nil {
			Logger.Info(fmt.Sprintf("%s failed\n", tx.Transaction.Signatures[0]))
			continue
		}

		var sb map[string]rpc.AccountKey = make(map[string]rpc.AccountKey)
		for _, ak := range tx.Transaction.Message.AccountKeys {
			sb[ak.Pubkey] = ak
		}
		for _, ix := range tx.Transaction.Message.Instructions {
			jd, err := json.Marshal(ix)
			if err != nil {
				Logger.Error(fmt.Sprintf("json marshal err:%s", err.Error()))
			}

			json.Unmarshal(jd, &ixF)
			if ixF.ProgramId == "opnb2LAfJYbRMAHHvqjCwQxanZn7ReEHp1k81EohpZb" {
				jd, err = json.Marshal(tx)
				Logger.Info(string(jd))

				var accounts []*ag_solanago.AccountMeta
				for _, x1 := range ixF.Accounts {
					accounts = append(accounts, &ag_solanago.AccountMeta{
						PublicKey:  ag_solanago.MustPublicKeyFromBase58(x1),
						IsWritable: sb[x1].Writable,
						IsSigner:   sb[x1].Signer,
					})
				}
				d, err := base58.Decode(ixF.Data)
				if err != nil {
					Logger.Error(fmt.Sprintf("decode err:%s", err.Error()))
				}
				ins, err := openbook_v2.DecodeInstruction(accounts, d)
				if err != nil {
					Logger.Error(fmt.Sprintf("decode instruction err:%s", err.Error()))
				}
				Logger.Info(fmt.Sprintf("%v", ins))
			}

		}
	}
	Logger.Info("i")
}
