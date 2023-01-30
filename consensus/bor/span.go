package bor

import (
	"context"

	"github.com/qydata/go-ctereum/common"
	"github.com/qydata/go-ctereum/consensus/bor/heimdall/span"
	"github.com/qydata/go-ctereum/consensus/bor/valset"
	"github.com/qydata/go-ctereum/core"
	"github.com/qydata/go-ctereum/core/state"
	"github.com/qydata/go-ctereum/core/types"
)

//go:generate mockgen -destination=./span_mock.go -package=bor . Spanner
type Spanner interface {
	GetCurrentSpan(ctx context.Context, headerHash common.Hash) (*span.Span, error)
	GetCurrentValidators(ctx context.Context, headerHash common.Hash, blockNumber uint64) ([]*valset.Validator, error)
	CommitSpan(ctx context.Context, heimdallSpan span.HeimdallSpan, state *state.StateDB, header *types.Header, chainContext core.ChainContext) error
}
