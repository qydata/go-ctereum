package authcontroller

import (
	"fmt"
	"github.com/ethereum/go-ctereum/accounts/abi/bind"
	"github.com/ethereum/go-ctereum/common"
	"github.com/ethereum/go-ctereum/contracts/authcontroller/contract"
	"github.com/ethereum/go-ctereum/ethclient"
	"testing"
)

func TestAuthBoth(t *testing.T) {
	// Initialize test accounts

	// Deploy registrar contract

	// 3 trusted signers, threshold 2
	contractAddr := "0x1E3b0AeC6C3210680915Cc36a62AE7437425923B"

	//连接本地的以太坊私链（一定要保证本地以太坊私链已经启动）
	conn, err := ethclient.Dial("http://ctblock.cn/blockChain")

	fmt.Println("connect to local geth node...", conn)
	if err != nil {
		fmt.Println("could not connect to local node: ", err)
	}
	fmt.Println("get the contract object...")

	token, err := contract.NewAuthController(common.HexToAddress(contractAddr), conn)
	if err != nil {
		fmt.Println("Failed to instantiate a Token contract: ", err)
	}
	fmt.Println("contract token======>:", token)
	auth, _ := token.AuthsSingle(&bind.CallOpts{Pending: true}, common.HexToAddress("0x39a1E670db3F586122150067F79937716Dd48230"))
	//fmt.Printf("the total data prices and desciption are: %s\n", info);
	fmt.Println(auth)
}
