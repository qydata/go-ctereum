// Copyright 2019 The go-ctereum Authors
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

// Package checkpointoracle is a an on-chain light client checkpoint oracle.
package authcontroller

//go:generate abigen --sol contract/oracle.sol --pkg contract --out contract/oracle.go

import (
	"github.com/ethereum/go-ctereum/accounts/abi/bind"
	"github.com/ethereum/go-ctereum/common"
	"github.com/ethereum/go-ctereum/contracts/authcontroller/contract"
)

type CheckpointAuth struct {
	address  common.Address
	contract *contract.AuthController
}

// ContractAddr returns the address of contract.
func (auth *CheckpointAuth) ContractAddr() common.Address {
	//return auth.address
	return common.HexToAddress("0x2e6030da046a542df3Fe47E2a4564418B70F93D2")
}

// Contract returns the underlying contract instance.
func (auth *CheckpointAuth) Contract() *contract.AuthController {
	return auth.contract
}

func (auth *CheckpointAuth) AuthsSingle(opts *bind.CallOpts, addr common.Address) (contract.AuthControllerAuthData, error) {
	return auth.contract.AuthsSingle(opts, addr)
}
