package consensus

import (
	"bytes"
	"math/big"

	"github.com/pkg/errors"
	"github.com/vechain/thor/block"
	"github.com/vechain/thor/chain"
	"github.com/vechain/thor/contracts"
	"github.com/vechain/thor/runtime"
	"github.com/vechain/thor/state"
	"github.com/vechain/thor/thor"
)

// Consensus check whether the block is verified,
// and predicate which trunk it belong to.
type Consensus struct {
	chain        *chain.Chain
	stateCreator func(thor.Hash) (*state.State, error)
}

// New is Consensus factory.
func New(chain *chain.Chain, stateCreator func(thor.Hash) (*state.State, error)) *Consensus {
	return &Consensus{
		chain:        chain,
		stateCreator: stateCreator}
}

// Consent is Consensus's main func.
func (c *Consensus) Consent(blk *block.Block, nowTime uint64) (isTrunk bool, err error) {
	if blk == nil {
		return false, errors.New("parameter is nil, must be *block.Block")
	}

	preHeader, err := newValidator(blk, c.chain).validate(nowTime)
	if err != nil {
		return false, err
	}

	if err = c.verify(blk, preHeader); err != nil {
		return false, err
	}

	return c.isTrunk(blk)
}

func (c *Consensus) verify(blk *block.Block, preHeader *block.Header) error {
	header := blk.Header()
	preHash := preHeader.StateRoot()
	state, err := c.stateCreator(preHash)
	if err != nil {
		return err
	}

	if err := newProposerHandler(c.chain, state, header, preHeader).handle(); err != nil {
		return err
	}

	rt := runtime.New(state,
		header.Beneficiary(),
		header.Number(),
		header.Timestamp(),
		header.GasLimit(),
		chain.NewBlockIDGetter(c.chain, preHash).GetID)

	energyUsed, err := newBlockProcessor(rt).process(blk)
	if err != nil {
		return err
	}

	rewardPercentage, err := getRewardPercentage(rt)
	if err != nil {
		return err
	}

	if output := handleClause(
		rt,
		contracts.Energy.PackCharge(
			header.Beneficiary(),
			new(big.Int).SetUint64(energyUsed*rewardPercentage))); output.VMErr != nil {
		return errors.Wrap(output.VMErr, "charge energy")
	}

	return checkState(state, header)
}

func (c *Consensus) isTrunk(block *block.Block) (bool, error) {
	bestBlock, err := c.chain.GetBestBlock()

	switch {
	case err != nil:
		return false, err
	case block.TotalScore() < bestBlock.TotalScore():
		return false, nil
	case block.TotalScore() == bestBlock.TotalScore():
		blockID := block.ID()
		bestID := bestBlock.ID()
		if bytes.Compare(blockID[:], bestID[:]) > 0 {
			return true, nil
		}
		return false, nil
	default:
		return true, nil
	}
}
