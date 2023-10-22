// the below is work in progress...it did not completely work, for different reasons
// currenrtly, i am running a mining geth node on a server with http and ws opened for testing

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
./build/bin/bootnode --nodekey=boot.key -addr :30313

// above gives this:
export ENODE=enode://b5576d7ca1a7960a661a6e9c7e350fd7db9b6a70d4fefb8ac9c1c530023ecbfe2627796f6db828ee126d548b02395a9c258cc02b05c5920753e80a2467c98c16@127.0.0.1:0?discport=30313

// need new terminal
tmux new -s miner

./build/bin/geth --datadir ~/chaindata --bootnodes $ENODE --mine --miner.etherbase 0x58c4c45c9b5954ee15E81C0FB7437DCaCEAD665e

// 
tmux new -s geth

./build/bin/geth --datadir ~/chaindata --networkid 196790 --port 30313 --http --http.addr 0.0.0.0 --http.port 8555 --http.corsdomain '*' --ws --ws.addr 0.0.0.0 --ws.port 8556 --ws.origins '*' --allow-insecure-unlock --unlock 0x58c4c45c9b5954ee15E81C0FB7437DCaCEAD665e --password ~/chaindata/password.txt --mine --miner.etherbase=0x58c4c45c9b5954ee15E81C0FB7437DCaCEAD665e


./build/bin/geth account new --datadir ~/chaindata
// add public address to genesis.json following https://geth.ethereum.org/docs/fundamentals/private-network#clique-example
./build/bin/geth init --datadir ~/chaindata ./tests/EVMPlus/genesis.json
./build/bin/geth --datadir ~/chaindata --networkid 196790 --port 30313 --nat extip:35.209.100.125 --unlock 0x58c4c45c9b5954ee15E81C0FB7437DCaCEAD665e --password ~/chaindata/password.txt --mine --miner.etherbase=0x58c4c45c9b5954ee15E81C0FB7437DCaCEAD665e

// remember to open HTTP port, any other?

./build/bin/geth attach --exec admin.nodeInfo.enr ~/chaindata/geth.ipc
enr:-KO4QIzzViglXefgjUBy1A2V_t1BrjZ1dgcn8Lu2Yes0ZUjBf-QJ7bp47KVUQil_WVtA89idZCnluKLkffQ0Ns0czYWGAYtVNkH0g2V0aMfGhKHCtOKAgmlkgnY0gmlwhCPRZH2Jc2VjcDI1NmsxoQNZ5N2y4Z2Ehb7UQVuKC24a1CWntX4b2OEBesNFVnaX9IRzbmFwwIN0Y3CCdmmDdWRwgnZp

// member node steps <- this part did not work

./build/bin/geth account new --datadir ~/chaindata
./build/bin/geth init --datadir ~/chaindata ./tests/EVMPlus/genesis.json
./build/bin/geth --datadir ~/chaindata --networkid 196790 --port 30313 --unlock 0xb316c8ca80e0dce73f3a81338ed97f31fe0a31eb --password ~/chaindata/password.txt --bootnodes enr:-KO4QIzzViglXefgjUBy1A2V_t1BrjZ1dgcn8Lu2Yes0ZUjBf-QJ7bp47KVUQil_WVtA89idZCnluKLkffQ0Ns0czYWGAYtVNkH0g2V0aMfGhKHCtOKAgmlkgnY0gmlwhCPRZH2Jc2VjcDI1NmsxoQNZ5N2y4Z2Ehb7UQVuKC24a1CWntX4b2OEBesNFVnaX9IRzbmFwwIN0Y3CCdmmDdWRwgnZp
./build/bin/geth attach ~/chaindata/geth.ipc

// attach

./build/bin/geth attach http://35.209.100.125:8555
./build/bin/geth --unlock 0xb316c8ca80e0dce73f3a81338ed97f31fe0a31eb --password ~/chaindata/password.txt attach http://35.209.100.125:8545