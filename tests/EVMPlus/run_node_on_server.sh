// server bootnode steps

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
./build/bin/geth --rpcapi eth,personal,net,admin,web3 --datadir ~/chaindata --networkid 196790 --port 30313 --nat extip:35.209.100.125

// remember to open HTTP port, any other?

./build/bin/geth attach --exec admin.nodeInfo.enr ~/chaindata/geth.ipc
enr:-KO4QAHtFu3-uVxR29yZAcfFxbOfGQCVDBz4Ld5BHSAMN6Mwe_jCO_JGl0VZ6GFSJk70T99JmBY0Wq5z_x70IWzUY8GGAYtU7X0Ug2V0aMfGhAE7G9WAgmlkgnY0gmlwhCPRZH2Jc2VjcDI1NmsxoQPTdT-sDAaR-zugRyw5nb6AFApvJND81gu1zmlDp2fi-oRzbmFwwIN0Y3CCdmmDdWRwgnZp

// write password from account new step into file
echo "password" >> ~/chaindata/password.txt
./build/bin/geth --unlock 0xa26c76a509795d921539c189c139e870666553a7 --password ~/chaindata/password.txt attach ~/chaindata/geth.ipc

// send gas to friends
eth.sendTransaction({from: eth.accounts[0], to: "0xd442f325d8B7491029417b87607e35DA9A8F4619", value: 55})

// member node steps

./build/bin/geth account new --datadir ~/chaindata
./build/bin/geth init --datadir ~/chaindata ./tests/EVMPlus/genesis.json
./build/bin/geth --datadir ~/chaindata --networkid 196790 --port 30313 --bootnodes enr:-KO4QAHtFu3-uVxR29yZAcfFxbOfGQCVDBz4Ld5BHSAMN6Mwe_jCO_JGl0VZ6GFSJk70T99JmBY0Wq5z_x70IWzUY8GGAYtU7X0Ug2V0aMfGhAE7G9WAgmlkgnY0gmlwhCPRZH2Jc2VjcDI1NmsxoQPTdT-sDAaR-zugRyw5nb6AFApvJND81gu1zmlDp2fi-oRzbmFwwIN0Y3CCdmmDdWRwgnZp
./build/bin/geth attach ~/chaindata/geth.ipc

