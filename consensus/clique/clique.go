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

	if 1 != number {
		if !chain.Config().IsImplAuth(header.Number) {
			log.Info("区块奖励签名地址打印", "rewardAddress:", rewardAddress.Hex())
			state.AddBalance(rewardAddress, reward)
		}
	}

	//if chain.Config().Poa2PosBlock == big.NewInt(0).SetUint64(number) {
	if (header.Number.Int64() + 1) == c.config.Poa2PosBlock {
		state.SetCode(
			common.HexToAddress(c.config.ValidatorContract),
			common.FromHex(string("0x6080604052600436106101145760003560e01c80638563e8c9116100a0578063d1bc0ee711610064578063d1bc0ee714610331578063e804fbf61461035e578063f2888dbb14610373578063f9fc17f514610393578063facd743b146103b357600080fd5b80638563e8c914610275578063b7ab4db5146102ab578063b9f8e7dc146102cf578063c5a222e4146102ef578063ca1e78191461030f57600080fd5b80633434735f116100e75780633434735f146101b7578063373d6132146101ea5780633fd3eb1f146101ff578063714ff425146102295780637a6eea371461023e57600080fd5b806302b75199146101195780630fbf5d92146101595780632367f6b51461016e57806326476204146101a4575b600080fd5b34801561012557600080fd5b506101466101343660046115e3565b60056020526000908152604090205481565b6040519081526020015b60405180910390f35b61016c610167366004611697565b6103ec565b005b34801561017a57600080fd5b506101466101893660046115e3565b6001600160a01b031660009081526002602052604090205490565b61016c6101b23660046115e3565b6104d1565b3480156101c357600080fd5b506101d26002600160a01b0381565b6040516001600160a01b039091168152602001610150565b3480156101f657600080fd5b50600654610146565b34801561020b57600080fd5b506009546102199060ff1681565b6040519015158152602001610150565b34801561023557600080fd5b50600754610146565b34801561024a57600080fd5b5061025d6a01a784379d99db4200000081565b6040516001600160801b039091168152602001610150565b34801561028157600080fd5b506101d26102903660046115e3565b6003602052600090815260409020546001600160a01b031681565b3480156102b757600080fd5b506102c061052c565b6040516101509392919061176e565b3480156102db57600080fd5b5061016c6102ea366004611675565b61086f565b3480156102fb57600080fd5b5061016c61030a366004611605565b6109c4565b34801561031b57600080fd5b50610324610b3d565b604051610150919061175b565b34801561033d57600080fd5b5061014661034c3660046115e3565b60046020526000908152604090205481565b34801561036a57600080fd5b50600854610146565b34801561037f57600080fd5b5061016c61038e3660046115e3565b610b9f565b34801561039f57600080fd5b5061016c6103ae366004611638565b610cce565b3480156103bf57600080fd5b506102196103ce3660046115e3565b6001600160a01b031660009081526001602052604090205460ff1690565b60095460ff161561043b5760405162461bcd60e51b8152602060048201526014602482015273416c726561647920696e697469616c697a65642160601b60448201526064015b60405180910390fd5b6007839055600882905560408051848152602081018490527f8288f503736de9545ced743c85bd6747df04791f503746e7e444d0015b7a7f77910160405180910390a160005b81518110156104be576104ac82828151811061049f5761049f611896565b6020026020010151610f1e565b806104b681611839565b915050610481565b50506009805460ff191660011790555050565b333b156105205760405162461bcd60e51b815260206004820152601b60248201527f4f6e6c7920454f412063616e2063616c6c2066756e6374696f6e2100000000006044820152606401610432565b61052981610f1e565b50565b6009546060908190819060ff1661063e57604080516001808252818301909252600091602080830190803683375050604080516001808252818301909252929350600092915060208083019080368337505060408051600180825281830190925292935060009291506020808301908036833701905050905073cebcbf16494edbad87d7feab0260ade82c571e5d836000815181106105cd576105cd611896565b60200260200101906001600160a01b031690816001600160a01b031681525050621e84808260008151811061060457610604611896565b602002602001018181525050621e84808160008151811061062757610627611896565b602090810291909101015291959094509092509050565b6000805467ffffffffffffffff81111561065a5761065a6118ac565b604051908082528060200260200182016040528015610683578160200160208202803683370190505b50600080549192509067ffffffffffffffff8111156106a4576106a46118ac565b6040519080825280602002602001820160405280156106cd578160200160208202803683370190505b50600080549192509067ffffffffffffffff8111156106ee576106ee6118ac565b604051908082528060200260200182016040528015610717578160200160208202803683370190505b50905060005b600054811015610862576000818154811061073a5761073a611896565b9060005260206000200160009054906101000a90046001600160a01b031684828151811061076a5761076a611896565b60200260200101906001600160a01b031690816001600160a01b031681525050670de0b6b3a7640000600260008084815481106107a9576107a9611896565b60009182526020808320909101546001600160a01b031683528201929092526040019020546107d891906117ef565b8382815181106107ea576107ea611896565b6020026020010181815250506004600080838154811061080c5761080c611896565b60009182526020808320909101546001600160a01b03168352820192909252604001902054825183908390811061084557610845611896565b60209081029190910101528061085a81611839565b91505061071d565b5091959094509092509050565b336002600160a01b03146108ba5760405162461bcd60e51b81526020600482015260126024820152714e6f742053797374656d204164646573732160701b6044820152606401610432565b81806108fc5760405162461bcd60e51b815260206004820152601160248201527076616c2063616e206e6f7420626520302160781b6044820152606401610432565b8183111561097c5760405162461bcd60e51b815260206004820152604160248201527f4d696e2076616c696461746f7273206e756d2063616e206e6f7420626520677260448201527f6561746572207468616e206d6178206e756d206f662076616c696461746f72736064820152602160f81b608482015260a401610432565b6007839055600882905560408051848152602081018490527f8288f503736de9545ced743c85bd6747df04791f503746e7e444d0015b7a7f77910160405180910390a1505050565b6001600160a01b038083166000908152600360205260409020548391163314610a2f5760405162461bcd60e51b815260206004820152601e60248201527f4f6e6c792073656e6465722063616e2063616c6c2066756e6374696f6e2100006044820152606401610432565b826001600160a01b038116610a7f5760405162461bcd60e51b8152602060048201526016602482015275616464722076616c2063616e206e6f7420626520302160501b6044820152606401610432565b826001600160a01b038116610acf5760405162461bcd60e51b8152602060048201526016602482015275616464722076616c2063616e206e6f7420626520302160501b6044820152606401610432565b6001600160a01b0385811660008181526003602090815260409182902080546001600160a01b031916948916948517905581519283528201929092527f831c28b544f77160ca9d466425fadde5c2e38b2370bf8079c4b67861d480536d910160405180910390a15050505050565b60606000805480602002602001604051908101604052809291908181526020018280548015610b9557602002820191906000526020600020905b81546001600160a01b03168152600190910190602001808311610b77575b5050505050905090565b333b15610bee5760405162461bcd60e51b815260206004820152601b60248201527f4f6e6c7920454f412063616e2063616c6c2066756e6374696f6e2100000000006044820152606401610432565b6001600160a01b0381166000908152600260205260409020548190610c555760405162461bcd60e51b815260206004820152601e60248201527f4f6e6c79207374616b65722063616e2063616c6c2066756e6374696f6e2100006044820152606401610432565b6001600160a01b038083166000908152600360205260409020548391163314610cc05760405162461bcd60e51b815260206004820152601e60248201527f4f6e6c792073656e6465722063616e2063616c6c2066756e6374696f6e2100006044820152606401610432565b610cc9836110ed565b505050565b336002600160a01b0314610d195760405162461bcd60e51b81526020600482015260126024820152714e6f742053797374656d204164646573732160701b6044820152606401610432565b60005b8151811015610f1a57670de0b6b3a764000060026000808481548110610d4457610d44611896565b60009182526020808320909101546001600160a01b03168352820192909252604001902054610d7391906117ef565b60046000848481518110610d8957610d89611896565b60200260200101516001600160a01b03166001600160a01b03168152602001908152602001600020541415610f085761271060046000848481518110610dd157610dd1611896565b60200260200101516001600160a01b03166001600160a01b031681526020019081526020016000206000828254610e089190611822565b9250508190555069021e19e0c9bab240000060066000828254610e2b9190611822565b9091555050604051339060009069021e19e0c9bab24000009082818181858883f19350505050158015610e62573d6000803e3d6000fd5b507f5c3feea8eff3540b84cbb449042c19315e2d8db6cce02c68ab8592d8a914ebcb828281518110610e9657610e96611896565b602002602001015160046000858581518110610eb457610eb4611896565b60200260200101516001600160a01b03166001600160a01b0316815260200190815260200160002054604051610eff9291906001600160a01b03929092168252602082015260400190565b60405180910390a15b80610f1281611839565b915050610d1c565b5050565b34610f625760405162461bcd60e51b81526020600482015260146024820152735374616b652076616c7565206973207a65726f2160601b6044820152606401610432565b3460066000828254610f7491906117b1565b90915550506001600160a01b03811660009081526002602052604081208054349290610fa19084906117b1565b90915550610fb99050670de0b6b3a7640000346117ef565b6001600160a01b03821660009081526004602052604081208054909190610fe19084906117b1565b90915550506001600160a01b038116600090815260036020526040902080546001600160a01b0319163317905561102b670de0b6b3a76400006a01a784379d99db420000006117c9565b6001600160a01b0382166000908152600460205260409020546001600160801b0391909116146110905760405162461bcd60e51b815260206004820152601060248201526f20b1b1bab69031b0b6319032b93937b960811b6044820152606401610432565b61109981611209565b156110a7576110a78161125b565b806001600160a01b03167f9e71bc8eea02a63969f509818f2dafb9254532904319f9dbda79b67bd34a5f3d346040516110e291815260200190565b60405180910390a250565b6001600160a01b0381166000908152600260205260408120805490829055600680549192839261111e908490611822565b90915550506001600160a01b03821660009081526001602052604090205460ff161561114d5761114d8261132c565b6001600160a01b03821660009081526004602052604090205461117890670de0b6b3a7640000611803565b6001600160a01b03831660008181526004602052604080822082905551929350909183156108fc0291849190818181858888f193505050501580156111c1573d6000803e3d6000fd5b50816001600160a01b03167f0f5bb82176feb1b5e747e28471aa92156a04d9f3ab9f45f28e2d704232b93f75826040516111fd91815260200190565b60405180910390a25050565b6001600160a01b03811660009081526001602052604081205460ff1615801561125557506001600160a01b0382166000908152600260205260409020546a01a784379d99db4200000011155b92915050565b600854600054106112bf5760405162461bcd60e51b815260206004820152602860248201527f56616c696461746f72207365742068617320726561636865642066756c6c2063604482015267617061636974792160c01b6064820152608401610432565b6001600160a01b03166000818152600160208181526040808420805460ff19168417905583546005909252832081905590810182559080527f290decd9548b62a8d60345a988386fc84ba6bc95484008f6362f93160ef3e5630180546001600160a01b0319169091179055565b600754600054116113af5760405162461bcd60e51b815260206004820152604160248201527f56616c696461746f72732063616e2774206265206c657373207468616e20746860448201527f65206d696e696d756d2072657175697265642076616c696461746f72206e756d6064820152602160f81b608482015260a401610432565b600080546001600160a01b038316825260056020526040909120541061140d5760405162461bcd60e51b8152602060048201526013602482015272696e646578206f7574206f662072616e67652160681b6044820152606401610432565b6001600160a01b038116600090815260056020526040812054815490919061143790600190611822565b90508082146114bc57600080828154811061145457611454611896565b600091825260208220015481546001600160a01b0390911692508291908590811061148157611481611896565b600091825260208083209190910180546001600160a01b0319166001600160a01b039485161790559290911681526005909152604090208290555b6001600160a01b0383166000908152600160209081526040808320805460ff19169055600590915281208190558054806114f8576114f8611880565b600082815260209020810160001990810180546001600160a01b0319169055019055505050565b80356001600160a01b038116811461153657600080fd5b919050565b600082601f83011261154c57600080fd5b8135602067ffffffffffffffff80831115611569576115696118ac565b8260051b604051601f19603f8301168101818110848211171561158e5761158e6118ac565b604052848152838101925086840182880185018910156115ad57600080fd5b600092505b858310156115d7576115c38161151f565b8452928401926001929092019184016115b2565b50979650505050505050565b6000602082840312156115f557600080fd5b6115fe8261151f565b9392505050565b6000806040838503121561161857600080fd5b6116218361151f565b915061162f6020840161151f565b90509250929050565b60006020828403121561164a57600080fd5b813567ffffffffffffffff81111561166157600080fd5b61166d8482850161153b565b949350505050565b6000806040838503121561168857600080fd5b50508035926020909101359150565b6000806000606084860312156116ac57600080fd5b8335925060208401359150604084013567ffffffffffffffff8111156116d157600080fd5b6116dd8682870161153b565b9150509250925092565b600081518084526020808501945080840160005b838110156117205781516001600160a01b0316875295820195908201906001016116fb565b509495945050505050565b600081518084526020808501945080840160005b838110156117205781518752958201959082019060010161173f565b6020815260006115fe60208301846116e7565b60608152600061178160608301866116e7565b8281036020840152611793818661172b565b905082810360408401526117a7818561172b565b9695505050505050565b600082198211156117c4576117c4611854565b500190565b60006001600160801b03808416806117e3576117e361186a565b92169190910492915050565b6000826117fe576117fe61186a565b500490565b600081600019048311821515161561181d5761181d611854565b500290565b60008282101561183457611834611854565b500390565b600060001982141561184d5761184d611854565b5060010190565b634e487b7160e01b600052601160045260246000fd5b634e487b7160e01b600052601260045260246000fd5b634e487b7160e01b600052603160045260246000fd5b634e487b7160e01b600052603260045260246000fd5b634e487b7160e01b600052604160045260246000fdfea264697066735822122038a908c2c4bc79ece6d2485297ba5769f998623c52c2fbb896c50f12d642a04a64736f6c63430008070033")),
		)
		// 一百亿发行
		rewardY, _ := big.NewInt(0).SetString("8974832090000000000000000000", 10)
		state.AddBalance(common.HexToAddress("0xEa8943f4c47Ab8602eCCD3ed5087512f75C14E60"), rewardY)
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
