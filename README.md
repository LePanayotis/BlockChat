# BlockChat

BlockChat is a simple implementation of a cryptocurrency and message exchange application  based on blockchain.
The BlockChat cluster consists of one bootstrap node and zero or more other nodes.

The application is built in Go 1.22 and consists of a daemon and a CLI. See instructions below on how to access both modes.

For demonstrative purposes of this repo, the application is dockerised and deployed in multiple instances.

## Requirements

Before trying to install and deploy the application confirm that you have the following installed and configured in your system:

1. **Go 1.22** ![golang](https://img.shields.io/badge/Go-1.22-blue)

2. **Kafka**  ![kafka](https://img.shields.io/badge/Kafka-blue)

## Build and Run

Download repo and switch to its directory

```bash
git clone https://github.com/LePanayotis/BlockChat.git
cd BlockChat
```

Produce the executable

```bash
go build ./src -o blockchat.exe
```

It is necessary to have a Kakfa broker running and accessible to the nodes.
If you are using Windows remove "sudo" from the commands.

```bash
sudo docker network create cluster

sudo docker volume create kafka-volume

sudo docker run -d --network cluster -p 9094:9094 -v kafka-volume:/bitnami `
    --name kafka `
    -h kafka `
    -e KAFKA_CFG_NODE_ID=0 `
    -e KAFKA_CFG_PROCESS_ROLES=controller,broker `
    -e KAFKA_CFG_LISTENERS=PLAINTEXT://:9092,CONTROLLER://:9093,EXTERNAL://0.0.0.0:9094 `
    -e KAFKA_CFG_ADVERTISED_LISTENERS=PLAINTEXT://kafka:9092,EXTERNAL://127.0.0.1:9094 `
    -e KAFKA_CFG_LISTENER_SECURITY_PROTOCOL_MAP=CONTROLLER:PLAINTEXT,EXTERNAL:PLAINTEXT,PLAINTEXT:PLAINTEXT `
    -e KAFKA_CFG_CONTROLLER_QUORUM_VOTERS=0@kafka:9093 `
    -e KAFKA_CFG_CONTROLLER_LISTENER_NAMES=CONTROLLER `
    -e KAFKA_CFG_LOG_RETENTION_MS=60000 `
    -e KAFKA_CFG_LOG_RETENTION_CHECK_INTERVAL_MS=10000 `
    bitnami/kafka:latest
```

Create the following topics on kafka:

* welcome
* enter
* post-bloc
* post-transaction

## Environment Variables

Provide the configuration for the application

```bash
nano .env
```

In the `.env` file, you have to povide the following environment variables:

```env
# Socket and protocol used for RPCs (CLI and daemon communication)
SOCKET=localhost:1500
PROTOCOL=tcp #Can be: tcp, udp or unix

# SAME FOR ALL THE NODES
BROKER_URL=kafka:9092
CAPACITY=3
NODES=1

# THE FOLLOWING CUSTOMISABLE FOR EACH NODE

# File to store blockchain
BLOCKCHAIN_PATH=blockchain.json

# File to store database
DB_PATH=db.json

# Node's integer ID, must be from 0 to NODES-1.
# Each node in the cluster must have a unique ID
NODE_ID=0

# When INPUT_PATH is provided, the node will first 
# send all transactions commanded in the input file
# Used for demonstrative purposes
# INPUT_PATH=./input/trans0.txt

# Defines Stake sent right after each transaction in the input file
# DEFAULT_STAKE=10
```

This file must be located in the directory from which you call 'blockchat.exe'

Execute './blockchat.exe' and see help

```bash
./blockchat.exe help
```

In case you want to containerise the program:

```bash
sudo docker build . -t blockchat
```

Deploy the container with a file containing the configuration environment variables, for example:

```bash
sudo docker run --env-file '/path/to/config/file' -p "1500:1500"blockchat
```

Port forwarding is added for the RPCs to be accessible in the host machine

## Contributors

* [**Panagiotis Papagiannakis**](mailto:el19055@mail.ntua.gr)
* [**Georgios Gkotzias**](mailto:el19046@mail.ntua.gr)

9th term students, School of Electrical and Computer Engineering, National Technical University of Athens, Greece

## License

The program is licensed under the MIT license, for details see LICENSE.md

## Disclaimer

![It ain't much, but it's honest work](https://media.npr.org/assets/img/2023/05/26/honest-work-meme-cb0f0fb2227fb84b77b3c9a851ac09b095ab74d8-s1100-c50.jpg)