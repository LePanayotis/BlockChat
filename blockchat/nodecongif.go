package blockchat

import (
	"log/slog"
	"os"

	"strconv"
	"time"

	"github.com/segmentio/kafka-go"
)

type nodeConfig struct {
	blockchainPath string
	capacity       int
	dbPath         string
	nodes          int
	socket         string
	protocol       string
	brokerURL      string
	id             int
	publicKey      string
	privateKey     string
	currentBlock   Block
	blockIndex     int
	lastHash       string
	nodeMap        map[string]int
	idArray        []string
	idString       string
	startTime      string
	headers        []kafka.Header
	blockchain     Blockchain
	myDB           DBmap
	writer         *kafka.Writer
	txConsumer     *kafka.Reader
	blockConsumer  *kafka.Reader
	outboundNonce  uint
}

// Necessary configuration for the module
var node *nodeConfig = &nodeConfig{
	blockchainPath: "blockchain.json",
	capacity:       3,
	dbPath:         "db.json",
	nodes:          1,
	socket:         ":1500",
	protocol:       "tcp",
	brokerURL:      "localhost:9093",
	id:             0,
	blockIndex:     0,
	lastHash:       genesisHash,
	idString:       "0",
	outboundNonce:  0,
}

func (node *nodeConfig) EnvironmentConfig() error {
	var err error
	v, found := os.LookupEnv("BROKER_URL")
	if found && v != "" {
		node.brokerURL = v
	}
	v, found = os.LookupEnv("SOCKET")
	if found && v != "" {
		node.socket = v
	}
	v, found = os.LookupEnv("PROTOCOL")
	if found && v != "" {
		node.protocol = v
	}
	v, found = os.LookupEnv("BLOCKCHAIN_PATH")
	if found && v != "" {
		node.blockchainPath = v
	}
	v, found = os.LookupEnv("DB_PATH")
	if found && v != "" {
		node.dbPath = v
	}
	v, found = os.LookupEnv("CAPACITY")
	if found && v != "" {

		node.capacity, err = strconv.Atoi(v)
		if err != nil {
			node.capacity = 3
		}
	}
	v, found = os.LookupEnv("NODE_ID")
	if found && v != "" {

		node.id, err = strconv.Atoi(v)
		if err != nil {
			node.id = 0
		}
		node.idString = v
	}
	v, found = os.LookupEnv("NODES")
	if found && v != "" {

		node.nodes, err = strconv.Atoi(v)
		if err != nil {
			node.nodes = 3
		}
	}
	return nil
}

func startConfig() {

	logger.Info("Starting configuring node")
	node.nodeMap = make(map[string]int)
	node.idArray = make([]string, node.nodes)
	node.generateKeysUpdate()
	node.startTime = time.Now().UTC().Format(timeFormat)
	node.lastHash = genesisHash
	node.idString = strconv.Itoa(node.id)

	node.headers = []kafka.Header{
		{
			Key:   "NodeId",
			Value: []byte(node.idString),
		},
		{
			Key:   "NodeWallet",
			Value: []byte(node.publicKey),
		},
	}
	node.writer = &kafka.Writer{
		Addr: kafka.TCP(node.brokerURL),
	}
}

var logger *slog.Logger = slog.Default()
