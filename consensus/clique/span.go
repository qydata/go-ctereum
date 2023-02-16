package clique

import (
	"context"

	"github.com/qydata/go-ctereum/common"
	"github.com/qydata/go-ctereum/consensus/clique/valset"
	"github.com/qydata/go-ctereum/core"
	"github.com/qydata/go-ctereum/core/state"
	"github.com/qydata/go-ctereum/core/types"
)

//go:generate mockgen -destination=./span_mock.go -package=clique . Spanner
type Spanner interface {
	GetCurrentValidators(ctx context.Context, headerHash common.Hash, blockNumber uint64) ([]*valset.Validator, error)
	CommitAccum(ctx context.Context, state *state.StateDB, header *types.Header, chainContext core.ChainContext, validators []common.Address) error
}
