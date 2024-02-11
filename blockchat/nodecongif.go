package blockchat

import (
	"log/slog"
	"os"

	"strconv"
	"time"

	"github.com/segmentio/kafka-go"
)

type nodeConfig struct {
	feePercentage       float64 `default:"0.03"`
	costPerChar         int     `default:"1"`
	blockchainPath      string
	keyLength           int     `default:"512"`
	timeFormat          string  `default:"02-01-2006 15:04:05.000"`
	initialBCC          float64 `default:"1000"`
	capacity            int     `default:"3"`
	dbPath              string
	nodes               int    `default:"1"`
	socket              string `default:":1500"`
	protocol            string `default:"tcp4"`
	genesisHash         string `default:"1"`
	brokerURL           string `default:"localhost:9093"`
	id                  int    `default:"0"`
	myPublicKey         string
	myPrivateKey        string
	currentBlock        Block
	blockIndex          int    `default:"0"`
	lastHash            string `default:"1"`
	nodeMap             map[string]int
	nodeIdArray         []string
	idString            string `default:"0"`
	startTime           string
	myHeaders           []kafka.Header
	myBlockchain        Blockchain
	myDB                DBmap
	writer              *kafka.Writer
	txConsumer          *kafka.Reader
	blockConsumer       *kafka.Reader
	outboundNonce       uint
}

// Necessary configuration for the module
var node *nodeConfig = &nodeConfig{
	keyLength:           512,
	timeFormat:          "02-01-2006 15:04:05.000",
	feePercentage:       0.03,
	costPerChar:         1,
	blockchainPath:      "blockchain.json",
	initialBCC:          1000,
	capacity:            3,
	dbPath:              "db.json",
	nodes:               1,
	socket:              ":1500",
	protocol:            "tcp",
	genesisHash:         "1",
	brokerURL:           "localhost:9093",
	id:                  0,
	blockIndex:          0,
	lastHash:            "1",
	idString:            "0",
	outboundNonce: 0,
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
	node.nodeIdArray = make([]string, node.nodes)
	node.generateKeysUpdate()
	node.startTime = time.Now().UTC().Format(node.timeFormat)
	node.lastHash = node.genesisHash
	node.idString = strconv.Itoa(node.id)

	node.myHeaders = []kafka.Header{
		{
			Key:   "NodeId",
			Value: []byte(node.idString),
		},
		{
			Key:   "NodeWallet",
			Value: []byte(node.myPublicKey),
		},
	}
}

var logger *slog.Logger = slog.Default()
