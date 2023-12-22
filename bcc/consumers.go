package bcc

import (
	"context"
	"errors"
	"github.com/segmentio/kafka-go"
)

var TxConsumer *kafka.Reader
var BlockConsumer *kafka.Reader

func GetNewBlock(r *kafka.Reader) (Block, string, error) {
	m, err := r.ReadMessage(context.Background())
	if err != nil {
		return Block{}, "", err
	}
	B, err := ParseBlockJSON(string(m.Value))
	if err != nil {
		return Block{}, "", err
	}
	stringId := string(m.Headers[1].Value)
	return B, stringId, nil
}

func GetNewTransaction(r *kafka.Reader) (Transaction, error) {
	tx := Transaction{}
	m , err := r.ReadMessage(context.Background())
	if err != nil {
		return tx, err
	}
	tx, err = ParseTransactionJSON(string(m.Value))
	if err != nil {
		return tx, err
	}
	if !tx.Verify() {
		return tx, errors.New("transaction not verified")
	}
	return tx,nil
}

