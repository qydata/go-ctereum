package api

import (
	"context"

	"github.com/qydata/go-ctereum/common/hexutil"
	"github.com/qydata/go-ctereum/internal/ethapi"
	"github.com/qydata/go-ctereum/rpc"
)

//go:generate mockgen -destination=./caller_mock.go -package=api . Caller
type Caller interface {
	Call(ctx context.Context, args ethapi.TransactionArgs, blockNrOrHash rpc.BlockNumberOrHash, overrides *ethapi.StateOverride) (hexutil.Bytes, error)
}
