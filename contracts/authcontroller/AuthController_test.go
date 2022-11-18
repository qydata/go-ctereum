package authcontroller

import (
	"fmt"
	"github.com/ethereum/go-ctereum/ethclient"
	"math/big"
	"testing"
)

func TestAuthBoth(t *testing.T) {
	// Initialize test accounts

	// Deploy registrar contract

	// 3 trusted signers, threshold 2

	//连接本地的以太坊私链（一定要保证本地以太坊私链已经启动）
	conn, err := ethclient.Dial("http://ctblock.cn/blockChain")

	fmt.Println("connect to local geth node...", conn)
	if err != nil {
		fmt.Println("could not connect to local node: ", err)
	}
	fmt.Println("get the contract object...")
	fmt.Println("isFlag:", big.NewInt(40).Cmp(big.NewInt(50)) <= 0)
	fmt.Println("isFlagq:", big.NewInt(4800000000000).Cmp(big.NewInt(5000100000000)) <= 0)
	if err != nil {
		fmt.Println("Failed to instantiate a Token contract: ", err)
	}
}
