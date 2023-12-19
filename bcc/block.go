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
	Validator     int
	Current_hash  string
	Previous_hash string
}

func (b *Block) GetConcat() string {
	s := strconv.Itoa(b.Index) + b.Timestamp + strconv.Itoa(b.Validator) + b.Previous_hash
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
		Validator:     0,
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
		Validator:     0,
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
	if len(b.Transactions) < CAPACITY {
		b.Transactions = append(b.Transactions, tx)
		return len(b.Transactions), nil
	}
	return -1, errors.New("capacity of block reached")
}

func (b *Block) JSONify() (string, error) {

	if !b.IsValid() {
		return "", errors.New("Transaction is not valid, can't be jsonified")
	}

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

func (b *Block) IsValid() bool {
	if b.Index < 0 || b.Validator < 0 || len(b.Transactions) > CAPACITY {
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
		return _current_hash == b.Current_hash
	}
	return false

}

func (b *Block) CorrectPrevHash(_prev_hash string) bool {
	return b.Previous_hash == _prev_hash
}

func (b *Block) CalcValidator() int {
	var NodeStakes [NODES]float64
	for i := range NodeStakes {
		NodeStakes[i] = DEFAULT_STAKE
	}
	for _, v := range b.Transactions {
		if v.Receiver_address == "0" {
			receiver_node := NodeMap[v.Receiver_address]
			NodeStakes[receiver_node] = DEFAULT_STAKE
		}
	}
	steaks_sum := 0.
	for _, v := range NodeStakes {
		steaks_sum += v
	}

	randomGenerator := rand.New(rand.NewSource(stringToSeed(b.Previous_hash)))
	rng := randomGenerator.Float64()
	
	
	temp := 0.
	for i := range NodeStakes {
		NodeStakes[i] += temp
		temp = NodeStakes[i]
		if temp - rng >= 0 {
			fmt.Println("Rng: ", rng, " index: ", i)
			return i
		}
	}
	return 0

}

func stringToSeed(s string) int64 {
	hash := fnv.New64a()
	hash.Write([]byte(s))
	return int64(hash.Sum64())
}