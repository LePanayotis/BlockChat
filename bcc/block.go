package bcc

import (
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"hash/fnv"
	"math/rand"
	"strconv"
	"time"
)



type Block struct {
	Index         int
	Timestamp     string
	Transactions  []Transaction
	Validator     string
	Current_hash  string
	Previous_hash string
}

func (b *Block) GetConcat() string {
	s := strconv.Itoa(b.Index) + b.Validator + b.Previous_hash
	for _, value := range b.Transactions {
		s = s + value.GetConcat()
	}
	return s
}

func (b *Block) GetHash() ([32]byte, error) {
	concat := []byte(b.GetConcat())
	hash := sha256.Sum256(concat)

	return hash, nil
}

func GenesisBlock(_public_key string, _priv_key string) Block {
	timestamp := time.Now().UTC().Format(TIME_FORMAT)

	t := NewTransferTransaction("0", _public_key, INITIAL_BCC*float64(NODES), 0, _priv_key)

	transactions := []Transaction{
		t,
	}

	b := Block{
		Index:         0,
		Timestamp:     timestamp,
		Transactions:  transactions,
		Validator:     _public_key,
		Previous_hash: "1",
	}

	hash_bytes, _ := b.GetHash()

	b.Current_hash = hex.EncodeToString(hash_bytes[:])
	return b

}

func NewBlock(_index int, _previous_hash string) Block {

	b := Block{
		Index:         _index,
		Transactions:  nil,
		Validator:     "",
		Previous_hash: _previous_hash,
	}
	return b

}

func (b *Block) CalcHash() {
	timestamp := time.Now().UTC().Format(TIME_FORMAT)
	b.Timestamp = timestamp
	hash_bytes, _ := b.GetHash()
	b.Current_hash = hex.EncodeToString(hash_bytes[:])

}

func (b *Block) AddTransaction(tx Transaction) (int, error) {
	if !tx.Verify() {
		return -1, errors.New("transaction verification failed")
	}
	if len(b.Transactions) < CAPACITY {
		b.Transactions = append(b.Transactions, tx)
		return len(b.Transactions), nil
	}
	return -1, errors.New("capacity of block reached")
}

func (b *Block) JSONify() (string, error) {
	jsonStringBytes, err := json.Marshal(b)
	if err != nil {
		return "", err
	}
	jsonString := string(jsonStringBytes)
	return jsonString, nil

}

func ParseBlockJSON(s string) (Block, error) {
	var b Block

	if err := json.Unmarshal([]byte(s), &b); err != nil {
		return b, err
	}
	return b, nil

}


//to-do add check for previous hash
func (b *Block) IsValid(_previous_hash string) bool {
	if b.Index < 0 || b.Validator == "" || len(b.Transactions) > CAPACITY {
		return false
	}
	temp := true
	for _, value := range b.Transactions {
		temp = temp && value.Verify()
		if !temp  {break}
	}
	if temp {
		_current_hash_bytes, _ := b.GetHash()
		_current_hash := hex.EncodeToString(_current_hash_bytes[:])
		temp = _current_hash == b.Current_hash
		temp = temp && b.CalcValidator() == b.Validator
		temp = temp && b.Previous_hash ==_previous_hash
	}
	return temp

}

func (b *Block) SetValidator() {
	b.Validator = b.CalcValidator()
}

func (b *Block) CalcValidator() string {
	var NodeStakes []float64 = make([]float64, NODES)
	var stakes int = 0
	for i := range NodeStakes {
		NodeStakes[i] = 0
	}
	for _, v := range b.Transactions {
		if v.Receiver_address == "0" {
			stakes++
			receiver_node := NodeMap[v.Sender_address]
			NodeStakes[receiver_node] = v.Amount
		}
	}
	//If no stakes have been made in the block, validator is node 0
	if stakes == 0 {
		return NodeIDArray[0]
	}
	
	//Calculate staking node
	steaks_sum := 0.
	for _, v := range NodeStakes {
		steaks_sum += v
	}

	randomGenerator := rand.New(rand.NewSource(stringToSeed(b.Previous_hash)))
	lucky := randomGenerator.Float64()*steaks_sum
	fmt.Println(lucky)

	temp := 0.
	for i := 0; i < NODES; i++ {
		NodeStakes[i] += temp
		fmt.Println(NodeStakes[i])
		temp = NodeStakes[i]
		if lucky < temp {
			return NodeIDArray[i]
		}
	}

	//if nothing goes right, validator is node 0
	return NodeIDArray[0]

}

func stringToSeed(s string) int64 {
	hash := fnv.New64a()
	hash.Write([]byte(s))
	return int64(hash.Sum64())
}