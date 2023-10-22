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
make all

tmux new -s bootnode

./build/bin/bootnode --genkey=boot.key
./build/bin/bootnode --nodekey=boot.key

./build/bin/geth account new --datadir ~/chaindata
// add public address to genesis.json following https://geth.ethereum.org/docs/fundamentals/private-network#clique-example
./build/bin/geth init --datadir ~/chaindata ./tests/EVMPlus/genesis.json
./build/bin/geth --datadir ~/chaindata --networkid 196790 --port 30313 --nat extip:35.209.100.125 --unlock 0x58c4c45c9b5954ee15E81C0FB7437DCaCEAD665e --password ~/chaindata/password.txt --mine --miner.etherbase=0x58c4c45c9b5954ee15E81C0FB7437DCaCEAD665e

// remember to open HTTP port, any other?

./build/bin/geth attach --exec admin.nodeInfo.enr ~/chaindata/geth.ipc
enr:-KO4QIzzViglXefgjUBy1A2V_t1BrjZ1dgcn8Lu2Yes0ZUjBf-QJ7bp47KVUQil_WVtA89idZCnluKLkffQ0Ns0czYWGAYtVNkH0g2V0aMfGhKHCtOKAgmlkgnY0gmlwhCPRZH2Jc2VjcDI1NmsxoQNZ5N2y4Z2Ehb7UQVuKC24a1CWntX4b2OEBesNFVnaX9IRzbmFwwIN0Y3CCdmmDdWRwgnZp

// member node steps

./build/bin/geth account new --datadir ~/chaindata
./build/bin/geth init --datadir ~/chaindata ./tests/EVMPlus/genesis.json
./build/bin/geth --datadir ~/chaindata --networkid 196790 --port 30313 --unlock 0xb316c8ca80e0dce73f3a81338ed97f31fe0a31eb --password ~/chaindata/password.txt --bootnodes enr:-KO4QIzzViglXefgjUBy1A2V_t1BrjZ1dgcn8Lu2Yes0ZUjBf-QJ7bp47KVUQil_WVtA89idZCnluKLkffQ0Ns0czYWGAYtVNkH0g2V0aMfGhKHCtOKAgmlkgnY0gmlwhCPRZH2Jc2VjcDI1NmsxoQNZ5N2y4Z2Ehb7UQVuKC24a1CWntX4b2OEBesNFVnaX9IRzbmFwwIN0Y3CCdmmDdWRwgnZp
./build/bin/geth attach ~/chaindata/geth.ipc

// write password from account new step into file
echo "password" >> ~/chaindata/password.txt
./build/bin/geth --unlock 0xb316c8ca80e0dce73f3a81338ed97f31fe0a31eb --password ~/chaindata/password.txt attach ~/chaindata/geth.ipc

// send gas to friends
eth.sendTransaction({from: eth.accounts[0], to: "0xd442f325d8B7491029417b87607e35DA9A8F4619", value: 55})
