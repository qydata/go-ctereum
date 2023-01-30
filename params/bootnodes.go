// Copyright 2015 The go-ctereum Authors
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

package params

import "github.com/qydata/go-ctereum/common"

// MainnetBootnodes are the enode URLs of the P2P bootstrap nodes running on
// the main Ethereum network.
var MainnetBootnodes = []string{
	// Ethereum Foundation Go Bootnodes
	"enode://1ecea9891dd889f9ffdaa391f46b4f43857a6b7e3eddedc834a9feb88ca27bc3e6c51c404b7d0e03dda8e92042b89f796b2bba67cbdd020dac037b4ec49b5c60@121.40.143.162:30300", // bootnode-aws-ap-southeast-1-001
	"enode://1ecea9891dd889f9ffdaa391f46b4f43857a6b7e3eddedc834a9feb88ca27bc3e6c51c404b7d0e03dda8e92042b89f796b2bba67cbdd020dac037b4ec49b5c60@172.31.100.129:30300", // bootnode-aws-ap-southeast-1-001
}

// RopstenBootnodes are the enode URLs of the P2P bootstrap nodes running on the
// Ropsten test network.
var RopstenBootnodes = []string{}

// SepoliaBootnodes are the enode URLs of the P2P bootstrap nodes running on the
// Sepolia test network.
var SepoliaBootnodes = []string{
	// geth
}

// RinkebyBootnodes are the enode URLs of the P2P bootstrap nodes running on the
// Rinkeby test network.
var RinkebyBootnodes = []string{}

// GoerliBootnodes are the enode URLs of the P2P bootstrap nodes running on the
// Görli test network.
var GoerliBootnodes = []string{
	// Upstream bootnodes
}

// BorMainnetBootnodes are the enode URLs of the P2P bootstrap nodes running on the
// main Bor network.
var BorMainnetBootnodes = []string{}

var KilnBootnodes = []string{}

var V5Bootnodes = []string{
	// Teku team's bootnode
}

const dnsPrefix = "enrtree://AKA3AM6LPBYEUDMVNU3BSVQJ5AD45Y7YPOHJLEF6W26QOE4VTUDPE@"

// KnownDNSNetwork returns the address of a public DNS-based node list for the given
// genesis hash and protocol. See https://github.com/ethereum/discv4-dns-lists for more
// information.
func KnownDNSNetwork(genesis common.Hash, protocol string) string {
	var net string
	switch genesis {
	case MainnetGenesisHash:
		net = "mainnet"
	case RopstenGenesisHash:
		net = "ropsten"
	case RinkebyGenesisHash:
		net = "rinkeby"
	case GoerliGenesisHash:
		net = "goerli"
	case BorMainnetGenesisHash:
		net = "bor-mainnet"
	case SepoliaGenesisHash:
		net = "sepolia"
	default:
		return ""
	}
	return dnsPrefix + protocol + "." + net + ".ethdisco.net"
}
