package blockchat

import (
	"context"
	"encoding/json"
	"github.com/segmentio/kafka-go"
	//"os"
)

type welcomeMessage struct {
	Bc      Blockchain `json:"blockchain"`
	NodesIn []string   `json:"nodesin"`
}

func (node *nodeConfig) sendTransaction(tx Transaction) error {

	//TODO: check if transaction can be performed

	jsonString, err := tx.JSONify()
	if err != nil {
		return err
	}
	err = node.writer.WriteMessages(context.Background(), kafka.Message{
		Topic:   "post-transaction",
		Value:   []byte(jsonString),
		Headers: node.headers,
	})
	return err
}

func (node *nodeConfig) declareExistence() error {
	err := node.writer.WriteMessages(context.Background(), kafka.Message{
		Topic:   "enter",
		Headers:  node.headers,
	})
	return err
}

func (node *nodeConfig) broadcastBlock(b Block) error {
	blockJson, err := b.JSONify()
	if err != nil {
		return err
	}
	byteMessage := []byte(blockJson)
	err = node.writer.WriteMessages(context.Background(), kafka.Message{
		Topic:   "post-block",
		Value:   byteMessage,
		Headers:  node.headers,
	})
	return err
}

// Broadcasts the initial blockchain and the nodes array of the cluster
func (node *nodeConfig) broadcastWelcome(W *kafka.Writer) error {


	block := node.NewBlock()
	node.blockIndex++
	
	for i := 1; i < len(node.idArray); i++ {

		node.outboundNonce = node.myDB.increaseNonce(node.id)
		tx := NewTransferTransaction(node.publicKey, node.idArray[i], node.initialBCC, node.outboundNonce, node.privateKey)
		block.AddTransaction(&tx)
	}
	
	block.Validator = block.CalcValidator()
	block.CalcHash()

	node.blockchain.AddBlock(&block)

	msg := welcomeMessage{
		Bc:      node.blockchain,
		NodesIn: node.idArray[:],
	}
	payload, _ := json.Marshal(msg)
	node.WriteBlockchain()
	
	node.MakeDB()
	node.WriteDB()

	W.WriteMessages(context.Background(), kafka.Message{
		Topic:   "welcome",
		Headers:  node.headers,
		Value:   payload,
	})
	logger.Info("Sent welcome message")
	return nil
}