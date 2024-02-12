package blockchat

import (
	"context"
	"encoding/json"
	"time"
	"github.com/segmentio/kafka-go"
)

func (node *nodeConfig) ordinaryNodeEnter() error {


	var r *kafka.Reader = kafka.NewReader(kafka.ReaderConfig{
		Brokers:     []string{node.brokerURL},
		Topic:       "welcome",
		StartOffset: kafka.LastOffset,
		GroupID:     node.idString,
	})

	logger.Info("Kafka producer and welcome consumer connected")

	err := node.declareExistence()
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

		t, _ := time.Parse(timeFormat, node.startTime)
		if m.Time.Before(t) {
			logger.Error("Received messages created before","node start time",node.startTime, "welcome received",m.Time.Format(timeFormat))
			continue
		}

		var welcomeMessage welcomeMessage
		err = json.Unmarshal(m.Value, &welcomeMessage)
		if err != nil {
			logger.Error("Parsing welcome message failed", err)
			return err
		}
		logger.Info("Successfully parsed welcome message")

		node.blockchain = welcomeMessage.Bc
		node.idArray = welcomeMessage.NodesIn[:]
		for id, key := range node.idArray {
			node.nodeMap[key] = id
		}

		node.WriteBlockchain()
		if err != nil {
			logger.Error("Could not write blockchain", err)
			return err
		}
		logger.Info("Successfully wrote blockchain to file")

		err = node.MakeDB()
		if err != nil {
			logger.Error("Creating database from blockchain failed", err)
			return err
		}
		logger.Info("Successfully created database")

		err = node.WriteDB()
		if err != nil {
			logger.Error("Could not write database to file", err)
			return err
		}
		logger.Info("Successfully stored database in file")

		break
	}

	go func() {
		r.Close()
	}()
	return nil
}

func (node *nodeConfig) bootstrapNodeEnter() error {
	var err error
	var myBlockchain *Blockchain = &node.blockchain
		//Creates genesis block
		genesis := node.GenesisBlock()
		logger.Info("Genesis block created")

		//Appends genesis to blockchain
		*myBlockchain = append(*myBlockchain, genesis)

		//Writes blockchain json to output file
		err = node.WriteBlockchain()
		if err != nil {
			logger.Error("Could not write genesis blockchain", err)
			return err
		}
		logger.Info("Blockchain written successfully")

		err = node.MakeDB()
		if err != nil {
			logger.Error("Could not make DB from genesis blockchain", err)
			return err
		}
		err = node.WriteDB()
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

func (node *nodeConfig) blockListener() error {

	block, validator, err := node.getNewBlock()
	if err != nil {
		logger.Warn("Block listener exiting")
		return err
	}
	logger.Info("Received new block")

	if node.currentBlock.CurrentHash == block.CurrentHash {

		node.blockchain = append(node.blockchain, block)
		node.currentBlock = node.NewBlock()

		logger.Info("New block accepted", "validator node:", validator)

		err = node.addBlockUndoStake(&block)
		if err != nil {
			logger.Error("Failed adding block to database", err)
			return err
		}

		err = node.WriteDB()
		if err != nil {
			logger.Error("Failed writing database", err)
			return err
		}
		err = node.WriteBlockchain()
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

func (node *nodeConfig) transactionListener() error {

	for {
		tx, err := node.getNewTransaction()
		if err != nil {
			logger.Warn("Transaction listener exiting")
			return err
		}
		
		if !tx.Verify() {
			logger.Warn("Transaction not verified")
			continue
		}
		logger.Info("New transaction received")
		if node.isTransactionPossible(&tx) {

			if len(node.currentBlock.Transactions) < node.capacity {

				_, err = node.addTransactionToDB(&tx)
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

					if node.currentBlock.Validator == node.id {
						logger.Info("The node is broadcaster")
						node.currentBlock.Timestamp = time.Now().UTC().Format(timeFormat)
						err = node.broadcastBlock(&node.currentBlock)
						if err != nil {
							logger.Error("Failed to broadcast new block", err)
							return err
						}
						logger.Info("Block broadcasted by me")
					}
					err = node.blockListener()
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
			if tx.SenderAddress == node.publicKey {
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
		err := node.bootstrapNodeEnter()
		if err != nil {
			logger.Error("Error in starting bootsrap node", err)
			return err
		}

	} else {
		logger.Info("Node is ordinary")
		err := node.ordinaryNodeEnter()
		if err != nil {
			logger.Error("Could not enter the cluster", err)
			return err
		}
	}
	//Start node kafka consumers and writers
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
	
	node.currentBlock = node.NewBlock()
	//node.blockIndex++
	go node.startRPC()
	node.transactionListener()
	return nil
}
