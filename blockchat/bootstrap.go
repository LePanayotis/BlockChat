package blockchat

import (
	"context"
	"github.com/segmentio/kafka-go"
	"strconv"
)

func collectNodesInfo() error {

	r := kafka.NewReader(kafka.ReaderConfig{
		Brokers:     []string{node.brokerURL},
		Topic:       "enter",
		StartOffset: kafka.LastOffset,
		GroupID:     node.idString,
	})
	
	w := &kafka.Writer{
		Addr: kafka.TCP(node.brokerURL),
	}

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
			i++
		}
	}
	broadcastWelcome(w)
	go func() {
		w.Close()
		r.Close()
	}()
	return nil
}
