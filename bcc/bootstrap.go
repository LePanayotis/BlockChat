package bcc

import(
	"strconv"
	"log"
	"github.com/segmentio/kafka-go"
	"context"
)

func collectNodesInfo() error {
	R := kafka.NewReader(kafka.ReaderConfig{
		Brokers:     []string{BROKER_URL},
		Topic:       "enter",
		StartOffset: kafka.LastOffset,
		GroupID:     NodeIDString,
	})
	var W *kafka.Writer = &kafka.Writer{
		Addr:  kafka.TCP(BROKER_URL),
	}
	i := 1
	for i < NODES {
		m, err := R.ReadMessage(context.Background())
		if err != nil {
			continue
		}
		strPublicKey := string(m.Headers[1].Value)
		intNodeId, _ := strconv.Atoi(string(m.Headers[0].Value))
		_, b := NodeMap[strPublicKey]
		if !b {
			NodeMap[strPublicKey] = intNodeId
			NodeIDArray[intNodeId] = strPublicKey
			i++
		}
	}
	log.Println("All nodes in")
	BroadcastWelcome(W)
	go func() {
		W.Close()
		R.Close()
	}()
	return nil
}