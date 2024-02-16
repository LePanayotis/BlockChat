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
	Index int `json:"i"`
	//Creation publication timestamp
	Timestamp string `json:"t"`
	//Array with transactions in the blockchain
	Transactions []Transaction `json:"tx"`
	//The public key of the block validator
	Validator int `json:"v"`
	//Hash produced from the result of GetConcat method
	CurrentHash string `json:"h"`
	//The hash of the previous block in the blockchain
	PreviousHash string `json:"p"`
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
func (b *Block) GetHash() [32]byte {
	concat := []byte(b.GetConcat())
	hash := sha256.Sum256(concat)
	return hash
}

// Calculates and sets the hash of the block
func (b *Block) CalcHash() {
	hashBytes := b.GetHash()
	b.CurrentHash = hex.EncodeToString(hashBytes[:])
}

// Appends transaction to the transaction list
func (b *Block) AppendTransaction(_tx *Transaction) int {
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

func (b *Block) IsValid(_previousHash string, _adressMap *map[string]int) bool {
	// Auxiliary boolean variable
	temp := true
	// Check all transactions are verified
	for _, value := range b.Transactions {
		temp = temp && value.Verify()
		if !temp {
			return false
		}
	}
	// Checks previous hash matches parameter
	if _previousHash != b.PreviousHash {
		return false
	}
	// Gets end encodes hash to string
	currenHashBytes := b.GetHash()
	currentHash := hex.EncodeToString(currenHashBytes[:])
	// Checks hash matches contained hash in block
	if currentHash != b.CurrentHash {
		return false
	}
	// Checks validator is expected
	return b.CalcValidator(_adressMap) == b.Validator

}

func (b* Block) CalcValidator(_adressMap *map[string]int) int {

	nodes := len(*_adressMap)

	var nodeStakes []float64 = make([]float64, nodes)
	var stakes int = 0
	// Initialise array entries to 0
	for i := range nodeStakes {
		nodeStakes[i] = 0
	}

	// Parse block's transactions
	for _, v := range b.Transactions {
		// If transaction is a stake (receiver wallet "0")
		if v.Receiver == "0" {
			// Increases stakes number
			stakes++
			// Maps sender address to node id
			receiverNode := (*_adressMap)[v.Sender]
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
	for i := 0; i < nodes; i++ {
		nodeStakes[i] += temp
		temp = nodeStakes[i]
		if lucky < temp {
			return i
		}
	}
	// If nothing goes right, validator is node 0
	return 0
}

func (b *Block) SetValidator(_adressMap *map[string]int) int{
	b.Validator = b.CalcValidator(_adressMap)
	return b.Validator
}

// Converts a block's hash into a seed for an RNG algorithm used in CalcValidator()
func stringToSeed(_s string) int64 {
	hash := fnv.New64a()
	_, err := hash.Write([]byte(_s))
	if err != nil {
		return int64(0)
	}
	return int64(hash.Sum64())
}

// Creates the instance of the Genesis Block,
// the first block of the blockchain with only one transaction
// to the bootstrap node.
func GenesisBlock(_publicKey string, _privateKey string, _nodes int) Block {
	timestamp := time.Now().UTC().Format(timeFormat)

	//The only transaction of the block granting initialBCC*(#number of nodes)
	t := NewTransferTransaction("0", _publicKey, initialBCC*float64(_nodes), 0, _privateKey)

	transactions := []Transaction{
		t,
	}

	// Block instance containing only t transaction
	b := Block{
		Index:        0,
		Timestamp:    timestamp,
		Transactions: transactions,
		Validator:    0,
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
		Index:        length,
		Transactions: []Transaction{},
		Validator:    -1,
		PreviousHash: node.blockchain[length-1].CurrentHash,
	}
	return b
}
