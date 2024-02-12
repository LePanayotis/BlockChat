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
//`SenderAddress` and `ReiceiverAddress` are the public keys of
//the sender and the receiver of the transaction respectively,
//whether it is a simple message or a transfer transaction.
//
//The `TypeOfTransaction` can have the values: "message" or "transfer" 
//The property `Message` bears the content of the message and the
//property `Amount` represents the amount of BlockChatCoins to be transferred
//
//`TransactionId` is the hash of the string representation of the transaction
//as defined in GetConcat method.
//`Signature` is the encrypted `TransactionId` with the private key of the sender
type Transaction struct {
	SenderAddress      string  `json:"sender_address"`
	ReceiverAddress    string  `json:"receiver_address"`
	TypeOfTransaction string  `json:"type_of_transaction"`
	Amount              float64 `json:"amount"`
	Message             string  `json:"message"`
	Nonce               uint    `json:"nonce"`
	TransactionId      string  `json:"transaction_id"`
	Signature           string  `json:"Signature"`
}

//Creates a new message transaction instance
//Sender's private key is required to sign the transaction
func NewMessageTransaction(_senderAddress string, _receiverAddress string, _message string,
	_nonce uint, _privateKey string) Transaction {

	return newTransactionInstance(_senderAddress, _receiverAddress, "message", 0, _message, _nonce, _privateKey)
}

//Creates a new transfer transaction instance
//Sender's private key is required to sign the transaction
func NewTransferTransaction(_senderAddress string, _receiverAddress string, _amount float64,
	_nonce uint, _privateKey string) Transaction {

	return newTransactionInstance(_senderAddress, _receiverAddress, "transfer", _amount, "", _nonce, _privateKey)
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
//`t`'s `SenderAddress` and `ReceiverAddress` are not empty strings
//
//`Amount` is not negative
//
//`TypeOfTransaction` is either "trasfer" or "message"
//
func (t *Transaction) IsValid() bool {
	if t.SenderAddress == "" ||
		t.ReceiverAddress == "" ||
		t.Amount < 0 {
		return false
	}
	if t.TypeOfTransaction == "transfer" &&
		t.Amount > 0 &&
		t.Message == "" {
			return true
	} else if t.TypeOfTransaction == "message" &&
		t.Message != "" &&
		t.Amount == 0 && (t.ReceiverAddress != "0" && t.SenderAddress != "0") {
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
func newTransactionInstance(_senderAddress string, _receiverAddress string,
	_typeOfTransaction string, _amount float64, _message string,
	_nonce uint, _private_key string) Transaction {

	var t Transaction
	t.SenderAddress = _senderAddress
	t.ReceiverAddress = _receiverAddress
	t.TypeOfTransaction = _typeOfTransaction
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
	concat := t.SenderAddress + t.ReceiverAddress + t.TypeOfTransaction + strconv.FormatFloat(t.Amount, 'f', -1, 64) + strconv.Itoa(int(t.Nonce)) + t.Message
	return concat
}

//Returns hash of the string concatenation of a transaction
func (t *Transaction) GetHash() ([32]byte, error) {
	concat := []byte(t.GetConcat())
	hash := sha256.Sum256(concat)

	return hash, nil
}


//Sets the `TransactionId`
func (t *Transaction) SetHash(_hash [32]byte) (string, error) {
	t.TransactionId = hex.EncodeToString(_hash[:])
	return t.TransactionId, nil
}


//Signs the transaction with the given private key
func (t *Transaction) Sign(_privateKey string) (string, error) {
	transactionHash, err := t.GetHash()
	if err != nil {
		return "", err
	}
	_, err = t.SetHash(transactionHash)
	if err != nil {
		return "", err
	}

	privateKeyArray, err := hex.DecodeString(_privateKey)
	if err != nil {
		return "", err
	}

	privateKey, err := x509.ParsePKCS1PrivateKey(privateKeyArray)
	if err != nil {
		return "", err
	}

	signature_bytes, err := rsa.SignPKCS1v15(rand.Reader, privateKey, crypto.SHA256, transactionHash[:])
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
	producedHash, err := t.GetHash()
	if err != nil {
		return false
	}

	signature, err := hex.DecodeString(t.Signature)
	if err != nil {
		return false
	}

	var publicKeyString string

	if t.SenderAddress == "0" {
		//If sender is genesis block, then t is signed with bootstrap node keys
		publicKeyString = t.ReceiverAddress
	} else {
		publicKeyString = t.SenderAddress
	}

	publicKeyArray, err := hex.DecodeString(publicKeyString)
	if err != nil {
		return false
	}

	publicKey, err := x509.ParsePKCS1PublicKey(publicKeyArray[:])
	if err != nil {
		return false
	}

	err = rsa.VerifyPKCS1v15(publicKey, crypto.SHA256, producedHash[:], signature[:])

	return err == nil
}



// var node only used here:
//Calculates the fee based on the type of the transaction
func (t *Transaction) CalcFee() float64 {
	if t.SenderAddress == "0" || t.ReceiverAddress == "0" {
		return 0
	} else if (t.TypeOfTransaction =="transfer") {
		return t.Amount*feePercentage
	} else if (t.TypeOfTransaction =="message"){
		return float64(len(t.Message)*costPerChar)
	}
	return 0
}

func (node *nodeConfig) NewMessageTransaction(_receiverAddress string, _message string) Transaction{
	node.outboundNonce++
	return newTransactionInstance(node.publicKey, _receiverAddress, "message", 0, _message, node.outboundNonce, node.privateKey)
}
func (node *nodeConfig) NewTransferTransaction(_receiverAddress string, _amount float64) Transaction{
	node.outboundNonce++
	return newTransactionInstance(node.publicKey, _receiverAddress, "transfer", _amount, "", node.outboundNonce, node.privateKey)
}