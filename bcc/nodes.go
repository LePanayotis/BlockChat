package bcc

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"strconv"
	"time"

	"github.com/segmentio/kafka-go"
)

type NodeConfig struct {
	feePercentage       float64 `default:"0.03"`
	costPerChar         int     `default:"1"`
	blockchainPath      string
	keyLength           int     `default:"512"`
	timeFormat          string  `default:"02-01-2006 15:04:05.000"`
	initialBCC          float64 `default:"1000"`
	capacity            int     `default:"3"`
	dbPath              string
	nodes               int    `default:"1"`
	socket              string `default:":1500"`
	protocol            string `default:"tcp4"`
	genesisHash         string `default:"1"`
	brokerURL           string `default:"localhost:9093"`
	id                  int    `default:"0"`
	myPublicKey         string
	myPrivateKey        string
	currentBlock        Block  `default:"0"`
	transactionsInBlock int    `default:"0"`
	blockIndex          int    `default:"0"`
	lastHash            string `default:"1"`
	nodeMap             map[string]int
	nodeIdArray         []string
	idString            string `default:"0"`
	nodeStartTime       string
	myHeaders           []kafka.Header
	ready               bool `default:"false"`
	detached            bool `default:"false"`
	currentPublicKey    string
	currentPrivateKey   string
	myBlockchain        Blockchain
	myDB                DBmap
	writer              *kafka.Writer
	txConsumer          *kafka.Reader
	blockConsumer       *kafka.Reader
}

var node *NodeConfig = &NodeConfig{
	keyLength:  512,
	timeFormat: "02-01-2006 15:04:05.000",
	protocol:   "tcp4",
}

var myNonce uint = 1
var CLI = false

func newNodeEnter() error {

	var w *kafka.Writer = &kafka.Writer{
		Addr: kafka.TCP(node.brokerURL),
	}
	var r *kafka.Reader = kafka.NewReader(kafka.ReaderConfig{
		Brokers:     []string{node.brokerURL},
		Topic:       "welcome",
		StartOffset: kafka.LastOffset,
		GroupID:     node.idString,
	})

	declareExistence(w)

	for {
		m, err := r.ReadMessage(context.Background())
		if err != nil {
			fmt.Println(err)
			continue
		}
		t, _ := time.Parse(node.timeFormat, node.nodeStartTime)
		if m.Time.Before(t) {
			log.Println("Message before node start time")
			continue
		}
		var welcomeMessage welcomeMessage
		err = json.Unmarshal(m.Value, &welcomeMessage)
		if err != nil {
			log.Println(err)
			return err
		}

		node.myBlockchain = welcomeMessage.Bc
		node.nodeIdArray = welcomeMessage.NodesIn[:]
		node.myBlockchain.WriteBlockchain()
		node.myDB, err = node.myBlockchain.MakeDB()
		if err != nil {
			log.Println(err)
			return err
		}

		err = node.myDB.WriteDB()
		if err != nil {
			log.Println(err)
			return err
		}

		break
	}
	go func() {
		r.Close()
		w.Close()
	}()
	return nil
}

func startEnv() {
	node.nodeMap = make(map[string]int)
	node.nodeIdArray = make([]string, node.nodes)
	node.generateKeysUpdate()
	node.nodeStartTime = time.Now().UTC().Format(node.timeFormat)
	node.lastHash = node.genesisHash
	node.idString = strconv.Itoa(node.id)

	node.myHeaders = []kafka.Header{
		{
			Key:   "NodeId",
			Value: []byte(node.idString),
		},
		{
			Key:   "NodeWallet",
			Value: []byte(node.myPublicKey),
		},
	}
}

func blockListener() error {
	_lasthash := node.myBlockchain[len(node.myBlockchain)-1].Current_hash
	_index := node.myBlockchain[len(node.myBlockchain)-1].Index + 1
	node.currentBlock = NewBlock(_index, _lasthash)
	for {
		block, validator, err := getNewBlock(node.blockConsumer)
		if err != nil {
			continue
		}
		if node.currentBlock.Current_hash == block.Current_hash && validator == node.currentBlock.Validator {
			node.myBlockchain = append(node.myBlockchain, block)
			node.currentBlock = NewBlock(block.Index+1, block.Current_hash)
			fmt.Printf("Block accepted\n>")
			fmt.Printf("Validator: %d\n>", node.nodeMap[validator])
			err = node.myDB.addBlockUndoStake(&block)
			if err != nil {
				fmt.Print(err)
				return err
			}
			node.myDB.WriteDB()
			node.myBlockchain.WriteBlockchain()
			node.transactionsInBlock = 0
		}
	}
}

func transactionListener() error {
	node.transactionsInBlock = 0
	for {
		tx, err := getNewTransaction(node.txConsumer)
		if err != nil {
			continue
		}
		fmt.Printf("Transaction received\n>")
		if node.myDB.isTransactionPossible(&tx) {
			if node.transactionsInBlock < node.capacity {
				node.myDB.addTransaction(&tx)
				node.currentBlock.AddTransaction(&tx)
				node.transactionsInBlock++
				if node.transactionsInBlock == node.capacity {
					node.currentBlock.SetValidator()
					node.currentBlock.CalcHash()
					if node.currentBlock.Validator == node.myPublicKey {
						node.currentBlock.Timestamp = time.Now().UTC().Format(node.timeFormat)
						broadcastBlock(node.writer, node.currentBlock)
						fmt.Printf("I broadcasted the block\n>")
					}
				}
			} else {
				fmt.Printf("Capacity Reached\n>")
			}

		} else {
			fmt.Printf("Transaction not possible\n>")
		}

	}
}

func StartNode() error {

	startEnv()

	var err error
	var myBlockchain *Blockchain = &node.myBlockchain

	if node.id == 0 {
		genesis := GenesisBlock(node.myPublicKey, node.myPrivateKey)
		*myBlockchain = append(*myBlockchain, genesis)
		myBlockchain.WriteBlockchain()

		if err != nil {
			return err
		}
		err = collectNodesInfo()

	} else {
		err = newNodeEnter()
	}

	//Start node kafka consumers and writers
	node.writer = &kafka.Writer{
		Addr: kafka.TCP(node.brokerURL),
	}
	node.txConsumer = kafka.NewReader(kafka.ReaderConfig{
		Brokers:     []string{node.brokerURL},
		Topic:       "post-transaction",
		StartOffset: kafka.LastOffset,
		GroupID:     node.idString,
	})
	node.blockConsumer = kafka.NewReader(kafka.ReaderConfig{
		Brokers:     []string{node.brokerURL},
		Topic:       "post-block",
		StartOffset: kafka.LastOffset,
		GroupID:     node.idString,
	})

	go blockListener()
	go start_rpc()
	go transactionListener()
	return nil
}
