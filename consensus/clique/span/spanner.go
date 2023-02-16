package span

import (
	"context"
	"math"
	"math/big"

	"github.com/qydata/go-ctereum/common"
	"github.com/qydata/go-ctereum/common/hexutil"
	"github.com/qydata/go-ctereum/consensus/clique/abi"
	"github.com/qydata/go-ctereum/consensus/clique/api"
	"github.com/qydata/go-ctereum/consensus/clique/statefull"
	"github.com/qydata/go-ctereum/consensus/clique/valset"
	"github.com/qydata/go-ctereum/core"
	"github.com/qydata/go-ctereum/core/state"
	"github.com/qydata/go-ctereum/core/types"
	"github.com/qydata/go-ctereum/internal/ethapi"
	"github.com/qydata/go-ctereum/log"
	"github.com/qydata/go-ctereum/params"
	"github.com/qydata/go-ctereum/rpc"
)

type ChainSpanner struct {
	ethAPI                   api.Caller
	staking                  abi.ABI
	chainConfig              *params.ChainConfig
	validatorContractAddress common.Address
}

func NewChainSpanner(ethAPI api.Caller, staking abi.ABI, chainConfig *params.ChainConfig, validatorContractAddress common.Address) *ChainSpanner {
	return &ChainSpanner{
		ethAPI:                   ethAPI,
		staking:                  staking,
		chainConfig:              chainConfig,
		validatorContractAddress: validatorContractAddress,
	}
}

// GetCurrentValidators get current validators
func (c *ChainSpanner) GetCurrentValidators(ctx context.Context, headerHash common.Hash, blockNumber uint64) ([]*valset.Validator, error) {
	ctx, cancel := context.WithCancel(ctx)
	defer cancel()

	// method
	const method = "getValidators"

	data, err := c.staking.Pack(method)
	if err != nil {
		log.Error("Unable to pack tx for getValidator", "error", err)
		return nil, err
	}

	// call
	msgData := (hexutil.Bytes)(data)
	toAddress := c.validatorContractAddress
	gas := (hexutil.Uint64)(uint64(math.MaxUint64 / 2))

	// block
	blockNr := rpc.BlockNumberOrHashWithHash(headerHash, false)
	//blockNr := rpc.BlockNumberOrHashWithNumber(rpc.BlockNumber(blockNumber))
	result, err := c.ethAPI.Call(ctx, ethapi.TransactionArgs{
		Gas:  &gas,
		To:   &toAddress,
		Data: &msgData,
	}, blockNr, nil)
	if err != nil {
		panic(err)
	}

	var (
		ret0 = new([]common.Address)
		ret1 = new([]*big.Int)
		ret2 = new([]*big.Int)
	)

	out := &[]interface{}{
		ret0,
		ret1,
		ret2,
	}

	if err := c.staking.UnpackIntoInterface(out, method, result); err != nil {
		return nil, err
	}

	valz := make([]*valset.Validator, len(*ret0))
	for i, a := range *ret0 {
		valz[i] = &valset.Validator{
			Address:          a,
			VotingPower:      (*ret1)[i].Int64(),
			ProposerPriority: (*ret2)[i].Int64(),
		}
	}

	return valz, nil
}

const method = "commitAccum"

func (c *ChainSpanner) CommitAccum(ctx context.Context, state *state.StateDB, header *types.Header, chainContext core.ChainContext, validators []common.Address) error {

	// get producers bytes
	log.Info("✅ Committing new accum",
		"Validators", validators,
	)

	data, err := c.staking.Pack(method,
		validators,
	)
	if err != nil {
		log.Error("Unable to pack tx for CommitAccum", "error", err)

		return err
	}

	// get system message
	msg := statefull.GetSystemMessage(c.validatorContractAddress, data)

	// apply message
	_, err = statefull.ApplyMessage(ctx, msg, state, header, c.chainConfig, chainContext)

	return err
}
