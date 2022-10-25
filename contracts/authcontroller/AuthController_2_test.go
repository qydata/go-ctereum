package authcontroller

import (
	"fmt"
	"github.com/ethereum/go-ctereum/accounts/abi"
	"github.com/ethereum/go-ctereum/common"
	"strings"
	"testing"
)

type AuthControllerAuthData struct {
	Caddress  common.Address
	Sender    common.Address
	Signature []byte
	IsAuth    bool
}

func TestAuthBoth2(t *testing.T) {
	// Initialize test accounts

	// Deploy registrar contract

	// 3 trusted signers, threshold 2
	//contractAddr := "0x1E3b0AeC6C3210680915Cc36a62AE7437425923B"

	//连接本地的以太坊私链（一定要保证本地以太坊私链已经启动）

	const AuthControllerABI = "[{\"inputs\":[],\"stateMutability\":\"nonpayable\",\"type\":\"constructor\"},{\"anonymous\":false,\"inputs\":[{\"indexed\":false,\"internalType\":\"address\",\"name\":\"\",\"type\":\"address\"}],\"name\":\"AddedToWhiteList\",\"type\":\"event\"},{\"anonymous\":false,\"inputs\":[{\"components\":[{\"internalType\":\"address\",\"name\":\"caddress\",\"type\":\"address\"},{\"internalType\":\"address\",\"name\":\"sender\",\"type\":\"address\"},{\"internalType\":\"bytes\",\"name\":\"signature\",\"type\":\"bytes\"},{\"internalType\":\"bool\",\"name\":\"isAuth\",\"type\":\"bool\"}],\"indexed\":false,\"internalType\":\"structAuthController.AuthData\",\"name\":\"\",\"type\":\"tuple\"}],\"name\":\"Authentication\",\"type\":\"event\"},{\"anonymous\":false,\"inputs\":[{\"indexed\":true,\"internalType\":\"address\",\"name\":\"previousOwner\",\"type\":\"address\"},{\"indexed\":true,\"internalType\":\"address\",\"name\":\"newOwner\",\"type\":\"address\"}],\"name\":\"OwnershipTransferred\",\"type\":\"event\"},{\"anonymous\":false,\"inputs\":[{\"indexed\":false,\"internalType\":\"address\",\"name\":\"\",\"type\":\"address\"}],\"name\":\"RemovedFromWhiteList\",\"type\":\"event\"},{\"inputs\":[],\"name\":\"AUTH_TYPEHASH\",\"outputs\":[{\"internalType\":\"bytes32\",\"name\":\"\",\"type\":\"bytes32\"}],\"stateMutability\":\"view\",\"type\":\"function\"},{\"inputs\":[{\"internalType\":\"address[]\",\"name\":\"_addresses\",\"type\":\"address[]\"}],\"name\":\"addToWhitelist\",\"outputs\":[],\"stateMutability\":\"nonpayable\",\"type\":\"function\"},{\"inputs\":[{\"components\":[{\"internalType\":\"address\",\"name\":\"caddress\",\"type\":\"address\"},{\"internalType\":\"address\",\"name\":\"sender\",\"type\":\"address\"},{\"internalType\":\"bytes\",\"name\":\"signature\",\"type\":\"bytes\"},{\"internalType\":\"bool\",\"name\":\"isAuth\",\"type\":\"bool\"}],\"internalType\":\"structAuthController.AuthData\",\"name\":\"auth\",\"type\":\"tuple\"},{\"internalType\":\"uint256\",\"name\":\"orderId\",\"type\":\"uint256\"}],\"name\":\"authentication\",\"outputs\":[],\"stateMutability\":\"nonpayable\",\"type\":\"function\"},{\"inputs\":[{\"components\":[{\"internalType\":\"address\",\"name\":\"caddress\",\"type\":\"address\"},{\"internalType\":\"address\",\"name\":\"sender\",\"type\":\"address\"},{\"internalType\":\"bytes\",\"name\":\"signature\",\"type\":\"bytes\"},{\"internalType\":\"bool\",\"name\":\"isAuth\",\"type\":\"bool\"}],\"internalType\":\"structAuthController.AuthData[]\",\"name\":\"auths\",\"type\":\"tuple[]\"},{\"internalType\":\"uint256[]\",\"name\":\"orderIds\",\"type\":\"uint256[]\"}],\"name\":\"authenticationBetch\",\"outputs\":[],\"stateMutability\":\"nonpayable\",\"type\":\"function\"},{\"inputs\":[{\"internalType\":\"address\",\"name\":\"addr\",\"type\":\"address\"}],\"name\":\"authsSingle\",\"outputs\":[{\"components\":[{\"internalType\":\"address\",\"name\":\"caddress\",\"type\":\"address\"},{\"internalType\":\"address\",\"name\":\"sender\",\"type\":\"address\"},{\"internalType\":\"bytes\",\"name\":\"signature\",\"type\":\"bytes\"},{\"internalType\":\"bool\",\"name\":\"isAuth\",\"type\":\"bool\"}],\"internalType\":\"structAuthController.AuthData\",\"name\":\"auth\",\"type\":\"tuple\"}],\"stateMutability\":\"view\",\"type\":\"function\"},{\"inputs\":[],\"name\":\"getWhitelist\",\"outputs\":[{\"internalType\":\"address[]\",\"name\":\"list\",\"type\":\"address[]\"}],\"stateMutability\":\"view\",\"type\":\"function\"},{\"inputs\":[{\"internalType\":\"uint256\",\"name\":\"\",\"type\":\"uint256\"}],\"name\":\"orders\",\"outputs\":[{\"internalType\":\"bool\",\"name\":\"\",\"type\":\"bool\"}],\"stateMutability\":\"view\",\"type\":\"function\"},{\"inputs\":[],\"name\":\"owner\",\"outputs\":[{\"internalType\":\"address\",\"name\":\"\",\"type\":\"address\"}],\"stateMutability\":\"view\",\"type\":\"function\"},{\"inputs\":[{\"internalType\":\"address[]\",\"name\":\"_addresses\",\"type\":\"address[]\"}],\"name\":\"removeFromWhitelist\",\"outputs\":[],\"stateMutability\":\"nonpayable\",\"type\":\"function\"},{\"inputs\":[],\"name\":\"renounceOwnership\",\"outputs\":[],\"stateMutability\":\"nonpayable\",\"type\":\"function\"},{\"inputs\":[{\"internalType\":\"address\",\"name\":\"newOwner\",\"type\":\"address\"}],\"name\":\"transferOwnership\",\"outputs\":[],\"stateMutability\":\"nonpayable\",\"type\":\"function\"},{\"inputs\":[{\"internalType\":\"address\",\"name\":\"_address\",\"type\":\"address\"}],\"name\":\"whitelisted\",\"outputs\":[{\"internalType\":\"bool\",\"name\":\"\",\"type\":\"bool\"}],\"stateMutability\":\"view\",\"type\":\"function\"}]"
	parsed, err := abi.JSON(strings.NewReader(AuthControllerABI))
	if err != nil {
		fmt.Println(err)
	}
	ret1, err := parsed.Pack("authsSingle", common.HexToAddress("0xb3a514a6b22abd8e5c2f19ba775038200daba087"))
	fmt.Println(err)
	fmt.Println(common.Bytes2Hex(ret1))

	//fmt.Println(out)
	ret, _ := parsed.Unpack("authsSingle", common.FromHex("0x000000000000000000000000000000000000000000000000000000000000002000000000000000000000000039a1e670db3f586122150067f79937716dd4823000000000000000000000000039a1e670db3f586122150067f79937716dd48230000000000000000000000000000000000000000000000000000000000000008000000000000000000000000000000000000000000000000000000000000000010000000000000000000000000000000000000000000000000000000000000000"))
	fmt.Println(ret)

	out0 := *abi.ConvertType(ret[0],
		new(AuthControllerAuthData)).(*AuthControllerAuthData)

	fmt.Println(out0.IsAuth)
}
