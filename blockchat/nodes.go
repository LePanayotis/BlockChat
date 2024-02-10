package blockchat

import (
	"context"
	"encoding/json"
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
	transactionsInBlock int    `default:"0"`
	blockIndex          int    `default:"0"`
	lastHash            string `default:"1"`
	nodeMap             map[string]int
	nodeIdArray         []string
	idString            string `default:"0"`
	startTime           string
	myHeaders           []kafka.Header
	currentPublicKey    string
	currentPrivateKey   string
	myBlockchain        Blockchain
	myDB                DBmap
	writer              *kafka.Writer
	txConsumer          *kafka.Reader
	blockConsumer       *kafka.Reader
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
	transactionsInBlock: 0,
	blockIndex:          0,
	lastHash:            "1",
	idString:            "0",
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

var myNonce uint = 1

var logger *slog.Logger = slog.Default()

func newNodeEnter() error {

	var w *kafka.Writer = &kafka.Writer{
		Addr: kafka.TCP(node.brokerURL),
	}
	var r *kafka.Reader = kafka.NewReader(kafka.ReaderConfig{
		Brokers:     []string{node.brokerURL},
		Topic:       "welcome",
		StartOffset: kafka.LastOffset,
		GroupID:     node.idString,
	})

	logger.Info("Kafka producer and welcome consumer connected")

	err := declareExistence(w)
	if err != nil {
		logger.Error("Could not declare existence to other nodes", err)
		return err
	}

	logger.Info("Broadcasted existence to other nodes")

	for {
		m, err := r.ReadMessage(context.Background())
		if err != nil {
			logger.Error("Node could not get welcome message", err)
			continue
		}
		logger.Info("Received welcome message")

		t, _ := time.Parse(node.timeFormat, node.startTime)
		if m.Time.Before(t) {
			logger.Error("Received messages created before: "+node.startTime, err)
			continue
		}

		var welcomeMessage welcomeMessage
		err = json.Unmarshal(m.Value, &welcomeMessage)
		if err != nil {
			logger.Error("Parsing welcome message failed", err)
			return err
		}
		logger.Info("Successfully parsed welcome message")

		node.myBlockchain = welcomeMessage.Bc
		node.nodeIdArray = welcomeMessage.NodesIn[:]

		node.myBlockchain.WriteBlockchain()
		if err != nil {
			logger.Error("Could not write blockchain", err)
			return err
		}
		logger.Info("Successfully wrote blockchain to file")

		node.myDB, err = node.myBlockchain.MakeDB()
		if err != nil {
			logger.Error("Creating database from blockchain failed", err)
			return err
		}
		logger.Info("Successfully created database")

		err = node.myDB.WriteDB()
		if err != nil {
			logger.Error("Could not write database to file", err)
			return err
		}
		logger.Info("Successfully stored database in file")

		break
	}

	go func() {
		r.Close()
		w.Close()
		logger.Info("Successfully stored database in file")
	}()
	return nil
}

func startEnv() {

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

func blockListener() error {

	_lasthash := node.myBlockchain[len(node.myBlockchain)-1].Current_hash

	_index := node.myBlockchain[len(node.myBlockchain)-1].Index + 1

	node.currentBlock = NewBlock(_index, _lasthash)

	for {
		block, validator, err := getNewBlock(node.blockConsumer)
		if err != nil {
			logger.Warn("Block listener exiting")
			return err
		}
		logger.Info("Received new block")

		if node.currentBlock.Current_hash == block.Current_hash && validator == node.currentBlock.Validator {

			node.myBlockchain = append(node.myBlockchain, block)
			node.currentBlock = NewBlock(block.Index+1, block.Current_hash)

			logger.Info("New block accepted", "validator node:", node.nodeMap[validator])

			err = node.myDB.addBlockUndoStake(&block)
			if err != nil {
				logger.Error("Failed adding block to database", err)
				return err
			}

			err = node.myDB.WriteDB()
			if err != nil {
				logger.Error("Failed writing database", err)
				return err
			}
			err = node.myBlockchain.WriteBlockchain()
			if err != nil {
				logger.Error("Failed writing blockchain", err)
				return err
			}
			node.transactionsInBlock = 0

			logger.Info("Block add routine completed")

		}
	}
}

func transactionListener() error {
	node.transactionsInBlock = 0
	for {
		tx, err := getNewTransaction(node.txConsumer)
		if err != nil {
			logger.Warn("Transaction listener exiting")
			return err
		}

		logger.Info("New ransaction received")

		if node.myDB.isTransactionPossible(&tx) {
			if node.transactionsInBlock < node.capacity {

				_, err = node.myDB.addTransaction(&tx)
				if err != nil {
					logger.Error("Failed adding transaction to database", err)
					return err
				}
				logger.Info("Transaction added to database")

				_, err = node.currentBlock.AddTransaction(&tx)
				if err != nil {
					logger.Error("Failed adding transaction to current Block", err)
					return err
				}

				node.transactionsInBlock++

				if node.transactionsInBlock == node.capacity {
					logger.Info("Block capacity reached")

					node.currentBlock.SetValidator()

					node.currentBlock.CalcHash()

					if node.currentBlock.Validator == node.myPublicKey {
						logger.Info("The node is broadcaster")
						node.currentBlock.Timestamp = time.Now().UTC().Format(node.timeFormat)
						err = broadcastBlock(node.writer, node.currentBlock)
						if err != nil {
							logger.Error("Failed to broadcast new block", err)
							return err
						}

						logger.Info("Block broadcasted by me")
					}
				}
			} else {
				logger.Warn("Capacity Reached")
			}

		} else {
			logger.Error("Transaction rejected")
		}

	}
}

func StartNode() error {
	startEnv()

	var err error
	var myBlockchain *Blockchain = &node.myBlockchain

	if node.id == 0 {
		logger.Info("Node is bootstrap node")
		genesis := GenesisBlock(node.myPublicKey, node.myPrivateKey)

		logger.Info("Genesis block created")
		*myBlockchain = append(*myBlockchain, genesis)

		err = myBlockchain.WriteBlockchain()
		if err != nil {
			logger.Error("Could not write genesis blockchain", err)
			return err
		}
		logger.Info("Blockchain written successfully")

		err = collectNodesInfo()
		if err != nil {
			logger.Error("Could not collect nodes info", err)
		}

	} else {
		logger.Info("Node is ordinary")
		err = newNodeEnter()
		if err != nil {
			logger.Error("Could not enter the cluster", err)
			return err
		}
	}

	//Start node kafka consumers and writers
	node.writer = &kafka.Writer{
		Addr: kafka.TCP(node.brokerURL),
	}
	node.txConsumer = kafka.NewReader(kafka.ReaderConfig{
		Brokers:     []string{node.brokerURL},
		Topic:       "post-transaction",
		StartOffset: kafka.LastOffset,
		GroupID:     node.idString,
	})
	node.blockConsumer = kafka.NewReader(kafka.ReaderConfig{
		Brokers:     []string{node.brokerURL},
		Topic:       "post-block",
		StartOffset: kafka.LastOffset,
		GroupID:     node.idString,
	})

	go blockListener()
	go startRPC()
	transactionListener()
	return nil
}
