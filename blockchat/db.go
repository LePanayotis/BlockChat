package blockchat

import (
	"encoding/json"
	"errors"
	"os"
)

// A database in memory is a map int->WalletData for easy manipulation
type DBmap map[int]WalletData

// Wallet data struct contains the information derived by a blockchain for a node's wallet
// Each wallet has a Balance in BlockChat Coins (float64), the current nonce,
// and a list of messages, represented by the Message struct below
type WalletData struct {
	Balance     float64   `json:"balance"`
	CurentNonce uint      `json:"curent_nonce"`
	Messages    []Message `json:"messages"`
}

// Each message has its sender (id of the sender node), the nonce of the transaction containing the message,
// and the content of the message
type Message struct {
	Sender  int    `json:"sender"`
	Nonce   uint   `json:"nonce"`
	Content string `json:"content"`
}

// Parses indicated file and returns a database
func LoadDB(_path string) (DBmap, error) {
	content, err := os.ReadFile(_path)
	if err != nil {
		return nil, err
	}

	// Makes initialy empty map and unmarshals json
	db := make(DBmap)
	if err := json.Unmarshal(content, &db); err != nil {
		return nil, err
	}
	return db, nil
}

// Writes json representation of database to indicated file
func (db *DBmap) WriteDB(_path string) error {
	file, err := os.Create(_path)
	if err != nil {
		return err
	}
	byteArray, err := json.Marshal(db)
	if err != nil {
		return err
	}
	_, err = file.Write(byteArray)
	if err != nil {
		return err
	}
	return nil
}

// Parses json file and sets node database to the one created from the file
func (node *nodeConfig) LoadDB() (DBmap, error) {
	var err error
	node.myDB, err = LoadDB(node.dbPath)
	return node.myDB, err
}

// Writes node's database to the file specified in node configuration
func (node *nodeConfig) WriteDB() error {
	return node.myDB.WriteDB(node.dbPath)
}

// Checks if an id has already an entry in the database
func (db *DBmap) accountExists(_accountId int) bool {
	return db.accountExistsAdd(_accountId, false)
}

// Checks if an id has already an entry in the database and adds it in the database
// according to _addIfNotExists
func (db *DBmap) accountExistsAdd(_accountId int, _addIfNotExists bool) bool {
	// Checks if account exists and its id does not equal -1 (->"0" wallet)
	_, exists := (*db)[_accountId]
	if !exists && _accountId != -1 {
		// Creates account if it doesn't exist according to _addIfNotExists
		if _addIfNotExists {
			(*db)[_accountId] = WalletData{
				Balance:  0,
				Messages: []Message{},
			}
		}
		// Account does not exist
		return false
	}
	// Account exists
	return true
}

// According to the pointed database data, checks if a transaction can be performed
// Meaning checks if the sender account has enough balance to send the transaction
// and its fees
func (db *DBmap) isTransactionPossible(_tx *Transaction, _accountId int) bool {
	// If sender is "0" wallet there's no need to check...they are definitely rich
	if _tx.SenderAddress == "0" {
		return true
	}
	// Amount + Fee <= Balance
	return _tx.CalcFee()+_tx.Amount <= (*db)[_accountId].Balance
}

// Changes an account's balance
func (db *DBmap) changeBalance(_accountId int, _amount float64) error {
	// If account is not "0" wallet (represented as -1)
	if _accountId != -1 {

		var newWallet WalletData = (*db)[_accountId]
		// Is balance sufficient?
		if newWallet.Balance+_amount < 0 {
			return errors.New("insufficient balance")
		}
		// Update wallet balance and database
		newWallet.Balance += _amount
		(*db)[_accountId] = newWallet
	}
	return nil
}

// Adds message transaction to database
// changeBalance needs to be performed before this
func (db *DBmap) addMessage(_accountId int, _m Message) {
	// Gets wallet from database
	newWallet := (*db)[_accountId]
	// Appends message to wallet
	newWallet.Messages = append(newWallet.Messages, _m)
	// Updates database entry
	(*db)[_accountId] = newWallet
}

// Adds transaction to database, provided translated sender and receiver ids
func (db *DBmap) addTransaction(_tx *Transaction, _senderId int, _receiverId int) (float64, error) {

	// Adds accounts if they don't already exist
	db.accountExistsAdd(_senderId, true)
	db.accountExistsAdd(_receiverId, true)

	// If sender is not "0" wallet
	if _senderId != -1 {
		//Check if the nonce of the transaction is the expected one as in the database
		if _tx.Nonce == (*db)[_senderId].CurentNonce+1 {
			// If nonce okey, increases the nonce in database
			db.increaseNonce(_senderId)
		} else {
			// Nonce inconsistent
			return 0, errors.New("Transaction nonce invalid")
		}
	}
	// Gets the fee
	fee := _tx.CalcFee()

	// Removes from sender's wallet the amount and the fee
	err := db.changeBalance(_senderId, 0-_tx.Amount-fee)
	if err != nil {
		return 0, err
	}

	// Credits the amount to the receiver wallet
	err = db.changeBalance(_receiverId, _tx.Amount)
	if err != nil {
		return 0, err
	}

	// Appends the message
	if _tx.TypeOfTransaction == "message" {
		db.addMessage(_receiverId, Message{
			Sender:  _senderId,
			Nonce:   _tx.Nonce,
			Content: _tx.Message,
		})
	}
	// returns the transaction fee
	return fee, nil

}

// Returns account's balance according to database
func (db *DBmap) getBalance(_accountId int) float64 {
	return (*db)[_accountId].Balance
}

// Increases an account's nonce by 1
func (db *DBmap) increaseNonce(_accountId int) uint {
	// If account exists
	if temp, b := (*db)[_accountId]; b {
		temp.CurentNonce++
		(*db)[_accountId] = temp
	}
	// Return increased nonce
	return (*db)[_accountId].CurentNonce
}

// Increases nonce of current node's account
func (node *nodeConfig) increaseNonce() uint {
	return node.myDB.increaseNonce(node.id)
}

// Gets nonce of account requested
func (db *DBmap) getNonce(_accountId int) uint {
	return (*db)[_accountId].CurentNonce
}

// Checks the node's database if a transaction is possible
func (node *nodeConfig) isTransactionPossible(_tx *Transaction) bool {
	// Maps key to node id
	accoundId := node.nodeMap[_tx.SenderAddress]
	// Checks
	return node.myDB.isTransactionPossible(_tx, accoundId)
}

// Adds transaction to node's database
func (node *nodeConfig) addTransactionToDB(_tx *Transaction) (float64, error) {

	var senderId, receiverId int
	// Maps wallets to node ids
	// If sender/receiver wallet is "0", it's mapped to -1
	if _tx.SenderAddress == "0" {
		senderId = -1
	} else {
		senderId = node.nodeMap[_tx.SenderAddress]
	}
	if _tx.ReceiverAddress == "0" {
		receiverId = -1
	} else {
		receiverId = node.nodeMap[_tx.ReceiverAddress]
	}
	// Adds transaction
	return node.myDB.addTransaction(_tx, senderId, receiverId)
}

// Adds block, and consequently all of its transactions to node's database
// Validity of the block and its transactions is not implemented in this stage
// TODO
func (node *nodeConfig) addBlockToDB(_block *Block) error {
	fee := float64(0)
	// For each transaction in the block
	for _, tx := range _block.Transactions {
		if !node.isTransactionPossible(&tx) {
			return errors.New("transaction from block not possible to add to database")
		}
		// TODO: i removed something from here earlier
		tmp, err := node.addTransactionToDB(&tx)
		if err != nil {
			return errors.Join(errors.New("failed to add block transaction to database"), err)
		}
		fee += tmp
	}
	// Credit the validator, the fees
	err := node.myDB.changeBalance(_block.Validator, fee)
	if err != nil {
		return errors.Join(errors.New("failed to credit fees to validator"), err)
	}
	// Credit the stakes back to their senders
	for _, tx := range _block.Transactions {
		if tx.ReceiverAddress == "0" {
			// Map wallet to node id
			senderId := node.nodeMap[tx.ReceiverAddress]
			// Change sender's balance with stake amount
			err := node.myDB.changeBalance(senderId, tx.Amount)
			if err != nil {
				return errors.Join(errors.New("failed to return stake from block transaction"), err)
			}
		}
	}

	return nil
}

// When a new block is received by the validator block, if it's hash is the same as the one
// recorded in the current node, then the database has already all the changes
// All that is left to do is to credit back the stakes back to their senders and the total fees to the validator
func (node *nodeConfig) addBlockUndoStake(_block *Block) error {

	fee := float64(0)
	// Parses all transactions
	for _, tx := range _block.Transactions {
		// If transaction is a stake
		if tx.ReceiverAddress == "0" {
			//Undos stakes
			// Maps wallet address to node id
			senderId := node.nodeMap[tx.SenderAddress]
			if tx.SenderAddress == "0" {
				senderId = -1
			}

			// Refunds amount
			err := node.myDB.changeBalance(senderId, tx.Amount)
			if err != nil {
				return errors.Join(errors.New("failed to refund stake"), err)
			}
			// Stakes have no fees
			continue
		}
		// Transaction fee is added to sum
		fee += tx.CalcFee()
	}
	// Sum of block fees is credited to the validator
	err := node.myDB.changeBalance(_block.Validator, fee)
	if err == nil {
		return nil
	}
	return errors.Join(errors.New("failed to credit fees to validator"), err)
}

// Sets the node database as instructed by its blockchain
func (node *nodeConfig) MakeDB() error {
	// Checks if node's blockchain is valid
	_, IsValid := node.IsBlockchainValid()
	// Return error if not valid
	if !IsValid {
		return errors.New("node's blockchain is not valid")
	}
	//  Create empty database
	node.myDB = make(DBmap)
	for _, block := range node.blockchain {
		// Adds block to database
		err := node.addBlockToDB(&block)
		if err != nil {
			return errors.Join(errors.New("failed to add block to database"), err)
		}
	}
	return nil
}
