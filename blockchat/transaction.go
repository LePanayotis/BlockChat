package blockchat

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

//The basic struct representing the transactions in BlockChat
//`Sender_address` and `Reiceiver_address` are the public keys of
//the sender and the receiver of the transaction respectively,
//whether it is a simple message or a transfer transaction.
//
//The `Type_of_transaction` can have the values: "message" or "transfer" 
//The property `message` bears the content of the message and the
//property `amount` represents the amount of BlockChatCoins to be transferred
//
//`Transaction_id` is the hash of the string representation of the transaction
//as defined in GetConcat method.
//`Signature` is the encrypted `Transaction_id` with the private key of the sender
type Transaction struct {
	Sender_address      string  `json:"sender_address"`
	Receiver_address    string  `json:"receiver_address"`
	Type_of_transaction string  `json:"type_of_transaction"`
	Amount              float64 `json:"amount"`
	Message             string  `json:"message"`
	Nonce               uint    `json:"nonce"`
	Transaction_id      string  `json:"transaction_id"`
	Signature           string  `json:"Signature"`
}

//Creates a new message transaction instance
//Sender's private key is required to sign the transaction
func NewMessageTransaction(_Sender_address string, _Receiver_address string, _message string,
	_nonce uint, _private_key string) Transaction {

	return newTransactionInstance(_Sender_address, _Receiver_address, "message", 0, _message, _nonce, _private_key)
}

//Creates a new transfer transaction instance
//Sender's private key is required to sign the transaction
func NewTransferTransaction(_Sender_address string, _Receiver_address string, _amount float64,
	_nonce uint, _private_key string) Transaction {

	return newTransactionInstance(_Sender_address, _Receiver_address, "transfer", _amount, "", _nonce, _private_key)
}

//Returns the transaction in json string representation
//and error if the parsing fails
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


//Checks if transaction `t` is valid. This means:
//
//`t`'s `Sender_address` and `Receiver_address` are not empty strings
//
//`Amount` is not negative
//
//`Type_of_transaction` is either "trasfer" or "message"
//
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


//Returns a Transaction instance from a json string
//Error if parsing fails
func ParseTransactionJSON(s string) (Transaction, error) {
	var t Transaction

	if err := json.Unmarshal([]byte(s), &t); err != nil {
		return t, err
	}
	return t, nil

}

//Fundamental function to create a new Transaction instance from the given parameters
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
	
	_, err := t.Sign(_private_key)
	if err != nil {
		return Transaction{}
	}
	return t

}


//Returns a concatenation of basic properties of the transactions
func (t *Transaction) GetConcat() string {
	concat := t.Sender_address + t.Receiver_address + t.Type_of_transaction + strconv.FormatFloat(t.Amount, 'f', -1, 64) + strconv.Itoa(int(t.Nonce)) + t.Message
	return concat
}

//Returns hash of the string concatenation of a transaction
func (t *Transaction) GetHash() ([32]byte, error) {
	concat := []byte(t.GetConcat())
	hash := sha256.Sum256(concat)

	return hash, nil
}


//Sets the `Transaction_id`
func (t *Transaction) SetHash(hash [32]byte) (string, error) {
	t.Transaction_id = hex.EncodeToString(hash[:])
	return t.Transaction_id, nil
}


//Signs the transaction with the given private key
func (t *Transaction) Sign(_private_key string) (string, error) {
	_transaction_hash, err := t.GetHash()
	if err != nil {
		return "", err
	}
	_, err = t.SetHash(_transaction_hash)
	if err != nil {
		return "", err
	}

	private_key_array, err := hex.DecodeString(_private_key)
	if err != nil {
		return "", err
	}

	private_key, err := x509.ParsePKCS1PrivateKey(private_key_array)
	if err != nil {
		return "", err
	}

	signature_bytes, err := rsa.SignPKCS1v15(rand.Reader, private_key, crypto.SHA256, _transaction_hash[:])
	if err != nil {
		return "", err
	}

	signature := hex.EncodeToString(signature_bytes)

	t.Signature = signature

	return signature, nil

}

//Verifies transaction:
//Checks if transaction properties values conform to the constraints
//Then verifies signature with the sender's public key
func (t *Transaction) Verify() (bool) {
	if (!t.IsValid()){
		return false
	}
	produced_hash, err := t.GetHash()
	if err != nil {
		return false
	}

	signature, err := hex.DecodeString(t.Signature)
	if err != nil {
		return false
	}

	var public_key_string string

	if t.Sender_address == "0" {
		//If sender is genesis block, then t is signed with bootstrap node keys
		public_key_string = t.Receiver_address
	} else {
		public_key_string = t.Sender_address
	}

	public_key_array, err := hex.DecodeString(public_key_string)
	if err != nil {
		return false
	}

	public_key, err := x509.ParsePKCS1PublicKey(public_key_array[:])
	if err != nil {
		return false
	}

	err = rsa.VerifyPKCS1v15(public_key, crypto.SHA256, produced_hash[:], signature[:])

	return err == nil
}



// var node only used here:
//Calculates the fee based on the type of the transaction
func (t *Transaction) CalcFee() float64 {
	if t.Sender_address == "0" || t.Receiver_address == "0" {
		return 0
	} else if (t.Type_of_transaction =="transfer") {
		return t.Amount*node.feePercentage
	} else if (t.Type_of_transaction =="message"){
		return float64(len(t.Message)*node.costPerChar)
	}
	return 0
}