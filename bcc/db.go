package bcc

import (
	"encoding/json"
	"errors"
	"os"
)

type DBmap map[string]WalletData

type Message struct {
	Sender  string `json:"sender"`
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
	_, exists := (*db)[_account_key]
	if !exists && _account_key != "0" {
		if _add_if_not_exists {
			(*db)[_account_key] = WalletData{
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
	return tx.CalcFee()+tx.Amount <= (*db)[tx.Sender_address].Balance
}

func (db *DBmap) changeBalance(_account_key string, _amount float64) error {
	if _account_key != "0" {
		var new_wallet WalletData = (*db)[_account_key]
		if new_wallet.Balance+_amount >= 0 {
			new_wallet.Balance += _amount
			(*db)[_account_key] = new_wallet
			return nil
		}
		return errors.New("insufficient balance")
	}
	return nil

}

func (db *DBmap) addMessage(_account_key string, m Message) {
	if _account_key == "" {
		return
	}
	var new_wallet WalletData = (*db)[_account_key]
	new_wallet.Messages = append(new_wallet.Messages, m)
	(*db)[_account_key] = new_wallet
}

func (db *DBmap) addTransaction(tx *Transaction) (float64, error) {
	if !tx.Verify() {
		return 0, errors.New("Transaction verification failed")
	}
	db.accountExistsAdd(tx.Sender_address, true)
	db.accountExistsAdd(tx.Receiver_address, true)
	if tx.Sender_address != "0" {
		if tx.Nonce == (*db)[tx.Sender_address].Curent_Nonce+1 {
			db.increaseNonce(tx.Sender_address)
		} else {
			return 0, errors.New("Transaction nonce invalid")
		}
	}

	fee := tx.CalcFee()
	err := db.changeBalance(tx.Sender_address, 0-tx.Amount-fee)
	if err == nil {
		db.changeBalance(tx.Receiver_address, tx.Amount)
		if tx.Type_of_transaction == "message" {
			db.addMessage(tx.Receiver_address, Message{
				Sender:  tx.Sender_address,
				Nonce:   tx.Nonce,
				Content: tx.Message,
			})
		}
		return fee, nil
	}
	return 0, err
}

func (db *DBmap) getBalance(_account_key string) float64 {
	return (*db)[_account_key].Balance
}

func (db *DBmap) increaseNonce(_account_key string) uint {
	if _, b := (*db)[_account_key]; b {
		temp := (*db)[_account_key]
		temp.Curent_Nonce++
		(*db)[_account_key] = temp
	}
	return (*db)[_account_key].Curent_Nonce
}

func (db *DBmap) getNonce(_account_key string) uint {
	return (*db)[_account_key].Curent_Nonce
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
