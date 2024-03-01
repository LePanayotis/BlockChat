package blockchat

import (
	"encoding/json"
	"errors"
	"fmt"
	"os"
)

// A database in memory is a map int->WalletData for easy manipulation
type Database map[int]Wallet

// Wallet data struct contains the information derived by a blockchain for a node's wallet
// Each wallet has a Balance in BlockChat Coins (float64), the current nonce,
// and a list of messages, represented by the Message struct below
type Wallet struct {
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
func LoadDB(_path string) (Database, error) {
	content, err := os.ReadFile(_path)
	if err != nil {
		return nil, err
	}

	// Makes initialy empty map and unmarshals json
	db := make(Database)
	if err := json.Unmarshal(content, &db); err != nil {
		return nil, err
	}
	return db, nil
}

// Writes json representation of database to indicated file
func (db *Database) WriteDB(_path string) error {
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

// Checks if an id has already an entry in the database and adds it in the database
// according to _addIfNotExists
func (db *Database) accountExistsAdd(_accountId int) bool {
	// Checks if account exists and its id does not equal -1 (->"0" wallet)
	_, exists := db.getWalletExists(_accountId)
	if !exists && _accountId != -1 {
		// Creates account if it doesn't exist

		db.setWallet(_accountId, Wallet{
			Balance:  0,
			Messages: []Message{},
		})

		// Account does not exist
		return false
	}
	// Account exists
	return true
}

// According to the pointed database data, checks if a transaction can be performed
// Meaning checks if the sender account has enough balance to send the transaction
// and its fees
func (db *Database) isTransactionPossibleSetNonce(_tx *Transaction, _accountId int) error {
	// Amount + Fee <= Balance

	wallet, exists := db.getWalletExists(_accountId)
	if _accountId == -1 {
		return nil
	}

	if !exists {
		return fmt.Errorf("exists no account for node %v, transaction not possible", _accountId)
	}

	withdraw := _tx.CalcFee() + _tx.Amount
	nonceInv, balanceIns := _tx.Nonce <= wallet.CurentNonce, withdraw > wallet.Balance
	var err error = nil
	if nonceInv {
		err = fmt.Errorf("transaction nonce expected greater than %v, got %v for node %v", wallet.CurentNonce, _tx.Nonce, _accountId)
	}

	//DANGEROUS POINT
	db.setNonce(_accountId, _tx.Nonce)
	if balanceIns {
		err = errors.Join(err, fmt.Errorf("insufficient balance: %v, requested amount: %v from node: %v", wallet.Balance, withdraw, _accountId))
	}
	return err
}

// Changes an account's balance
func (db *Database) addToBalance(_accountId int, _amount float64) {
	if _accountId == -1 {
		return
	}
	wallet := db.getWallet(_accountId)
	// Update wallet balance and database
	wallet.Balance += _amount
	db.setWallet(_accountId, wallet)

}

// Adds message transaction to database
// changeBalance needs to be performed before this
func (db *Database) addMessage(_accountId int, _m Message) {
	// Gets wallet from database
	newWallet := db.getWallet(_accountId)
	// Appends message to wallet
	newWallet.Messages = append(newWallet.Messages, _m)
	// Updates database entry
	db.setWallet(_accountId, newWallet)
}

// Adds transaction to database, provided translated sender and receiver ids
// just the transaction is added, NO CHECK IS DONE
func (db *Database) addTransaction(_tx *Transaction, _senderId int, _receiverId int) float64 {

	// Gets the fee
	fee := _tx.CalcFee()

	// Removes from sender's wallet the amount and the fee

	db.addToBalance(_senderId, 0-_tx.Amount-fee)
	// Appends the message
	if _tx.Type == "message" {
		db.addMessage(_receiverId, Message{
			Sender:  _senderId,
			Nonce:   _tx.Nonce,
			Content: _tx.Message,
		})
	} else {
		// Credits the amount to the receiver wallet
		db.addToBalance(_receiverId, _tx.Amount)
	}
	return fee
}

// Returns account's balance according to database
func (db *Database) getBalance(_accountId int) float64 {
	return (*db)[_accountId].Balance
}

func (db *Database) setWallet(_accountId int, wallet Wallet) {
	(*db)[_accountId] = wallet
}

func (db *Database) getWalletExists(_accountId int) (Wallet, bool) {
	wallet, exists := (*db)[_accountId]
	return wallet, exists
}

func (db *Database) getWallet(_accountId int) Wallet {
	return (*db)[_accountId]
}

func (db *Database) setNonce(_accountId int, _nonce uint) {
	// If account exists
	if wallet, b := db.getWalletExists(_accountId); b {
		wallet.CurentNonce = _nonce
		db.setWallet(_accountId, wallet)

	}
}

// Gets nonce of account requested
func (db *Database) getNonce(_accountId int) uint {
	return (*db)[_accountId].CurentNonce
}

// Checks the node's database if a transaction is possible
func (node *nodeConfig) isTransactionPossibleSetNonce(_tx *Transaction) error {
	// Maps key to node id
	accoundId := node.nodeMap[_tx.Sender]
	// Checks
	return node.myDB.isTransactionPossibleSetNonce(_tx, accoundId)
}

// Adds transaction to node's database
func (node *nodeConfig) addTransactionToDB(_tx *Transaction) {

	senderId, receiverId := addressToId(&node.nodeMap, _tx.Sender), addressToId(&node.nodeMap, _tx.Receiver)
	// Adds transaction
	node.myDB.addTransaction(_tx, senderId, receiverId)
}

func addressToId(addressMap *map[string]int, _address string) int {
	if _address == "0" {
		return -1
	}
	return (*addressMap)[_address]
}

func (db *Database) addBlock(_block *Block, _adressMap *map[string]int) {
	fee := float64(0)
	// For each transaction in the block
	for _, tx := range _block.Transactions {
		senderId := addressToId(_adressMap, tx.Sender)
		receiverId := addressToId(_adressMap, tx.Receiver)

		// Adds accounts if they don't already exist
		db.accountExistsAdd(senderId)
		db.accountExistsAdd(receiverId)
		if senderId != -1 {
			err := db.isTransactionPossibleSetNonce(&tx, senderId)
			if err != nil {
				logger.Error("Could not add transaction from block", "error", err)
			}
		}
		fee = fee + db.addTransaction(&tx, senderId, receiverId)

	}
	// Credit the validator, the fees
	db.addToBalance(_block.Validator, fee)

	// Credit the stakes back to their senders
	for _, tx := range _block.Transactions {
		if tx.Receiver == "0" {
			// Map wallet to node id
			senderId := addressToId(_adressMap, tx.Sender)
			// Change sender's balance with stake amount
			db.addToBalance(senderId, tx.Amount)
		}
	}
}

// Adds block, and consequently all of its transactions to node's database
// Validity of the block and its transactions is not implemented in this stage
// TODO
func (node *nodeConfig) addBlockToDB(_block *Block) {
	node.myDB.addBlock(_block, &node.nodeMap)
}

// When a new block is received by the validator block, if it's hash is the same as the one
// recorded in the current node, then the database has already all the changes
// All that is left to do is to credit back the stakes back to their senders and the total fees to the validator
func (db *Database) addBlockUndoStake(_block *Block, _addressMap *map[string]int) {

	fee := float64(0)
	// Parses all transactions
	for _, tx := range _block.Transactions {
		// If transaction is a stake
		if tx.Receiver == "0" {
			//Undos stakes
			// Maps wallet address to node id
			senderId := addressToId(_addressMap, tx.Sender)

			// Refunds amount
			db.addToBalance(senderId, tx.Amount)
			// Stakes have no fees
			continue
		}
		// Transaction fee is added to sum
		fee += tx.CalcFee()
	}
	// Sum of block fees is credited to the validator
	db.addToBalance(_block.Validator, fee)
}

func (node *nodeConfig) addBlockUndoStake(_block *Block) {
	node.myDB.addBlockUndoStake(_block, &node.nodeMap)
}

// Sets the node database as instructed by its blockchain
func (node *nodeConfig) MakeDB() error {
	// Checks if node's blockchain is valid
	IsValid := node.IsBlockchainValid()
	// Return error if not valid
	if !IsValid {
		return errors.New("node's blockchain is not valid")
	}
	//  Create empty database
	node.myDB = make(Database)
	for _, block := range node.blockchain {
		// Adds block to database
		node.addBlockToDB(&block)
	}
	return nil
}

// Parses json file and sets node database to the one created from the file
func (node *nodeConfig) LoadDB() (Database, error) {
	var err error
	node.myDB, err = LoadDB(node.dbPath)
	return node.myDB, err
}

// Writes node's database to the file specified in node configuration
func (node *nodeConfig) WriteDB() error {
	return node.myDB.WriteDB(node.dbPath)
}
