package blockchat

import (
	"encoding/json"
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

func (node *nodeConfig) IsBlockchainValid() bool {
	previousHash := genesisHash	
	boolState := true
	i := 0

	// Checks each block in the blockchain
	for _, block := range node.blockchain {
		// Checks if block is valid and has the expected index
		boolState = boolState && block.IsValid(previousHash, &node.nodeMap) && block.Index == i
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
	return boolState
}