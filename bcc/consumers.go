package bcc

import (
	"context"
	"errors"
	"fmt"
	"strconv"

	"github.com/segmentio/kafka-go"
)

func GetBlockMessage(r *kafka.Reader) (Block, int, error) {
	m, err := r.ReadMessage(context.Background())
	if err != nil {
		return Block{}, -1, err
	}
	B, err := ParseBlockJSON(string(m.Value))
	if err != nil {
		return Block{}, -1, err
	}
	stringId := string(m.Headers[0].Value)
	fmt.Println(m.Headers[0].Key)
	node, _ := strconv.Atoi(stringId)
	return B, node, nil
}

func GetNewTransaction(r *kafka.Reader) error {
	m , err := r.ReadMessage(context.Background())
	if err != nil {
		return err
	}
	tx, err := ParseTransactionJSON(string(m.Value))
	if err != nil {
		return err
	}
	if !tx.Verify() {
		return errors.New("transaction not verified")
	}
	return nil
}

