package blockchat

import (
	"context"
	"encoding/json"
	"github.com/segmentio/kafka-go"
)

// Struct to represent welcome message sent by the bootstrap node
type welcomeMessage struct {
	Bc      Blockchain `json:"blockchain"`
	NodesIn []string   `json:"nodesin"`
}


// Method used to send a transaction to kafka topic
// Does not perform any checks on the transaction
func (node *nodeConfig) sendTransaction(tx *Transaction) error {

	// Produces json string representation of the transaction
	jsonString, err := tx.JSONify()
	if err != nil {
		return err
	}
	// Sends serialised json representation to kafka topic
	err = node.writer.WriteMessages(context.Background(), kafka.Message{
		Topic:   "post-transaction",
		Value:   []byte(jsonString),
		Headers: node.headers,
	})
	return err
}


// Method implements ordinary node declaring its existence to the boootstrap node 
// by posting message to kafka.
// Sends its headers: node public kay and node id
func (node *nodeConfig) declareExistence() error {
	err := node.writer.WriteMessages(context.Background(), kafka.Message{
		Topic:   "enter",
		Headers:  node.headers,
	})
	return err
}


// Broadcasts block to the cluster via kafka broker
// Used when the node is the validator of the block
func (node *nodeConfig) broadcastBlock(b *Block) error {

	// Produces json representation of block
	blockJson, err := b.JSONify()
	if err != nil {
		return err
	}

	// Serialises json string to byte array
	byteMessage := []byte(blockJson)
	// Forwards block to the  cluster
	err = node.writer.WriteMessages(context.Background(), kafka.Message{
		Topic:   "post-block",
		Value:   byteMessage,
		Headers:  node.headers,
	})
	return err
}

// Broadcasts the initial blockchain and the nodes array of the cluster
func (node *nodeConfig) broadcastWelcome() error {

	// Creates new block to share genesis block amount with other nodes
	block := node.NewBlock()
	// to remove: node.blockIndex++

	for i := 1; i < len(node.idArray); i++ {
		// Gets node's wallet address
		receiver := node.idArray[i]
		// Creates transaction granting node initialBCC
		// outbound nonce is increased in node.NewTransferTransaction
		tx := node.NewTransferTransaction(receiver,initialBCC)

		// Increases nonce in database
		node.increaseNonce()
		// Adds transaction to the curret block
		block.AddTransaction(&tx)
	}
	
	// Calculates validator and hash
	block.Validator = block.CalcValidator()
	block.CalcHash()

	// Appends block to blockchain
	node.blockchain.AddBlock(&block)

	// Creates message struct to broadcast
	// blockchain to the other nodes
	msg := welcomeMessage{
		Bc:      node.blockchain,
		NodesIn: node.idArray[:],
	}

	// Turns message into json
	payload, err := json.Marshal(msg)
	if err != nil {
		logger.Error("Failed to marshal welcome message", "error",err)
		return err
	}

	// Writes blockchain to file
	err = node.WriteBlockchain()
	if err != nil {
		logger.Error("Failed to store blockchain to file", "error",err)
		return err
	}

	// Creates in-memory database as indicated by previous blockchain
	err = node.MakeDB()
	if err != nil {
		logger.Error("Failed to produce memory database", "error",err)
		return err
	}

	// Writes the database to file
	err = node.WriteDB()
	if err != nil {
		logger.Error("Failed to store database to file", "error",err)
		return err
	}

	// Broadcasts the welcome message to the cluster via kafka
	err = node.writer.WriteMessages(context.Background(), kafka.Message{
		Topic:   "welcome",
		Headers:  node.headers,
		Value:   payload,
	})
	if err != nil {
		logger.Error("Failed to send welcome message", "error",err)
		return err
	}

	// Successfully sent welcome message
	logger.Info("Sent welcome message")
	return nil
}