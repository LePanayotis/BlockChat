package blockchat

import (
	"encoding/json"
	"errors"
	"os"
)

type Blockchain []Block

func (node *nodeConfig) LoadBlockchain() (Blockchain, error) {
	content, err := os.ReadFile(node.blockchainPath)
	if err != nil {
		return nil, err
	}
	var blockchain Blockchain
	if err := json.Unmarshal(content, &blockchain); err != nil {
		return nil, err
	}
	_, isValid := node.IsBlockchainValid()
	if isValid {
		return blockchain, nil
	}
	return blockchain, errors.New("Blockchain not valid")

}

func (B *Blockchain) WriteBlockchain(path string) error {

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




func (node *nodeConfig) WriteBlockchain() error {
	err := node.blockchain.WriteBlockchain(node.blockchainPath)
	if err != nil {
		return err
	}
	return nil

}

func  (B *Blockchain) IsValid(previousHash string)(int, bool) {
	bool_state := true
	i := 0
	for _, block := range *B {
		bool_state = bool_state && block.IsValid(previousHash) && block.Index == i
		previousHash = block.Current_hash
		i++
		if !bool_state {
			i--
			break
		}
	}
	return i, bool_state
}

func  (node * nodeConfig) IsBlockchainValid()(int, bool) {
	return node.blockchain.IsValid(node.genesisHash)
}

// Need to change this
func (node * nodeConfig) MakeDB()  error {
	index, _ := node.IsBlockchainValid()
	B := &node.blockchain
	node.myDB = make(DBmap)
	for i := 0; i < index; i++ {
		block := (*B)[i]
		err := node.addBlockToDB(&block)
		if err != nil {
			return err
		}
	}
	return nil
}

// Appends valid block to blockchain
func (B *Blockchain) AddBlock(block *Block) error {
	if block.IsValid((*B)[len(*B)-1].Current_hash) && block.Index == len(*B) {
		*B = append(*B, *block)
		return nil
	}
	return errors.New("Block not valid")
}
