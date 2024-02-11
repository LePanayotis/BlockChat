package blockchat

import (
	"encoding/json"
	"errors"
	"os"
)

type DBmap map[int]WalletData

type Message struct {
	Sender  int `json:"sender"`
	Nonce   uint   `json:"nonce"`
	Content string `json:"content"`
}

type WalletData struct {
	Balance      float64   `json:"balance"`
	Curent_Nonce uint      `json:"curent_nonce"`
	Messages     []Message `json:"messages"`
}

func LoadDB(path string) (DBmap, error) {
	content, err := os.ReadFile(path)
	if err != nil {
		return nil, err
	}
	db := make(DBmap)
	if err := json.Unmarshal(content, &db); err != nil {
		return nil, err
	}
	return db, nil
}


func (db *DBmap) WriteDB(path string) error {

	file, err := os.Create(path)
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

func (node *nodeConfig) LoadDB() (DBmap, error) {
	return LoadDB(node.dbPath)
}

func (node *nodeConfig) WriteDB() error { 
	return node.myDB.WriteDB(node.dbPath)
}


func (db *DBmap) accountExists(accountId int) bool {
	return db.accountExistsAdd(accountId, false)
}

func (db *DBmap) accountExistsAdd(accountId int, _add_if_not_exists bool) bool {
	_, exists := (*db)[accountId]
	if !exists && accountId != -1 {
		if _add_if_not_exists {
			(*db)[accountId] = WalletData{
				Balance:  0,
				Messages: []Message{},
			}
		}
		return false
	}
	return true
}

func (db *DBmap) isTransactionPossible(tx *Transaction, accountId int) bool {
	if tx.Sender_address == "0" {
		return true
	}
	logger.Info("Transaction details:","fee",tx.CalcFee(),"amount",tx.Amount,"balance",(*db)[accountId].Balance)
	return tx.CalcFee()+tx.Amount <= (*db)[accountId].Balance
}

func (node *nodeConfig) isTransactionPossible(tx *Transaction) bool {
	accoundId := node.nodeMap[tx.Sender_address]
	return node.myDB.isTransactionPossible(tx, accoundId)

}

func (db *DBmap) changeBalance(_accountId int, _amount float64) error {
	if _accountId != -1 {
		var new_wallet WalletData = (*db)[_accountId]
		if new_wallet.Balance+_amount >= 0 {
			new_wallet.Balance += _amount
			(*db)[_accountId] = new_wallet
			return nil
		}
		return errors.New("insufficient balance")
	}
	return nil
}

func (db *DBmap) addMessage(_accountId int, m Message) error{
	if _accountId < 0 {
		return errors.New("account non existent")
	}
	var newWallet WalletData = (*db)[_accountId]
	newWallet.Messages = append(newWallet.Messages, m)
	(*db)[_accountId] = newWallet
	return nil
}


func (db *DBmap) addTransaction(tx *Transaction, _senderId int, _receiverId int) (float64, error) {

	db.accountExistsAdd(_senderId, true)
	db.accountExistsAdd(_receiverId, true)
	
	if _senderId != -1 {
		if tx.Nonce == (*db)[_senderId].Curent_Nonce+1 {
			db.increaseNonce(_senderId)
		} else {
			return 0, errors.New("Transaction nonce invalid")
		}
	}
	fee := tx.CalcFee()
	err := db.changeBalance(_senderId, 0-tx.Amount-fee)
	if err == nil {
		db.changeBalance(_receiverId, tx.Amount)
		if tx.Type_of_transaction == "message" {			
			db.addMessage(_receiverId, Message{
				Sender:  _senderId,
				Nonce:   tx.Nonce,
				Content: tx.Message,
			})
		}
		return fee, nil
	}
	return 0, err
}

func (db *DBmap) getBalance(_accountId int) float64 {
	return (*db)[_accountId].Balance
}

func (db *DBmap) increaseNonce(_accountId int) uint {
	if _, b := (*db)[_accountId]; b {
		temp := (*db)[_accountId]
		temp.Curent_Nonce++
		(*db)[_accountId] = temp
	}
	return (*db)[_accountId].Curent_Nonce
}

func (db *DBmap) getNonce(_accountId int) uint {
	return (*db)[_accountId].Curent_Nonce
}


func (node *nodeConfig) addTransactionToDB(tx *Transaction) (float64, error) {
	var _senderId, _receiverId int
	if tx.Sender_address == "0" {
		_senderId = -1
	} else {
		_senderId = node.nodeMap[tx.Sender_address]
	}
	if tx.Receiver_address == "0" {
		_receiverId = -1
	}else {
		_receiverId = node.nodeMap[tx.Receiver_address]
	}
	
	return node.myDB.addTransaction(tx, _senderId, _receiverId)
}


func (node * nodeConfig) addBlockToDB(block *Block) error {
	fee := float64(0)
	db :=  &node.myDB
	for _, tx := range block.Transactions {
		if !node.isTransactionPossible(&tx) {
			continue
		}
		if tx.Receiver_address == "0" {
			continue
		}
		tmp, _ := node.addTransactionToDB(&tx)
		fee += tmp
	}
	db.changeBalance(block.Validator, fee)
	return nil
}

func (node *nodeConfig) addBlockUndoStake(block *Block) error {
	fee := float64(0)
	for _, tx := range block.Transactions {
		if tx.Receiver_address == "0" {
			//Undos stakes
			senderId := node.nodeMap[tx.Sender_address]
			if tx.Sender_address == "0" {
				senderId = -1	
			}
			node.myDB.changeBalance(senderId, tx.Amount)
			continue
		}
		fee += tx.CalcFee()
	}
	node.myDB.changeBalance(block.Validator, fee)
	return nil
}
