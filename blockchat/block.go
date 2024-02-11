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
	Current_hash string `json:"current_hash"`
	//The hash of the previous block in the blockchain
	Previous_hash string `json:"previous_hash"`
}

// Returns concatenation of key properties of the block
// TODO: use transaction hash instead of concatenation
func (b *Block) GetConcat() string {
	s := strconv.Itoa(b.Index) + strconv.Itoa(b.Validator) + b.Previous_hash
	for _, value := range b.Transactions {
		s = s + value.GetConcat()
	}
	return s
}

// Returns hash256 of the above concatenation
func (b *Block) GetHash() ([32]byte, error) {
	concat := []byte(b.GetConcat())
	hash := sha256.Sum256(concat)
	return hash, nil
}

// Creates the instance of the Genesis Block,
// the first block of the blockchain with only one transaction
// to the bootstrap node.
// `_public_key` and `_private_key` are the ones of the bootstrap node
func (node *nodeConfig) GenesisBlock() Block {
	//Timestamp in UTC in the format indicated in the TIME_FORMAT
	timestamp := time.Now().UTC().Format(node.timeFormat)
	_public_key , _priv_key := node.publicKey, node.privateKey
	node.blockIndex = 0

	//The only transaction of the block granting INITIAL_BCC*(#number of nodes)
	t := NewTransferTransaction("0", _public_key, node.initialBCC*float64(node.nodes), 0, _priv_key)

	transactions := []Transaction{
		t,
	}

	b := Block{
		Index:         0,
		Timestamp:     timestamp,
		Transactions:  transactions,
		Validator:     0,
		Previous_hash: "1",
	}

	//produces and sets the hash
	hash_bytes, _ := b.GetHash()
	b.Current_hash = hex.EncodeToString(hash_bytes[:])

	return b
}

// Creates new block with input parameters its index and the hash of the previous block
func (node *nodeConfig) NewBlock() Block {
	//revise
	if len(node.blockchain) == 0 {
		return *new(Block)
	}


	b := Block{
		Index:         node.blockIndex+1,
		Transactions:  nil,
		Validator:     -1,
		Previous_hash: node.blockchain[node.blockIndex].Current_hash,
	}
	node.blockIndex++
	return b
}



// Calculates and sets the hash of the block
func (b *Block) CalcHash() {
	hash_bytes, _ := b.GetHash()
	b.Current_hash = hex.EncodeToString(hash_bytes[:])
}

// Appends transaction to the transaction list
func (b *Block) AddTransaction(tx *Transaction) int{
	b.Transactions = append(b.Transactions, *tx)
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
func (b *Block) IsValid(_previous_hash string) bool {
	if b.Index < 0 || b.Validator < 0 {
		return false
	}
	temp := true
	for _, value := range b.Transactions {
		temp = temp && value.Verify()
		if !temp {
			break
		}
	}
	if temp {
		_current_hash_bytes, _ := b.GetHash()
		_current_hash := hex.EncodeToString(_current_hash_bytes[:])
		temp = _current_hash == b.Current_hash
		temp = temp && b.CalcValidator() == b.Validator
		temp = temp && b.Previous_hash == _previous_hash
	}
	return temp

}

// First calculates then sets the validator of the block
func (b *Block) SetValidator() {
	b.Validator = b.CalcValidator()
}

// Calculates the validator of the current block based on the stakes
// A stake is a transfer to the "0" wallet
func (b *Block) CalcValidator() int {
	var NodeStakes []float64 = make([]float64, node.nodes)
	var stakes int = 0
	for i := range NodeStakes {
		NodeStakes[i] = 0
	}
	for _, v := range b.Transactions {
		if v.Receiver_address == "0" {
			stakes++
			receiver_node := node.nodeMap[v.Sender_address]
			NodeStakes[receiver_node] = v.Amount
		}
	}
	//If no stakes have been made in the block, validator is node 0
	if stakes == 0 {
		return 0
	}

	//Calculate staking node
	steaks_sum := 0.
	for _, v := range NodeStakes {
		steaks_sum += v
	}

	randomGenerator := rand.New(rand.NewSource(stringToSeed(b.Previous_hash)))
	lucky := randomGenerator.Float64() * steaks_sum
	temp := 0.
	for i := 0; i < node.nodes; i++ {
		NodeStakes[i] += temp
		temp = NodeStakes[i]
		if lucky < temp {
			return i
			//return node.nodeIdArray[i]
		}
	}

	//if nothing goes right, validator is node 0
	return 0

}

// Converts a block's hash into a seed for an RNG algorithm used in CalcValidator()
func stringToSeed(s string) int64 {
	hash := fnv.New64a()
	hash.Write([]byte(s))
	return int64(hash.Sum64())
}
