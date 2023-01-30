rm -rf build/bin/
go build ./cmd/geth
go build ./cmd/bootnode
mv geth ../ct-network-config/client/
mv bootnode ../ct-network-config/client/
