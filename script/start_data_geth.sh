nohup ./geth --datadir /data/nodedata/datanode \
--port 17402 \
--rpcapi eth,net,txpool,admin,debug,clique,web3 \
--rpc \
--rpcport 7402 \
--rpcaddr 0.0.0.0 \
--syncmode=full \
--gcmode=archive \
--allow-insecure-unlock \
--rpccorsdomain "*" \
--bootnodes "enode://1ecea9891dd889f9ffdaa391f46b4f43857a6b7e3eddedc834a9feb88ca27bc3e6c51c404b7d0e03dda8e92042b89f796b2bba67cbdd020dac037b4ec49b5c60@121.40.143.162:30300,enode://1ecea9891dd889f9ffdaa391f46b4f43857a6b7e3eddedc834a9feb88ca27bc3e6c51c404b7d0e03dda8e92042b89f796b2bba67cbdd020dac037b4ec49b5c60@172.31.100.129:30300,enode://1ecea9891dd889f9ffdaa391f46b4f43857a6b7e3eddedc834a9feb88ca27bc3e6c51c404b7d0e03dda8e92042b89f796b2bba67cbdd020dac037b4ec49b5c60@118.31.15.145:30300,enode://1ecea9891dd889f9ffdaa391f46b4f43857a6b7e3eddedc834a9feb88ca27bc3e6c51c404b7d0e03dda8e92042b89f796b2bba67cbdd020dac037b4ec49b5c60@172.16.169.162:30300" \
> datanode.out 2>&1 &