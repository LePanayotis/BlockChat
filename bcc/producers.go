package bcc

import (
	"context"
	"encoding/json"
	"os"
	"strconv"
	"github.com/segmentio/kafka-go"
)

var Writer *kafka.Writer

type WelcomeMessage struct{
	Bc Blockchain `json:"blockchain"`
	NodesIn []string `json:"nodesin"`	
}

func SendTransaction(w *kafka.Writer, tx Transaction) error {
	jsonString, err := tx.JSONify()
	if err != nil {
		return err
	}
	err = w.WriteMessages(context.Background(), kafka.Message{
		Topic: "post-transaction",
		Value: []byte(jsonString),
		Headers: MyHeaders,
	})
	return err
}

func declareExistence(w *kafka.Writer) error {
	err := w.WriteMessages(context.Background(), kafka.Message{
		Topic: "enter",
		Headers: MyHeaders,
	})
	return err
}

func BroadcastBlock(w *kafka.Writer, b Block) error {
	blockJson, err := b.JSONify()
	if err != nil {
		return err
	}
	byteMessage := []byte(blockJson)
	err = w.WriteMessages(context.Background(), kafka.Message{
		Topic: "post-block",
		Value: byteMessage,
		Headers: MyHeaders,
	})
	return err
}

//Broadcasts the initial blockchain and the nodes array of the cluster
func BroadcastWelcome(W *kafka.Writer) error {
	BlockIndex++
	_prev_blockchain := MyBlockchain[len(MyBlockchain)-1].Current_hash
	block := NewBlock(BlockIndex,_prev_blockchain)

	for i := 1; i < len(NodeIDArray); i++ {
		tx := NewTransferTransaction(MyPublicKey,NodeIDArray[i],INITIAL_BCC,myNonce,MyPrivateKey)
		block.AddTransaction(&tx)
		myNonce++
	}
	block.Validator = block.CalcValidator()
	block.CalcHash()

	MyBlockchain.AddBlock(&block)

	msg := WelcomeMessage{
		Bc:      MyBlockchain,
		NodesIn: NodeIDArray[:],
	}
	payload, _ := json.Marshal(msg)
	MyBlockchain.WriteBlockchain()
	ValidDB, _ = MyBlockchain.MakeDB()
	ValidDB.WriteDB()
	W.WriteMessages(context.Background(), kafka.Message{
		Topic: "welcome",
		Headers: MyHeaders,
		Value: payload,
	})
	return nil
}


func TransmitBlockChain(w *kafka.Writer) error {
	content, err := os.ReadFile(BLOCKCHAIN_PATH)
	if err != nil {
		return err
	}
	nodeIdBytes := []byte(strconv.Itoa(NodeID))
	err = w.WriteMessages(context.Background(), kafka.Message{
		Topic: "blockchain",
		Value: content,
		Headers: []kafka.Header{{
			Key:   "SenderNode",
			Value: nodeIdBytes,
		}},
	})
    return err
}