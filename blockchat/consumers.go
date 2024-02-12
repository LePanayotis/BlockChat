package blockchat

import (
	"context"
	"fmt"
	"strconv"
	"github.com/segmentio/kafka-go"
)

// Kafka consumer to receive newly posted blocks
func (node* nodeConfig) getNewBlock() (Block, int, error) {
	// Gets message from broker
	m, err := node.blockConsumer.ReadMessage(context.Background())
	if err != nil {
		return Block{}, -1, err
	}

	// Creates block from message
	B, err := ParseBlockJSON(string(m.Value))
	if err != nil {
		return Block{}, -1, err
	}

	// Checks sender is also the validator described in the block
	validator, err := strconv.Atoi(string(m.Headers[0].Value))
	if err != nil {
		return Block{}, -1, err
	}
	if validator != B.Validator {
		return Block{}, -1, fmt.Errorf("invalid validator: in block %v, received by %v", B.Validator, validator)
	}
	// Returns block, validator
	return B, validator, nil
}


// Kafka consumer for new transactions
func (node *nodeConfig) getNewTransaction() (Transaction, error) {
	
	tx := Transaction{}
	// Gets transaction message
	m, err := node.txConsumer.ReadMessage(context.Background())
	if err != nil {
		return tx, err
	}
	
	// Parses message and creates new transaction
	tx, err = ParseTransactionJSON(string(m.Value))
	if err != nil {
		return tx, err
	}

	// Returns the transaction
	return tx, nil
}

// Method implemented by the bootstrap node to collect information about
// other nodes in the cluster
func (node * nodeConfig) collectNodesInfo() error {

	// Checks number of nodes in the cluster
	// If alone, no need to collect information
	if node.nodes == 1 {
		logger.Info("I'm an alone node :'(")
		return nil
	}

	// Creates kafka reader instance on topic 'enter'
	r := kafka.NewReader(kafka.ReaderConfig{
		Brokers:     []string{node.brokerURL},
		Topic:       "enter",
		StartOffset: kafka.LastOffset,
		GroupID:     node.idString,
	})
	
	logger.Info("Kafka producer and enter consumer connected")

	// Loops until receives enter messages from all expected nodes
	i := 1
	for i < node.nodes {
		// Gets enter message
		m, err := r.ReadMessage(context.Background())
		if err != nil {
			logger.Error("Error while reading enter messages", "error",err)
			return err
		}

		// Gets node's public key from header
		strPublicKey := string(m.Headers[1].Value)
		// Gets node's id from header
		intNodeId, err := strconv.Atoi(string(m.Headers[0].Value))
		if err != nil {
			logger.Warn("Could not parse node's id from header","error",err,"header",m.Headers[0].Value)
			continue
		}

		// Checks whether the public key exists in the bootstrap's node map
		_, b := node.nodeMap[strPublicKey]
		if !b {
			// Updates node map with public key - node id
			node.nodeMap[strPublicKey] = intNodeId
			// Updates idArray with public key - node id
			node.idArray[intNodeId] = strPublicKey
			logger.Info("Node in","nodeId",intNodeId)
			
			// One more node successfully in!
			i++
		} else {
			logger.Warn("Node has already entered cluster","nodeId",intNodeId)
		}
	}

	// Initiates parallel execution
	// Closes kafka reader on 'enter' topic, won't be used anymore
	go func() {
		r.Close()
	}()

	//When all nodes have entered cluster bootstrap sends welcome message
	return 	node.broadcastWelcome()
}
