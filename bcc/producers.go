package bcc

import (
	"context"
	"errors"
	"fmt"
	"github.com/segmentio/kafka-go"
	"os"
	"strconv"
)

var W *kafka.Writer = &kafka.Writer{
	Addr:  kafka.TCP("localhost:9092"),
	Topic: "my-topic",
}

func SendTransaction(w *kafka.Writer, tx Transaction) error {
	if w.Topic != "post-transaction" {
		return errors.New("writer does not post to correct topic")
	}
	jsonString, err := tx.JSONify()
	if err != nil {
		return err
	}
	err = w.WriteMessages(context.Background(), kafka.Message{
		Value: []byte(jsonString),
		Headers: []kafka.Header{{
			Key:   "SenderNode",
			Value: []byte(strconv.Itoa(NodeID)),
		}},
	})

	return err
}

func DeclareExistence(w *kafka.Writer) error {
	if w.Topic != "declare-self" {
		return errors.New("writer does not post to correct topic")
	}
	byteMessage := []byte(MyPublicKey)
	err := w.WriteMessages(context.Background(), kafka.Message{
		Value: byteMessage,
		Headers: []kafka.Header{{
			Key:   "SenderNode",
			Value: byteMessage,
		}},
	})
	return err
}

func BroadcastBlock(w *kafka.Writer, b Block) error {
	if w.Topic != "post-block" {
		return errors.New("writer does not post to correct topic")
	}
	blockJson, err := b.JSONify()
	if err != nil {
		return err
	}
	byteMessage := []byte(blockJson)
	nodeIdBytes := []byte(strconv.Itoa(NodeID))
	err = w.WriteMessages(context.Background(), kafka.Message{
		Value: byteMessage,
		Headers: []kafka.Header{{
			Key:   "SenderNode",
			Value: nodeIdBytes,
		}},
	})
	return err
}

func TransmitBlockChain(w *kafka.Writer) error {
	if w.Topic != "transmit-blockchain" {
		return errors.New("writer does not post to correct topic")
	}
	content, err := os.ReadFile(BLOCKCHAIN_PATH)
	if err != nil {
		fmt.Println(err)
	}
	nodeIdBytes := []byte(strconv.Itoa(NodeID))
	err = w.WriteMessages(context.Background(), kafka.Message{
		Value: content,
		Headers: []kafka.Header{{
			Key:   "SenderNode",
			Value: nodeIdBytes,
		}},
	})
    return err
}

