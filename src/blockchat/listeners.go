package blockchat

import (
	"errors"
	"time"
)

// Method that performs logic of reveiving broadcasted transactions
// and executing necessary actions after it
func (node *nodeConfig) transactionListener() error {
	// Loops
	for {
		// Gets new transaction
		tx, err := node.getNewTransaction()
		if err != nil {
			logger.Error("Transaction listener exiting", "error",err)
			return err
		}
		
		node.logTransaction("Received new transaction",&tx)
		// Verifies transaction and goes to next iteration
		if !tx.Verify() {
			logger.Warn("Transaction not verified")
			continue
		}

		// Checks nodes database if transaction is possible
		err = node.isTransactionPossibleSetNonce(&tx)
		if err == nil {
			// If block capacity not yet reached
			
			if  len(node.currentBlock.Transactions) < node.capacity+1 {
				// Adds new transaction to database
				node.addTransactionToDB(&tx)
				logger.Info("Transaction added to database")
				// Adds transaction to current block
				node.currentBlock.AppendTransaction(&tx)

				// If block's capacity reached
				if len(node.currentBlock.Transactions) == node.capacity {
					logger.Info("Block capacity reached")

					// Sets current block's validator
					node.currentBlock.SetValidator(&node.nodeMap)
					// Sets current block's hash
					node.currentBlock.CalcHash()

					// If current node is validator
					if node.currentBlock.Validator == node.id {
						logger.Info("The node is broadcaster")
						// Sets block creation timestamp
						node.currentBlock.Timestamp = time.Now().UTC().Format(timeFormat)
						// Broadcasts block το cluster
						err = node.broadcastBlock(&node.currentBlock)
						if err != nil {
							logger.Error("Failed to broadcast new block", "error", err)
							return err
						}
						logger.Info("Block broadcasted by me")
					}

					// Initiates block listener method to receive broadcasted block
					err = node.blockListener()
					if err != nil {
						logger.Error("Block listener failed", "error", err)
						return err
					}
				}
			} else {
				// This case is not wished, returns error
				logger.Error("Capacity Reached")
				return errors.New("block capacity reached")
			}

		} else {
			// In case transaction is not possible basd to database data, it's rejected
			logger.Warn("Transaction rejected","error", err)
		}
	}
}

// Method that performs logic of reveiving broadcasted block
// and executing necessary actions after it
func (node *nodeConfig) blockListener() error {

	// Gets block
	block, validator, err := node.getNewBlock()
	if err != nil {
		logger.Error("Block listener exiting","error",err)
		return err
	}
	logger.Info("Received new block")

	// Checks if  received block's hash matches the one of the current block
	// meaning their contents are the same
	if node.currentBlock.CurrentHash == block.CurrentHash {

		// Appends received block to blockchain
		node.blockchain = append(node.blockchain, block)
		// Current block is a new empty block
		node.currentBlock = node.NewBlock()
		logger.Info("New block accepted", "validator", validator)

		// Corrects database, primarily returning stakes and crediting fees to validator
		node.addBlockUndoStake(&block)

		// Writes updated database to file
		err = node.WriteDB()
		if err != nil {
			logger.Error("Failed writing database", "error", err)
			return err
		}

		// Writes updated blockchain to file
		err = node.WriteBlockchain()
		if err != nil {
			logger.Error("Failed writing blockchain", "error", err)
			return err
		}
		logger.Info("Block add routine completed")

	} else {
		// Case block rejected (hashes do not match)
		logger.Warn("Block rejected")
	}
	return nil
}