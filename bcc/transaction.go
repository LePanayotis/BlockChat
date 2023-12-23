package bcc

import (
	"crypto"
	"crypto/rand"
	"crypto/rsa"
	"crypto/sha256"
	"crypto/x509"
	"encoding/hex"
	"encoding/json"
	"errors"
	"strconv"
)

type Transaction struct {
	Sender_address      string  `json:"sender_address"`
	Receiver_address    string  `json:"receiver_address"`
	Type_of_transaction string  `json:"type_of_transaction"`
	Amount              float64 `json:"amount"`
	Message             string  `json:"message"`
	Nonce               uint     `json:"nonce"`
	Transaction_id      string  `json:"transaction_id"`
	Signature           string  `json:"Signature"`
}

func NewMessageTransaction(_Sender_address string, _Receiver_address string, _message string,
	_nonce uint, _private_key string) Transaction {

	return newTransactionInstance(_Sender_address, _Receiver_address, "message", 0, _message, _nonce, _private_key)
}

func NewTransferTransaction(_Sender_address string, _Receiver_address string, _amount float64,
	_nonce uint, _private_key string) Transaction {

	return newTransactionInstance(_Sender_address, _Receiver_address, "transfer", _amount, "", _nonce, _private_key)
}

func (t *Transaction) JSONify() (string, error) {

	if !t.IsValid() {
		return "", errors.New("Transaction is not valid, can't be jsonified")
	}

	jsonStringBytes, err := json.Marshal(t)
	if err != nil {
		return "", err
	}

	jsonString := string(jsonStringBytes)

	return jsonString, nil

}

func (t *Transaction) IsValid() bool {
	if t.Sender_address == "" ||
		t.Receiver_address == "" ||
		t.Amount < 0 {
		return false
	}
	if t.Type_of_transaction == "transfer" &&
		t.Amount > 0 &&
		t.Message == "" {
			return true
	} else if t.Type_of_transaction == "message" &&
		t.Message != "" &&
		t.Amount == 0 && (t.Receiver_address != "0" && t.Sender_address != "0") {
			return true
	}
	return false
}

func ParseTransactionJSON(s string) (Transaction, error) {
	var t Transaction

	if err := json.Unmarshal([]byte(s), &t); err != nil {
		return t, err
	}
	return t, nil

}

func newTransactionInstance(_Sender_address string, _Receiver_address string,
	_Type_of_transaction string, _amount float64, _message string,
	_nonce uint, _private_key string) Transaction {

	var t Transaction
	t.Sender_address = _Sender_address
	t.Receiver_address = _Receiver_address
	t.Type_of_transaction = _Type_of_transaction
	t.Amount = _amount
	t.Nonce = _nonce
	t.Message = _message
	
	t.Sign(_private_key)
	return t

}

func (t *Transaction) GetConcat() string {
	concat := t.Sender_address + t.Receiver_address + t.Type_of_transaction + strconv.FormatFloat(t.Amount, 'f', -1, 64) + strconv.Itoa(int(t.Nonce)) + t.Message
	return concat
}

func (t *Transaction) GetHash() ([32]byte, error) {
	concat := []byte(t.GetConcat())
	hash := sha256.Sum256(concat)

	return hash, nil
}

func (t *Transaction) SetHash(hash [32]byte) (string, error) {
	t.Transaction_id = hex.EncodeToString(hash[:])
	return t.Transaction_id, nil
}

func (t *Transaction) Sign(_private_key string) (string, error) {
	_transaction_hash, _ := t.GetHash()

	t.SetHash(_transaction_hash)

	private_key_array, _ := hex.DecodeString(_private_key)

	private_key, _ := x509.ParsePKCS1PrivateKey(private_key_array)

	signature_bytes, _ := rsa.SignPKCS1v15(rand.Reader, private_key, crypto.SHA256, _transaction_hash[:])

	signature := hex.EncodeToString(signature_bytes)

	t.Signature = signature

	return signature, nil

}

func (t *Transaction) Verify() (bool) {
	if (!t.IsValid()){
		return false
	}
	produced_hash, _ := t.GetHash()

	signature, _ := hex.DecodeString(t.Signature)

	var public_key_string string
	if t.Sender_address == "0" {
		//If sender is genesis block, then t is signed with bootstrap node keys
		public_key_string = t.Receiver_address
	} else {
		public_key_string = t.Sender_address
	}

	public_key_array, _ := hex.DecodeString(public_key_string)

	public_key, _ := x509.ParsePKCS1PublicKey(public_key_array[:])

	err := rsa.VerifyPKCS1v15(public_key, crypto.SHA256, produced_hash[:], signature[:])
	return err == nil
}

func (t *Transaction) CalcFee() float64 {
	if t.Sender_address == "0" || t.Receiver_address == "0" {
		return 0
	} else if (t.Type_of_transaction =="transfer") {
		return t.Amount*FEE_PERCENTAGE
	} else if (t.Type_of_transaction =="message"){
		return float64(len(t.Message)*COST_PER_CHAR)
	}
	return 0
}