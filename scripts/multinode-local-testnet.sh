#!/bin/bash
set -e

# always returns true so set -e doesn't exit if it is not running.
killall blackfuryd || true
rm -rf $HOME/.blackfuryd/

# make four blackfury directories
mkdir $HOME/.blackfuryd
mkdir $HOME/.blackfuryd/validator1
mkdir $HOME/.blackfuryd/validator2
mkdir $HOME/.blackfuryd/validator3

# init all three validators
blackfuryd init --chain-id=testing validator1 --home=$HOME/.blackfuryd/validator1
blackfuryd init --chain-id=testing validator2 --home=$HOME/.blackfuryd/validator2
blackfuryd init --chain-id=testing validator3 --home=$HOME/.blackfuryd/validator3
# create keys for all three validators
blackfuryd keys add validator1 --keyring-backend=test --home=$HOME/.blackfuryd/validator1
blackfuryd keys add validator2 --keyring-backend=test --home=$HOME/.blackfuryd/validator2
blackfuryd keys add validator3 --keyring-backend=test --home=$HOME/.blackfuryd/validator3

# create validator node with tokens to transfer to the three other nodes
blackfuryd add-genesis-account $(blackfuryd keys show validator1 -a --keyring-backend=test --home=$HOME/.blackfuryd/validator1) 100000000000stake --home=$HOME/.blackfuryd/validator1
blackfuryd gentx validator1 500000000stake --keyring-backend=test --home=$HOME/.blackfuryd/validator1 --chain-id=testing
blackfuryd collect-gentxs --home=$HOME/.blackfuryd/validator1

# change app.toml values
VALIDATOR2_APP_TOML=$HOME/.blackfuryd/validator2/config/app.toml
VALIDATOR3_APP_TOML=$HOME/.blackfuryd/validator3/config/app.toml

# validator2
sed -i -E 's|tcp://0.0.0.0:1317|tcp://0.0.0.0:1316|g' $VALIDATOR2_APP_TOML
sed -i -E 's|0.0.0.0:9090|0.0.0.0:9088|g' $VALIDATOR2_APP_TOML
sed -i -E 's|0.0.0.0:9091|0.0.0.0:9089|g' $VALIDATOR2_APP_TOML

# validator3
sed -i -E 's|tcp://0.0.0.0:1317|tcp://0.0.0.0:1315|g' $VALIDATOR3_APP_TOML
sed -i -E 's|0.0.0.0:9090|0.0.0.0:9086|g' $VALIDATOR3_APP_TOML
sed -i -E 's|0.0.0.0:9091|0.0.0.0:9087|g' $VALIDATOR3_APP_TOML


# change config.toml values
VALIDATOR1_CONFIG=$HOME/.blackfuryd/validator1/config/config.toml
VALIDATOR2_CONFIG=$HOME/.blackfuryd/validator2/config/config.toml
VALIDATOR3_CONFIG=$HOME/.blackfuryd/validator3/config/config.toml

# validator1
sed -i -E 's|allow_duplicate_ip = false|allow_duplicate_ip = true|g' $VALIDATOR1_CONFIG
# validator2
sed -i -E 's|tcp://127.0.0.1:26658|tcp://127.0.0.1:26655|g' $VALIDATOR2_CONFIG
sed -i -E 's|tcp://127.0.0.1:26657|tcp://127.0.0.1:26654|g' $VALIDATOR2_CONFIG
sed -i -E 's|tcp://0.0.0.0:26656|tcp://0.0.0.0:26653|g' $VALIDATOR2_CONFIG
sed -i -E 's|tcp://0.0.0.0:26656|tcp://0.0.0.0:26650|g' $VALIDATOR2_CONFIG
sed -i -E 's|allow_duplicate_ip = false|allow_duplicate_ip = true|g' $VALIDATOR2_CONFIG
# validator3
sed -i -E 's|tcp://127.0.0.1:26658|tcp://127.0.0.1:26652|g' $VALIDATOR3_CONFIG
sed -i -E 's|tcp://127.0.0.1:26657|tcp://127.0.0.1:26651|g' $VALIDATOR3_CONFIG
sed -i -E 's|tcp://0.0.0.0:26656|tcp://0.0.0.0:26650|g' $VALIDATOR3_CONFIG
sed -i -E 's|tcp://0.0.0.0:26656|tcp://0.0.0.0:26650|g' $VALIDATOR3_CONFIG
sed -i -E 's|allow_duplicate_ip = false|allow_duplicate_ip = true|g' $VALIDATOR3_CONFIG

# copy validator1 genesis file to validator2-3
cp $HOME/.blackfuryd/validator1/config/genesis.json $HOME/.blackfuryd/validator2/config/genesis.json
cp $HOME/.blackfuryd/validator1/config/genesis.json $HOME/.blackfuryd/validator3/config/genesis.json

# copy tendermint node id of validator1 to persistent peers of validator2-3
sed -i -E "s|persistent_peers = \"\"|persistent_peers = \"$(blackfuryd tendermint show-node-id --home=$HOME/.blackfuryd/validator1)@localhost:26656\"|g" $HOME/.blackfuryd/validator2/config/config.toml
sed -i -E "s|persistent_peers = \"\"|persistent_peers = \"$(blackfuryd tendermint show-node-id --home=$HOME/.blackfuryd/validator1)@localhost:26656\"|g" $HOME/.blackfuryd/validator3/config/config.toml


# start all three validators
tmux new -s validator1 -d blackfuryd start --home=$HOME/.blackfuryd/validator1
tmux new -s validator2 -d blackfuryd start --home=$HOME/.blackfuryd/validator2
tmux new -s validator3 -d blackfuryd start --home=$HOME/.blackfuryd/validator3

# send stake from first validator to second validator
echo "Waiting 7 seconds to send funds to validators 2 and 3..."
sleep 7
blackfuryd tx bank send validator1 $(blackfuryd keys show validator2 -a --keyring-backend=test --home=$HOME/.blackfuryd/validator2) 500000000stake --keyring-backend=test --home=$HOME/.blackfuryd/validator1 --chain-id=testing --broadcast-mode block --node http://localhost:26657 --yes
blackfuryd tx bank send validator1 $(blackfuryd keys show validator3 -a --keyring-backend=test --home=$HOME/.blackfuryd/validator3) 400000000stake --keyring-backend=test --home=$HOME/.blackfuryd/validator1 --chain-id=testing --broadcast-mode block --node http://localhost:26657 --yes
