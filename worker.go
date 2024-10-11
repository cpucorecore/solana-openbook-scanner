package main

import (
	"context"
	"encoding/json"
	"fmt"
	"github.com/blocto/solana-go-sdk/rpc"
	"github.com/mr-tron/base58"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/v2/mongo"
	"os"
	"solana-openbook-scanner/openbook_v2"

	ag_solanago "github.com/gagliardetto/solana-go"
)

type Worker struct {
	cli    rpc.RpcClient
	client *mongo.Client
}

func NewWorker(address string, client *mongo.Client) *Worker {
	return &Worker{
		cli:    rpc.NewRpcClient(address),
		client: client,
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
			//Logger.Info(fmt.Sprintf("%s failed\n", tx.Transaction.Signatures[0]))
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

				worker.client.Database("openbook_v2").Collection("tx").InsertOne(ctx, tx)

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
				Logger.Info(fmt.Sprintf("%s: %v", openbook_v2.InstructionIDToName(ins.TypeID), ins))

				r, err := worker.client.Database("openbook_v2").Collection("ix_index").InsertOne(ctx, bson.M{"ins": openbook_v2.InstructionIDToName(ins.TypeID), "signature": tx.Transaction.Signatures})
				if err != nil {
					Logger.Error(fmt.Sprintf("insert err:%s", err.Error()))
				}
				Logger.Info(fmt.Sprintf("%v", r))

				switch ins.TypeID {
				//case openbook_v2.Instruction_CreateMarket:
				//case openbook_v2.Instruction_CloseMarket:
				//case openbook_v2.Instruction_CreateOpenOrdersIndexer:
				//case openbook_v2.Instruction_CloseOpenOrdersIndexer:
				//case openbook_v2.Instruction_CreateOpenOrdersAccount:
				//case openbook_v2.Instruction_CloseOpenOrdersAccount:
				//case openbook_v2.Instruction_PlaceOrder:
				//case openbook_v2.Instruction_EditOrder:
				//case openbook_v2.Instruction_EditOrderPegged:
				//case openbook_v2.Instruction_PlaceOrders:
				case openbook_v2.Instruction_CancelAllAndPlaceOrders:
					x, ok := ins.Impl.(*openbook_v2.CancelAllAndPlaceOrders)
					if ok {
						Logger.Info(fmt.Sprintf("%v", x))
						r, err = worker.client.Database("openbook_v2").Collection("ix_can").InsertOne(ctx, *x)
						if err != nil {
							Logger.Error(fmt.Sprintf("insert err:%s", err.Error()))
						}
					} else {
						Logger.Info(fmt.Sprintf("%v", ins.TypeID))
					}
					break
				//case openbook_v2.Instruction_PlaceOrderPegged:
				//case openbook_v2.Instruction_PlaceTakeOrder:
				//case openbook_v2.Instruction_ConsumeEvents:
				//case openbook_v2.Instruction_ConsumeGivenEvents:
				//case openbook_v2.Instruction_CancelOrder:
				//case openbook_v2.Instruction_CancelOrderByClientOrderId:
				//case openbook_v2.Instruction_CancelAllOrders:
				//case openbook_v2.Instruction_Deposit:
				//case openbook_v2.Instruction_Refill:
				case openbook_v2.Instruction_SettleFunds:
					x, ok := ins.Impl.(*openbook_v2.SettleFunds)
					if ok {
						Logger.Info(fmt.Sprintf("%v", x))
						r, err = worker.client.Database("openbook_v2").Collection("ix_sett").InsertOne(ctx, *x)
						if err != nil {
							Logger.Error(fmt.Sprintf("insert err:%s", err.Error()))
						}
					} else {
						Logger.Info(fmt.Sprintf("%v", ins.TypeID))
					}
					break
				//case openbook_v2.Instruction_SettleFundsExpired:
				//case openbook_v2.Instruction_SweepFees:
				//case openbook_v2.Instruction_SetDelegate:
				//case openbook_v2.Instruction_SetMarketExpired:
				//case openbook_v2.Instruction_PruneOrders:
				//case openbook_v2.Instruction_StubOracleCreate:
				//case openbook_v2.Instruction_StubOracleClose:
				//case openbook_v2.Instruction_StubOracleSet:
				default:
					break
				}
			}
		}
	}
}
