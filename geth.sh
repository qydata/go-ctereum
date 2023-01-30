#go run ./cmd/geth account import \
#--password /mnt/d/Shqy/server/projects/2022/chain/ct-network-config/execution/geth_password.txt \
#/mnt/d/Shqy/server/projects/2022/chain/ct-network-config/execution/account_geth_privateKey
#--datadir /mnt/d/Shqy/server/projects/2022/chain/ct-network-config/_execution \
go run ./cmd/geth \
--http.api "eth,net,personal,txpool,admin,debug,miner,clique,web3" \
--http.addr=0.0.0.0 \
--http.port=8545 \
--authrpc.vhosts=* \
--authrpc.addr=0.0.0.0 \
--authrpc.jwtsecret=/mnt/d/Shqy/server/projects/2022/chain/ct-network-config/execution/jwtsecret \
--allow-insecure-unlock \
--unlock=0x123463a4b065722e99115d6c222f267d9cabb524 \
--password=/mnt/d/Shqy/server/projects/2022/chain/ct-network-config/execution-pw.txt \
--nodiscover \
--syncmode=full \
--mine \
console