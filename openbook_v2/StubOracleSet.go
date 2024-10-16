// Code generated by https://github.com/gagliardetto/anchor-go. DO NOT EDIT.

package openbook_v2

import (
	"errors"
	ag_binary "github.com/gagliardetto/binary"
	ag_solanago "github.com/gagliardetto/solana-go"
	ag_format "github.com/gagliardetto/solana-go/text/format"
	ag_treeout "github.com/gagliardetto/treeout"
)

// StubOracleSet is the `stubOracleSet` instruction.
type StubOracleSet struct {
	Price *int64

	// [0] = [SIGNER] owner
	//
	// [1] = [WRITE] oracle
	ag_solanago.AccountMetaSlice `bin:"-"`
}

// NewStubOracleSetInstructionBuilder creates a new `StubOracleSet` instruction builder.
func NewStubOracleSetInstructionBuilder() *StubOracleSet {
	nd := &StubOracleSet{
		AccountMetaSlice: make(ag_solanago.AccountMetaSlice, 2),
	}
	return nd
}

// SetPrice sets the "price" parameter.
func (inst *StubOracleSet) SetPrice(price int64) *StubOracleSet {
	inst.Price = &price
	return inst
}

// SetOwnerAccount sets the "owner" account.
func (inst *StubOracleSet) SetOwnerAccount(owner ag_solanago.PublicKey) *StubOracleSet {
	inst.AccountMetaSlice[0] = ag_solanago.Meta(owner).SIGNER()
	return inst
}

// GetOwnerAccount gets the "owner" account.
func (inst *StubOracleSet) GetOwnerAccount() *ag_solanago.AccountMeta {
	return inst.AccountMetaSlice.Get(0)
}

// SetOracleAccount sets the "oracle" account.
func (inst *StubOracleSet) SetOracleAccount(oracle ag_solanago.PublicKey) *StubOracleSet {
	inst.AccountMetaSlice[1] = ag_solanago.Meta(oracle).WRITE()
	return inst
}

// GetOracleAccount gets the "oracle" account.
func (inst *StubOracleSet) GetOracleAccount() *ag_solanago.AccountMeta {
	return inst.AccountMetaSlice.Get(1)
}

func (inst StubOracleSet) Build() *Instruction {
	return &Instruction{BaseVariant: ag_binary.BaseVariant{
		Impl:   inst,
		TypeID: Instruction_StubOracleSet,
	}}
}

// ValidateAndBuild validates the instruction parameters and accounts;
// if there is a validation error, it returns the error.
// Otherwise, it builds and returns the instruction.
func (inst StubOracleSet) ValidateAndBuild() (*Instruction, error) {
	if err := inst.Validate(); err != nil {
		return nil, err
	}
	return inst.Build(), nil
}

func (inst *StubOracleSet) Validate() error {
	// Check whether all (required) parameters are set:
	{
		if inst.Price == nil {
			return errors.New("Price parameter is not set")
		}
	}

	// Check whether all (required) accounts are set:
	{
		if inst.AccountMetaSlice[0] == nil {
			return errors.New("accounts.Owner is not set")
		}
		if inst.AccountMetaSlice[1] == nil {
			return errors.New("accounts.Oracle is not set")
		}
	}
	return nil
}

func (inst *StubOracleSet) EncodeToTree(parent ag_treeout.Branches) {
	parent.Child(ag_format.Program(ProgramName, ProgramID)).
		//
		ParentFunc(func(programBranch ag_treeout.Branches) {
			programBranch.Child(ag_format.Instruction("StubOracleSet")).
				//
				ParentFunc(func(instructionBranch ag_treeout.Branches) {

					// Parameters of the instruction:
					instructionBranch.Child("Params[len=1]").ParentFunc(func(paramsBranch ag_treeout.Branches) {
						paramsBranch.Child(ag_format.Param("Price", *inst.Price))
					})

					// Accounts of the instruction:
					instructionBranch.Child("Accounts[len=2]").ParentFunc(func(accountsBranch ag_treeout.Branches) {
						accountsBranch.Child(ag_format.Meta(" owner", inst.AccountMetaSlice.Get(0)))
						accountsBranch.Child(ag_format.Meta("oracle", inst.AccountMetaSlice.Get(1)))
					})
				})
		})
}

func (obj StubOracleSet) MarshalWithEncoder(encoder *ag_binary.Encoder) (err error) {
	// Serialize `Price` param:
	err = encoder.Encode(obj.Price)
	if err != nil {
		return err
	}
	return nil
}
func (obj *StubOracleSet) UnmarshalWithDecoder(decoder *ag_binary.Decoder) (err error) {
	// Deserialize `Price`:
	err = decoder.Decode(&obj.Price)
	if err != nil {
		return err
	}
	return nil
}

// NewStubOracleSetInstruction declares a new StubOracleSet instruction with the provided parameters and accounts.
func NewStubOracleSetInstruction(
	// Parameters:
	price int64,
	// Accounts:
	owner ag_solanago.PublicKey,
	oracle ag_solanago.PublicKey) *StubOracleSet {
	return NewStubOracleSetInstructionBuilder().
		SetPrice(price).
		SetOwnerAccount(owner).
		SetOracleAccount(oracle)
}
