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
./build/bin/geth --datadir ~/chaindata --networkid 196790 --port 30313 --nat extip:35.209.100.125

// remember to open HTTP port, any other?

./build/bin/geth attach --exec admin.nodeInfo.enr ~/chaindata/geth.ipc
enr:-KO4QMDjuzf4kyaxHbu6erV9l6ekJHKTUGCoK5nmAryLsluSAwWYbpXtJNocI6T8ePAivTTwL7e2zSnHpvqMSSDdzOuGAYtU32_5g2V0aMfGhLFOWi2AgmlkgnY0gmlwhCPRZH2Jc2VjcDI1NmsxoQP-8pqTWte8Djic2jDtFQ_iwB9NTUw3crn-IV4YCpUvDoRzbmFwwIN0Y3CCdl-DdWRwgnZf

// member node steps

./build/bin/geth init --datadir ~/Downloads/chaindata ./tests/EVMPlus/genesis.json
./build/bin/geth account new --datadir ~/Downloads/chaindata
./build/bin/geth --datadir ~/Downloads/chaindata --networkid 196790 --port 30303 --bootnodes enr:-KO4QMDjuzf4kyaxHbu6erV9l6ekJHKTUGCoK5nmAryLsluSAwWYbpXtJNocI6T8ePAivTTwL7e2zSnHpvqMSSDdzOuGAYtU32_5g2V0aMfGhLFOWi2AgmlkgnY0gmlwhCPRZH2Jc2VjcDI1NmsxoQP-8pqTWte8Djic2jDtFQ_iwB9NTUw3crn-IV4YCpUvDoRzbmFwwIN0Y3CCdl-DdWRwgnZf --authrpc.port 8551
./build/bin/geth attach ~/Downloads/chaindata/geth.ipc

--unlock 0xC1B2c0dFD381e6aC08f34816172d6343Decbb12b --password node1/password.txt
personal.unlockAccount(eth.accounts[0], "", 300)
eth.sendTransaction({from: eth.accounts[0], to: "0xb3360a6ef50f5a540c5d6aa99fe5d2467a1d342a", value: 55})
