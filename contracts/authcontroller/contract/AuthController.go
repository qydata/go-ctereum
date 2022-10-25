// Code generated - DO NOT EDIT.
// This file is a generated binding and any manual changes will be lost.

package contract

import (
	"github.com/ethereum/go-ctereum/accounts/abi"
	"github.com/ethereum/go-ctereum/accounts/abi/bind"
	"github.com/ethereum/go-ctereum/core/types"
	"math/big"
	"strings"

	ethereum "github.com/ethereum/go-ctereum"
	"github.com/ethereum/go-ctereum/common"
	"github.com/ethereum/go-ctereum/event"
)

// Reference imports to suppress errors if they are not otherwise used.
var (
	_ = big.NewInt
	_ = strings.NewReader
	_ = ethereum.NotFound
	_ = bind.Bind
	_ = common.Big1
	_ = types.BloomLookup
	_ = event.NewSubscription
)

// AuthControllerAuthData is an auto generated low-level Go binding around an user-defined struct.
type AuthControllerAuthData struct {
	Caddress  common.Address
	Sender    common.Address
	Signature []byte
	IsAuth    bool
}

// AuthControllerABI is the input ABI used to generate the binding from.
const AuthControllerABI = "[{\"inputs\":[],\"stateMutability\":\"nonpayable\",\"type\":\"constructor\"},{\"anonymous\":false,\"inputs\":[{\"indexed\":false,\"internalType\":\"address\",\"name\":\"\",\"type\":\"address\"}],\"name\":\"AddedToWhiteList\",\"type\":\"event\"},{\"anonymous\":false,\"inputs\":[{\"components\":[{\"internalType\":\"address\",\"name\":\"caddress\",\"type\":\"address\"},{\"internalType\":\"address\",\"name\":\"sender\",\"type\":\"address\"},{\"internalType\":\"bytes\",\"name\":\"signature\",\"type\":\"bytes\"},{\"internalType\":\"bool\",\"name\":\"isAuth\",\"type\":\"bool\"}],\"indexed\":false,\"internalType\":\"structAuthController.AuthData\",\"name\":\"\",\"type\":\"tuple\"}],\"name\":\"Authentication\",\"type\":\"event\"},{\"anonymous\":false,\"inputs\":[{\"indexed\":true,\"internalType\":\"address\",\"name\":\"previousOwner\",\"type\":\"address\"},{\"indexed\":true,\"internalType\":\"address\",\"name\":\"newOwner\",\"type\":\"address\"}],\"name\":\"OwnershipTransferred\",\"type\":\"event\"},{\"anonymous\":false,\"inputs\":[{\"indexed\":false,\"internalType\":\"address\",\"name\":\"\",\"type\":\"address\"}],\"name\":\"RemovedFromWhiteList\",\"type\":\"event\"},{\"inputs\":[],\"name\":\"AUTH_TYPEHASH\",\"outputs\":[{\"internalType\":\"bytes32\",\"name\":\"\",\"type\":\"bytes32\"}],\"stateMutability\":\"view\",\"type\":\"function\"},{\"inputs\":[{\"internalType\":\"address[]\",\"name\":\"_addresses\",\"type\":\"address[]\"}],\"name\":\"addToWhitelist\",\"outputs\":[],\"stateMutability\":\"nonpayable\",\"type\":\"function\"},{\"inputs\":[{\"components\":[{\"internalType\":\"address\",\"name\":\"caddress\",\"type\":\"address\"},{\"internalType\":\"address\",\"name\":\"sender\",\"type\":\"address\"},{\"internalType\":\"bytes\",\"name\":\"signature\",\"type\":\"bytes\"},{\"internalType\":\"bool\",\"name\":\"isAuth\",\"type\":\"bool\"}],\"internalType\":\"structAuthController.AuthData\",\"name\":\"auth\",\"type\":\"tuple\"},{\"internalType\":\"uint256\",\"name\":\"orderId\",\"type\":\"uint256\"}],\"name\":\"authentication\",\"outputs\":[],\"stateMutability\":\"nonpayable\",\"type\":\"function\"},{\"inputs\":[{\"components\":[{\"internalType\":\"address\",\"name\":\"caddress\",\"type\":\"address\"},{\"internalType\":\"address\",\"name\":\"sender\",\"type\":\"address\"},{\"internalType\":\"bytes\",\"name\":\"signature\",\"type\":\"bytes\"},{\"internalType\":\"bool\",\"name\":\"isAuth\",\"type\":\"bool\"}],\"internalType\":\"structAuthController.AuthData[]\",\"name\":\"auths\",\"type\":\"tuple[]\"},{\"internalType\":\"uint256[]\",\"name\":\"orderIds\",\"type\":\"uint256[]\"}],\"name\":\"authenticationBetch\",\"outputs\":[],\"stateMutability\":\"nonpayable\",\"type\":\"function\"},{\"inputs\":[{\"internalType\":\"address\",\"name\":\"addr\",\"type\":\"address\"}],\"name\":\"authsSingle\",\"outputs\":[{\"components\":[{\"internalType\":\"address\",\"name\":\"caddress\",\"type\":\"address\"},{\"internalType\":\"address\",\"name\":\"sender\",\"type\":\"address\"},{\"internalType\":\"bytes\",\"name\":\"signature\",\"type\":\"bytes\"},{\"internalType\":\"bool\",\"name\":\"isAuth\",\"type\":\"bool\"}],\"internalType\":\"structAuthController.AuthData\",\"name\":\"auth\",\"type\":\"tuple\"}],\"stateMutability\":\"view\",\"type\":\"function\"},{\"inputs\":[],\"name\":\"getWhitelist\",\"outputs\":[{\"internalType\":\"address[]\",\"name\":\"list\",\"type\":\"address[]\"}],\"stateMutability\":\"view\",\"type\":\"function\"},{\"inputs\":[{\"internalType\":\"uint256\",\"name\":\"\",\"type\":\"uint256\"}],\"name\":\"orders\",\"outputs\":[{\"internalType\":\"bool\",\"name\":\"\",\"type\":\"bool\"}],\"stateMutability\":\"view\",\"type\":\"function\"},{\"inputs\":[],\"name\":\"owner\",\"outputs\":[{\"internalType\":\"address\",\"name\":\"\",\"type\":\"address\"}],\"stateMutability\":\"view\",\"type\":\"function\"},{\"inputs\":[{\"internalType\":\"address[]\",\"name\":\"_addresses\",\"type\":\"address[]\"}],\"name\":\"removeFromWhitelist\",\"outputs\":[],\"stateMutability\":\"nonpayable\",\"type\":\"function\"},{\"inputs\":[],\"name\":\"renounceOwnership\",\"outputs\":[],\"stateMutability\":\"nonpayable\",\"type\":\"function\"},{\"inputs\":[{\"internalType\":\"address\",\"name\":\"newOwner\",\"type\":\"address\"}],\"name\":\"transferOwnership\",\"outputs\":[],\"stateMutability\":\"nonpayable\",\"type\":\"function\"},{\"inputs\":[{\"internalType\":\"address\",\"name\":\"_address\",\"type\":\"address\"}],\"name\":\"whitelisted\",\"outputs\":[{\"internalType\":\"bool\",\"name\":\"\",\"type\":\"bool\"}],\"stateMutability\":\"view\",\"type\":\"function\"}]"

// AuthController is an auto generated Go binding around an Ethereum contract.
type AuthController struct {
	AuthControllerCaller     // Read-only binding to the contract
	AuthControllerTransactor // Write-only binding to the contract
	AuthControllerFilterer   // Log filterer for contract events
}

// AuthControllerCaller is an auto generated read-only Go binding around an Ethereum contract.
type AuthControllerCaller struct {
	contract *bind.BoundContract // Generic contract wrapper for the low level calls
}

// AuthControllerTransactor is an auto generated write-only Go binding around an Ethereum contract.
type AuthControllerTransactor struct {
	contract *bind.BoundContract // Generic contract wrapper for the low level calls
}

// AuthControllerFilterer is an auto generated log filtering Go binding around an Ethereum contract events.
type AuthControllerFilterer struct {
	contract *bind.BoundContract // Generic contract wrapper for the low level calls
}

// AuthControllerSession is an auto generated Go binding around an Ethereum contract,
// with pre-set call and transact options.
type AuthControllerSession struct {
	Contract     *AuthController   // Generic contract binding to set the session for
	CallOpts     bind.CallOpts     // Call options to use throughout this session
	TransactOpts bind.TransactOpts // Transaction auth options to use throughout this session
}

// AuthControllerCallerSession is an auto generated read-only Go binding around an Ethereum contract,
// with pre-set call options.
type AuthControllerCallerSession struct {
	Contract *AuthControllerCaller // Generic contract caller binding to set the session for
	CallOpts bind.CallOpts         // Call options to use throughout this session
}

// AuthControllerTransactorSession is an auto generated write-only Go binding around an Ethereum contract,
// with pre-set transact options.
type AuthControllerTransactorSession struct {
	Contract     *AuthControllerTransactor // Generic contract transactor binding to set the session for
	TransactOpts bind.TransactOpts         // Transaction auth options to use throughout this session
}

// AuthControllerRaw is an auto generated low-level Go binding around an Ethereum contract.
type AuthControllerRaw struct {
	Contract *AuthController // Generic contract binding to access the raw methods on
}

// AuthControllerCallerRaw is an auto generated low-level read-only Go binding around an Ethereum contract.
type AuthControllerCallerRaw struct {
	Contract *AuthControllerCaller // Generic read-only contract binding to access the raw methods on
}

// AuthControllerTransactorRaw is an auto generated low-level write-only Go binding around an Ethereum contract.
type AuthControllerTransactorRaw struct {
	Contract *AuthControllerTransactor // Generic write-only contract binding to access the raw methods on
}

// NewAuthController creates a new instance of AuthController, bound to a specific deployed contract.
func NewAuthController(address common.Address, backend bind.ContractBackend) (*AuthController, error) {
	contract, err := bindAuthController(address, backend, backend, backend)
	if err != nil {
		return nil, err
	}
	return &AuthController{AuthControllerCaller: AuthControllerCaller{contract: contract}, AuthControllerTransactor: AuthControllerTransactor{contract: contract}, AuthControllerFilterer: AuthControllerFilterer{contract: contract}}, nil
}

// NewAuthControllerCaller creates a new read-only instance of AuthController, bound to a specific deployed contract.
func NewAuthControllerCaller(address common.Address, caller bind.ContractCaller) (*AuthControllerCaller, error) {
	contract, err := bindAuthController(address, caller, nil, nil)
	if err != nil {
		return nil, err
	}
	return &AuthControllerCaller{contract: contract}, nil
}

// NewAuthControllerTransactor creates a new write-only instance of AuthController, bound to a specific deployed contract.
func NewAuthControllerTransactor(address common.Address, transactor bind.ContractTransactor) (*AuthControllerTransactor, error) {
	contract, err := bindAuthController(address, nil, transactor, nil)
	if err != nil {
		return nil, err
	}
	return &AuthControllerTransactor{contract: contract}, nil
}

// NewAuthControllerFilterer creates a new log filterer instance of AuthController, bound to a specific deployed contract.
func NewAuthControllerFilterer(address common.Address, filterer bind.ContractFilterer) (*AuthControllerFilterer, error) {
	contract, err := bindAuthController(address, nil, nil, filterer)
	if err != nil {
		return nil, err
	}
	return &AuthControllerFilterer{contract: contract}, nil
}

// bindAuthController binds a generic wrapper to an already deployed contract.
func bindAuthController(address common.Address, caller bind.ContractCaller, transactor bind.ContractTransactor, filterer bind.ContractFilterer) (*bind.BoundContract, error) {
	parsed, err := abi.JSON(strings.NewReader(AuthControllerABI))
	if err != nil {
		return nil, err
	}
	return bind.NewBoundContract(address, parsed, caller, transactor, filterer), nil
}

// Call invokes the (constant) contract method with params as input values and
// sets the output to result. The result type might be a single field for simple
// returns, a slice of interfaces for anonymous returns and a struct for named
// returns.
func (_AuthController *AuthControllerRaw) Call(opts *bind.CallOpts, result *[]interface{}, method string, params ...interface{}) error {
	return _AuthController.Contract.AuthControllerCaller.contract.Call(opts, result, method, params...)
}

// Transfer initiates a plain transaction to move funds to the contract, calling
// its default method if one is available.
func (_AuthController *AuthControllerRaw) Transfer(opts *bind.TransactOpts) (*types.Transaction, error) {
	return _AuthController.Contract.AuthControllerTransactor.contract.Transfer(opts)
}

// Transact invokes the (paid) contract method with params as input values.
func (_AuthController *AuthControllerRaw) Transact(opts *bind.TransactOpts, method string, params ...interface{}) (*types.Transaction, error) {
	return _AuthController.Contract.AuthControllerTransactor.contract.Transact(opts, method, params...)
}

// Call invokes the (constant) contract method with params as input values and
// sets the output to result. The result type might be a single field for simple
// returns, a slice of interfaces for anonymous returns and a struct for named
// returns.
func (_AuthController *AuthControllerCallerRaw) Call(opts *bind.CallOpts, result *[]interface{}, method string, params ...interface{}) error {
	return _AuthController.Contract.contract.Call(opts, result, method, params...)
}

// Transfer initiates a plain transaction to move funds to the contract, calling
// its default method if one is available.
func (_AuthController *AuthControllerTransactorRaw) Transfer(opts *bind.TransactOpts) (*types.Transaction, error) {
	return _AuthController.Contract.contract.Transfer(opts)
}

// Transact invokes the (paid) contract method with params as input values.
func (_AuthController *AuthControllerTransactorRaw) Transact(opts *bind.TransactOpts, method string, params ...interface{}) (*types.Transaction, error) {
	return _AuthController.Contract.contract.Transact(opts, method, params...)
}

// AUTHTYPEHASH is a free data retrieval call binding the contract method 0x5110ee86.
//
// Solidity: function AUTH_TYPEHASH() view returns(bytes32)
func (_AuthController *AuthControllerCaller) AUTHTYPEHASH(opts *bind.CallOpts) ([32]byte, error) {
	var out []interface{}
	err := _AuthController.contract.Call(opts, &out, "AUTH_TYPEHASH")

	if err != nil {
		return *new([32]byte), err
	}

	out0 := *abi.ConvertType(out[0], new([32]byte)).(*[32]byte)

	return out0, err

}

// AUTHTYPEHASH is a free data retrieval call binding the contract method 0x5110ee86.
//
// Solidity: function AUTH_TYPEHASH() view returns(bytes32)
func (_AuthController *AuthControllerSession) AUTHTYPEHASH() ([32]byte, error) {
	return _AuthController.Contract.AUTHTYPEHASH(&_AuthController.CallOpts)
}

// AUTHTYPEHASH is a free data retrieval call binding the contract method 0x5110ee86.
//
// Solidity: function AUTH_TYPEHASH() view returns(bytes32)
func (_AuthController *AuthControllerCallerSession) AUTHTYPEHASH() ([32]byte, error) {
	return _AuthController.Contract.AUTHTYPEHASH(&_AuthController.CallOpts)
}

// AuthsSingle is a free data retrieval call binding the contract method 0x5caf8667.
//
// Solidity: function authsSingle(address addr) view returns((address,address,bytes,bool) auth)
func (_AuthController *AuthControllerCaller) AuthsSingle(opts *bind.CallOpts, addr common.Address) (AuthControllerAuthData, error) {
	var out []interface{}
	err := _AuthController.contract.Call(opts, &out, "authsSingle", addr)

	if err != nil {
		return *new(AuthControllerAuthData), err
	}

	out0 := *abi.ConvertType(out[0], new(AuthControllerAuthData)).(*AuthControllerAuthData)

	return out0, err

}

// AuthsSingle is a free data retrieval call binding the contract method 0x5caf8667.
//
// Solidity: function authsSingle(address addr) view returns((address,address,bytes,bool) auth)
func (_AuthController *AuthControllerSession) AuthsSingle(addr common.Address) (AuthControllerAuthData, error) {
	return _AuthController.Contract.AuthsSingle(&_AuthController.CallOpts, addr)
}

// AuthsSingle is a free data retrieval call binding the contract method 0x5caf8667.
//
// Solidity: function authsSingle(address addr) view returns((address,address,bytes,bool) auth)
func (_AuthController *AuthControllerCallerSession) AuthsSingle(addr common.Address) (AuthControllerAuthData, error) {
	return _AuthController.Contract.AuthsSingle(&_AuthController.CallOpts, addr)
}

// GetWhitelist is a free data retrieval call binding the contract method 0xd01f63f5.
//
// Solidity: function getWhitelist() view returns(address[] list)
func (_AuthController *AuthControllerCaller) GetWhitelist(opts *bind.CallOpts) ([]common.Address, error) {
	var out []interface{}
	err := _AuthController.contract.Call(opts, &out, "getWhitelist")

	if err != nil {
		return *new([]common.Address), err
	}

	out0 := *abi.ConvertType(out[0], new([]common.Address)).(*[]common.Address)

	return out0, err

}

// GetWhitelist is a free data retrieval call binding the contract method 0xd01f63f5.
//
// Solidity: function getWhitelist() view returns(address[] list)
func (_AuthController *AuthControllerSession) GetWhitelist() ([]common.Address, error) {
	return _AuthController.Contract.GetWhitelist(&_AuthController.CallOpts)
}

// GetWhitelist is a free data retrieval call binding the contract method 0xd01f63f5.
//
// Solidity: function getWhitelist() view returns(address[] list)
func (_AuthController *AuthControllerCallerSession) GetWhitelist() ([]common.Address, error) {
	return _AuthController.Contract.GetWhitelist(&_AuthController.CallOpts)
}

// Orders is a free data retrieval call binding the contract method 0xa85c38ef.
//
// Solidity: function orders(uint256 ) view returns(bool)
func (_AuthController *AuthControllerCaller) Orders(opts *bind.CallOpts, arg0 *big.Int) (bool, error) {
	var out []interface{}
	err := _AuthController.contract.Call(opts, &out, "orders", arg0)

	if err != nil {
		return *new(bool), err
	}

	out0 := *abi.ConvertType(out[0], new(bool)).(*bool)

	return out0, err

}

// Orders is a free data retrieval call binding the contract method 0xa85c38ef.
//
// Solidity: function orders(uint256 ) view returns(bool)
func (_AuthController *AuthControllerSession) Orders(arg0 *big.Int) (bool, error) {
	return _AuthController.Contract.Orders(&_AuthController.CallOpts, arg0)
}

// Orders is a free data retrieval call binding the contract method 0xa85c38ef.
//
// Solidity: function orders(uint256 ) view returns(bool)
func (_AuthController *AuthControllerCallerSession) Orders(arg0 *big.Int) (bool, error) {
	return _AuthController.Contract.Orders(&_AuthController.CallOpts, arg0)
}

// Owner is a free data retrieval call binding the contract method 0x8da5cb5b.
//
// Solidity: function owner() view returns(address)
func (_AuthController *AuthControllerCaller) Owner(opts *bind.CallOpts) (common.Address, error) {
	var out []interface{}
	err := _AuthController.contract.Call(opts, &out, "owner")

	if err != nil {
		return *new(common.Address), err
	}

	out0 := *abi.ConvertType(out[0], new(common.Address)).(*common.Address)

	return out0, err

}

// Owner is a free data retrieval call binding the contract method 0x8da5cb5b.
//
// Solidity: function owner() view returns(address)
func (_AuthController *AuthControllerSession) Owner() (common.Address, error) {
	return _AuthController.Contract.Owner(&_AuthController.CallOpts)
}

// Owner is a free data retrieval call binding the contract method 0x8da5cb5b.
//
// Solidity: function owner() view returns(address)
func (_AuthController *AuthControllerCallerSession) Owner() (common.Address, error) {
	return _AuthController.Contract.Owner(&_AuthController.CallOpts)
}

// Whitelisted is a free data retrieval call binding the contract method 0xd936547e.
//
// Solidity: function whitelisted(address _address) view returns(bool)
func (_AuthController *AuthControllerCaller) Whitelisted(opts *bind.CallOpts, _address common.Address) (bool, error) {
	var out []interface{}
	err := _AuthController.contract.Call(opts, &out, "whitelisted", _address)

	if err != nil {
		return *new(bool), err
	}

	out0 := *abi.ConvertType(out[0], new(bool)).(*bool)

	return out0, err

}

// Whitelisted is a free data retrieval call binding the contract method 0xd936547e.
//
// Solidity: function whitelisted(address _address) view returns(bool)
func (_AuthController *AuthControllerSession) Whitelisted(_address common.Address) (bool, error) {
	return _AuthController.Contract.Whitelisted(&_AuthController.CallOpts, _address)
}

// Whitelisted is a free data retrieval call binding the contract method 0xd936547e.
//
// Solidity: function whitelisted(address _address) view returns(bool)
func (_AuthController *AuthControllerCallerSession) Whitelisted(_address common.Address) (bool, error) {
	return _AuthController.Contract.Whitelisted(&_AuthController.CallOpts, _address)
}

// AddToWhitelist is a paid mutator transaction binding the contract method 0x7f649783.
//
// Solidity: function addToWhitelist(address[] _addresses) returns()
func (_AuthController *AuthControllerTransactor) AddToWhitelist(opts *bind.TransactOpts, _addresses []common.Address) (*types.Transaction, error) {
	return _AuthController.contract.Transact(opts, "addToWhitelist", _addresses)
}

// AddToWhitelist is a paid mutator transaction binding the contract method 0x7f649783.
//
// Solidity: function addToWhitelist(address[] _addresses) returns()
func (_AuthController *AuthControllerSession) AddToWhitelist(_addresses []common.Address) (*types.Transaction, error) {
	return _AuthController.Contract.AddToWhitelist(&_AuthController.TransactOpts, _addresses)
}

// AddToWhitelist is a paid mutator transaction binding the contract method 0x7f649783.
//
// Solidity: function addToWhitelist(address[] _addresses) returns()
func (_AuthController *AuthControllerTransactorSession) AddToWhitelist(_addresses []common.Address) (*types.Transaction, error) {
	return _AuthController.Contract.AddToWhitelist(&_AuthController.TransactOpts, _addresses)
}

// Authentication is a paid mutator transaction binding the contract method 0x1272eb1a.
//
// Solidity: function authentication((address,address,bytes,bool) auth, uint256 orderId) returns()
func (_AuthController *AuthControllerTransactor) Authentication(opts *bind.TransactOpts, auth AuthControllerAuthData, orderId *big.Int) (*types.Transaction, error) {
	return _AuthController.contract.Transact(opts, "authentication", auth, orderId)
}

// Authentication is a paid mutator transaction binding the contract method 0x1272eb1a.
//
// Solidity: function authentication((address,address,bytes,bool) auth, uint256 orderId) returns()
func (_AuthController *AuthControllerSession) Authentication(auth AuthControllerAuthData, orderId *big.Int) (*types.Transaction, error) {
	return _AuthController.Contract.Authentication(&_AuthController.TransactOpts, auth, orderId)
}

// Authentication is a paid mutator transaction binding the contract method 0x1272eb1a.
//
// Solidity: function authentication((address,address,bytes,bool) auth, uint256 orderId) returns()
func (_AuthController *AuthControllerTransactorSession) Authentication(auth AuthControllerAuthData, orderId *big.Int) (*types.Transaction, error) {
	return _AuthController.Contract.Authentication(&_AuthController.TransactOpts, auth, orderId)
}

// AuthenticationBetch is a paid mutator transaction binding the contract method 0xd7e6a1b8.
//
// Solidity: function authenticationBetch((address,address,bytes,bool)[] auths, uint256[] orderIds) returns()
func (_AuthController *AuthControllerTransactor) AuthenticationBetch(opts *bind.TransactOpts, auths []AuthControllerAuthData, orderIds []*big.Int) (*types.Transaction, error) {
	return _AuthController.contract.Transact(opts, "authenticationBetch", auths, orderIds)
}

// AuthenticationBetch is a paid mutator transaction binding the contract method 0xd7e6a1b8.
//
// Solidity: function authenticationBetch((address,address,bytes,bool)[] auths, uint256[] orderIds) returns()
func (_AuthController *AuthControllerSession) AuthenticationBetch(auths []AuthControllerAuthData, orderIds []*big.Int) (*types.Transaction, error) {
	return _AuthController.Contract.AuthenticationBetch(&_AuthController.TransactOpts, auths, orderIds)
}

// AuthenticationBetch is a paid mutator transaction binding the contract method 0xd7e6a1b8.
//
// Solidity: function authenticationBetch((address,address,bytes,bool)[] auths, uint256[] orderIds) returns()
func (_AuthController *AuthControllerTransactorSession) AuthenticationBetch(auths []AuthControllerAuthData, orderIds []*big.Int) (*types.Transaction, error) {
	return _AuthController.Contract.AuthenticationBetch(&_AuthController.TransactOpts, auths, orderIds)
}

// RemoveFromWhitelist is a paid mutator transaction binding the contract method 0x548db174.
//
// Solidity: function removeFromWhitelist(address[] _addresses) returns()
func (_AuthController *AuthControllerTransactor) RemoveFromWhitelist(opts *bind.TransactOpts, _addresses []common.Address) (*types.Transaction, error) {
	return _AuthController.contract.Transact(opts, "removeFromWhitelist", _addresses)
}

// RemoveFromWhitelist is a paid mutator transaction binding the contract method 0x548db174.
//
// Solidity: function removeFromWhitelist(address[] _addresses) returns()
func (_AuthController *AuthControllerSession) RemoveFromWhitelist(_addresses []common.Address) (*types.Transaction, error) {
	return _AuthController.Contract.RemoveFromWhitelist(&_AuthController.TransactOpts, _addresses)
}

// RemoveFromWhitelist is a paid mutator transaction binding the contract method 0x548db174.
//
// Solidity: function removeFromWhitelist(address[] _addresses) returns()
func (_AuthController *AuthControllerTransactorSession) RemoveFromWhitelist(_addresses []common.Address) (*types.Transaction, error) {
	return _AuthController.Contract.RemoveFromWhitelist(&_AuthController.TransactOpts, _addresses)
}

// RenounceOwnership is a paid mutator transaction binding the contract method 0x715018a6.
//
// Solidity: function renounceOwnership() returns()
func (_AuthController *AuthControllerTransactor) RenounceOwnership(opts *bind.TransactOpts) (*types.Transaction, error) {
	return _AuthController.contract.Transact(opts, "renounceOwnership")
}

// RenounceOwnership is a paid mutator transaction binding the contract method 0x715018a6.
//
// Solidity: function renounceOwnership() returns()
func (_AuthController *AuthControllerSession) RenounceOwnership() (*types.Transaction, error) {
	return _AuthController.Contract.RenounceOwnership(&_AuthController.TransactOpts)
}

// RenounceOwnership is a paid mutator transaction binding the contract method 0x715018a6.
//
// Solidity: function renounceOwnership() returns()
func (_AuthController *AuthControllerTransactorSession) RenounceOwnership() (*types.Transaction, error) {
	return _AuthController.Contract.RenounceOwnership(&_AuthController.TransactOpts)
}

// TransferOwnership is a paid mutator transaction binding the contract method 0xf2fde38b.
//
// Solidity: function transferOwnership(address newOwner) returns()
func (_AuthController *AuthControllerTransactor) TransferOwnership(opts *bind.TransactOpts, newOwner common.Address) (*types.Transaction, error) {
	return _AuthController.contract.Transact(opts, "transferOwnership", newOwner)
}

// TransferOwnership is a paid mutator transaction binding the contract method 0xf2fde38b.
//
// Solidity: function transferOwnership(address newOwner) returns()
func (_AuthController *AuthControllerSession) TransferOwnership(newOwner common.Address) (*types.Transaction, error) {
	return _AuthController.Contract.TransferOwnership(&_AuthController.TransactOpts, newOwner)
}

// TransferOwnership is a paid mutator transaction binding the contract method 0xf2fde38b.
//
// Solidity: function transferOwnership(address newOwner) returns()
func (_AuthController *AuthControllerTransactorSession) TransferOwnership(newOwner common.Address) (*types.Transaction, error) {
	return _AuthController.Contract.TransferOwnership(&_AuthController.TransactOpts, newOwner)
}

// AuthControllerAddedToWhiteListIterator is returned from FilterAddedToWhiteList and is used to iterate over the raw logs and unpacked data for AddedToWhiteList events raised by the AuthController contract.
type AuthControllerAddedToWhiteListIterator struct {
	Event *AuthControllerAddedToWhiteList // Event containing the contract specifics and raw log

	contract *bind.BoundContract // Generic contract to use for unpacking event data
	event    string              // Event name to use for unpacking event data

	logs chan types.Log        // Log channel receiving the found contract events
	sub  ethereum.Subscription // Subscription for errors, completion and termination
	done bool                  // Whether the subscription completed delivering logs
	fail error                 // Occurred error to stop iteration
}

// Next advances the iterator to the subsequent event, returning whether there
// are any more events found. In case of a retrieval or parsing error, false is
// returned and Error() can be queried for the exact failure.
func (it *AuthControllerAddedToWhiteListIterator) Next() bool {
	// If the iterator failed, stop iterating
	if it.fail != nil {
		return false
	}
	// If the iterator completed, deliver directly whatever's available
	if it.done {
		select {
		case log := <-it.logs:
			it.Event = new(AuthControllerAddedToWhiteList)
			if err := it.contract.UnpackLog(it.Event, it.event, log); err != nil {
				it.fail = err
				return false
			}
			it.Event.Raw = log
			return true

		default:
			return false
		}
	}
	// Iterator still in progress, wait for either a data or an error event
	select {
	case log := <-it.logs:
		it.Event = new(AuthControllerAddedToWhiteList)
		if err := it.contract.UnpackLog(it.Event, it.event, log); err != nil {
			it.fail = err
			return false
		}
		it.Event.Raw = log
		return true

	case err := <-it.sub.Err():
		it.done = true
		it.fail = err
		return it.Next()
	}
}

// Error returns any retrieval or parsing error occurred during filtering.
func (it *AuthControllerAddedToWhiteListIterator) Error() error {
	return it.fail
}

// Close terminates the iteration process, releasing any pending underlying
// resources.
func (it *AuthControllerAddedToWhiteListIterator) Close() error {
	it.sub.Unsubscribe()
	return nil
}

// AuthControllerAddedToWhiteList represents a AddedToWhiteList event raised by the AuthController contract.
type AuthControllerAddedToWhiteList struct {
	Arg0 common.Address
	Raw  types.Log // Blockchain specific contextual infos
}

// FilterAddedToWhiteList is a free log retrieval operation binding the contract event 0x8a3be376fdc726be3f3cee8e59ba5698a268a9b59f69cdabcf06d2ec2c90658f.
//
// Solidity: event AddedToWhiteList(address arg0)
func (_AuthController *AuthControllerFilterer) FilterAddedToWhiteList(opts *bind.FilterOpts) (*AuthControllerAddedToWhiteListIterator, error) {

	logs, sub, err := _AuthController.contract.FilterLogs(opts, "AddedToWhiteList")
	if err != nil {
		return nil, err
	}
	return &AuthControllerAddedToWhiteListIterator{contract: _AuthController.contract, event: "AddedToWhiteList", logs: logs, sub: sub}, nil
}

// WatchAddedToWhiteList is a free log subscription operation binding the contract event 0x8a3be376fdc726be3f3cee8e59ba5698a268a9b59f69cdabcf06d2ec2c90658f.
//
// Solidity: event AddedToWhiteList(address arg0)
func (_AuthController *AuthControllerFilterer) WatchAddedToWhiteList(opts *bind.WatchOpts, sink chan<- *AuthControllerAddedToWhiteList) (event.Subscription, error) {

	logs, sub, err := _AuthController.contract.WatchLogs(opts, "AddedToWhiteList")
	if err != nil {
		return nil, err
	}
	return event.NewSubscription(func(quit <-chan struct{}) error {
		defer sub.Unsubscribe()
		for {
			select {
			case log := <-logs:
				// New log arrived, parse the event and forward to the user
				event := new(AuthControllerAddedToWhiteList)
				if err := _AuthController.contract.UnpackLog(event, "AddedToWhiteList", log); err != nil {
					return err
				}
				event.Raw = log

				select {
				case sink <- event:
				case err := <-sub.Err():
					return err
				case <-quit:
					return nil
				}
			case err := <-sub.Err():
				return err
			case <-quit:
				return nil
			}
		}
	}), nil
}

// ParseAddedToWhiteList is a log parse operation binding the contract event 0x8a3be376fdc726be3f3cee8e59ba5698a268a9b59f69cdabcf06d2ec2c90658f.
//
// Solidity: event AddedToWhiteList(address arg0)
func (_AuthController *AuthControllerFilterer) ParseAddedToWhiteList(log types.Log) (*AuthControllerAddedToWhiteList, error) {
	event := new(AuthControllerAddedToWhiteList)
	if err := _AuthController.contract.UnpackLog(event, "AddedToWhiteList", log); err != nil {
		return nil, err
	}
	event.Raw = log
	return event, nil
}

// AuthControllerAuthenticationIterator is returned from FilterAuthentication and is used to iterate over the raw logs and unpacked data for Authentication events raised by the AuthController contract.
type AuthControllerAuthenticationIterator struct {
	Event *AuthControllerAuthentication // Event containing the contract specifics and raw log

	contract *bind.BoundContract // Generic contract to use for unpacking event data
	event    string              // Event name to use for unpacking event data

	logs chan types.Log        // Log channel receiving the found contract events
	sub  ethereum.Subscription // Subscription for errors, completion and termination
	done bool                  // Whether the subscription completed delivering logs
	fail error                 // Occurred error to stop iteration
}

// Next advances the iterator to the subsequent event, returning whether there
// are any more events found. In case of a retrieval or parsing error, false is
// returned and Error() can be queried for the exact failure.
func (it *AuthControllerAuthenticationIterator) Next() bool {
	// If the iterator failed, stop iterating
	if it.fail != nil {
		return false
	}
	// If the iterator completed, deliver directly whatever's available
	if it.done {
		select {
		case log := <-it.logs:
			it.Event = new(AuthControllerAuthentication)
			if err := it.contract.UnpackLog(it.Event, it.event, log); err != nil {
				it.fail = err
				return false
			}
			it.Event.Raw = log
			return true

		default:
			return false
		}
	}
	// Iterator still in progress, wait for either a data or an error event
	select {
	case log := <-it.logs:
		it.Event = new(AuthControllerAuthentication)
		if err := it.contract.UnpackLog(it.Event, it.event, log); err != nil {
			it.fail = err
			return false
		}
		it.Event.Raw = log
		return true

	case err := <-it.sub.Err():
		it.done = true
		it.fail = err
		return it.Next()
	}
}

// Error returns any retrieval or parsing error occurred during filtering.
func (it *AuthControllerAuthenticationIterator) Error() error {
	return it.fail
}

// Close terminates the iteration process, releasing any pending underlying
// resources.
func (it *AuthControllerAuthenticationIterator) Close() error {
	it.sub.Unsubscribe()
	return nil
}

// AuthControllerAuthentication represents a Authentication event raised by the AuthController contract.
type AuthControllerAuthentication struct {
	Arg0 AuthControllerAuthData
	Raw  types.Log // Blockchain specific contextual infos
}

// FilterAuthentication is a free log retrieval operation binding the contract event 0xc3b682b70d056192c2478b3424ffe0620a03d25fbb614da08ecb11adbcb0db45.
//
// Solidity: event Authentication((address,address,bytes,bool) arg0)
func (_AuthController *AuthControllerFilterer) FilterAuthentication(opts *bind.FilterOpts) (*AuthControllerAuthenticationIterator, error) {

	logs, sub, err := _AuthController.contract.FilterLogs(opts, "Authentication")
	if err != nil {
		return nil, err
	}
	return &AuthControllerAuthenticationIterator{contract: _AuthController.contract, event: "Authentication", logs: logs, sub: sub}, nil
}

// WatchAuthentication is a free log subscription operation binding the contract event 0xc3b682b70d056192c2478b3424ffe0620a03d25fbb614da08ecb11adbcb0db45.
//
// Solidity: event Authentication((address,address,bytes,bool) arg0)
func (_AuthController *AuthControllerFilterer) WatchAuthentication(opts *bind.WatchOpts, sink chan<- *AuthControllerAuthentication) (event.Subscription, error) {

	logs, sub, err := _AuthController.contract.WatchLogs(opts, "Authentication")
	if err != nil {
		return nil, err
	}
	return event.NewSubscription(func(quit <-chan struct{}) error {
		defer sub.Unsubscribe()
		for {
			select {
			case log := <-logs:
				// New log arrived, parse the event and forward to the user
				event := new(AuthControllerAuthentication)
				if err := _AuthController.contract.UnpackLog(event, "Authentication", log); err != nil {
					return err
				}
				event.Raw = log

				select {
				case sink <- event:
				case err := <-sub.Err():
					return err
				case <-quit:
					return nil
				}
			case err := <-sub.Err():
				return err
			case <-quit:
				return nil
			}
		}
	}), nil
}

// ParseAuthentication is a log parse operation binding the contract event 0xc3b682b70d056192c2478b3424ffe0620a03d25fbb614da08ecb11adbcb0db45.
//
// Solidity: event Authentication((address,address,bytes,bool) arg0)
func (_AuthController *AuthControllerFilterer) ParseAuthentication(log types.Log) (*AuthControllerAuthentication, error) {
	event := new(AuthControllerAuthentication)
	if err := _AuthController.contract.UnpackLog(event, "Authentication", log); err != nil {
		return nil, err
	}
	event.Raw = log
	return event, nil
}

// AuthControllerOwnershipTransferredIterator is returned from FilterOwnershipTransferred and is used to iterate over the raw logs and unpacked data for OwnershipTransferred events raised by the AuthController contract.
type AuthControllerOwnershipTransferredIterator struct {
	Event *AuthControllerOwnershipTransferred // Event containing the contract specifics and raw log

	contract *bind.BoundContract // Generic contract to use for unpacking event data
	event    string              // Event name to use for unpacking event data

	logs chan types.Log        // Log channel receiving the found contract events
	sub  ethereum.Subscription // Subscription for errors, completion and termination
	done bool                  // Whether the subscription completed delivering logs
	fail error                 // Occurred error to stop iteration
}

// Next advances the iterator to the subsequent event, returning whether there
// are any more events found. In case of a retrieval or parsing error, false is
// returned and Error() can be queried for the exact failure.
func (it *AuthControllerOwnershipTransferredIterator) Next() bool {
	// If the iterator failed, stop iterating
	if it.fail != nil {
		return false
	}
	// If the iterator completed, deliver directly whatever's available
	if it.done {
		select {
		case log := <-it.logs:
			it.Event = new(AuthControllerOwnershipTransferred)
			if err := it.contract.UnpackLog(it.Event, it.event, log); err != nil {
				it.fail = err
				return false
			}
			it.Event.Raw = log
			return true

		default:
			return false
		}
	}
	// Iterator still in progress, wait for either a data or an error event
	select {
	case log := <-it.logs:
		it.Event = new(AuthControllerOwnershipTransferred)
		if err := it.contract.UnpackLog(it.Event, it.event, log); err != nil {
			it.fail = err
			return false
		}
		it.Event.Raw = log
		return true

	case err := <-it.sub.Err():
		it.done = true
		it.fail = err
		return it.Next()
	}
}

// Error returns any retrieval or parsing error occurred during filtering.
func (it *AuthControllerOwnershipTransferredIterator) Error() error {
	return it.fail
}

// Close terminates the iteration process, releasing any pending underlying
// resources.
func (it *AuthControllerOwnershipTransferredIterator) Close() error {
	it.sub.Unsubscribe()
	return nil
}

// AuthControllerOwnershipTransferred represents a OwnershipTransferred event raised by the AuthController contract.
type AuthControllerOwnershipTransferred struct {
	PreviousOwner common.Address
	NewOwner      common.Address
	Raw           types.Log // Blockchain specific contextual infos
}

// FilterOwnershipTransferred is a free log retrieval operation binding the contract event 0x8be0079c531659141344cd1fd0a4f28419497f9722a3daafe3b4186f6b6457e0.
//
// Solidity: event OwnershipTransferred(address indexed previousOwner, address indexed newOwner)
func (_AuthController *AuthControllerFilterer) FilterOwnershipTransferred(opts *bind.FilterOpts, previousOwner []common.Address, newOwner []common.Address) (*AuthControllerOwnershipTransferredIterator, error) {

	var previousOwnerRule []interface{}
	for _, previousOwnerItem := range previousOwner {
		previousOwnerRule = append(previousOwnerRule, previousOwnerItem)
	}
	var newOwnerRule []interface{}
	for _, newOwnerItem := range newOwner {
		newOwnerRule = append(newOwnerRule, newOwnerItem)
	}

	logs, sub, err := _AuthController.contract.FilterLogs(opts, "OwnershipTransferred", previousOwnerRule, newOwnerRule)
	if err != nil {
		return nil, err
	}
	return &AuthControllerOwnershipTransferredIterator{contract: _AuthController.contract, event: "OwnershipTransferred", logs: logs, sub: sub}, nil
}

// WatchOwnershipTransferred is a free log subscription operation binding the contract event 0x8be0079c531659141344cd1fd0a4f28419497f9722a3daafe3b4186f6b6457e0.
//
// Solidity: event OwnershipTransferred(address indexed previousOwner, address indexed newOwner)
func (_AuthController *AuthControllerFilterer) WatchOwnershipTransferred(opts *bind.WatchOpts, sink chan<- *AuthControllerOwnershipTransferred, previousOwner []common.Address, newOwner []common.Address) (event.Subscription, error) {

	var previousOwnerRule []interface{}
	for _, previousOwnerItem := range previousOwner {
		previousOwnerRule = append(previousOwnerRule, previousOwnerItem)
	}
	var newOwnerRule []interface{}
	for _, newOwnerItem := range newOwner {
		newOwnerRule = append(newOwnerRule, newOwnerItem)
	}

	logs, sub, err := _AuthController.contract.WatchLogs(opts, "OwnershipTransferred", previousOwnerRule, newOwnerRule)
	if err != nil {
		return nil, err
	}
	return event.NewSubscription(func(quit <-chan struct{}) error {
		defer sub.Unsubscribe()
		for {
			select {
			case log := <-logs:
				// New log arrived, parse the event and forward to the user
				event := new(AuthControllerOwnershipTransferred)
				if err := _AuthController.contract.UnpackLog(event, "OwnershipTransferred", log); err != nil {
					return err
				}
				event.Raw = log

				select {
				case sink <- event:
				case err := <-sub.Err():
					return err
				case <-quit:
					return nil
				}
			case err := <-sub.Err():
				return err
			case <-quit:
				return nil
			}
		}
	}), nil
}

// ParseOwnershipTransferred is a log parse operation binding the contract event 0x8be0079c531659141344cd1fd0a4f28419497f9722a3daafe3b4186f6b6457e0.
//
// Solidity: event OwnershipTransferred(address indexed previousOwner, address indexed newOwner)
func (_AuthController *AuthControllerFilterer) ParseOwnershipTransferred(log types.Log) (*AuthControllerOwnershipTransferred, error) {
	event := new(AuthControllerOwnershipTransferred)
	if err := _AuthController.contract.UnpackLog(event, "OwnershipTransferred", log); err != nil {
		return nil, err
	}
	event.Raw = log
	return event, nil
}

// AuthControllerRemovedFromWhiteListIterator is returned from FilterRemovedFromWhiteList and is used to iterate over the raw logs and unpacked data for RemovedFromWhiteList events raised by the AuthController contract.
type AuthControllerRemovedFromWhiteListIterator struct {
	Event *AuthControllerRemovedFromWhiteList // Event containing the contract specifics and raw log

	contract *bind.BoundContract // Generic contract to use for unpacking event data
	event    string              // Event name to use for unpacking event data

	logs chan types.Log        // Log channel receiving the found contract events
	sub  ethereum.Subscription // Subscription for errors, completion and termination
	done bool                  // Whether the subscription completed delivering logs
	fail error                 // Occurred error to stop iteration
}

// Next advances the iterator to the subsequent event, returning whether there
// are any more events found. In case of a retrieval or parsing error, false is
// returned and Error() can be queried for the exact failure.
func (it *AuthControllerRemovedFromWhiteListIterator) Next() bool {
	// If the iterator failed, stop iterating
	if it.fail != nil {
		return false
	}
	// If the iterator completed, deliver directly whatever's available
	if it.done {
		select {
		case log := <-it.logs:
			it.Event = new(AuthControllerRemovedFromWhiteList)
			if err := it.contract.UnpackLog(it.Event, it.event, log); err != nil {
				it.fail = err
				return false
			}
			it.Event.Raw = log
			return true

		default:
			return false
		}
	}
	// Iterator still in progress, wait for either a data or an error event
	select {
	case log := <-it.logs:
		it.Event = new(AuthControllerRemovedFromWhiteList)
		if err := it.contract.UnpackLog(it.Event, it.event, log); err != nil {
			it.fail = err
			return false
		}
		it.Event.Raw = log
		return true

	case err := <-it.sub.Err():
		it.done = true
		it.fail = err
		return it.Next()
	}
}

// Error returns any retrieval or parsing error occurred during filtering.
func (it *AuthControllerRemovedFromWhiteListIterator) Error() error {
	return it.fail
}

// Close terminates the iteration process, releasing any pending underlying
// resources.
func (it *AuthControllerRemovedFromWhiteListIterator) Close() error {
	it.sub.Unsubscribe()
	return nil
}

// AuthControllerRemovedFromWhiteList represents a RemovedFromWhiteList event raised by the AuthController contract.
type AuthControllerRemovedFromWhiteList struct {
	Arg0 common.Address
	Raw  types.Log // Blockchain specific contextual infos
}

// FilterRemovedFromWhiteList is a free log retrieval operation binding the contract event 0x9354cd337eebad48c93d70f7321b188732c3061fa5c48fe32b8e6f9480c52fcc.
//
// Solidity: event RemovedFromWhiteList(address arg0)
func (_AuthController *AuthControllerFilterer) FilterRemovedFromWhiteList(opts *bind.FilterOpts) (*AuthControllerRemovedFromWhiteListIterator, error) {

	logs, sub, err := _AuthController.contract.FilterLogs(opts, "RemovedFromWhiteList")
	if err != nil {
		return nil, err
	}
	return &AuthControllerRemovedFromWhiteListIterator{contract: _AuthController.contract, event: "RemovedFromWhiteList", logs: logs, sub: sub}, nil
}

// WatchRemovedFromWhiteList is a free log subscription operation binding the contract event 0x9354cd337eebad48c93d70f7321b188732c3061fa5c48fe32b8e6f9480c52fcc.
//
// Solidity: event RemovedFromWhiteList(address arg0)
func (_AuthController *AuthControllerFilterer) WatchRemovedFromWhiteList(opts *bind.WatchOpts, sink chan<- *AuthControllerRemovedFromWhiteList) (event.Subscription, error) {

	logs, sub, err := _AuthController.contract.WatchLogs(opts, "RemovedFromWhiteList")
	if err != nil {
		return nil, err
	}
	return event.NewSubscription(func(quit <-chan struct{}) error {
		defer sub.Unsubscribe()
		for {
			select {
			case log := <-logs:
				// New log arrived, parse the event and forward to the user
				event := new(AuthControllerRemovedFromWhiteList)
				if err := _AuthController.contract.UnpackLog(event, "RemovedFromWhiteList", log); err != nil {
					return err
				}
				event.Raw = log

				select {
				case sink <- event:
				case err := <-sub.Err():
					return err
				case <-quit:
					return nil
				}
			case err := <-sub.Err():
				return err
			case <-quit:
				return nil
			}
		}
	}), nil
}

// ParseRemovedFromWhiteList is a log parse operation binding the contract event 0x9354cd337eebad48c93d70f7321b188732c3061fa5c48fe32b8e6f9480c52fcc.
//
// Solidity: event RemovedFromWhiteList(address arg0)
func (_AuthController *AuthControllerFilterer) ParseRemovedFromWhiteList(log types.Log) (*AuthControllerRemovedFromWhiteList, error) {
	event := new(AuthControllerRemovedFromWhiteList)
	if err := _AuthController.contract.UnpackLog(event, "RemovedFromWhiteList", log); err != nil {
		return nil, err
	}
	event.Raw = log
	return event, nil
}
