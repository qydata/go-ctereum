package core

import (
	"github.com/qydata/go-ctereum/common"
	"github.com/qydata/go-ctereum/core/rawdb"
	"github.com/qydata/go-ctereum/core/types"
)

// GetBorReceiptByHash retrieves the bor block receipt in a given block.
func (bc *BlockChain) GetBorReceiptByHash(hash common.Hash) *types.Receipt {
	if receipt, ok := bc.borReceiptsCache.Get(hash); ok {
		return receipt.(*types.Receipt)
	}

	// read header from hash
	number := rawdb.ReadHeaderNumber(bc.db, hash)
	if number == nil {
		return nil
	}

	// read bor reciept by hash and number
	receipt := rawdb.ReadBorReceipt(bc.db, hash, *number)
	if receipt == nil {
		return nil
	}

	// add into bor receipt cache
	bc.borReceiptsCache.Add(hash, receipt)
	return receipt
}
