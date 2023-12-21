package bcc

import (
	"context"
	"encoding/json"
	"fmt"
	"github.com/segmentio/kafka-go"
	"log"
	"strconv"
	"time"
)

var MyPublicKey, MyPrivateKey string
var MyBlockchain Blockchain
var NodeStartTime time.Time
var ValidDB, TempDB DBmap
var NodeID int = 0
var NodeIDString string = strconv.Itoa(NodeID)
var BlockIndex int = 0
var Last_hash string = GENESIS_HASH
var NodeMap map[string]int = make(map[string]int)
var NodeIDArray []string
var myNonce uint = 0
var MyHeaders []kafka.Header

func newNodeEnter() error {
	var W *kafka.Writer = &kafka.Writer{
		Addr:  kafka.TCP(BROKER_URL),
	}
	var R *kafka.Reader = kafka.NewReader(kafka.ReaderConfig{
		Brokers:     []string{BROKER_URL},
		Topic:       "welcome",
		StartOffset: kafka.LastOffset,
		GroupID:     NodeIDString,
	})
	
	DeclareExistence(W)

	log.Println("My stringId is", NodeIDString)

	for {
		m, err := R.ReadMessage(context.Background())
		if err != nil {
			fmt.Println(err)
			continue
		}
		if m.Time.Before(NodeStartTime) {
			log.Println("Message before NodeStartTime")
			continue
		}
		var welcomeMessage WelcomeMessage
		err = json.Unmarshal(m.Value, &welcomeMessage)
		if err != nil {
			log.Println(err)
		}
		MyBlockchain = welcomeMessage.Bc
		NodeIDArray = welcomeMessage.NodesIn[:]
		MyBlockchain.WriteBlockchain()
		break
	}
	go func() {
		R.Close()
		W.Close()
	}()
	return nil
}

func StartNode() error {

	NodeStartTime = time.Now()

	StartEnv()

	MyPublicKey, MyPrivateKey = GenerateKeysUpdate()

	MyBlockchain = Blockchain{}

	var err error
	if NodeID == 0 {
		genesis := GenesisBlock(MyPublicKey, MyPrivateKey)
		MyBlockchain = append(MyBlockchain, genesis)
		MyBlockchain.WriteBlockchain()
		//ValidDB, err = LoadDB()
		if err != nil {
			return err
		}
		TempDB = ValidDB
		err = collectNodesInfo()

	} else {
		err = newNodeEnter()
	}
	return err
}
