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

func LoadDB() (DBmap, error) {
	content, err := os.ReadFile(node.dbPath)
	if err != nil {
		return nil, err
	}
	db := make(DBmap)
	if err := json.Unmarshal(content, &db); err != nil {
		return nil, err
	}
	return db, nil
}

func (db *DBmap) WriteDB() error {

	file, err := os.Create(node.dbPath)
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

func (db *DBmap) accountExists(_account_key string) bool {
	return db.accountExistsAdd(_account_key, false)
}

func (db *DBmap) accountExistsAdd(_account_key string, _add_if_not_exists bool) bool {

	accountId := node.nodeMap[_account_key]

	_, exists := (*db)[accountId]
	if !exists && _account_key != "0" {
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

func (db *DBmap) isTransactionPossible(tx *Transaction) bool {
	if tx.Sender_address == "0" {
		return true
	}
	accountId := node.nodeMap[tx.Sender_address]
	logger.Info("Transaction details:","fee",tx.CalcFee(),"amount",tx.Amount,"balance",(*db)[accountId].Balance)
	return tx.CalcFee()+tx.Amount <= (*db)[accountId].Balance
}

func (db *DBmap) changeBalance(_account_key string, _amount float64) error {
	accountId := node.nodeMap[_account_key]
	if _account_key != "0" {
		var new_wallet WalletData = (*db)[accountId]
		if new_wallet.Balance+_amount >= 0 {
			new_wallet.Balance += _amount
			(*db)[accountId] = new_wallet
			return nil
		}
		return errors.New("insufficient balance")
	}
	return nil

}

func (db *DBmap) addMessage(_account_key string, m Message) {
	accountId := node.nodeMap[_account_key]
	if _account_key == "" {
		return
	}
	var new_wallet WalletData = (*db)[accountId]
	new_wallet.Messages = append(new_wallet.Messages, m)
	(*db)[accountId] = new_wallet
}


func (db *DBmap) addTransaction(tx *Transaction) (float64, error) {
	db.accountExistsAdd(tx.Sender_address, true)
	db.accountExistsAdd(tx.Receiver_address, true)
	if tx.Sender_address != "0" {
		db.increaseNonce(tx.Sender_address)
		// if tx.Nonce == (*db)[tx.Sender_address].Curent_Nonce+1 {
			
		// } else {
		// 	return 0, errors.New("Transaction nonce invalid")
		// }
	}

	fee := tx.CalcFee()
	err := db.changeBalance(tx.Sender_address, 0-tx.Amount-fee)
	if err == nil {
		db.changeBalance(tx.Receiver_address, tx.Amount)
		if tx.Type_of_transaction == "message" {

			senderId := node.nodeMap[tx.Sender_address]
			db.addMessage(tx.Receiver_address, Message{
				Sender:  senderId,
				Nonce:   tx.Nonce,
				Content: tx.Message,
			})
		}
		return fee, nil
	}
	return 0, err
}

func (db *DBmap) getBalance(_account_key string) float64 {
	accountId := node.nodeMap[_account_key]
	return (*db)[accountId].Balance
}

func (db *DBmap) increaseNonce(_account_key string) uint {
	accountId := node.nodeMap[_account_key]
	if _, b := (*db)[accountId]; b {
		temp := (*db)[accountId]
		temp.Curent_Nonce++
		(*db)[accountId] = temp
	}
	return (*db)[accountId].Curent_Nonce
}

func (db *DBmap) getNonce(_account_key string) uint {
	accountId := node.nodeMap[_account_key]
	return (*db)[accountId].Curent_Nonce
}

func (db *DBmap) addBlock(block *Block) error {
	fee := float64(0)
	for _, tx := range block.Transactions {
		if !db.isTransactionPossible(&tx) {
			continue
		}
		if tx.Receiver_address == "0" {
			continue
		}
		tmp, _ := db.addTransaction(&tx)
		fee += tmp
	}
	db.changeBalance(block.Validator, fee)
	return nil
}

func (db *DBmap) addBlockUndoStake(block *Block) error {
	fee := float64(0)
	for _, tx := range block.Transactions {
		if tx.Receiver_address == "0" {
			db.changeBalance(tx.Sender_address, tx.Amount)
			continue
		}
		fee += tx.CalcFee()
	}
	db.changeBalance(block.Validator, fee)
	return nil
}
