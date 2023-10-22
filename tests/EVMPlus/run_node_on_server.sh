sudo apt update
sudo apt install git -y
sudo apt install make -y
sudo apt install tmux -y

(https://go.dev/doc/install)
wget https://go.dev/dl/go1.21.3.linux-amd64.tar.gz
sudo rm -rf /usr/local/go && sudo tar -C /usr/local -xzf go1.21.3.linux-amd64.tar.gz
nano .profile
// add export PATH=$PATH:/usr/local/go/bin to .profile
source .profile

git clone https://github.com/1m1-github/go-ethereum-plus.git
cd go-ethereum-plus
make geth

tmux new -s geth

./build/bin/geth account new --datadir ~/chaindata
// add public address to genesis.json following https://geth.ethereum.org/docs/fundamentals/private-network#clique-example
./build/bin/geth init --datadir ~/chaindata ./tests/EVMPlus/genesis.json
./build/bin/geth --datadir ~/chaindata --networkid 196790 --nat extip:35.209.100.125

// remember to open HTTP port, any other?

./build/bin/geth attach --exec admin.nodeInfo.enr ~/chaindata/geth.ipc
enr:-KO4QCnVVlXYZrs6y8EIOjUlR6pqtO0BAkLxKQxvW3B72muNCmbpnqrXx8KOGcoBsyWsiCJW9h32uTe7SHMmejkFRG2GAYtUsPIFg2V0aMfGhDgzR7mAgmlkgnY0gmlwhCPRZH2Jc2VjcDI1NmsxoQNU9riSVzaP3AwzmhA3-GH78YP7TTFwo_KP70rB5BWZJIRzbmFwwIN0Y3CCdl-DdWRwgnZf

./build/bin/geth init --datadir ~/Downloads/chaindata1 ./tests/EVMPlus/genesis.json
./build/bin/geth --datadir ~/Downloads/chaindata1 --networkid 196790 --port 30303 --bootnodes enr:-KO4QCnVVlXYZrs6y8EIOjUlR6pqtO0BAkLxKQxvW3B72muNCmbpnqrXx8KOGcoBsyWsiCJW9h32uTe7SHMmejkFRG2GAYtUsPIFg2V0aMfGhDgzR7mAgmlkgnY0gmlwhCPRZH2Jc2VjcDI1NmsxoQNU9riSVzaP3AwzmhA3-GH78YP7TTFwo_KP70rB5BWZJIRzbmFwwIN0Y3CCdl-DdWRwgnZf --authrpc.port 8551
./build/bin/geth --datadir ~/Downloads/chaindata2 --networkid 196790 --port 30304 --bootnodes enr:-KO4QCnVVlXYZrs6y8EIOjUlR6pqtO0BAkLxKQxvW3B72muNCmbpnqrXx8KOGcoBsyWsiCJW9h32uTe7SHMmejkFRG2GAYtUsPIFg2V0aMfGhDgzR7mAgmlkgnY0gmlwhCPRZH2Jc2VjcDI1NmsxoQNU9riSVzaP3AwzmhA3-GH78YP7TTFwo_KP70rB5BWZJIRzbmFwwIN0Y3CCdl-DdWRwgnZf --authrpc.port 8552

./build/bin/geth attach ~/Downloads/chaindata1/geth.ipc