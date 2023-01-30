./geth --http \
--http.api "db,eth,net,personal,txpool,admin,debug,miner,clique,web3,trace" \
--http.addr=0.0.0.0 \
--authrpc.vhosts=* \
--authrpc.addr=0.0.0.0 \
--authrpc.jwtsecret=/root/ct-network-config/execution/jwtsecret \
--allow-insecure-unlock \
--unlock=0x123463a4b065722e99115d6c222f267d9cabb524 \
--password=/root/ct-network-config/execution/geth_password.txt \
--nodiscover \
--syncmode=full \
--mine \
console