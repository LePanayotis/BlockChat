package blockchat

import (
	"context"
	"encoding/json"
	"time"

	"github.com/segmentio/kafka-go"
)

func ordinaryNodeEnter() error {

	var w *kafka.Writer = &kafka.Writer{
		Addr: kafka.TCP(node.brokerURL),
	}
	var r *kafka.Reader = kafka.NewReader(kafka.ReaderConfig{
		Brokers:     []string{node.brokerURL},
		Topic:       "welcome",
		StartOffset: kafka.LastOffset,
		GroupID:     node.idString,
	})

	logger.Info("Kafka producer and welcome consumer connected")

	err := declareExistence(w)
	if err != nil {
		logger.Error("Could not declare existence to other nodes", err)
		return err
	}

	logger.Info("Broadcasted existence to other nodes")

	for {

		m, err := r.ReadMessage(context.Background())
		if err != nil {
			logger.Error("Node could not get welcome message", err)
			continue
		}
		logger.Info("Received welcome message")

		t, _ := time.Parse(node.timeFormat, node.startTime)
		if m.Time.Before(t) {
			logger.Error("Received messages created before","node start time",node.startTime, "welcome received",m.Time.Format(node.timeFormat))
			continue
		}

		var welcomeMessage welcomeMessage
		err = json.Unmarshal(m.Value, &welcomeMessage)
		if err != nil {
			logger.Error("Parsing welcome message failed", err)
			return err
		}
		logger.Info("Successfully parsed welcome message")

		node.myBlockchain = welcomeMessage.Bc
		node.nodeIdArray = welcomeMessage.NodesIn[:]
		for id, key := range node.nodeIdArray {
			node.nodeMap[key] = id
		}

		node.myBlockchain.WriteBlockchain()
		if err != nil {
			logger.Error("Could not write blockchain", err)
			return err
		}
		logger.Info("Successfully wrote blockchain to file")

		node.myDB, err = node.myBlockchain.MakeDB()
		if err != nil {
			logger.Error("Creating database from blockchain failed", err)
			return err
		}
		logger.Info("Successfully created database")

		err = node.myDB.WriteDB()
		if err != nil {
			logger.Error("Could not write database to file", err)
			return err
		}
		logger.Info("Successfully stored database in file")

		break
	}

	go func() {
		r.Close()
		w.Close()
	}()
	return nil
}

func bootstrapNodeEnter() error {
	var err error
	var myBlockchain *Blockchain = &node.myBlockchain
		//Creates genesis block
		genesis := GenesisBlock(node.myPublicKey, node.myPrivateKey)
		logger.Info("Genesis block created")

		//Appends genesis to blockchain
		*myBlockchain = append(*myBlockchain, genesis)

		//Writes blockchain json to output file
		err = myBlockchain.WriteBlockchain()
		if err != nil {
			logger.Error("Could not write genesis blockchain", err)
			return err
		}
		logger.Info("Blockchain written successfully")

		node.myDB, err = node.myBlockchain.MakeDB()
		if err != nil {
			logger.Error("Could not make DB from genesis blockchain", err)
			return err
		}
		err = node.myDB.WriteDB()
		if err != nil {
			logger.Error("Could not store bootstrap db to file", err)
			return err
		}

		err = node.collectNodesInfo()
		if err != nil {
			logger.Error("Could not collect nodes info", err)
		}
		return nil
}

func blockListener() error {

	block, validator, err := getNewBlock(node.blockConsumer)
	if err != nil {
		logger.Warn("Block listener exiting")
		return err
	}
	logger.Info("Received new block")

	// node_temp := *node
	// logger.Info("Check this hash:", "hash", node_temp.currentBlock.Current_hash)
	logger.Warn("Check this","my_hash",node.currentBlock.Current_hash,"received_hash",block.Current_hash)

	if node.currentBlock.Current_hash == block.Current_hash /*&& validator == node.currentBlock.Validator*/ {

		node.myBlockchain = append(node.myBlockchain, block)
		node.currentBlock = NewBlock(block.Index+1, block.Current_hash)

		logger.Info("New block accepted", "validator node:", node.nodeMap[validator])

		err = node.myDB.addBlockUndoStake(&block)
		if err != nil {
			logger.Error("Failed adding block to database", err)
			return err
		}

		err = node.myDB.WriteDB()
		if err != nil {
			logger.Error("Failed writing database", err)
			return err
		}
		err = node.myBlockchain.WriteBlockchain()
		if err != nil {
			logger.Error("Failed writing blockchain", err)
			return err
		}

		logger.Info("Block add routine completed")

	} else {
		logger.Warn("Block rejected")
	}
	return nil
}

func transactionListener() error {

	for {
		tx, err := getNewTransaction(node.txConsumer)
		if err != nil {
			logger.Warn("Transaction listener exiting")
			return err
		}
		
		if !tx.Verify() {
			logger.Warn("Transaction not verified")
			continue
		}
		logger.Info("New transaction received")
		if node.myDB.isTransactionPossible(&tx) {

			if len(node.currentBlock.Transactions) < node.capacity {

				_, err = node.myDB.addTransaction(&tx)
				if err != nil {
					logger.Error("Failed adding transaction to database", err)
					return err
				}
				logger.Info("Transaction added to database")

				transactionsInBlock := node.currentBlock.AddTransaction(&tx)

				if transactionsInBlock == node.capacity {
					logger.Info("Block capacity reached")

					node.currentBlock.SetValidator()
					node.currentBlock.CalcHash()

					if node.currentBlock.Validator == node.myPublicKey {
						logger.Info("The node is broadcaster")
						node.currentBlock.Timestamp = time.Now().UTC().Format(node.timeFormat)
						err = broadcastBlock(node.writer, node.currentBlock)
						if err != nil {
							logger.Error("Failed to broadcast new block", err)
							return err
						}
						logger.Info("Block broadcasted by me")
					}
					err = blockListener()
					if err != nil {
						logger.Error("Experiment failed")
						return err
					}
				}
			} else {
				logger.Warn("Capacity Reached")
				return nil
			}

		} else {
			logger.Warn("Transaction rejected")
			if tx.Sender_address == node.myPublicKey {
				logger.Warn("My nonce is decreased by one")
				node.outboundNonce--

			}
		}

	}
}

func StartNode() error {
	startConfig()
	if node.id == 0 {
		//Case bootstrap node
		logger.Info("Node is bootstrap node")
		err := bootstrapNodeEnter()
		if err != nil {
			logger.Error("Error in starting bootsrap node", err)
			return err
		}

	} else {
		logger.Info("Node is ordinary")
		err := ordinaryNodeEnter()
		if err != nil {
			logger.Error("Could not enter the cluster", err)
			return err
		}
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

	logger.Info("Created post-transaction and post-block readers")
	
	_lasthash := node.myBlockchain[len(node.myBlockchain)-1].Current_hash

	_index := node.myBlockchain[len(node.myBlockchain)-1].Index + 1

	node.currentBlock = NewBlock(_index, _lasthash)

	go startRPC()
	transactionListener()
	return nil
}
