package bcc

import (
	"encoding/json"
	"errors"
	"fmt"
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
	content, err := os.ReadFile(DB_PATH)
	if err != nil {
		fmt.Println(err)
	}
	db := make(DBmap)
	if err := json.Unmarshal(content, &db); err != nil {
		fmt.Println(err)
		return nil, err
	}

	return db, nil
}

func (db *DBmap) WriteDB() error {

	file, err := os.Create(DB_PATH)
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

func (db *DBmap) AccountExists(_account_key string) bool {
	return db.AccountExistsAdd(_account_key, false)
}

func (db *DBmap) AccountExistsAdd(_account_key string, _add_if_not_exists bool) bool {
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

func (db *DBmap) IsTransactionPossible(tx *Transaction) bool {
	if tx.Sender_address == "0" {
		return true
	}
	fmt.Printf("Fee : %f\nBalance: %f\n>", tx.CalcFee()+tx.Amount,(*db)[tx.Sender_address].Balance )
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
	if !tx.IsValid() {
		return 0, errors.New("Transaction verification failed")
	}
	db.AccountExistsAdd(tx.Sender_address, true)
	db.AccountExistsAdd(tx.Receiver_address, true)

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
	return 0, nil
}


func (db *DBmap) GetBalance(_account_key string) float64 {
	return (*db)[_account_key].Balance
}

func (db *DBmap) IncreaseNonce(_account_key string) uint {
	if _, b := (*db)[_account_key]; !b {
		temp := (*db)[_account_key]
		temp.Curent_Nonce++
		(*db)[_account_key] = temp
	}
	return (*db)[_account_key].Curent_Nonce
}