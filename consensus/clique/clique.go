// Copyright 2017 The go-ctereum Authors
// This file is part of the go-ctereum library.
//
// The go-ctereum library is free software: you can redistribute it and/or modify
// it under the terms of the GNU Lesser General Public License as published by
// the Free Software Foundation, either version 3 of the License, or
// (at your option) any later version.
//
// The go-ctereum library is distributed in the hope that it will be useful,
// but WITHOUT ANY WARRANTY; without even the implied warranty of
// MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE. See the
// GNU Lesser General Public License for more details.
//
// You should have received a copy of the GNU Lesser General Public License
// along with the go-ctereum library. If not, see <http://www.gnu.org/licenses/>.

// Package clique implements the proof-of-authority consensus engine.
package clique

import (
	"bytes"
	"context"
	"errors"
	"fmt"
	"github.com/qydata/go-ctereum/consensus/clique/statefull"
	"io"
	"math/big"
	"math/rand"
	"sync"
	"time"

	lru "github.com/hashicorp/golang-lru"
	"github.com/qydata/go-ctereum/accounts"
	"github.com/qydata/go-ctereum/common"
	"github.com/qydata/go-ctereum/common/hexutil"
	"github.com/qydata/go-ctereum/consensus"
	"github.com/qydata/go-ctereum/consensus/misc"
	"github.com/qydata/go-ctereum/core/state"
	"github.com/qydata/go-ctereum/core/types"
	"github.com/qydata/go-ctereum/crypto"
	"github.com/qydata/go-ctereum/ethdb"
	"github.com/qydata/go-ctereum/log"
	"github.com/qydata/go-ctereum/params"
	"github.com/qydata/go-ctereum/rlp"
	"github.com/qydata/go-ctereum/rpc"
	"github.com/qydata/go-ctereum/trie"
	"golang.org/x/crypto/sha3"
)

const (
	checkpointInterval = 1024 // Number of blocks after which to save the vote snapshot to the database
	inmemorySnapshots  = 128  // Number of recent vote snapshots to keep in memory
	inmemorySignatures = 4096 // Number of recent block signatures to keep in memory

	wiggleTime = 500 * time.Millisecond // Random delay (per signer) to allow concurrent signers
)

// Clique proof-of-authority protocol constants.
var (
	BlockReward = big.NewInt(5e+18) // Block reward in wei for successfully mining a block upward from Byzantium

	epochLength = uint64(30000) // Default number of blocks after which to checkpoint and reset the pending votes

	extraVanity = 32                     // Fixed number of extra-data prefix bytes reserved for signer vanity
	extraSeal   = crypto.SignatureLength // Fixed number of extra-data suffix bytes reserved for signer seal

	nonceAuthVote = hexutil.MustDecode("0xffffffffffffffff") // Magic nonce number to vote on adding a new signer
	nonceDropVote = hexutil.MustDecode("0x0000000000000000") // Magic nonce number to vote on removing a signer.

	uncleHash = types.CalcUncleHash(nil) // Always Keccak256(RLP([])) as uncles are meaningless outside of PoW.

	diffInTurn = big.NewInt(2) // Block difficulty for in-turn signatures
	diffNoTurn = big.NewInt(1) // Block difficulty for out-of-turn signatures
)

// Various error messages to mark blocks invalid. These should be private to
// prevent engine specific errors from being referenced in the remainder of the
// codebase, inherently breaking if the engine is swapped out. Please put common
// error types into the consensus package.
var (
	// errUnknownBlock is returned when the list of signers is requested for a block
	// that is not part of the local blockchain.
	errUnknownBlock = errors.New("unknown block")

	// errInvalidCheckpointBeneficiary is returned if a checkpoint/epoch transition
	// block has a beneficiary set to non-zeroes.
	errInvalidCheckpointBeneficiary = errors.New("beneficiary in checkpoint block non-zero")

	// errInvalidVote is returned if a nonce value is something else that the two
	// allowed constants of 0x00..0 or 0xff..f.
	errInvalidVote = errors.New("vote nonce not 0x00..0 or 0xff..f")

	// errInvalidCheckpointVote is returned if a checkpoint/epoch transition block
	// has a vote nonce set to non-zeroes.
	errInvalidCheckpointVote = errors.New("vote nonce in checkpoint block non-zero")

	// errMissingVanity is returned if a block's extra-data section is shorter than
	// 32 bytes, which is required to store the signer vanity.
	errMissingVanity = errors.New("extra-data 32 byte vanity prefix missing")

	// errMissingSignature is returned if a block's extra-data section doesn't seem
	// to contain a 65 byte secp256k1 signature.
	errMissingSignature = errors.New("extra-data 65 byte signature suffix missing")

	// errExtraSigners is returned if non-checkpoint block contain signer data in
	// their extra-data fields.
	errExtraSigners = errors.New("non-checkpoint block contains extra signer list")

	// errInvalidCheckpointSigners is returned if a checkpoint block contains an
	// invalid list of signers (i.e. non divisible by 20 bytes).
	errInvalidCheckpointSigners = errors.New("invalid signer list on checkpoint block")

	// errMismatchingCheckpointSigners is returned if a checkpoint block contains a
	// list of signers different than the one the local node calculated.
	errMismatchingCheckpointSigners = errors.New("mismatching signer list on checkpoint block")

	// errInvalidMixDigest is returned if a block's mix digest is non-zero.
	errInvalidMixDigest = errors.New("non-zero mix digest")

	// errInvalidUncleHash is returned if a block contains an non-empty uncle list.
	errInvalidUncleHash = errors.New("non empty uncle hash")

	// errInvalidDifficulty is returned if the difficulty of a block neither 1 or 2.
	errInvalidDifficulty = errors.New("invalid difficulty")

	// errWrongDifficulty is returned if the difficulty of a block doesn't match the
	// turn of the signer.
	errWrongDifficulty = errors.New("wrong difficulty")

	// errInvalidTimestamp is returned if the timestamp of a block is lower than
	// the previous block's timestamp + the minimum block period.
	errInvalidTimestamp = errors.New("invalid timestamp")

	// errInvalidVotingChain is returned if an authorization list is attempted to
	// be modified via out-of-range or non-contiguous headers.
	errInvalidVotingChain = errors.New("invalid voting chain")

	// errUnauthorizedSigner is returned if a header is signed by a non-authorized entity.
	errUnauthorizedSigner = errors.New("unauthorized signer")

	// errRecentlySigned is returned if a header is signed by an authorized entity
	// that already signed a header recently, thus is temporarily not allowed to.
	errRecentlySigned    = errors.New("recently signed")
	errUnknownValidators = errors.New("unknown validators")
)

// SignerFn hashes and signs the data to be signed by a backing account.
type SignerFn func(signer accounts.Account, mimeType string, message []byte) ([]byte, error)

// ecrecover extracts the Ethereum account address from a signed header.
func ecrecover(header *types.Header, sigcache *lru.ARCCache) (common.Address, error) {
	// If the signature's already cached, return that
	hash := header.Hash()
	if address, known := sigcache.Get(hash); known {
		return address.(common.Address), nil
	}
	// Retrieve the signature from the header extra-data
	if len(header.Extra) < extraSeal {
		return common.Address{}, errMissingSignature
	}
	signature := header.Extra[len(header.Extra)-extraSeal:]

	// Recover the public key and the Ethereum address
	pubkey, err := crypto.Ecrecover(SealHash(header).Bytes(), signature)
	if err != nil {
		return common.Address{}, err
	}
	var signer common.Address
	copy(signer[:], crypto.Keccak256(pubkey[1:])[12:])

	sigcache.Add(hash, signer)
	return signer, nil
}

// Clique is the proof-of-authority consensus engine proposed to support the
// Ethereum testnet following the Ropsten attacks.
type Clique struct {
	config *params.CliqueConfig // Consensus engine configuration parameters
	db     ethdb.Database       // Database to store and retrieve snapshot checkpoints

	recents    *lru.ARCCache // Snapshots for recent block to speed up reorgs
	signatures *lru.ARCCache // Signatures of recent blocks to speed up mining

	proposals map[common.Address]bool // Current list of proposals we are pushing

	signer common.Address // Ethereum address of the signing key
	signFn SignerFn       // Signer function to authorize hashes with
	lock   sync.RWMutex   // Protects the signer and proposals fields

	// The fields below are for testing only
	fakeDiff bool // Skip difficulty verifications

	spanner Spanner
}

// New creates a Clique proof-of-authority consensus engine with the initial
// signers set to the ones provided by the user.
func New(config *params.CliqueConfig, db ethdb.Database, spanner Spanner) *Clique {
	// Set any missing consensus parameters to their defaults
	conf := *config
	if conf.Epoch == 0 {
		conf.Epoch = epochLength
	}
	// Allocate the snapshot caches and create the engine
	recents, _ := lru.NewARC(inmemorySnapshots)
	signatures, _ := lru.NewARC(inmemorySignatures)

	return &Clique{
		config:     &conf,
		db:         db,
		recents:    recents,
		signatures: signatures,
		proposals:  make(map[common.Address]bool),
		spanner:    spanner,
	}
}

// Author implements consensus.Engine, returning the Ethereum address recovered
// from the signature in the header's extra-data section.
func (c *Clique) Author(header *types.Header) (common.Address, error) {
	return ecrecover(header, c.signatures)
}

// VerifyHeader checks whether a header conforms to the consensus rules.
func (c *Clique) VerifyHeader(chain consensus.ChainHeaderReader, header *types.Header, seal bool) error {
	return c.verifyHeader(chain, header, nil)
}

// VerifyHeaders is similar to VerifyHeader, but verifies a batch of headers. The
// method returns a quit channel to abort the operations and a results channel to
// retrieve the async verifications (the order is that of the input slice).
func (c *Clique) VerifyHeaders(chain consensus.ChainHeaderReader, headers []*types.Header, seals []bool) (chan<- struct{}, <-chan error) {
	abort := make(chan struct{})
	results := make(chan error, len(headers))

	go func() {
		for i, header := range headers {
			err := c.verifyHeader(chain, header, headers[:i])

			select {
			case <-abort:
				return
			case results <- err:
			}
		}
	}()
	return abort, results
}

// verifyHeader checks whether a header conforms to the consensus rules.The
// caller may optionally pass in a batch of parents (ascending order) to avoid
// looking those up from the database. This is useful for concurrently verifying
// a batch of new headers.
func (c *Clique) verifyHeader(chain consensus.ChainHeaderReader, header *types.Header, parents []*types.Header) error {
	if header.Number == nil {
		return errUnknownBlock
	}
	number := header.Number.Uint64()

	// Don't waste time checking blocks from the future
	if header.Time > uint64(time.Now().Unix()) {
		return consensus.ErrFutureBlock
	}
	// Checkpoint blocks need to enforce zero beneficiary
	checkpoint := (number % c.config.Epoch) == 0
	if checkpoint && header.Coinbase != (common.Address{}) {
		return errInvalidCheckpointBeneficiary
	}
	// Nonces must be 0x00..0 or 0xff..f, zeroes enforced on checkpoints
	if !bytes.Equal(header.Nonce[:], nonceAuthVote) && !bytes.Equal(header.Nonce[:], nonceDropVote) {
		return errInvalidVote
	}
	if checkpoint && !bytes.Equal(header.Nonce[:], nonceDropVote) {
		return errInvalidCheckpointVote
	}
	// Check that the extra-data contains both the vanity and signature
	if len(header.Extra) < extraVanity {
		return errMissingVanity
	}
	if len(header.Extra) < extraVanity+extraSeal {
		return errMissingSignature
	}
	// Ensure that the extra-data contains a signer list on checkpoint, but none otherwise
	signersBytes := len(header.Extra) - extraVanity - extraSeal
	if !checkpoint && signersBytes != 0 {
		return errExtraSigners
	}
	if checkpoint && signersBytes%common.AddressLength != 0 {
		return errInvalidCheckpointSigners
	}
	// Ensure that the mix digest is zero as we don't have fork protection currently
	if header.MixDigest != (common.Hash{}) {
		return errInvalidMixDigest
	}
	// Ensure that the block doesn't contain any uncles which are meaningless in PoA
	if header.UncleHash != uncleHash {
		return errInvalidUncleHash
	}
	// Ensure that the block's difficulty is meaningful (may not be correct at this point)
	if number > 0 {
		if header.Difficulty == nil || (header.Difficulty.Cmp(diffInTurn) != 0 && header.Difficulty.Cmp(diffNoTurn) != 0) {
			return errInvalidDifficulty
		}
	}
	// Verify that the gas limit is <= 2^63-1
	if header.GasLimit > params.MaxGasLimit {
		return fmt.Errorf("invalid gasLimit: have %v, max %v", header.GasLimit, params.MaxGasLimit)
	}
	// If all checks passed, validate any special fields for hard forks
	if err := misc.VerifyForkHashes(chain.Config(), header, false); err != nil {
		return err
	}
	// All basic checks passed, verify cascading fields
	return c.verifyCascadingFields(chain, header, parents)
}

// verifyCascadingFields verifies all the header fields that are not standalone,
// rather depend on a batch of previous headers. The caller may optionally pass
// in a batch of parents (ascending order) to avoid looking those up from the
// database. This is useful for concurrently verifying a batch of new headers.
func (c *Clique) verifyCascadingFields(chain consensus.ChainHeaderReader, header *types.Header, parents []*types.Header) error {
	// The genesis block is the always valid dead-end
	number := header.Number.Uint64()
	if number == 0 {
		return nil
	}
	// Ensure that the block's timestamp isn't too close to its parent
	var parent *types.Header
	if len(parents) > 0 {
		parent = parents[len(parents)-1]
	} else {
		parent = chain.GetHeader(header.ParentHash, number-1)
	}
	if parent == nil || parent.Number.Uint64() != number-1 || parent.Hash() != header.ParentHash {
		return consensus.ErrUnknownAncestor
	}
	if parent.Time+c.config.Period > header.Time {
		return errInvalidTimestamp
	}
	// Verify that the gasUsed is <= gasLimit
	if header.GasUsed > header.GasLimit {
		return fmt.Errorf("invalid gasUsed: have %d, gasLimit %d", header.GasUsed, header.GasLimit)
	}
	if !chain.Config().IsLondon(header.Number) {
		// Verify BaseFee not present before EIP-1559 fork.
		if header.BaseFee != nil {
			return fmt.Errorf("invalid baseFee before fork: have %d, want <nil>", header.BaseFee)
		}
		if err := misc.VerifyGaslimit(parent.GasLimit, header.GasLimit); err != nil {
			return err
		}
	} else if err := misc.VerifyEip1559Header(chain.Config(), parent, header); err != nil {
		// Verify the header's EIP-1559 attributes.
		return err
	}
	// Retrieve the snapshot needed to verify this header and cache it
	snap, err := c.snapshot(chain, number-1, header.ParentHash, parents)
	if err != nil {
		return err
	}
	// If the block is a checkpoint block, verify the signer list
	if number%c.config.Epoch == 0 {
		signers := make([]byte, len(snap.Signers)*common.AddressLength)

		for i, signer := range snap.signers() {
			copy(signers[i*common.AddressLength:], signer[:])
		}
		extraSuffix := len(header.Extra) - extraSeal
		if !bytes.Equal(header.Extra[extraVanity:extraSuffix], signers) {
			return errMismatchingCheckpointSigners
		}
	}
	// All basic checks passed, verify the seal and return
	return c.verifySeal(snap, header, parents)
}

// snapshot retrieves the authorization snapshot at a given point in time.
func (c *Clique) snapshot(chain consensus.ChainHeaderReader, number uint64, hash common.Hash, parents []*types.Header) (*Snapshot, error) {
	// Search for a snapshot in memory or on disk for checkpoints
	var (
		headers []*types.Header
		snap    *Snapshot
	)
	for snap == nil {
		// If an in-memory snapshot was found, use that
		if s, ok := c.recents.Get(hash); ok {
			snap = s.(*Snapshot)
			break
		}
		// If an on-disk checkpoint snapshot can be found, use that
		if number%checkpointInterval == 0 {
			if s, err := loadSnapshot(c.config, c.signatures, c.db, hash); err == nil {
				log.Trace("Loaded voting snapshot from disk", "number", number, "hash", hash)
				snap = s
				break
			}
		}
		// If we're at the genesis, snapshot the initial state. Alternatively if we're
		// at a checkpoint block without a parent (light client CHT), or we have piled
		// up more headers than allowed to be reorged (chain reinit from a freezer),
		// consider the checkpoint trusted and snapshot it.
		if number == 0 || (number%c.config.Epoch == 0 && (len(headers) > params.FullImmutabilityThreshold || chain.GetHeaderByNumber(number-1) == nil)) {
			checkpoint := chain.GetHeaderByNumber(number)
			if checkpoint != nil {
				hash := checkpoint.Hash()

				signers := make([]common.Address, (len(checkpoint.Extra)-extraVanity-extraSeal)/common.AddressLength)
				for i := 0; i < len(signers); i++ {
					copy(signers[i][:], checkpoint.Extra[extraVanity+i*common.AddressLength:])
				}
				snap = newSnapshot(c.config, c.signatures, number, hash, signers)
				if err := snap.store(c.db); err != nil {
					return nil, err
				}
				log.Info("Stored checkpoint snapshot to disk", "number", number, "hash", hash)
				break
			}
		}
		// No snapshot for this header, gather the header and move backward
		var header *types.Header
		if len(parents) > 0 {
			// If we have explicit parents, pick from there (enforced)
			header = parents[len(parents)-1]
			if header.Hash() != hash || header.Number.Uint64() != number {
				return nil, consensus.ErrUnknownAncestor
			}
			parents = parents[:len(parents)-1]
		} else {
			// No explicit parents (or no more left), reach out to the database
			header = chain.GetHeader(hash, number)
			if header == nil {
				return nil, consensus.ErrUnknownAncestor
			}
		}
		headers = append(headers, header)
		number, hash = number-1, header.ParentHash
	}
	// Previous snapshot found, apply any pending headers on top of it
	for i := 0; i < len(headers)/2; i++ {
		headers[i], headers[len(headers)-1-i] = headers[len(headers)-1-i], headers[i]
	}

	snap, err := snap.apply(headers)
	if err != nil {
		return nil, err
	}
	c.recents.Add(snap.Hash, snap)

	// If we've generated a new checkpoint snapshot, save to disk
	if snap.Number%checkpointInterval == 0 && len(headers) > 0 {
		if err = snap.store(c.db); err != nil {
			return nil, err
		}
		log.Trace("Stored voting snapshot to disk", "number", snap.Number, "hash", snap.Hash)
	}
	return snap, err
}

// VerifyUncles implements consensus.Engine, always returning an error for any
// uncles as this consensus mechanism doesn't permit uncles.
func (c *Clique) VerifyUncles(chain consensus.ChainReader, block *types.Block) error {
	if len(block.Uncles()) > 0 {
		return errors.New("uncles not allowed")
	}
	return nil
}

// verifySeal checks whether the signature contained in the header satisfies the
// consensus protocol requirements. The method accepts an optional list of parent
// headers that aren't yet part of the local blockchain to generate the snapshots
// from.
func (c *Clique) verifySeal(snap *Snapshot, header *types.Header, parents []*types.Header) error {
	// Verifying the genesis block is not supported
	number := header.Number.Uint64()
	if number == 0 {
		return errUnknownBlock
	}
	// Resolve the authorization key and check against signers
	signer, err := ecrecover(header, c.signatures)
	if err != nil {
		return err
	}
	if _, ok := snap.Signers[signer]; !ok {
		return errUnauthorizedSigner
	}
	for seen, recent := range snap.Recents {
		if recent == signer {
			// Signer is among recents, only fail if the current block doesn't shift it out
			if limit := uint64(len(snap.Signers)/2 + 1); seen > number-limit {
				return errRecentlySigned
			}
		}
	}
	// Ensure that the difficulty corresponds to the turn-ness of the signer
	if !c.fakeDiff {
		inturn := snap.inturn(header.Number.Uint64(), signer)
		if inturn && header.Difficulty.Cmp(diffInTurn) != 0 {
			return errWrongDifficulty
		}
		if !inturn && header.Difficulty.Cmp(diffNoTurn) != 0 {
			return errWrongDifficulty
		}
	}
	return nil
}

// Prepare implements consensus.Engine, preparing all the consensus fields of the
// header for running the transactions on top.
func (c *Clique) Prepare(chain consensus.ChainHeaderReader, header *types.Header) error {
	// If the block isn't a checkpoint, cast a random vote (good enough for now)
	header.Coinbase = common.Address{}
	header.Nonce = types.BlockNonce{}

	number := header.Number.Uint64()
	// Assemble the voting snapshot to check which votes make sense
	snap, err := c.snapshot(chain, number-1, header.ParentHash, nil)
	if err != nil {
		return err
	}
	c.lock.RLock()
	if number%c.config.Epoch != 0 {
		if chain.Config().IsPoa2Pos(big.NewInt(0).SetUint64(number)) {

			newValidators, err := c.spanner.GetCurrentValidators(context.Background(), header.ParentHash, number+1)
			if err1 := snap.updateSigners(newValidators, c); err1 != nil {
				log.Info("updateSigners", "Err:", err1)
				//}
			}
			if err != nil {
				log.Info("Prepare", "err:", err)
				return errUnknownValidators
			}

		}
		// Gather all the proposals that make sense voting on
		addresses := make([]common.Address, 0, len(c.proposals))
		for address, authorize := range c.proposals {
			if snap.validVote(address, authorize) {
				addresses = append(addresses, address)
			}
		}
		// If there's pending proposals, cast a vote on them
		if len(addresses) > 0 {
			header.Coinbase = addresses[rand.Intn(len(addresses))]
			if c.proposals[header.Coinbase] {
				copy(header.Nonce[:], nonceAuthVote)
			} else {
				copy(header.Nonce[:], nonceDropVote)
			}
		}
	}

	// Copy signer protected by mutex to avoid race condition
	signer := c.signer
	c.lock.RUnlock()

	// Set the correct difficulty
	header.Difficulty = calcDifficulty(snap, signer)

	// Ensure the extra data has all its components
	if len(header.Extra) < extraVanity {
		header.Extra = append(header.Extra, bytes.Repeat([]byte{0x00}, extraVanity-len(header.Extra))...)
	}
	header.Extra = header.Extra[:extraVanity]

	if number%c.config.Epoch == 0 {
		for _, signer := range snap.signers() {
			header.Extra = append(header.Extra, signer[:]...)
		}
	}
	header.Extra = append(header.Extra, make([]byte, extraSeal)...)

	// Mix digest is reserved for now, set to empty
	header.MixDigest = common.Hash{}

	// Ensure the timestamp has the correct delay
	parent := chain.GetHeader(header.ParentHash, number-1)
	if parent == nil {
		return consensus.ErrUnknownAncestor
	}
	header.Time = parent.Time + c.config.Period
	if header.Time < uint64(time.Now().Unix()) {
		header.Time = uint64(time.Now().Unix())
	}
	return nil
}

// Finalize implements consensus.Engine, ensuring no uncles are set, nor block
// rewards given.
func (c *Clique) Finalize(chain consensus.ChainHeaderReader, header *types.Header, state *state.StateDB, txs []*types.Transaction, uncles []*types.Header) {
	//iozhaq  加入矿工奖励
	blockReward := BlockReward
	reward := new(big.Int).Set(blockReward)
	number := header.Number.Uint64()
	//log.Info("区块奖励签名地址打印number:", number)
	snap, err := c.snapshot(chain, number-1, header.ParentHash, nil)
	if err != nil {
		log.Info("Finalize", "err", err)
	}

	rewardAddress := snap.Recents[number-1]

	if !chain.Config().IsImplAuth(header.Number) {
		log.Info("区块奖励签名地址打印", "rewardAddress:", rewardAddress.Hex())
		state.AddBalance(rewardAddress, reward)
	}

	//if chain.Config().Poa2PosBlock == big.NewInt(0).SetUint64(number) {
	if (header.Number.Int64() + 1) == c.config.Poa2PosBlock {
		state.SetCode(
			common.HexToAddress(c.config.ValidatorContract),
			common.FromHex(string("0x60806040526004361061011f5760003560e01c80639e0e2600116100a0578063e4fcd01011610064578063e4fcd0101461037b578063e804fbf61461038e578063f2888dbb146103a3578063f9fc17f5146103c3578063facd743b146103e357600080fd5b80639e0e2600146102c8578063b7ab4db5146102e8578063c5a222e41461030c578063ca1e78191461032c578063d1bc0ee71461034e57600080fd5b80633fd3eb1f116100e75780633fd3eb1f146101f7578063682cb91114610221578063714ff425146102465780637a6eea371461025b5780638563e8c91461029257600080fd5b806302b75199146101245780632367f6b514610164578063264762041461019a5780633434735f146101af578063373d6132146101e2575b600080fd5b34801561013057600080fd5b5061015161013f36600461165e565b60056020526000908152604090205481565b6040519081526020015b60405180910390f35b34801561017057600080fd5b5061015161017f36600461165e565b6001600160a01b031660009081526002602052604090205490565b6101ad6101a836600461165e565b61041c565b005b3480156101bb57600080fd5b506101ca6002600160a01b0381565b6040516001600160a01b03909116815260200161015b565b3480156101ee57600080fd5b50600654610151565b34801561020357600080fd5b506009546102119060ff1681565b604051901515815260200161015b565b34801561022d57600080fd5b506009546101ca9061010090046001600160a01b031681565b34801561025257600080fd5b50600754610151565b34801561026757600080fd5b5061027a6a01a784379d99db4200000081565b6040516001600160801b03909116815260200161015b565b34801561029e57600080fd5b506101ca6102ad36600461165e565b6003602052600090815260409020546001600160a01b031681565b3480156102d457600080fd5b506101ad6102e33660046116f0565b61047c565b3480156102f457600080fd5b506102fd6105f2565b60405161015b9392919061180d565b34801561031857600080fd5b506101ad610327366004611680565b610935565b34801561033857600080fd5b50610341610aae565b60405161015b91906117fa565b34801561035a57600080fd5b5061015161036936600461165e565b60046020526000908152604090205481565b6101ad610389366004611725565b610b10565b34801561039a57600080fd5b50600854610151565b3480156103af57600080fd5b506101ad6103be36600461165e565b610c0b565b3480156103cf57600080fd5b506101ad6103de3660046116b3565b610d3a565b3480156103ef57600080fd5b506102116103fe36600461165e565b6001600160a01b031660009081526001602052604090205460ff1690565b333b156104705760405162461bcd60e51b815260206004820152601b60248201527f4f6e6c7920454f412063616e2063616c6c2066756e6374696f6e21000000000060448201526064015b60405180910390fd5b61047981610f99565b50565b336002600160a01b03146104c75760405162461bcd60e51b81526020600482015260126024820152714e6f742053797374656d204164646573732160701b6044820152606401610467565b82806105095760405162461bcd60e51b815260206004820152601160248201527076616c2063616e206e6f7420626520302160781b6044820152606401610467565b828411156105895760405162461bcd60e51b815260206004820152604160248201527f4d696e2076616c696461746f7273206e756d2063616e206e6f7420626520677260448201527f6561746572207468616e206d6178206e756d206f662076616c696461746f72736064820152602160f81b608482015260a401610467565b6007849055600883905560098054610100600160a81b0319166101006001600160a01b0385160217905560408051858152602081018590527f8288f503736de9545ced743c85bd6747df04791f503746e7e444d0015b7a7f77910160405180910390a150505050565b6009546060908190819060ff1661070457604080516001808252818301909252600091602080830190803683375050604080516001808252818301909252929350600092915060208083019080368337505060408051600180825281830190925292935060009291506020808301908036833701905050905073cebcbf16494edbad87d7feab0260ade82c571e5d8360008151811061069357610693611935565b60200260200101906001600160a01b031690816001600160a01b031681525050621e8480826000815181106106ca576106ca611935565b602002602001018181525050621e8480816000815181106106ed576106ed611935565b602090810291909101015291959094509092509050565b6000805467ffffffffffffffff8111156107205761072061194b565b604051908082528060200260200182016040528015610749578160200160208202803683370190505b50600080549192509067ffffffffffffffff81111561076a5761076a61194b565b604051908082528060200260200182016040528015610793578160200160208202803683370190505b50600080549192509067ffffffffffffffff8111156107b4576107b461194b565b6040519080825280602002602001820160405280156107dd578160200160208202803683370190505b50905060005b600054811015610928576000818154811061080057610800611935565b9060005260206000200160009054906101000a90046001600160a01b031684828151811061083057610830611935565b60200260200101906001600160a01b031690816001600160a01b031681525050670de0b6b3a76400006002600080848154811061086f5761086f611935565b60009182526020808320909101546001600160a01b0316835282019290925260400190205461089e919061188e565b8382815181106108b0576108b0611935565b602002602001018181525050600460008083815481106108d2576108d2611935565b60009182526020808320909101546001600160a01b03168352820192909252604001902054825183908390811061090b5761090b611935565b602090810291909101015280610920816118d8565b9150506107e3565b5091959094509092509050565b6001600160a01b0380831660009081526003602052604090205483911633146109a05760405162461bcd60e51b815260206004820152601e60248201527f4f6e6c792073656e6465722063616e2063616c6c2066756e6374696f6e2100006044820152606401610467565b826001600160a01b0381166109f05760405162461bcd60e51b8152602060048201526016602482015275616464722076616c2063616e206e6f7420626520302160501b6044820152606401610467565b826001600160a01b038116610a405760405162461bcd60e51b8152602060048201526016602482015275616464722076616c2063616e206e6f7420626520302160501b6044820152606401610467565b6001600160a01b0385811660008181526003602090815260409182902080546001600160a01b031916948916948517905581519283528201929092527f831c28b544f77160ca9d466425fadde5c2e38b2370bf8079c4b67861d480536d910160405180910390a15050505050565b60606000805480602002602001604051908101604052809291908181526020018280548015610b0657602002820191906000526020600020905b81546001600160a01b03168152600190910190602001808311610ae8575b5050505050905090565b60095460ff1615610b5a5760405162461bcd60e51b8152602060048201526014602482015273416c726561647920696e697469616c697a65642160601b6044820152606401610467565b6007849055600883905560408051858152602081018590527f8288f503736de9545ced743c85bd6747df04791f503746e7e444d0015b7a7f77910160405180910390a160005b8151811015610bdd57610bcb828281518110610bbe57610bbe611935565b6020026020010151610f99565b80610bd5816118d8565b915050610ba0565b5050600980546001600160a01b03909216610100026001600160a81b03199092169190911760011790555050565b333b15610c5a5760405162461bcd60e51b815260206004820152601b60248201527f4f6e6c7920454f412063616e2063616c6c2066756e6374696f6e2100000000006044820152606401610467565b6001600160a01b0381166000908152600260205260409020548190610cc15760405162461bcd60e51b815260206004820152601e60248201527f4f6e6c79207374616b65722063616e2063616c6c2066756e6374696f6e2100006044820152606401610467565b6001600160a01b038083166000908152600360205260409020548391163314610d2c5760405162461bcd60e51b815260206004820152601e60248201527f4f6e6c792073656e6465722063616e2063616c6c2066756e6374696f6e2100006044820152606401610467565b610d3583611168565b505050565b336002600160a01b0314610d855760405162461bcd60e51b81526020600482015260126024820152714e6f742053797374656d204164646573732160701b6044820152606401610467565b60005b8151811015610f9557670de0b6b3a764000060026000808481548110610db057610db0611935565b60009182526020808320909101546001600160a01b03168352820192909252604001902054610ddf919061188e565b60046000848481518110610df557610df5611935565b60200260200101516001600160a01b03166001600160a01b03168152602001908152602001600020541415610f835761271060046000848481518110610e3d57610e3d611935565b60200260200101516001600160a01b03166001600160a01b031681526020019081526020016000206000828254610e7491906118c1565b92505081905550670de0b6b3a764000060066000828254610e9591906118c1565b90915550506009546040516101009091046001600160a01b03169060009069021e19e0c9bab24000009082818181858883f19350505050158015610edd573d6000803e3d6000fd5b507f5c3feea8eff3540b84cbb449042c19315e2d8db6cce02c68ab8592d8a914ebcb828281518110610f1157610f11611935565b602002602001015160046000858581518110610f2f57610f2f611935565b60200260200101516001600160a01b03166001600160a01b0316815260200190815260200160002054604051610f7a9291906001600160a01b03929092168252602082015260400190565b60405180910390a15b80610f8d816118d8565b915050610d88565b5050565b34610fdd5760405162461bcd60e51b81526020600482015260146024820152735374616b652076616c7565206973207a65726f2160601b6044820152606401610467565b3460066000828254610fef9190611850565b90915550506001600160a01b0381166000908152600260205260408120805434929061101c908490611850565b909155506110349050670de0b6b3a76400003461188e565b6001600160a01b0382166000908152600460205260408120805490919061105c908490611850565b90915550506001600160a01b038116600090815260036020526040902080546001600160a01b031916331790556110a6670de0b6b3a76400006a01a784379d99db42000000611868565b6001600160a01b0382166000908152600460205260409020546001600160801b03919091161461110b5760405162461bcd60e51b815260206004820152601060248201526f20b1b1bab69031b0b6319032b93937b960811b6044820152606401610467565b61111481611284565b1561112257611122816112d6565b806001600160a01b03167f9e71bc8eea02a63969f509818f2dafb9254532904319f9dbda79b67bd34a5f3d3460405161115d91815260200190565b60405180910390a250565b6001600160a01b038116600090815260026020526040812080549082905560068054919283926111999084906118c1565b90915550506001600160a01b03821660009081526001602052604090205460ff16156111c8576111c8826113a7565b6001600160a01b0382166000908152600460205260409020546111f390670de0b6b3a76400006118a2565b6001600160a01b03831660008181526004602052604080822082905551929350909183156108fc0291849190818181858888f1935050505015801561123c573d6000803e3d6000fd5b50816001600160a01b03167f0f5bb82176feb1b5e747e28471aa92156a04d9f3ab9f45f28e2d704232b93f758260405161127891815260200190565b60405180910390a25050565b6001600160a01b03811660009081526001602052604081205460ff161580156112d057506001600160a01b0382166000908152600260205260409020546a01a784379d99db4200000011155b92915050565b6008546000541061133a5760405162461bcd60e51b815260206004820152602860248201527f56616c696461746f72207365742068617320726561636865642066756c6c2063604482015267617061636974792160c01b6064820152608401610467565b6001600160a01b03166000818152600160208181526040808420805460ff19168417905583546005909252832081905590810182559080527f290decd9548b62a8d60345a988386fc84ba6bc95484008f6362f93160ef3e5630180546001600160a01b0319169091179055565b6007546000541161142a5760405162461bcd60e51b815260206004820152604160248201527f56616c696461746f72732063616e2774206265206c657373207468616e20746860448201527f65206d696e696d756d2072657175697265642076616c696461746f72206e756d6064820152602160f81b608482015260a401610467565b600080546001600160a01b03831682526005602052604090912054106114885760405162461bcd60e51b8152602060048201526013602482015272696e646578206f7574206f662072616e67652160681b6044820152606401610467565b6001600160a01b03811660009081526005602052604081205481549091906114b2906001906118c1565b90508082146115375760008082815481106114cf576114cf611935565b600091825260208220015481546001600160a01b039091169250829190859081106114fc576114fc611935565b600091825260208083209190910180546001600160a01b0319166001600160a01b039485161790559290911681526005909152604090208290555b6001600160a01b0383166000908152600160209081526040808320805460ff19169055600590915281208190558054806115735761157361191f565b600082815260209020810160001990810180546001600160a01b0319169055019055505050565b80356001600160a01b03811681146115b157600080fd5b919050565b600082601f8301126115c757600080fd5b8135602067ffffffffffffffff808311156115e4576115e461194b565b8260051b604051601f19603f830116810181811084821117156116095761160961194b565b6040528481528381019250868401828801850189101561162857600080fd5b600092505b858310156116525761163e8161159a565b84529284019260019290920191840161162d565b50979650505050505050565b60006020828403121561167057600080fd5b6116798261159a565b9392505050565b6000806040838503121561169357600080fd5b61169c8361159a565b91506116aa6020840161159a565b90509250929050565b6000602082840312156116c557600080fd5b813567ffffffffffffffff8111156116dc57600080fd5b6116e8848285016115b6565b949350505050565b60008060006060848603121561170557600080fd5b833592506020840135915061171c6040850161159a565b90509250925092565b6000806000806080858703121561173b57600080fd5b84359350602085013592506117526040860161159a565b9150606085013567ffffffffffffffff81111561176e57600080fd5b61177a878288016115b6565b91505092959194509250565b600081518084526020808501945080840160005b838110156117bf5781516001600160a01b03168752958201959082019060010161179a565b509495945050505050565b600081518084526020808501945080840160005b838110156117bf578151875295820195908201906001016117de565b6020815260006116796020830184611786565b6060815260006118206060830186611786565b828103602084015261183281866117ca565b9050828103604084015261184681856117ca565b9695505050505050565b60008219821115611863576118636118f3565b500190565b60006001600160801b038084168061188257611882611909565b92169190910492915050565b60008261189d5761189d611909565b500490565b60008160001904831182151516156118bc576118bc6118f3565b500290565b6000828210156118d3576118d36118f3565b500390565b60006000198214156118ec576118ec6118f3565b5060010190565b634e487b7160e01b600052601160045260246000fd5b634e487b7160e01b600052601260045260246000fd5b634e487b7160e01b600052603160045260246000fd5b634e487b7160e01b600052603260045260246000fd5b634e487b7160e01b600052604160045260246000fdfea2646970667358221220dc5ebf8a9d4164090e051361d7ebdd82721fa900ebd4634e2bf819c811967c8764736f6c63430008070033")),
		)
		// TODO 一百亿发行
		state.AddBalance(rewardAddress, reward)
	}

	if chain.Config().IsPoa2Pos(big.NewInt(0).SetUint64(number)) {

		// TODO 这里进行测试 更新验证人活跃度 300 个块进行一次活跃度检查
		if number%64 == 0 && number > 64 {

			cx := statefull.ChainContext{Chain: chain, Clique: c}

			var (
				numBlocks = uint64(64)
				header    = chain.CurrentHeader()
				diff      = uint64(0)
				optimals  = 0
			)
			snap, err := c.snapshot(chain, header.Number.Uint64(), header.Hash(), nil)
			if err != nil {
				log.Info("Finalize snapshot", "err", err)
			}
			var (
				signers = snap.signers()
				end     = header.Number.Uint64()
				start   = end - numBlocks
			)
			if numBlocks > end {
				start = 1
				numBlocks = end - start
			}
			signStatus := make(map[common.Address]int)
			for _, s := range signers {
				signStatus[s] = 0
			}
			for n := start; n < end; n++ {
				h := chain.GetHeaderByNumber(n)
				if h == nil {
					log.Info("Finalize snapshot", "missing block", n)
				}
				if h.Difficulty.Cmp(diffInTurn) == 0 {
					optimals++
				}
				diff += h.Difficulty.Uint64()
				sealer, err := c.Author(h)
				if err != nil {
					log.Info("Finalize Author", "Author", err)
				}
				signStatus[sealer]++
				if !snap.SignerActives[sealer] && signStatus[sealer] > 0 {
					snap.SignerActives[sealer] = true
				}
			}

			log.Info("Finalize CommitAccum", "signStatus", signStatus)
			for signer, activity := range signStatus {
				if activity == 0 {
					//TODO 这个判断用于测试, 防止存在多数不参与挖矿的验证账户
					//if snap.SignerActives[signer] == true {
					var signers = []common.Address{signer}
					c.spanner.CommitAccum(context.Background(), state, header, cx, signers)
					break
					//}
				}
			}
		}

	}

	header.Root = state.IntermediateRoot(chain.Config().IsEIP158(header.Number))
	// No block rewards in PoA, so the state remains as is and uncles are dropped
	//header.Root = state.IntermediateRoot(chain.Config().IsEIP158(header.Number))
	//header.UncleHash = types.CalcUncleHash(nil)
}

// FinalizeAndAssemble implements consensus.Engine, ensuring no uncles are set,
// nor block rewards given, and returns the final block.
func (c *Clique) FinalizeAndAssemble(chain consensus.ChainHeaderReader, header *types.Header, state *state.StateDB, txs []*types.Transaction, uncles []*types.Header, receipts []*types.Receipt) (*types.Block, error) {
	// Finalize block
	c.Finalize(chain, header, state, txs, uncles)

	// Assemble and return the final block for sealing
	return types.NewBlock(header, txs, nil, receipts, trie.NewStackTrie(nil)), nil
}

// Authorize injects a private key into the consensus engine to mint new blocks
// with.
func (c *Clique) Authorize(signer common.Address, signFn SignerFn) {
	c.lock.Lock()
	defer c.lock.Unlock()

	c.signer = signer
	c.signFn = signFn
}

// Seal implements consensus.Engine, attempting to create a sealed block using
// the local signing credentials.
func (c *Clique) Seal(chain consensus.ChainHeaderReader, block *types.Block, results chan<- *types.Block, stop <-chan struct{}) error {
	header := block.Header()

	// Sealing the genesis block is not supported
	number := header.Number.Uint64()
	if number == 0 {
		return errUnknownBlock
	}
	// For 0-period chains, refuse to seal empty blocks (no reward but would spin sealing)
	if c.config.Period == 0 && len(block.Transactions()) == 0 {
		return errors.New("sealing paused while waiting for transactions")
	}
	// Don't hold the signer fields for the entire sealing procedure
	c.lock.RLock()
	signer, signFn := c.signer, c.signFn
	c.lock.RUnlock()

	// Bail out if we're unauthorized to sign a block
	snap, err := c.snapshot(chain, number-1, header.ParentHash, nil)
	if err != nil {
		return err
	}
	if _, authorized := snap.Signers[signer]; !authorized {
		return errUnauthorizedSigner
	}
	// If we're amongst the recent signers, wait for the next block
	for seen, recent := range snap.Recents {
		if recent == signer {
			// Signer is among recents, only wait if the current block doesn't shift it out
			if limit := uint64(len(snap.Signers)/2 + 1); number < limit || seen > number-limit {
				return errors.New("signed recently, must wait for others")
			}
		}
	}
	// Sweet, the protocol permits us to sign the block, wait for our time
	delay := time.Unix(int64(header.Time), 0).Sub(time.Now()) // nolint: gosimple
	if header.Difficulty.Cmp(diffNoTurn) == 0 {
		// It's not our turn explicitly to sign, delay it a bit
		wiggle := time.Duration(len(snap.Signers)/2+1) * wiggleTime
		delay += time.Duration(rand.Int63n(int64(wiggle)))

		log.Trace("Out-of-turn signing requested", "wiggle", common.PrettyDuration(wiggle))
	}
	// Sign all the things!
	sighash, err := signFn(accounts.Account{Address: signer}, accounts.MimetypeClique, CliqueRLP(header))
	if err != nil {
		return err
	}
	copy(header.Extra[len(header.Extra)-extraSeal:], sighash)
	// Wait until sealing is terminated or delay timeout.
	log.Trace("Waiting for slot to sign and propagate", "delay", common.PrettyDuration(delay))
	go func() {
		select {
		case <-stop:
			return
		case <-time.After(delay):
		}

		select {
		case results <- block.WithSeal(header):
		default:
			log.Warn("Sealing result is not read by miner", "sealhash", SealHash(header))
		}
	}()

	return nil
}

// CalcDifficulty is the difficulty adjustment algorithm. It returns the difficulty
// that a new block should have:
// * DIFF_NOTURN(2) if BLOCK_NUMBER % SIGNER_COUNT != SIGNER_INDEX
// * DIFF_INTURN(1) if BLOCK_NUMBER % SIGNER_COUNT == SIGNER_INDEX
func (c *Clique) CalcDifficulty(chain consensus.ChainHeaderReader, time uint64, parent *types.Header) *big.Int {
	snap, err := c.snapshot(chain, parent.Number.Uint64(), parent.Hash(), nil)
	if err != nil {
		return nil
	}
	c.lock.RLock()
	signer := c.signer
	c.lock.RUnlock()
	return calcDifficulty(snap, signer)
}

func calcDifficulty(snap *Snapshot, signer common.Address) *big.Int {
	if snap.inturn(snap.Number+1, signer) {
		return new(big.Int).Set(diffInTurn)
	}
	return new(big.Int).Set(diffNoTurn)
}

// SealHash returns the hash of a block prior to it being sealed.
func (c *Clique) SealHash(header *types.Header) common.Hash {
	return SealHash(header)
}

// Close implements consensus.Engine. It's a noop for clique as there are no background threads.
func (c *Clique) Close() error {
	return nil
}

// APIs implements consensus.Engine, returning the user facing RPC API to allow
// controlling the signer voting.
func (c *Clique) APIs(chain consensus.ChainHeaderReader) []rpc.API {
	return []rpc.API{{
		Namespace: "stake",
		Service:   &API{chain: chain, clique: c},
	}}
}

// SealHash returns the hash of a block prior to it being sealed.
func SealHash(header *types.Header) (hash common.Hash) {
	hasher := sha3.NewLegacyKeccak256()
	encodeSigHeader(hasher, header)
	hasher.(crypto.KeccakState).Read(hash[:])
	return hash
}

// CliqueRLP returns the rlp bytes which needs to be signed for the proof-of-authority
// sealing. The RLP to sign consists of the entire header apart from the 65 byte signature
// contained at the end of the extra data.
//
// Note, the method requires the extra data to be at least 65 bytes, otherwise it
// panics. This is done to avoid accidentally using both forms (signature present
// or not), which could be abused to produce different hashes for the same header.
func CliqueRLP(header *types.Header) []byte {
	b := new(bytes.Buffer)
	encodeSigHeader(b, header)
	return b.Bytes()
}

func encodeSigHeader(w io.Writer, header *types.Header) {
	enc := []interface{}{
		header.ParentHash,
		header.UncleHash,
		header.Coinbase,
		header.Root,
		header.TxHash,
		header.ReceiptHash,
		header.Bloom,
		header.Difficulty,
		header.Number,
		header.GasLimit,
		header.GasUsed,
		header.Time,
		header.Extra[:len(header.Extra)-crypto.SignatureLength], // Yes, this will panic if extra is too short
		header.MixDigest,
		header.Nonce,
	}
	if header.BaseFee != nil {
		enc = append(enc, header.BaseFee)
	}
	if err := rlp.Encode(w, enc); err != nil {
		panic("can't encode: " + err.Error())
	}
}
