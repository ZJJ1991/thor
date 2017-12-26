package vm

import (
	"math/big"

	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/core/types"
	"github.com/ethereum/go-ethereum/params"
	"github.com/vechain/thor/acc"
	"github.com/vechain/thor/cry"
	"github.com/vechain/thor/state"
	"github.com/vechain/thor/vm/account"
	"github.com/vechain/thor/vm/evm"
	"github.com/vechain/thor/vm/snapshot"
	"github.com/vechain/thor/vm/state"
	"github.com/vechain/thor/vm/vmlog"
)

// Config is ref to evm.Config.
type Config evm.Config

// Output contains the execution return value.
type Output struct {
	Value           []byte
	DirtiedAccounts []*account.Account
	Preimages       map[cry.Hash][]byte
	Log             []*types.Log
	VMErr           error        // VMErr identify the execution result of the contract function, not evm function's err.
	ContractAddress *acc.Address // if create a new contract, or is nil.
}

// VM is a facade for ethEvm.
type VM struct {
	evm   *evm.EVM
	state *state.State
}

var chainConfig = &params.ChainConfig{
	ChainId:        big.NewInt(0),
	HomesteadBlock: big.NewInt(0),
	DAOForkBlock:   big.NewInt(0),
	DAOForkSupport: false,
	EIP150Block:    big.NewInt(0),
	EIP150Hash:     common.Hash{},
	EIP155Block:    big.NewInt(0),
	EIP158Block:    big.NewInt(0),
	ByzantiumBlock: big.NewInt(0),
	Ethash:         nil,
	Clique:         nil,
}

// Context for VM runtime.
type Context struct {
	Origin      acc.Address
	Beneficiary acc.Address
	BlockNumber *big.Int
	Time        *big.Int
	GasLimit    *big.Int
	GasPrice    *big.Int
	TxHash      cry.Hash
	ClauseIndex uint64
	GetHash     func(uint64) cry.Hash
}

// The only purpose of this func separate definition is to be compatible with evm.context.
func canTransfer(db evm.StateDB, addr common.Address, amount *big.Int) bool {
	return db.GetBalance(addr).Cmp(amount) >= 0
}

// The only purpose of this func separate definition is to be compatible with evm.Context.
func transfer(db evm.StateDB, sender, recipient common.Address, amount *big.Int) {
	db.SubBalance(sender, amount)
	db.AddBalance(recipient, amount)
}

// NewVM retutrns a new EVM . The returned EVM is not thread safe and should
// only ever be used *once*.
func NewVM(ctx Context, stateReader account.StateReader, vmConfig Config) *VM {
	tGetHash := func(n uint64) common.Hash {
		return common.Hash(ctx.GetHash(n))
	}

	am := account.NewManager(stateReader)

	evmCtx := evm.Context{
		CanTransfer: canTransfer,
		Transfer:    transfer,
		GetHash:     tGetHash,
		Difficulty:  new(big.Int),

		Origin:      common.Address(ctx.Origin),
		Coinbase:    common.Address(ctx.Beneficiary),
		BlockNumber: ctx.BlockNumber,
		Time:        ctx.Time,
		GasLimit:    ctx.GasLimit,
		GasPrice:    ctx.GasPrice,
		TxHash:      common.Hash(ctx.TxHash),
	}

	sm := snapshot.New()
	vl := vmlog.New()
	state := state.New(am, sm, vl)
	evm := evm.NewEVM(evmCtx, state, chainConfig, evm.Config(vmConfig))
	return &VM{evm: evm, state: state}
}

// Cancel cancels any running EVM operation.
// This may be called concurrently and it's safe to be called multiple times.
func (vm *VM) Cancel() {
	vm.evm.Cancel()
}

// Call executes the contract associated with the addr with the given input as parameters.
// It also handles any necessary value transfer required and takes the necessary steps to
// create accounts and reverses the state in case of an execution error or failed value transfer.
func (vm *VM) Call(caller acc.Address, addr acc.Address, input []byte, gas uint64, value *big.Int) (*Output, uint64, *big.Int) {
	ret, leftOverGas, vmErr := vm.evm.Call(&vmContractRef{caller}, common.Address(addr), input, gas, value)
	return &Output{
		Value:           ret,
		DirtiedAccounts: vm.state.GetDirtyAccounts(),
		Preimages:       vm.state.Preimages(),
		Log:             vm.state.GetLogs(),
		VMErr:           vmErr,
		ContractAddress: nil,
	}, leftOverGas, vm.state.GetRefund()
}

// CallCode executes the contract associated with the addr with the given input as parameters.
// It also handles any necessary value transfer required and takes the necessary steps to create
// accounts and reverses the state in case of an execution error or failed value transfer.
//
// CallCode differs from Call in the sense that it executes the given address'
// code with the caller as context.
func (vm *VM) CallCode(caller acc.Address, addr acc.Address, input []byte, gas uint64, value *big.Int) (*Output, uint64, *big.Int) {
	ret, leftOverGas, vmErr := vm.evm.CallCode(&vmContractRef{caller}, common.Address(addr), input, gas, value)
	return &Output{
		Value:           ret,
		DirtiedAccounts: vm.state.GetDirtyAccounts(),
		Preimages:       vm.state.Preimages(),
		Log:             vm.state.GetLogs(),
		VMErr:           vmErr,
		ContractAddress: nil,
	}, leftOverGas, vm.state.GetRefund()
}

// DelegateCall executes the contract associated with the addr with the given input as parameters.
// It reverses the state in case of an execution error.
//
// DelegateCall differs from CallCode in the sense that it executes the given address' code with
// the caller as context and the caller is set to the caller of the caller.
func (vm *VM) DelegateCall(caller acc.Address, addr acc.Address, input []byte, gas uint64) (*Output, uint64, *big.Int) {
	ret, leftOverGas, vmErr := vm.evm.DelegateCall(&vmContractRef{caller}, common.Address(addr), input, gas)
	return &Output{
		Value:           ret,
		DirtiedAccounts: vm.state.GetDirtyAccounts(),
		Preimages:       vm.state.Preimages(),
		Log:             vm.state.GetLogs(),
		VMErr:           vmErr,
		ContractAddress: nil,
	}, leftOverGas, vm.state.GetRefund()
}

// StaticCall executes the contract associated with the addr with the given input as parameters
// while disallowing any modifications to the state during the call.
//
// Opcodes that attempt to perform such modifications will result in exceptions instead of performing
// the modifications.
func (vm *VM) StaticCall(caller acc.Address, addr acc.Address, input []byte, gas uint64) (*Output, uint64, *big.Int) {
	ret, leftOverGas, vmErr := vm.evm.StaticCall(&vmContractRef{caller}, common.Address(addr), input, gas)
	return &Output{
		Value:           ret,
		DirtiedAccounts: vm.state.GetDirtyAccounts(),
		Preimages:       vm.state.Preimages(),
		Log:             vm.state.GetLogs(),
		VMErr:           vmErr,
		ContractAddress: nil,
	}, leftOverGas, vm.state.GetRefund()
}

// Create creates a new contract using code as deployment code.
func (vm *VM) Create(caller acc.Address, code []byte, gas uint64, value *big.Int) (*Output, uint64, *big.Int) {
	ret, contractAddr, leftOverGas, vmErr := vm.evm.Create(&vmContractRef{caller}, code, gas, value)
	ContractAddress := acc.Address(contractAddr)
	return &Output{
		Value:           ret,
		DirtiedAccounts: vm.state.GetDirtyAccounts(),
		Preimages:       vm.state.Preimages(),
		Log:             vm.state.GetLogs(),
		VMErr:           vmErr,
		ContractAddress: &ContractAddress,
	}, leftOverGas, vm.state.GetRefund()
}

// ChainConfig returns the evmironment's chain configuration
func (vm *VM) ChainConfig() *params.ChainConfig {
	return vm.evm.ChainConfig()
}
