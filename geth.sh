go run ./cmd/geth --http \
--http.api "eth,net,engine,admin,web3,debug,txpool" \
--http.addr=0.0.0.0 \
--authrpc.vhosts=* \
--authrpc.addr=0.0.0.0 \
--authrpc.jwtsecret=/mnt/d/Shqy/server/projects/2022/chain/ct-network-config/execution/jwtsecret \
--allow-insecure-unlock \
--unlock=0x123463a4b065722e99115d6c222f267d9cabb524 \
--password=/mnt/d/Shqy/server/projects/2022/chain/ct-network-config/execution/geth_password.txt \
--nodiscover \
--syncmode=full \
--mine \
console
