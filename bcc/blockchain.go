package bcc

import (
	"encoding/json"
	"errors"
	"fmt"
	"os"
)

type Blockchain []Block

func LoadBlockchain() (Blockchain, error) {
	content, err := os.ReadFile(BLOCKCHAIN_PATH)
	if err != nil {
		fmt.Println(err)
		return nil, err
	}
	var blockchain Blockchain
	if err := json.Unmarshal(content, &blockchain); err != nil {
		fmt.Println(err)
		return nil, err
	}
	_, isValid := blockchain.IsValid()
	if isValid {
		return blockchain, nil
	}
	return blockchain, errors.New("Blockchain not valid")

}

func (B *Blockchain) WriteBlockchain() error {
	return B.WriteBlockchainAt(BLOCKCHAIN_PATH)
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
	prev_hash := GENESIS_HASH
	i := 0
	for _, block := range *B {
		bool_state = bool_state && block.Previous_hash == prev_hash && block.IsValid() && block.Index == i
		prev_hash = block.Current_hash
		i++
		if !bool_state {
			i--
			break
		}
	}
	return i, bool_state
}

func (B *Blockchain) MakeDB() (DBmap, error) {
	index, _ := B.IsValid()

	var dbmap DBmap = make(DBmap)
	fee := float64(0)
	for i := 0; i < index; i++ {
		block := (*B)[i]
		for _, tx := range block.Transactions {
			tmp, _ := dbmap.addTransaction(&tx)
			fee += tmp
		}
		validator := block.CalcValidator()
		
		tmp := NodeIDArray[validator]
		if tmp == "" {
			tmp = NodeIDArray[0]
		}
		fmt.Println(tmp)
		fmt.Println(validator)
		dbmap.changeBalance(tmp, fee)

	}
	return dbmap, nil
}

func (B *Blockchain) AddBlock(block *Block) error {

	if block.Index == len(*B) {
		*B = append(*B, *block)
		return nil
	}

	return errors.New("Block not valid")
}
