package main

import (
	"context"
	"encoding/json"
	"fmt"
	"os"

	"github.com/blocto/solana-go-sdk/rpc"
	ag_solanago "github.com/gagliardetto/solana-go"
	"github.com/mr-tron/base58"
	"go.mongodb.org/mongo-driver/v2/bson"
	"solana-openbook-scanner/openbook_v2"
)

type BlockProcessor interface {
	name() string
	process(*rpc.GetBlock) error
	done()
}

type BlockProcessorAdmin interface {
	run(ctx context.Context)
}

type blockProcessorAdmin struct {
	blockCh    chan *rpc.GetBlock
	processors []BlockProcessor
}

func (b *blockProcessorAdmin) run(ctx context.Context) {
	for {
		select {
		case <-ctx.Done():
			return

		default:
			for block := range b.blockCh {
				for _, processor := range b.processors {
					err := processor.process(block)
					if err != nil {
						Logger.Error(fmt.Sprintf("%s process block:%d failed", processor.name(), block.BlockHeight))
						// TODO exit
					}
				}
			}
		}
	}

}

func NewBlockProcessorAdmin(blockCh chan *rpc.GetBlock, txRawChan chan string, ixRawChan chan string, ixIndexCh chan bson.M, ixCh chan bson.M) BlockProcessorAdmin {
	bpf, err := newBlockProcessorFile()
	if err != nil {
		Logger.Fatal(fmt.Sprintf("newBlockProcessorFile err:%v", err))
	}

	bpp, err := newBlockProcessorParser(txRawChan, ixRawChan, ixIndexCh, ixCh)
	if err != nil {
		Logger.Fatal(fmt.Sprintf("newBlockProcessorParser err:%v", err))
	}

	processors := []BlockProcessor{
		bpf,
		bpp,
	}

	return &blockProcessorAdmin{
		blockCh:    blockCh,
		processors: processors,
	}
}

type BlockProcessorFile struct {
	fd *os.File
}

var _ BlockProcessor = &BlockProcessorFile{}

const (
	BlocksFilePath = "blocks.json"
)

func newBlockProcessorFile() (bpf *BlockProcessorFile, err error) {
	f, err := os.Create(BlocksFilePath)
	if err != nil {
		Logger.Error(fmt.Sprintf("create file err:%s", err.Error()))
		return nil, err
	}
	return &BlockProcessorFile{
		fd: f,
	}, nil
}

func (bpf *BlockProcessorFile) name() string {
	return "file"
}

func (bpf *BlockProcessorFile) process(block *rpc.GetBlock) error {
	blockData, err := json.Marshal(block)
	if err != nil {
		Logger.Error(fmt.Sprintf(""))
		return err
	}

	_, err = bpf.fd.Write(blockData)
	return err
}

func (bpf *BlockProcessorFile) done() {
	bpf.fd.Close()
}

type BlockProcessorParser struct {
	txRawChan chan string
	ixRawChan chan string
	ixIndexCh chan bson.M
	ixCh      chan bson.M
}

func newBlockProcessorParser(txRawChan chan string, ixRawChan chan string, ixIndexCh chan bson.M, ixCh chan bson.M) (bpp *BlockProcessorParser, err error) {
	return &BlockProcessorParser{
		txRawChan: txRawChan,
		ixRawChan: ixRawChan,
		ixIndexCh: ixIndexCh,
		ixCh:      ixCh,
	}, nil
}

func (bpp *BlockProcessorParser) name() string {
	return "parser"
}

const (
	OpenbookV2AddressMainnet = "opnb2LAfJYbRMAHHvqjCwQxanZn7ReEHp1k81EohpZb"
)

func (bpp *BlockProcessorParser) process(block *rpc.GetBlock) error {
	for _, tx := range block.Transactions {
		if tx.Meta.Err != nil {
			Logger.Info(fmt.Sprintf("skip failed tx Signatures:%v", tx.Transaction.Signatures))
			continue
		}

		accountKeys := make(map[string]rpc.AccountKey)
		for _, accountKey := range tx.Transaction.Message.AccountKeys {
			accountKeys[accountKey.Pubkey] = accountKey
		}

		var ixF rpc.InstructionFull
		for _, instruction := range tx.Transaction.Message.Instructions {
			ixJson, err := json.Marshal(instruction)
			if err != nil {
				Logger.Error(fmt.Sprintf("json marshal err:%s", err.Error()))
				// TODO exit
			}

			err = json.Unmarshal(ixJson, &ixF)
			if err != nil {
				Logger.Error(fmt.Sprintf("json unmarshal err:%s", err.Error()))
				// TODO exit
			}

			if ixF.ProgramId == OpenbookV2AddressMainnet {
				txJson, err := json.Marshal(tx)
				if err != nil {
					Logger.Error(fmt.Sprintf("json unmarshal err:%s", err.Error()))
					// TODO exit
				}
				bpp.txRawChan <- string(txJson)

				var accounts []*ag_solanago.AccountMeta
				for _, account := range ixF.Accounts {
					accounts = append(accounts, &ag_solanago.AccountMeta{
						PublicKey:  ag_solanago.MustPublicKeyFromBase58(account),
						IsWritable: accountKeys[account].Writable,
						IsSigner:   accountKeys[account].Signer,
					})
				}

				ixData, err := base58.Decode(ixF.Data)
				if err != nil {
					Logger.Error(fmt.Sprintf("decode err:%s", err.Error()))
				}

				ix, err := openbook_v2.DecodeInstruction(accounts, ixData)
				if err != nil {
					Logger.Error(fmt.Sprintf("decode instruction err:%s", err.Error()))
				}

				bpp.ixIndexCh <- bson.M{"ins": openbook_v2.InstructionIDToName(ix.TypeID), "signature": tx.Transaction.Signatures}

				switch ix.TypeID {
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
					ixNude, ok := ix.Impl.(*openbook_v2.CancelAllAndPlaceOrders)
					if ok {
						ixJson, err = json.Marshal(ixNude)
						if err != nil {
							Logger.Error(fmt.Sprintf("json marshal err:%s", err.Error()))
						}

						bpp.ixCh <- bson.M{"data": string(ixJson)}
					} else {
						Logger.Error(fmt.Sprintf("instruction:%s assertion failed", openbook_v2.InstructionIDToName(ix.TypeID)))
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
					ixNude, ok := ix.Impl.(*openbook_v2.SettleFunds)
					if ok {
						ixJson, err = json.Marshal(ixNude)
						if err != nil {
							Logger.Error(fmt.Sprintf("json marshal err:%s", err.Error()))
						}

						bpp.ixCh <- bson.M{"data": string(ixJson)}
					} else {
						Logger.Error(fmt.Sprintf("instruction:%s assertion failed", openbook_v2.InstructionIDToName(ix.TypeID)))
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
				}
			}
		}
	}

	return nil
}

func (bpp *BlockProcessorParser) done() {
	close(bpp.txRawChan)
	close(bpp.ixRawChan)
	close(bpp.ixIndexCh)
	close(bpp.ixCh)
}

var _ BlockProcessor = &BlockProcessorParser{}
