package bcc

import (
	"context"
	"fmt"
	"github.com/segmentio/kafka-go"
	"os"
	"strconv"
)

// var W *kafka.Writer = &kafka.Writer{
// 	Addr:  kafka.TCP("localhost:9092"),
// 	Topic: "my-topic",
// }

func SendTransaction(w *kafka.Writer, tx Transaction) error {
	jsonString, err := tx.JSONify()
	if err != nil {
		return err
	}
	err = w.WriteMessages(context.Background(), kafka.Message{
		Topic: "post-transaction",
		Value: []byte(jsonString),
		Headers: []kafka.Header{{
			Key:   "SenderNode",
			Value: []byte(strconv.Itoa(NodeID)),
		}},
	})

	return err
}

func DeclareExistence(w *kafka.Writer) error {
	byteMessage := []byte(MyPublicKey)
	err := w.WriteMessages(context.Background(), kafka.Message{
		Topic: "declare-self",
		Value: byteMessage,
		Headers: []kafka.Header{{
			Key:   "SenderNode",
			Value: byteMessage,
		}},
	})
	return err
}

func BroadcastBlock(w *kafka.Writer, b Block) error {
	blockJson, err := b.JSONify()
	if err != nil {
		return err
	}
	byteMessage := []byte(blockJson)
	nodeIdBytes := []byte(strconv.Itoa(NodeID))
	err = w.WriteMessages(context.Background(), kafka.Message{
		Topic: "post-block",
		Value: byteMessage,
		Headers: []kafka.Header{{
			Key:   "SenderNode",
			Value: nodeIdBytes,
		}},
	})
	return err
}

func TransmitBlockChain(w *kafka.Writer) error {
	content, err := os.ReadFile(BLOCKCHAIN_PATH)
	if err != nil {
		fmt.Println(err)
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