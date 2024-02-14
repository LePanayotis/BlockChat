package blockchat

import (
	"os"
	"strconv"
	"time"

	"github.com/segmentio/kafka-go"
)

// Essential struct for the module which represents the node
// and the way the process is running.
type nodeConfig struct {
	// The following fields should be remain unchanged after initial configuration

	// Instance essentials
	blockchainPath string // The path where the blockchain json representation is stored, can be absolute or relative
	dbPath         string // The path where the database json representation is stored, can be absolute or relative
	inputPath      string // The path where the transactions input is stored, can be absolute or relative
	logPath string // The path where the
	defaultStake float64

	id       int            // Node id varying from 0 to nodes-1 (number of nodes in cluster-1)
	idString string         // String representation of node id (above)
	headers  []kafka.Header // The headers sent with every message to kafka. Includes node id and its public key

	startTime string // The timestamp when the node service instance started, format set by timeFormat constant

	publicKey  string // Node's public key
	privateKey string // Node's private key

	capacity int // Maximum number of transactions that can exist in a node before validator node broadcasts it
	nodes    int // Number of nodes in the cluster

	nodeMap map[string]int // Maps wallet addresses to node ids
	idArray []string       // Maps node ids to wallet addresses (nodes' public keys)

	protocol  string // Protocol used for remote procedure calls (RPCs), possible values: "unix","tcp","tcp4","tcp6","udp"
	socket    string // Socket used for remote procedure calls (RPCs), possible values: "ipaddres:port","/path/to/unix/socket",":port" (port->number, ipaddress->ip such as 127.0.0.1)
	brokerURL string // URL of the kafka message broker (host:port such as "localhost:9094")

	// The following fields are subject to change by the code
	currentBlock Block      // Current block being processed by node instance
	blockchain   Blockchain // An arrary of blocks
	myDB         DBmap      // Includes in memory wallet data as generated from blockchain

	outboundNonce uint   // Temporary nonce of node
	lastHash      string // Hash of block before current block in blockchain

	useCLI bool // *************************
	// Pointers to kafka producer and consumers
	writer        *kafka.Writer
	txConsumer    *kafka.Reader
	blockConsumer *kafka.Reader
}

// Default configuration of initial instance
// Changed by environment variables and command line arguments
func DefaultNodeConfig() *nodeConfig {
	return &nodeConfig{
		blockchainPath: "blockchain.json",
		capacity:       3,
		dbPath:         "db.json",
		nodes:          1,
		socket:         "localhost:1500",
		protocol:       "tcp",
		brokerURL:      "localhost:9094",
		id:             0,
		lastHash:       genesisHash,
		idString:       "0",
		outboundNonce:  0,
		useCLI:         false,
	}
}

// Takes existing environment variables and overrides node configuration
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
	v, found = os.LookupEnv("INPUT_PATH")
	if found && v != "" {
		node.inputPath = v
	}
	v, found = os.LookupEnv("LOG_PATH")
	if found && v != "" {
		node.logPath = v
	}
	v, found = os.LookupEnv("CAPACITY")
	if found && v != "" {
		// If CAPACITY not integer, default capacity 3 is used
		node.capacity, err = strconv.Atoi(v)
		if err != nil {
			node.capacity = 3
		}
	}
	v, found = os.LookupEnv("NODE_ID")
	if found && v != "" {
		node.id, err = strconv.Atoi(v)
		if err != nil {
			// If NODE_ID not integer, 0 is used
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
	v, found = os.LookupEnv("DEFAULT_STAKE")
	if found && v != "" {

		node.defaultStake, err = strconv.ParseFloat(v,64)
		if err != nil {
			node.defaultStake = 1
		}
	}
	return nil
}

// After configuring environment and in the case the process calles
// is meant to be a service, starts configuration of other necessary
// parameters to function as node (not CLI)
func (node *nodeConfig) initialConfig() {

	// Creates empty node map and id array
	node.nodeMap = make(map[string]int)
	node.idArray = make([]string, node.nodes)

	// Generates keys and updates above data structures
	node.generateKeysUpdate()

	// Sets start time
	node.startTime = time.Now().UTC().Format(timeFormat)
	// Initial last hash is genesis hash
	node.lastHash = genesisHash

	node.idString = strconv.Itoa(node.id)

	// Sets node headers for kafka messages
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

	// Creates kafka producer instance
	node.writer = &kafka.Writer{
		Addr: kafka.TCP(node.brokerURL),
	}

	logger.Info("Initial configuration done")
}

