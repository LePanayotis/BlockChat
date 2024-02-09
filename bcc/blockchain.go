package bcc

import (
	"encoding/json"
	"errors"
	"os"
)

type Blockchain []Block

func (B *Blockchain) GetBlockchainBytes() []byte {
	byteArray, err := json.Marshal(B)
	if err != nil {
		return nil
	}
	return byteArray
}

func LoadBlockchain() (Blockchain, error) {
	content, err := os.ReadFile(node.blockchainPath)
	if err != nil {
		return nil, err
	}
	var blockchain Blockchain
	if err := json.Unmarshal(content, &blockchain); err != nil {
		return nil, err
	}
	_, isValid := blockchain.IsValid()
	if isValid {
		return blockchain, nil
	}
	return blockchain, errors.New("Blockchain not valid")

}

func (B *Blockchain) WriteBlockchain() error {
	return B.WriteBlockchainAt(node.blockchainPath)
}

func (B *Blockchain) WriteBlockchainAt(path string) error {

	file, err := os.Create(path)
	if err != nil {
		return err
	}
	byteArray, err := json.Marshal(B)
	if err != nil {
		return err
	}

	_, err = file.Write(byteArray)
	if err != nil {
		return err
	}
	return nil
}

func (B *Blockchain) IsValid() (int, bool) {
	bool_state := true
	prev_hash := node.genesisHash
	i := 0
	for _, block := range *B {
		bool_state = bool_state && block.IsValid(prev_hash) && block.Index == i
		prev_hash = block.Current_hash
		i++
		if !bool_state {
			i--
			break
		}
	}
	return i, bool_state
}

// Need to change this
func (B *Blockchain) MakeDB() (DBmap, error) {
	index, _ := B.IsValid()
	var dbmap DBmap = make(DBmap)
	for i := 0; i < index; i++ {
		block := (*B)[i]
		dbmap.addBlock(&block)
	}
	return dbmap, nil
}

// Appends valid block to blockchain
func (B *Blockchain) AddBlock(block *Block) error {
	if block.IsValid((*B)[len(*B)-1].Current_hash) && block.Index == len(*B) {
		*B = append(*B, *block)
		return nil
	}
	return errors.New("Block not valid")
}
