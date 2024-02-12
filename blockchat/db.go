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
	CurentNonce uint      `json:"curent_nonce"`
	Messages     []Message `json:"messages"`
}

func LoadDB(_path string) (DBmap, error) {
	content, err := os.ReadFile(_path)
	if err != nil {
		return nil, err
	}
	db := make(DBmap)
	if err := json.Unmarshal(content, &db); err != nil {
		return nil, err
	}
	return db, nil
}


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

func (node *nodeConfig) LoadDB() (DBmap, error) {
	return LoadDB(node.dbPath)
}

func (node *nodeConfig) WriteDB() error { 
	return node.myDB.WriteDB(node.dbPath)
}


func (db *DBmap) accountExists(_accountId int) bool {
	return db.accountExistsAdd(_accountId, false)
}

func (db *DBmap) accountExistsAdd(_accountId int, _addIfNotExists bool) bool {
	_, exists := (*db)[_accountId]
	if !exists && _accountId != -1 {
		if _addIfNotExists {
			(*db)[_accountId] = WalletData{
				Balance:  0,
				Messages: []Message{},
			}
		}
		return false
	}
	return true
}

func (db *DBmap) isTransactionPossible(_tx *Transaction, _accountId int) bool {
	if _tx.SenderAddress == "0" {
		return true
	}
	logger.Info("Transaction details:","fee",_tx.CalcFee(),"amount",_tx.Amount,"balance",(*db)[_accountId].Balance)
	return _tx.CalcFee()+_tx.Amount <= (*db)[_accountId].Balance
}

func (db *DBmap) changeBalance(_accountId int, _amount float64) error {
	if _accountId != -1 {
		var newWallet WalletData = (*db)[_accountId]
		if newWallet.Balance+_amount >= 0 {
			newWallet.Balance += _amount
			(*db)[_accountId] = newWallet
			return nil
		}
		return errors.New("insufficient balance")
	}
	return nil
}

func (db *DBmap) addMessage(_accountId int, _m Message) error{
	if _accountId < 0 {
		return errors.New("account non existent")
	}
	var newWallet WalletData = (*db)[_accountId]
	newWallet.Messages = append(newWallet.Messages, _m)
	(*db)[_accountId] = newWallet
	return nil
}


func (db *DBmap) addTransaction(_tx *Transaction, _senderId int, _receiverId int) (float64, error) {

	db.accountExistsAdd(_senderId, true)
	db.accountExistsAdd(_receiverId, true)
	
	if _senderId != -1 {
		if _tx.Nonce == (*db)[_senderId].CurentNonce+1 {
			db.increaseNonce(_senderId)
		} else {
			return 0, errors.New("Transaction nonce invalid")
		}
	}
	fee := _tx.CalcFee()
	err := db.changeBalance(_senderId, 0-_tx.Amount-fee)
	if err == nil {
		db.changeBalance(_receiverId, _tx.Amount)
		if _tx.TypeOfTransaction == "message" {			
			db.addMessage(_receiverId, Message{
				Sender:  _senderId,
				Nonce:   _tx.Nonce,
				Content: _tx.Message,
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
		temp.CurentNonce++
		(*db)[_accountId] = temp
	}
	return (*db)[_accountId].CurentNonce
}

func (node *nodeConfig) increaseNonce() uint {
	return node.myDB.increaseNonce(node.id)
}

func (db *DBmap) getNonce(_accountId int) uint {
	return (*db)[_accountId].CurentNonce
}

func (node *nodeConfig) isTransactionPossible(_tx *Transaction) bool {
	accoundId := node.nodeMap[_tx.SenderAddress]
	return node.myDB.isTransactionPossible(_tx, accoundId)

}

func (node *nodeConfig) addTransactionToDB(_tx *Transaction) (float64, error) {
	var senderId, receiverId int
	if _tx.SenderAddress == "0" {
		senderId = -1
	} else {
		senderId = node.nodeMap[_tx.SenderAddress]
	}
	if _tx.ReceiverAddress == "0" {
		receiverId = -1
	}else {
		receiverId = node.nodeMap[_tx.ReceiverAddress]
	}
	
	return node.myDB.addTransaction(_tx, senderId, receiverId)
}


func (node * nodeConfig) addBlockToDB(_block *Block) error {
	fee := float64(0)
	db :=  &node.myDB
	for _, tx := range _block.Transactions {
		if !node.isTransactionPossible(&tx) {
			continue
		}
		if tx.ReceiverAddress == "0" {
			continue
		}
		tmp, _ := node.addTransactionToDB(&tx)
		fee += tmp
	}
	db.changeBalance(_block.Validator, fee)
	return nil
}

func (node *nodeConfig) addBlockUndoStake(_block *Block) error {
	fee := float64(0)
	for _, tx := range _block.Transactions {
		if tx.ReceiverAddress == "0" {
			//Undos stakes
			senderId := node.nodeMap[tx.SenderAddress]
			if tx.SenderAddress == "0" {
				senderId = -1	
			}
			node.myDB.changeBalance(senderId, tx.Amount)
			continue
		}
		fee += tx.CalcFee()
	}
	node.myDB.changeBalance(_block.Validator, fee)
	return nil
}
