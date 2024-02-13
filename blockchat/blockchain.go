package blockchat

import (
	"encoding/json"
	"errors"
	"os"
)

// The Blockchain is just a chain of blocks:
// We name it blockchain for convenience
type Blockchain []Block

// Method that writes the blockchain to the indicated path
func (B *Blockchain) WriteBlockchain(_path string) error {

	file, err := os.Create(_path)
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

// Writes blockchain to file indicated in node configuration
func (node *nodeConfig) WriteBlockchain() error {
	return node.blockchain.WriteBlockchain(node.blockchainPath)
}

func (node *nodeConfig) IsBlockchainValid() (int, bool) {
	previousHash := genesisHash	
	boolState := true
	i := 0

	// Checks each block in the blockchain
	for _, block := range node.blockchain {
		// Checks if block is valid and has the expected index
		boolState = boolState && node.IsBlockValid(&block, previousHash) && block.Index == i
		// New previous hash
		previousHash = block.CurrentHash

		// Increases expected index
		i++

		// If block not valid break and set i to index of last valid block
		if !boolState {
			i--
			break
		}
	}
	// Returns validity and last valid block in blockchain
	return i, boolState
}

// Appends valid block to blockchain
func (node *nodeConfig) addBlock(_block *Block) error {

	B := &node.blockchain

	// If blockchain empty and block index equals 0, append block. Check its own previous hash
	if len(*B) == 0 && _block.Index == 0 && node.IsBlockValid(_block, _block.PreviousHash) {
		*B = append(*B, *_block)
		return nil

	// If blockchain not empty and block index equals length of blockchain, append block
	} else if node.IsBlockValid(_block, (*B)[len(*B)-1].CurrentHash) && _block.Index == len(*B) {
		*B = append(*B, *_block)
		return nil
	}
	return errors.New("Block not valid")
}
