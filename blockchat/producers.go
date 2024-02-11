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

func sendTransaction(w *kafka.Writer, tx Transaction) error {

	//TODO: check if transaction can be performed

	jsonString, err := tx.JSONify()
	if err != nil {
		return err
	}
	err = w.WriteMessages(context.Background(), kafka.Message{
		Topic:   "post-transaction",
		Value:   []byte(jsonString),
		Headers: node.myHeaders,
	})
	return err
}

func declareExistence(w *kafka.Writer) error {
	err := w.WriteMessages(context.Background(), kafka.Message{
		Topic:   "enter",
		Headers:  node.myHeaders,
	})
	return err
}

func broadcastBlock(w *kafka.Writer, b Block) error {
	blockJson, err := b.JSONify()
	if err != nil {
		return err
	}
	byteMessage := []byte(blockJson)
	err = w.WriteMessages(context.Background(), kafka.Message{
		Topic:   "post-block",
		Value:   byteMessage,
		Headers:  node.myHeaders,
	})
	return err
}

// Broadcasts the initial blockchain and the nodes array of the cluster
func (node *nodeConfig) broadcastWelcome(W *kafka.Writer) error {

	node.blockIndex++
	_prev_blockchain := node.myBlockchain[len(node.myBlockchain)-1].Current_hash
	block := NewBlock(node.blockIndex, _prev_blockchain)

	for i := 1; i < len(node.nodeIdArray); i++ {

		node.outboundNonce = node.myDB.increaseNonce(node.myPublicKey)
		tx := NewTransferTransaction(node.myPublicKey, node.nodeIdArray[i], node.initialBCC, node.outboundNonce, node.myPrivateKey)
		block.AddTransaction(&tx)
	}
	
	block.Validator = block.CalcValidator()
	block.CalcHash()

	node.myBlockchain.AddBlock(&block)

	msg := welcomeMessage{
		Bc:      node.myBlockchain,
		NodesIn: node.nodeIdArray[:],
	}
	payload, _ := json.Marshal(msg)
	node.myBlockchain.WriteBlockchain()
	node.myDB, _ = node.myBlockchain.MakeDB()
	node.myDB.WriteDB()

	W.WriteMessages(context.Background(), kafka.Message{
		Topic:   "welcome",
		Headers:  node.myHeaders,
		Value:   payload,
	})
	logger.Info("Sent welcome message")
	return nil
}

// func transmitBlockChain(w *kafka.Writer) error {
// 	content, err := os.ReadFile(node.blockchainPath)
// 	if err != nil {
// 		return err
// 	}
// 	nodeIdBytes := []byte(node.idString)
// 	err = w.WriteMessages(context.Background(), kafka.Message{
// 		Topic: "blockchain",
// 		Value: content,
// 		Headers: []kafka.Header{{
// 			Key:   "SenderNode",
// 			Value: nodeIdBytes,
// 		}},
// 	})
// 	return err
// }
