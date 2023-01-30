package bor

import (
	"math/big"

	"github.com/ethereum/go-ctereum/consensus/bor/clerk"
	"github.com/ethereum/go-ctereum/consensus/bor/statefull"
	"github.com/ethereum/go-ctereum/core/state"
	"github.com/ethereum/go-ctereum/core/types"
)

//go:generate mockgen -destination=./genesis_contract_mock.go -package=bor . GenesisContract
type GenesisContract interface {
	CommitState(event *clerk.EventRecordWithTime, state *state.StateDB, header *types.Header, chCtx statefull.ChainContext) (uint64, error)
	LastStateId(snapshotNumber uint64) (*big.Int, error)
}
