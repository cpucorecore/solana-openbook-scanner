// Code generated by https://github.com/gagliardetto/anchor-go. DO NOT EDIT.

package openbook_v2

import (
	"errors"
	ag_binary "github.com/gagliardetto/binary"
	ag_solanago "github.com/gagliardetto/solana-go"
	ag_format "github.com/gagliardetto/solana-go/text/format"
	ag_treeout "github.com/gagliardetto/treeout"
)

// Cancel up to `limit` orders, optionally filtering by side
type CancelAllOrders struct {
	SideOption *Side `bin:"optional"`
	Limit      *uint8

	// [0] = [SIGNER] signer
	//
	// [1] = [WRITE] openOrdersAccount
	//
	// [2] = [] market
	//
	// [3] = [WRITE] bids
	//
	// [4] = [WRITE] asks
	ag_solanago.AccountMetaSlice `bin:"-"`
}

// NewCancelAllOrdersInstructionBuilder creates a new `CancelAllOrders` instruction builder.
func NewCancelAllOrdersInstructionBuilder() *CancelAllOrders {
	nd := &CancelAllOrders{
		AccountMetaSlice: make(ag_solanago.AccountMetaSlice, 5),
	}
	return nd
}

// SetSideOption sets the "sideOption" parameter.
func (inst *CancelAllOrders) SetSideOption(sideOption Side) *CancelAllOrders {
	inst.SideOption = &sideOption
	return inst
}

// SetLimit sets the "limit" parameter.
func (inst *CancelAllOrders) SetLimit(limit uint8) *CancelAllOrders {
	inst.Limit = &limit
	return inst
}

// SetSignerAccount sets the "signer" account.
func (inst *CancelAllOrders) SetSignerAccount(signer ag_solanago.PublicKey) *CancelAllOrders {
	inst.AccountMetaSlice[0] = ag_solanago.Meta(signer).SIGNER()
	return inst
}

// GetSignerAccount gets the "signer" account.
func (inst *CancelAllOrders) GetSignerAccount() *ag_solanago.AccountMeta {
	return inst.AccountMetaSlice.Get(0)
}

// SetOpenOrdersAccountAccount sets the "openOrdersAccount" account.
func (inst *CancelAllOrders) SetOpenOrdersAccountAccount(openOrdersAccount ag_solanago.PublicKey) *CancelAllOrders {
	inst.AccountMetaSlice[1] = ag_solanago.Meta(openOrdersAccount).WRITE()
	return inst
}

// GetOpenOrdersAccountAccount gets the "openOrdersAccount" account.
func (inst *CancelAllOrders) GetOpenOrdersAccountAccount() *ag_solanago.AccountMeta {
	return inst.AccountMetaSlice.Get(1)
}

// SetMarketAccount sets the "market" account.
func (inst *CancelAllOrders) SetMarketAccount(market ag_solanago.PublicKey) *CancelAllOrders {
	inst.AccountMetaSlice[2] = ag_solanago.Meta(market)
	return inst
}

// GetMarketAccount gets the "market" account.
func (inst *CancelAllOrders) GetMarketAccount() *ag_solanago.AccountMeta {
	return inst.AccountMetaSlice.Get(2)
}

// SetBidsAccount sets the "bids" account.
func (inst *CancelAllOrders) SetBidsAccount(bids ag_solanago.PublicKey) *CancelAllOrders {
	inst.AccountMetaSlice[3] = ag_solanago.Meta(bids).WRITE()
	return inst
}

// GetBidsAccount gets the "bids" account.
func (inst *CancelAllOrders) GetBidsAccount() *ag_solanago.AccountMeta {
	return inst.AccountMetaSlice.Get(3)
}

// SetAsksAccount sets the "asks" account.
func (inst *CancelAllOrders) SetAsksAccount(asks ag_solanago.PublicKey) *CancelAllOrders {
	inst.AccountMetaSlice[4] = ag_solanago.Meta(asks).WRITE()
	return inst
}

// GetAsksAccount gets the "asks" account.
func (inst *CancelAllOrders) GetAsksAccount() *ag_solanago.AccountMeta {
	return inst.AccountMetaSlice.Get(4)
}

func (inst CancelAllOrders) Build() *Instruction {
	return &Instruction{BaseVariant: ag_binary.BaseVariant{
		Impl:   inst,
		TypeID: Instruction_CancelAllOrders,
	}}
}

// ValidateAndBuild validates the instruction parameters and accounts;
// if there is a validation error, it returns the error.
// Otherwise, it builds and returns the instruction.
func (inst CancelAllOrders) ValidateAndBuild() (*Instruction, error) {
	if err := inst.Validate(); err != nil {
		return nil, err
	}
	return inst.Build(), nil
}

func (inst *CancelAllOrders) Validate() error {
	// Check whether all (required) parameters are set:
	{
		if inst.Limit == nil {
			return errors.New("Limit parameter is not set")
		}
	}

	// Check whether all (required) accounts are set:
	{
		if inst.AccountMetaSlice[0] == nil {
			return errors.New("accounts.Signer is not set")
		}
		if inst.AccountMetaSlice[1] == nil {
			return errors.New("accounts.OpenOrdersAccount is not set")
		}
		if inst.AccountMetaSlice[2] == nil {
			return errors.New("accounts.Market is not set")
		}
		if inst.AccountMetaSlice[3] == nil {
			return errors.New("accounts.Bids is not set")
		}
		if inst.AccountMetaSlice[4] == nil {
			return errors.New("accounts.Asks is not set")
		}
	}
	return nil
}

func (inst *CancelAllOrders) EncodeToTree(parent ag_treeout.Branches) {
	parent.Child(ag_format.Program(ProgramName, ProgramID)).
		//
		ParentFunc(func(programBranch ag_treeout.Branches) {
			programBranch.Child(ag_format.Instruction("CancelAllOrders")).
				//
				ParentFunc(func(instructionBranch ag_treeout.Branches) {

					// Parameters of the instruction:
					instructionBranch.Child("Params[len=2]").ParentFunc(func(paramsBranch ag_treeout.Branches) {
						paramsBranch.Child(ag_format.Param("SideOption (OPT)", inst.SideOption))
						paramsBranch.Child(ag_format.Param("     Limit", *inst.Limit))
					})

					// Accounts of the instruction:
					instructionBranch.Child("Accounts[len=5]").ParentFunc(func(accountsBranch ag_treeout.Branches) {
						accountsBranch.Child(ag_format.Meta("    signer", inst.AccountMetaSlice.Get(0)))
						accountsBranch.Child(ag_format.Meta("openOrders", inst.AccountMetaSlice.Get(1)))
						accountsBranch.Child(ag_format.Meta("    market", inst.AccountMetaSlice.Get(2)))
						accountsBranch.Child(ag_format.Meta("      bids", inst.AccountMetaSlice.Get(3)))
						accountsBranch.Child(ag_format.Meta("      asks", inst.AccountMetaSlice.Get(4)))
					})
				})
		})
}

func (obj CancelAllOrders) MarshalWithEncoder(encoder *ag_binary.Encoder) (err error) {
	// Serialize `SideOption` param (optional):
	{
		if obj.SideOption == nil {
			err = encoder.WriteBool(false)
			if err != nil {
				return err
			}
		} else {
			err = encoder.WriteBool(true)
			if err != nil {
				return err
			}
			err = encoder.Encode(obj.SideOption)
			if err != nil {
				return err
			}
		}
	}
	// Serialize `Limit` param:
	err = encoder.Encode(obj.Limit)
	if err != nil {
		return err
	}
	return nil
}
func (obj *CancelAllOrders) UnmarshalWithDecoder(decoder *ag_binary.Decoder) (err error) {
	// Deserialize `SideOption` (optional):
	{
		ok, err := decoder.ReadBool()
		if err != nil {
			return err
		}
		if ok {
			err = decoder.Decode(&obj.SideOption)
			if err != nil {
				return err
			}
		}
	}
	// Deserialize `Limit`:
	err = decoder.Decode(&obj.Limit)
	if err != nil {
		return err
	}
	return nil
}

// NewCancelAllOrdersInstruction declares a new CancelAllOrders instruction with the provided parameters and accounts.
func NewCancelAllOrdersInstruction(
	// Parameters:
	sideOption Side,
	limit uint8,
	// Accounts:
	signer ag_solanago.PublicKey,
	openOrdersAccount ag_solanago.PublicKey,
	market ag_solanago.PublicKey,
	bids ag_solanago.PublicKey,
	asks ag_solanago.PublicKey) *CancelAllOrders {
	return NewCancelAllOrdersInstructionBuilder().
		SetSideOption(sideOption).
		SetLimit(limit).
		SetSignerAccount(signer).
		SetOpenOrdersAccountAccount(openOrdersAccount).
		SetMarketAccount(market).
		SetBidsAccount(bids).
		SetAsksAccount(asks)
}
