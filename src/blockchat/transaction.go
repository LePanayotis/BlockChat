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
	"fmt"
	"strconv"
)

// The basic struct representing the transactions in BlockChat
// `Sender` and `ReiceiverAddress` are the public keys of
// the sender and the receiver of the transaction respectively,
// whether it is a simple message or a transfer transaction.
// The `Type` can have the values: "message" or "transfer"
// The property `Message` bears the content of the message and the
// property `Amount` represents the amount of BlockChatCoins to be transferred
// `Id` is the hash of the string representation of the transaction
// as defined in GetConcat method.
// `Signature` is the encrypted `Id` with the private key of the sender
type Transaction struct {
	Sender    string  `json:"sender"`
	Receiver  string  `json:"receiver"`
	Type      string  `json:"type"`
	Amount    float64 `json:"amount"`
	Message   string  `json:"message"`
	Nonce     uint    `json:"nonce"`
	Id        string  `json:"id"`
	Signature string  `json:"sig"`
}

// Creates a new message transaction instance
// Sender's private key is required to sign the transaction
func NewMessageTransaction(_senderAddress string, _receiverAddress string, _message string,
	_nonce uint, _privateKey string) Transaction {

	return newTransactionInstance(_senderAddress, _receiverAddress, "message", 0, _message, _nonce, _privateKey)
}

// Creates a new transfer transaction instance
// Sender's private key is required to sign the transaction
func NewTransferTransaction(_senderAddress string, _receiverAddress string, _amount float64,
	_nonce uint, _privateKey string) Transaction {

	return newTransactionInstance(_senderAddress, _receiverAddress, "transfer", _amount, "", _nonce, _privateKey)
}

// Returns the transaction in json string representation
// and error if the parsing fails
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

// Checks if transaction `t` is valid. This means:
// `t`'s `Sender` and `Receiver` are not empty strings
// `Amount` is not negative
// `Type` is either "trasfer" or "message"
func (t *Transaction) IsValid() bool {
	// Checks addresses not empty and amount non negative
	if t.Sender == "" ||
		t.Receiver == "" ||
		t.Amount < 0 {
		return false
	}

	// Transfers have positive amounts and empty string messages
	if t.Type == "transfer" &&
		t.Amount > 0 &&
		t.Message == "" {
		return true
	} else if t.Type == "message" &&
		// Messages have non-empty message content and zero amount, and can't be sent or received by "0" wallet
		t.Message != "" &&
		t.Amount == 0 && (t.Receiver != "0" && t.Sender != "0") {
		return true
	}
	return false
}

// Returns a Transaction instance from a json string
// Error if parsing fails
func ParseTransactionJSON(s string) (Transaction, error) {
	var t Transaction
	if err := json.Unmarshal([]byte(s), &t); err != nil {
		return t, err
	}
	return t, nil

}

// Fundamental function to create a new Transaction instance from the given parameters
func newTransactionInstance(_senderAddress string, _receiverAddress string,
	_Type string, _amount float64, _message string,
	_nonce uint, _privateKey string) Transaction {

	var t Transaction
	// Sets instance's fields
	t.Sender = _senderAddress
	t.Receiver = _receiverAddress
	t.Type = _Type
	t.Amount = _amount
	t.Nonce = _nonce
	t.Message = _message

	// Signs transaction
	_, err := t.Sign(_privateKey)
	if err != nil {
		return Transaction{}
	}
	return t
}

// Returns a concatenation of basic properties of the transactions
func (t *Transaction) GetConcat() string {
	concat := t.Sender + t.Receiver + t.Type + strconv.FormatFloat(t.Amount, 'f', -1, 64) + strconv.Itoa(int(t.Nonce)) + t.Message
	return concat
}

// Returns hash of the string concatenation of a transaction
func (t *Transaction) GetHash() ([32]byte, error) {
	concat := []byte(t.GetConcat())
	hash := sha256.Sum256(concat)

	return hash, nil
}

// Sets the `Id`
func (t *Transaction) SetHash(_hash [32]byte) (string, error) {
	t.Id = hex.EncodeToString(_hash[:])
	return t.Id, nil
}

// Signs the transaction with the given private key
func (t *Transaction) Sign(_privateKey string) (string, error) {
	// Get's transaction's hash
	transactionHash, err := t.GetHash()
	if err != nil {
		return "", err
	}
	// Assigns hash to appropriate fields
	_, err = t.SetHash(transactionHash)
	if err != nil {
		return "", err
	}

	// Decodes private key string to bytes
	privateKeyArray, err := hex.DecodeString(_privateKey)
	if err != nil {
		return "", err
	}

	// Makes key out of private key byte array
	privateKey, err := x509.ParsePKCS1PrivateKey(privateKeyArray)
	if err != nil {
		return "", err
	}

	// Signs hash
	signatureBytes, err := rsa.SignPKCS1v15(rand.Reader, privateKey, crypto.SHA256, transactionHash[:])
	if err != nil {
		return "", err
	}

	// Gets signature string
	signature := hex.EncodeToString(signatureBytes)
	// Assigns signature to field
	t.Signature = signature

	return signature, nil

}

// Verifies transaction:
// Checks if transaction properties values conform to the constraints
// Then verifies signature with the sender's public key
func (t *Transaction) Verify() bool {
	// A Verified transaction must be valid (obey the constraints)
	if !t.IsValid() {
		return false
	}
	// Calculates hash
	producedHash, err := t.GetHash()
	if err != nil {
		return false
	}

	signature, err := hex.DecodeString(t.Signature)
	if err != nil {
		return false
	}

	var publicKeyString string
	if t.Sender == "0" {
		//If sender is genesis block, then t is signed with bootstrap node keys
		publicKeyString = t.Receiver
	} else {
		publicKeyString = t.Sender
	}

	publicKeyArray, err := hex.DecodeString(publicKeyString)
	if err != nil {
		return false
	}
	// Makes public key
	publicKey, err := x509.ParsePKCS1PublicKey(publicKeyArray[:])
	if err != nil {
		return false
	}

	// Verifies that the decrypted signature equals the produced hash
	err = rsa.VerifyPKCS1v15(publicKey, crypto.SHA256, producedHash[:], signature[:])

	return err == nil
}

// Calculates the fee based on the type of the transaction
func (t *Transaction) CalcFee() float64 {
	if t.Sender == "0" || t.Receiver == "0" {
		return 0
	} else if t.Type == "transfer" {
		return t.Amount * feePercentage
	} else if t.Type == "message" {
		return float64(len(t.Message) * costPerChar)
	}
	return 0
}

// Creates message transaction with node's configuration
func (node *nodeConfig) NewMessageTransaction(_receiverAddress string, _message string) Transaction {
	node.outboundNonce++
	return newTransactionInstance(node.publicKey, _receiverAddress, "message", 0, _message, node.outboundNonce, node.privateKey)
}

// Creates transfer transaction with node's configuration
func (node *nodeConfig) NewTransferTransaction(_receiverAddress string, _amount float64) Transaction {
	node.outboundNonce++
	return newTransactionInstance(node.publicKey, _receiverAddress, "transfer", _amount, "", node.outboundNonce, node.privateKey)
}

func (node *nodeConfig) logTransaction(_logInfo string, _tx *Transaction) {

	nonce, amount, message, id := _tx.Nonce, _tx.Amount, _tx.Message, _tx.Id
	var sender, receiver string

	fmt.Print()
	senderId, b := node.nodeMap[_tx.Sender]
	if !b && _tx.Sender != "0" {
		logger.Error("Requested to log transaction with unresoluted sender wallet address owner node", "sender", _tx.Sender)
		return
	}
	sender = strconv.Itoa(senderId)
	receiverId, b := node.nodeMap[_tx.Receiver]
	if !b && _tx.Receiver != "0" {
		logger.Error("Requested to log transaction with unresoluted receiver node wallet addres", "receiver", _tx.Receiver)
		return
	}
	receiver = strconv.Itoa(receiverId)
	if _tx.Sender == "0" {
		logger.Info("Genesis transaction", "to", receiver, "amount", amount, "nonce", nonce, "tid", id)
		return
	}
	if _tx.Receiver == "0" {
		logger.Info(_logInfo, "type", "stake", "from", sender, "amount", amount, "nonce", nonce, "tid", id)
		return
	}
	logger.Info(_logInfo, "type", _tx.Type, "from", sender, "to", receiver, "amount", amount, "message", message, "nonce", nonce, "tid", id)

}
