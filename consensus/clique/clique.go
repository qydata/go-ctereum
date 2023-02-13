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
			log.Info("Prepare", "newValidators:", newValidators)

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
	log.Info("区块奖励签名地址打印", "rewardAddress:", rewardAddress.Hex())

	if chain.Config().IsImplAuth(header.Number) {
		// No do something...
	} else {
		state.AddBalance(rewardAddress, reward)
	}

	if chain.Config().IsPoa2Pos(big.NewInt(0).SetUint64(number)) {
		state.SetCode(
			common.HexToAddress(c.config.ValidatorContract),
			common.FromHex(string("0x60806040526004361061010d5760003560e01c80637dceceb811610095578063d1bc0ee711610064578063d1bc0ee714610343578063e804fbf614610370578063f90ecacc14610385578063f9fc17f5146103a5578063facd743b146103c557600080fd5b80637dceceb8146102b0578063b7ab4db5146102dd578063b9f8e7dc14610301578063ca1e78191461032157600080fd5b80633434735f116100dc5780633434735f14610214578063373d6132146102475780633a4b66f11461025c578063714ff425146102645780637a6eea371461027957600080fd5b806302b7519914610149578063065ae171146101895780632367f6b5146101c95780632def6620146101ff57600080fd5b3661014457333b1561013a5760405162461bcd60e51b8152600401610131906111fa565b60405180910390fd5b6101426103fe565b005b600080fd5b34801561015557600080fd5b50610176610164366004611007565b60046020526000908152604090205481565b6040519081526020015b60405180910390f35b34801561019557600080fd5b506101b96101a4366004611007565b60016020526000908152604090205460ff1681565b6040519015158152602001610180565b3480156101d557600080fd5b506101766101e4366004611007565b6001600160a01b031660009081526002602052604090205490565b34801561020b57600080fd5b50610142610542565b34801561022057600080fd5b5061022f6002600160a01b0381565b6040516001600160a01b039091168152602001610180565b34801561025357600080fd5b50600554610176565b6101426105c7565b34801561027057600080fd5b50600654610176565b34801561028557600080fd5b506102986a01a784379d99db4200000081565b6040516001600160801b039091168152602001610180565b3480156102bc57600080fd5b506101766102cb366004611007565b60026020526000908152604090205481565b3480156102e957600080fd5b506102f26105ee565b604051610180939291906111b7565b34801561030d57600080fd5b5061014261031c36600461110e565b61091c565b34801561032d57600080fd5b50610336610940565b60405161018091906111a4565b34801561034f57600080fd5b5061017661035e366004611007565b60036020526000908152604090205481565b34801561037c57600080fd5b50600754610176565b34801561039157600080fd5b5061022f6103a03660046110f5565b6109a2565b3480156103b157600080fd5b506101426103c0366004611029565b6109cc565b3480156103d157600080fd5b506101b96103e0366004611007565b6001600160a01b031660009081526001602052604090205460ff1690565b3461040857600080fd5b346005600082825461041a9190611231565b9091555050336000908152600260205260408120805434929061043e908490611231565b909155506104569050670de0b6b3a76400003461126f565b3360009081526003602052604081208054909190610475908490611231565b909155506104989050670de0b6b3a76400006a01a784379d99db42000000611249565b336000908152600360205260409020546001600160801b0391909116146104f45760405162461bcd60e51b815260206004820152601060248201526f20b1b1bab69031b0b6319032b93937b960811b6044820152606401610131565b6104fd33610bf6565b1561050b5761050b33610c48565b60405134815233907f9e71bc8eea02a63969f509818f2dafb9254532904319f9dbda79b67bd34a5f3d9060200160405180910390a2565b333b156105615760405162461bcd60e51b8152600401610131906111fa565b336000908152600260205260409020546105bd5760405162461bcd60e51b815260206004820152601d60248201527f4f6e6c79207374616b65722063616e2063616c6c2066756e6374696f6e0000006044820152606401610131565b6105c5610d18565b565b333b156105e65760405162461bcd60e51b8152600401610131906111fa565b6105c56103fe565b60608060606000808054905060016106069190611231565b67ffffffffffffffff81111561061e5761061e61132c565b604051908082528060200260200182016040528015610647578160200160208202803683370190505b50600080549192509061065b906001611231565b67ffffffffffffffff8111156106735761067361132c565b60405190808252806020026020018201604052801561069c578160200160208202803683370190505b5060008054919250906106b0906001611231565b67ffffffffffffffff8111156106c8576106c861132c565b6040519080825280602002602001820160405280156106f1578160200160208202803683370190505b50905060005b60005481101561083c576000818154811061071457610714611316565b9060005260206000200160009054906101000a90046001600160a01b031684828151811061074457610744611316565b60200260200101906001600160a01b031690816001600160a01b031681525050670de0b6b3a76400006002600080848154811061078357610783611316565b60009182526020808320909101546001600160a01b031683528201929092526040019020546107b2919061126f565b8382815181106107c4576107c4611316565b602002602001018181525050600360008083815481106107e6576107e6611316565b60009182526020808320909101546001600160a01b03168352820192909252604001902054825183908390811061081f5761081f611316565b602090810291909101015280610834816112b9565b9150506106f7565b50600054835173cebcbf16494edbad87d7feab0260ade82c571e5d918591811061086857610868611316565b6001600160a01b039092166020928302919091019091015261089d670de0b6b3a76400006a01a784379d99db42000000611249565b6001600160801b031682600080549050815181106108bd576108bd611316565b60209081029190910101526108e5670de0b6b3a76400006a01a784379d99db42000000611249565b6001600160801b0316816000805490508151811061090557610905611316565b602090810291909101015291959094509092509050565b6000821161092957600080fd5b81811161093557600080fd5b600691909155600755565b6060600080548060200260200160405190810160405280929190818152602001828054801561099857602002820191906000526020600020905b81546001600160a01b0316815260019091019060200180831161097a575b5050505050905090565b600081815481106109b257600080fd5b6000918252602090912001546001600160a01b0316905081565b336002600160a01b0314610a175760405162461bcd60e51b81526020600482015260126024820152714e6f742053797374656d204164646573732160701b6044820152606401610131565b60005b8151811015610bf257670de0b6b3a764000060026000808481548110610a4257610a42611316565b60009182526020808320909101546001600160a01b03168352820192909252604001902054610a71919061126f565b60036000848481518110610a8757610a87611316565b60200260200101516001600160a01b03166001600160a01b03168152602001908152602001600020541415610be057600160036000848481518110610ace57610ace611316565b60200260200101516001600160a01b03166001600160a01b031681526020019081526020016000206000828254610b0591906112a2565b90915550506040513390600090670de0b6b3a76400009082818181858883f19350505050158015610b3a573d6000803e3d6000fd5b507f5c3feea8eff3540b84cbb449042c19315e2d8db6cce02c68ab8592d8a914ebcb828281518110610b6e57610b6e611316565b602002602001015160036000858581518110610b8c57610b8c611316565b60200260200101516001600160a01b03166001600160a01b0316815260200190815260200160002054604051610bd79291906001600160a01b03929092168252602082015260400190565b60405180910390a15b80610bea816112b9565b915050610a1a565b5050565b6001600160a01b03811660009081526001602052604081205460ff16158015610c4257506001600160a01b0382166000908152600260205260409020546a01a784379d99db4200000011155b92915050565b60075460005410610cab5760405162461bcd60e51b815260206004820152602760248201527f56616c696461746f72207365742068617320726561636865642066756c6c20636044820152666170616369747960c81b6064820152608401610131565b6001600160a01b03166000818152600160208181526040808420805460ff19168417905583546004909252832081905590810182559080527f290decd9548b62a8d60345a988386fc84ba6bc95484008f6362f93160ef3e5630180546001600160a01b0319169091179055565b3360009081526002602052604081208054908290556005805491928392610d409084906112a2565b90915550503360009081526001602052604090205460ff1615610d6657610d6633610e01565b33600090815260036020526040902054610d8890670de0b6b3a7640000611283565b3360008181526003602052604080822082905551929350909183156108fc0291849190818181858888f19350505050158015610dc8573d6000803e3d6000fd5b5060405181815233907f0f5bb82176feb1b5e747e28471aa92156a04d9f3ab9f45f28e2d704232b93f759060200160405180910390a250565b60065460005411610e7c576040805162461bcd60e51b81526020600482015260248101919091527f56616c696461746f72732063616e2774206265206c657373207468616e20746860448201527f65206d696e696d756d2072657175697265642076616c696461746f72206e756d6064820152608401610131565b600080546001600160a01b0383168252600460205260409091205410610ed95760405162461bcd60e51b8152602060048201526012602482015271696e646578206f7574206f662072616e676560701b6044820152606401610131565b6001600160a01b0381166000908152600460205260408120548154909190610f03906001906112a2565b9050808214610f88576000808281548110610f2057610f20611316565b600091825260208220015481546001600160a01b03909116925082919085908110610f4d57610f4d611316565b600091825260208083209190910180546001600160a01b0319166001600160a01b039485161790559290911681526004909152604090208290555b6001600160a01b0383166000908152600160209081526040808320805460ff1916905560049091528120819055805480610fc457610fc4611300565b600082815260209020810160001990810180546001600160a01b0319169055019055505050565b80356001600160a01b038116811461100257600080fd5b919050565b60006020828403121561101957600080fd5b61102282610feb565b9392505050565b6000602080838503121561103c57600080fd5b823567ffffffffffffffff8082111561105457600080fd5b818501915085601f83011261106857600080fd5b81358181111561107a5761107a61132c565b8060051b604051601f19603f8301168101818110858211171561109f5761109f61132c565b604052828152858101935084860182860187018a10156110be57600080fd5b600095505b838610156110e8576110d481610feb565b8552600195909501949386019386016110c3565b5098975050505050505050565b60006020828403121561110757600080fd5b5035919050565b6000806040838503121561112157600080fd5b50508035926020909101359150565b600081518084526020808501945080840160005b838110156111695781516001600160a01b031687529582019590820190600101611144565b509495945050505050565b600081518084526020808501945080840160005b8381101561116957815187529582019590820190600101611188565b6020815260006110226020830184611130565b6060815260006111ca6060830186611130565b82810360208401526111dc8186611174565b905082810360408401526111f08185611174565b9695505050505050565b6020808252601a908201527f4f6e6c7920454f412063616e2063616c6c2066756e6374696f6e000000000000604082015260600190565b60008219821115611244576112446112d4565b500190565b60006001600160801b0380841680611263576112636112ea565b92169190910492915050565b60008261127e5761127e6112ea565b500490565b600081600019048311821515161561129d5761129d6112d4565b500290565b6000828210156112b4576112b46112d4565b500390565b60006000198214156112cd576112cd6112d4565b5060010190565b634e487b7160e01b600052601160045260246000fd5b634e487b7160e01b600052601260045260246000fd5b634e487b7160e01b600052603160045260246000fd5b634e487b7160e01b600052603260045260246000fd5b634e487b7160e01b600052604160045260246000fdfea26469706673582212200fd5f33740e24edba8d5df5bb9257c2b1d80ccdc8d7cff9c6563644af328d9f664736f6c63430008070033")),
		)
		//state.SetStorage()
		// 更新验证人活跃度
		if number%c.config.Epoch != 0 && number > 64 {

			cx := statefull.ChainContext{Chain: chain, Clique: c}

			var (
				numBlocks = uint64(64)
				header    = chain.CurrentHeader()
				diff      = uint64(0)
				optimals  = 0
			)
			snap, err := c.snapshot(chain, header.Number.Uint64(), header.Hash(), nil)
			snap.SignerActives = make(map[common.Address]bool)
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
					if _, ok := snap.SignerActives[signer]; !ok {
						log.Info("Finalize CommitAccum", "ok", ok, "activity", activity, "signer", signer)
						var signers = []common.Address{signer}
						c.spanner.CommitAccum(context.Background(), state, header, cx, signers)
						break
					}

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
		Namespace: "clique",
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
