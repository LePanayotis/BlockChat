package blockchat

import (
	"context"
	"encoding/json"
	"github.com/segmentio/kafka-go"
	"time"
)

// Actions performed when an ordinary node (id != 0) enters the cluster
func (node *nodeConfig) ordinaryNodeEnter() error {
	// Creates consumer for "welcome" topic
	var r *kafka.Reader = kafka.NewReader(kafka.ReaderConfig{
		Brokers:     []string{node.brokerURL},
		Topic:       "welcome",
		StartOffset: kafka.LastOffset,
		GroupID:     node.idString,
	})

	logger.Info("Kafka welcome consumer connected")

	// Posts node existence to bootstrap node
	err := node.declareExistence()
	if err != nil {
		logger.Error("Could not declare existence to other nodes", "error", err)
		return err
	}
	logger.Info("Successfully broadcasted existence")

	// Loop wait for valid welcome message
	for {
		// Get welcome message
		m, err := r.ReadMessage(context.Background())
		if err != nil {
			logger.Error("Node could not get welcome message", "error", err)
			continue
		}
		logger.Info("Received welcome message")

		// Check welcome message created after this node started
		t, err := time.Parse(timeFormat, node.startTime)
		if err != nil {
			logger.Warn("Could not parse start time, skips transaction")
			continue
		}
		if m.Time.Before(t) {
			// If received old welcome, continue to next iteration
			logger.Warn("Received messages created before", "node start time", node.startTime, "welcome received", m.Time.Format(timeFormat))
			continue
		}

		// Parse welcome message containing blockchain and node-wallet addresses
		var welcomeMessage welcomeMessage
		err = json.Unmarshal(m.Value, &welcomeMessage)
		if err != nil {
			logger.Error("Parsing welcome message failed", "error", err)
			return err
		}
		logger.Info("Successfully parsed welcome message")

		// Set node blockchain to received one
		node.blockchain = welcomeMessage.Bc
		// Set id array to received one
		node.idArray = welcomeMessage.NodesIn[:]
		// Update the map with received mapping
		for id, key := range node.idArray {
			node.nodeMap[key] = id
		}

		// Write blockchain to file
		err = node.WriteBlockchain()
		if err != nil {
			logger.Error("Could not write blockchain", "error", err)
			return err
		}
		logger.Info("Successfully wrote blockchain to file")

		// Turn blockchain to database
		err = node.MakeDB()
		if err != nil {
			logger.Error("Creating database from blockchain failed", "error", err)
			return err
		}
		logger.Info("Successfully created database")

		// Write database to file
		err = node.WriteDB()
		if err != nil {
			logger.Error("Could not write database to file", "error", err)
			return err
		}
		logger.Info("Successfully stored database in file")

		break
	}

	// Close welcome reader in parallel
	go r.Close()

	return nil
}

// Actions performed when the bootstrap node (id == 0) enters
func (node *nodeConfig) bootstrapNodeEnter() error {
	var err error

	// Creates genesis block
	genesis := GenesisBlock(node.publicKey, node.privateKey, node.nodes)
	// Appends genesis to blockchain
	node.blockchain = append(node.blockchain, genesis)
	logger.Info("Genesis block created")

	// Writes blockchain json to output file
	err = node.WriteBlockchain()
	if err != nil {
		logger.Error("Could not write genesis blockchain", "error", err)
		return err
	}
	logger.Info("Blockchain written successfully")

	// Makes database from blockchain
	err = node.MakeDB()
	if err != nil {
		logger.Error("Could not make DB from genesis blockchain", "error", err)
		return err
	}

	// Stores database to file
	err = node.WriteDB()
	if err != nil {
		logger.Error("Could not store bootstrap db to file", "error", err)
		return err
	}

	// Starts collect node procedure
	err = node.collectNodesInfo()
	if err != nil {
		logger.Error("Could not collect nodes info", "error", err)
	}
	return nil
}

// Initial point of node process
func (node *nodeConfig) Start() error {

	// Set remaining parameters necessary for function as node (not CLI)
	node.initialConfig()

	// Check whether node is bootsrap or ordinary
	if node.id == 0 {
		// Case bootstrap node
		logger.Info("Node is bootstrap node")
		// Perform bootstrap stuff
		err := node.bootstrapNodeEnter()
		if err != nil {
			logger.Error("Error in starting bootsrap node", "error", err)
			return err
		}
	} else {
		// Case ordinary node
		logger.Info("Node is ordinary")
		// Perform ordinary stuff
		err := node.ordinaryNodeEnter()
		if err != nil {
			logger.Error("Could not enter the cluster", "error", err)
			return err
		}
	}
	return node.startListeners()
}

// Performed by all nodes when the welcome message is sent by bootstrap
func (node *nodeConfig) startListeners() error {
	// Start kafka consumers
	// Consumer for "post-transaction" topic
	node.txConsumer = kafka.NewReader(kafka.ReaderConfig{
		Brokers:     []string{node.brokerURL},
		Topic:       "post-transaction",
		StartOffset: kafka.LastOffset,
		GroupID:     node.idString,
	})
	// On exit, close consumer
	defer node.txConsumer.Close()
	// Consumer for "post-block" topic
	node.blockConsumer = kafka.NewReader(kafka.ReaderConfig{
		Brokers:     []string{node.brokerURL},
		Topic:       "post-block",
		StartOffset: kafka.LastOffset,
		GroupID:     node.idString,
	})
	// On exit, close consumer
	defer node.blockConsumer.Close()

	logger.Info("Created post-transaction and post-block consumers")
	// Create's node new current block
	node.currentBlock = node.NewBlock()

	// Listens in parallel for remote procedure calls by CLI
	go node.startRPC()
	if node.inputPath != "" {
		go node.sendInputTransactions()
	}
	if node.useCLI {
		go node.startCLI()
	}

	// Listen for transactions posted to the kafka broker
	return node.transactionListener()

}
