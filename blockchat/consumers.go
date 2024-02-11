package blockchat

import (
	"context"
	"strconv"
	"github.com/segmentio/kafka-go"
)
func getNewBlock(r *kafka.Reader) (Block, string, error) {
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

func getNewTransaction(r *kafka.Reader) (Transaction, error) {
	tx := Transaction{}
	m, err := r.ReadMessage(context.Background())
	if err != nil {
		return tx, err
	}
	tx, err = ParseTransactionJSON(string(m.Value))
	if err != nil {
		return tx, err
	}
	return tx, nil
}




func (node * nodeConfig) collectNodesInfo() error {

	if node.nodes == 1 {
		logger.Info("I'm an alone node :'(")
		return nil
	}

	r := kafka.NewReader(kafka.ReaderConfig{
		Brokers:     []string{node.brokerURL},
		Topic:       "enter",
		StartOffset: kafka.LastOffset,
		GroupID:     node.idString,
	})
	
	w := &kafka.Writer{
		Addr: kafka.TCP(node.brokerURL),
	}
	logger.Info("Kafka producer and enter consumer connected")

	i := 1
	for i < node.nodes {
		m, err := r.ReadMessage(context.Background())
		if err != nil {
			continue
		}

		strPublicKey := string(m.Headers[1].Value)
		intNodeId, _ := strconv.Atoi(string(m.Headers[0].Value))
		_, b := node.nodeMap[strPublicKey]
		if !b {
			node.nodeMap[strPublicKey] = intNodeId
			node.nodeIdArray[intNodeId] = strPublicKey
			logger.Info("Node in","node",intNodeId)
			i++
		}
	}
	node.broadcastWelcome(w)
	go func() {
		w.Close()
		r.Close()
	}()
	return nil
}
