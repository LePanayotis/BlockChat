package bcc

import (
	"context"
	"encoding/json"
	"fmt"
	"github.com/segmentio/kafka-go"
	"log"
	"time"
)

//Node's public and private RSA keys for transactions
var MyPublicKey, MyPrivateKey string

//Node's blockchain struct
var MyBlockchain Blockchain

//Block object currently in creation
//
var Current_block Block
var Transactions_in_block int

var NodeStartTime time.Time
var ValidDB DBmap
var NodeID int
var NodeIDString string
var BlockIndex int = 0
var Last_hash string = GENESIS_HASH
var NodeMap map[string]int = make(map[string]int)
var NodeIDArray []string
var myNonce uint = 1

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
	
	declareExistence(W)

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
	go func(){
		_lasthash := MyBlockchain[len(MyBlockchain)-1].Current_hash
		_index := MyBlockchain[len(MyBlockchain)-1].Index + 1
		Current_block = NewBlock(_index, _lasthash)
		for {
			block, validator, err := GetNewBlock(BlockConsumer)
			if err != nil {
				continue
			}
			if Current_block.Current_hash == block.Current_hash &&  validator == Current_block.Validator {
				MyBlockchain = append(MyBlockchain, block)
				Current_block = NewBlock(block.Index+1, block.Current_hash)
				fmt.Printf("Block accepted\n>")
				fmt.Printf("Validator: %d\n>", NodeMap[validator])
				err = ValidDB.AddBlockUndoStake(&block)
				if err != nil {
					fmt.Print(err)
				}
				ValidDB.WriteDB()
				MyBlockchain.WriteBlockchain()
				Transactions_in_block = 0
			}
		}


	}()

	go func(){
		Transactions_in_block = 0
		for {
			tx, err := GetNewTransaction(TxConsumer)
			if err != nil {
				continue
			}
			if ValidDB.IsTransactionPossible(&tx) {
				if(Transactions_in_block < CAPACITY){
					ValidDB.addTransaction(&tx)
					Current_block.AddTransaction(&tx)
					Transactions_in_block++
					if Transactions_in_block == CAPACITY {
						Current_block.SetValidator()
						Current_block.CalcHash()
						if Current_block.Validator == MyPublicKey{
							Current_block.Timestamp = time.Now().UTC().Format(TIME_FORMAT)
							BroadcastBlock(Writer, Current_block)
							fmt.Printf("I broadcasted the block\n>")
						}
					}
				} else {
					fmt.Printf("Capacity Reached\n>")
				}
				
			}
			fmt.Printf("Transaction received\n>")
		}
	}()
	StartCLI()
	return nil
}
