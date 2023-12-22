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
var Current_block Block
var Transactions_in_block int
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
		ValidDB, _ = MyBlockchain.MakeDB()
		TempDB = ValidDB
		ValidDB.WriteDB()
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
	Writer =  &kafka.Writer{
		Addr: kafka.TCP(BROKER_URL),
	}
	TxConsumer = kafka.NewReader(kafka.ReaderConfig{
		Brokers:     []string{BROKER_URL},
		Topic:       "post-transaction",
		StartOffset: kafka.LastOffset,
		GroupID:     NodeIDString,
	})
	BlockConsumer = kafka.NewReader(kafka.ReaderConfig{
		Brokers:     []string{BROKER_URL},
		Topic:       "post-block",
		StartOffset: kafka.LastOffset,
		GroupID:     NodeIDString,
	})
	fmt.Println("Consumers and producers connected")

	go func(){
		_lasthash := MyBlockchain[len(MyBlockchain)-1].Current_hash
		_index := MyBlockchain[len(MyBlockchain)-1].Index + 1
		Current_block = NewBlock(_index, _lasthash)
		for {
			block, validator, err := GetNewBlock(BlockConsumer)
			if err != nil {
				continue
			}
			fmt.Printf("\n%s \n = \n %s\n>", validator, Current_block.Validator)
			fmt.Printf("Block received\n>")
			if Current_block.Current_hash == block.Current_hash &&  validator == Current_block.Validator {
				MyBlockchain = append(MyBlockchain, block)
				Current_block = NewBlock(block.Index+1, block.Current_hash)
				fmt.Printf("\nBlock accepted\n>")
				ValidDB, _ = MyBlockchain.MakeDB()
				TempDB = ValidDB
				ValidDB.WriteDB()
				MyBlockchain.WriteBlockchain()
				Transactions_in_block = 0
			}
		}


	}()

	go func(){
		Transactions_in_block = 0
		fmt.Println("new thread started")
		for {
			tx, err := GetNewTransaction(TxConsumer)
			if err != nil {
				continue
			}
			if TempDB.IsTransactionPossible(&tx) {
				if(Transactions_in_block < CAPACITY){
					TempDB.addTransaction(&tx)
					Current_block.AddTransaction(tx)
					Transactions_in_block++
					if Transactions_in_block == CAPACITY {
						Current_block.SetValidator()
						Current_block.CalcHash()
						if Current_block.Validator == MyPublicKey{
							BroadcastBlock(Writer, Current_block)
							fmt.Printf("\nBlock sent\n>")
						}
					}
				} else {
					fmt.Printf("\nCapacity Reached\n>")
				}
				
			}
			fmt.Printf("\nTransaction received\n>")
		}
	}()
	StartCLI()
	defer Writer.Close()
	defer TxConsumer.Close()
	defer BlockConsumer.Close()
	return nil
}
