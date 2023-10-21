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

./build/bin/geth init --datadir ~/chaindata ./tests/EVMPlus/genesis.json
./build/bin/geth --http --http.api web3,eth,net --datadir ~/chaindata --networkid 196790 --nat extip:35.209.100.125

// remember to open HTTP port, any other?


./build/bin/geth init --datadir ~/Downloads/chaindata ./tests/EVMPlus/genesis.json
./build/bin/geth --http --http.api web3,eth,net --datadir ~/Downloads/chaindata --networkid 196790 --dev --nat extip:35.209.100.125
./build/bin/geth --http --http.api web3,eth,debug,personal,net --vmdebug --datadir ~/Downloads/chaindata --dev --nodiscover console