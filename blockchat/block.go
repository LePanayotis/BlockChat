package blockchat

import (
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"hash/fnv"
	"math/rand"
	"strconv"
	"time"
)

// Basic struct that represents the blocks of the Blockchain
type Block struct {
	//The increasing index of the block in the blockchain
	Index int `json:"index"`
	//Creation publication timestamp
	Timestamp string `json:"timestamp"`
	//Array with transactions in the blockchain
	Transactions []Transaction `json:"transactions"`
	//The public key of the block validator
	Validator int `json:"validator"`
	//Hash produced from the result of GetConcat method
	CurrentHash string `json:"current_hash"`
	//The hash of the previous block in the blockchain
	PreviousHash string `json:"previous_hash"`
}

// Returns concatenation of key properties of the block
// TODO: use transaction hash instead of concatenation
func (b *Block) GetConcat() string {
	s := strconv.Itoa(b.Index) + strconv.Itoa(b.Validator) + b.PreviousHash
	for _, value := range b.Transactions {
		s = s + value.GetConcat()
	}
	return s
}

// Returns hash256 of the above concatenation
func (b *Block) GetHash() ([32]byte) {
	concat := []byte(b.GetConcat())
	hash := sha256.Sum256(concat)
	return hash
}

// Creates the instance of the Genesis Block,
// the first block of the blockchain with only one transaction
// to the bootstrap node.
func (node *nodeConfig) GenesisBlock() Block {
	//Timestamp in UTC in the format indicated in the timeFormat
	timestamp := time.Now().UTC().Format(timeFormat)

	//The only transaction of the block granting initialBCC*(#number of nodes)
	t := NewTransferTransaction("0", node.publicKey, initialBCC*float64(node.nodes), 0, node.privateKey)

	transactions := []Transaction{
		t,
	}

	// Block instance containing only t transaction
	b := Block{
		Index:         0,
		Timestamp:     timestamp,
		Transactions:  transactions,
		Validator:     0,
		PreviousHash: "1",
	}

	// Sets the hash
	hashBytes := b.GetHash()
	b.CurrentHash = hex.EncodeToString(hashBytes[:])
	// Returns genesis block
	return b
}

// Creates new block with input parameters its index and the hash of the previous block
func (node *nodeConfig) NewBlock() Block {

	// Checks node's blockchain length
	// In case of 0, returns empty default block
	length := len(node.blockchain)
	if length == 0 {
		return *new(Block)
	}
	// Creates new block instance
	// Its index is the length of the blockchain
	// Previoush hash is the hash of the previous block
	b := Block{
		Index:         length,
		Transactions:  nil,
		Validator:     -1,
		PreviousHash: node.blockchain[length-1].CurrentHash,
	}
	return b
}



// Calculates and sets the hash of the block
func (b *Block) CalcHash() {
	hashBytes := b.GetHash()
	b.CurrentHash = hex.EncodeToString(hashBytes[:])
}

// Appends transaction to the transaction list
func (b *Block) AddTransaction(_tx *Transaction) int{
	b.Transactions = append(b.Transactions, *_tx)
	return len(b.Transactions)
}

// Return the string of the JSON representation of the block
func (b *Block) JSONify() (string, error) {
	jsonStringBytes, err := json.Marshal(b)
	if err != nil {
		return "", err
	}
	jsonString := string(jsonStringBytes)
	return jsonString, nil
}

// Creates block instance from its JSON representation
func ParseBlockJSON(s string) (Block, error) {
	var b Block
	if err := json.Unmarshal([]byte(s), &b); err != nil {
		return b, err
	}
	return b, nil

}

// Checks whether is valid or not, provided the hash of the previous block
func (node *nodeConfig) IsBlockValid(_b *Block, _previousHash string) bool {
	// Block index and validator must be greater than zero
	if _b.Index < 0 || _b.Validator < 0 {
		return false
	}

	// Auxiliary boolean variable
	temp := true
	// Check all transactions are verified
	for _, value := range _b.Transactions {
		temp = temp && value.Verify()
		if !temp {
			break
		}
	}
	
	// No need to check hash and validator if contains non verified transactions
	if temp {
		// Gets end encodes hash to string
		currenHashBytes:= _b.GetHash()
		currentHash := hex.EncodeToString(currenHashBytes[:])
		
		// Checks hash matches contained hash in block
		temp = currentHash == _b.CurrentHash
		// Checks validator is expected
		temp = temp && node.CalcValidator(_b) == _b.Validator
		// Checks previous hash matches parameter
		temp = temp && _b.PreviousHash == _previousHash
	}
	// If all validity criteria are met returns true
	return temp
}

// First calculates then sets the validator of the block
func (node *nodeConfig) SetValidator(b *Block) {
	b.Validator = node.CalcValidator(b)
}

// Calculates the validator of the current block based on the stakes
// A stake is a transfer to the "0" wallet
func (node *nodeConfig) CalcValidator(b *Block) int {
	// Array to store each node's stakes
	var nodeStakes []float64 = make([]float64, node.nodes)
	var stakes int = 0
	// Initialise array entries to 0
	for i := range nodeStakes {
		nodeStakes[i] = 0
	}

	// Parse block's transactions
	for _, v := range b.Transactions {
		// If transaction is a stake (receiver wallet "0")
		if v.ReceiverAddress == "0" {
			// Increases stakes number
			stakes++
			// Maps sender address to node id
			receiverNode := node.nodeMap[v.SenderAddress]
			// Increases stake amount of the node
			nodeStakes[receiverNode] += v.Amount
		}
	}
	// If no stakes have been made in the block, validator is node 0
	if stakes == 0 {
		return 0
	}

	// Calculate sum of stake
	steaksSum := 0.
	for _, v := range nodeStakes {
		steaksSum += v
	}

	// Create RNG with the block's previous hash as seed
	randomGenerator := rand.New(rand.NewSource(stringToSeed(b.PreviousHash)))
	
	// Gets the lucky coin
	lucky := randomGenerator.Float64() * steaksSum
	
	// Finds which node has the lucky coin
	temp := 0.
	for i := 0; i < node.nodes; i++ {
		nodeStakes[i] += temp
		temp = nodeStakes[i]
		if lucky < temp {
			return i
		}
	}

	// If nothing goes right, validator is node 0
	return 0

}

// Converts a block's hash into a seed for an RNG algorithm used in CalcValidator()
func stringToSeed(_s string) int64 {
	hash := fnv.New64a()
	_ , err :=  hash.Write([]byte(_s))
	if err != nil {
		return int64(0)
	}
	return int64(hash.Sum64())
}
