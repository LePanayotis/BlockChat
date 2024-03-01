package blockchat

import (
	"bufio"
	"fmt"
	"os"
	"strconv"
	"strings"
	"time"
)

func (node *nodeConfig) sendInputTransactions() error {

	inputFile, err := os.Open(node.inputPath)
	if err != nil {
		logger.Error("Could not open input file","error",err)
		return err
	}
	defer inputFile.Close()

	scanner := bufio.NewScanner(inputFile)
	i := 0
	startTime := time.Now()
	for scanner.Scan() {
		input := scanner.Text()

		input = strings.TrimSpace(input)

		inputSlice := strings.SplitN(input, " ", 2)
		if len(inputSlice) < 2 {
			logger.Warn("Skipping empty line")
			continue
		}

		message := inputSlice[1]
		receiverId, err := strconv.Atoi(inputSlice[0][2:])
		if err != nil {
			logger.Warn("Could not parse node id, moving to next entry", "error", err)
			continue
		}
		if receiverId < 0 || receiverId >= node.nodes {
			continue
		}
		if len(message) == 0 {
			continue
		}
		err = node.postInputTransaction(receiverId, message)
		if err == nil {
			i++
		}

	}
	duration := time.Since(startTime)
	seconds:= duration.Seconds()

	logger.Info(fmt.Sprintf("Successfully sent %d transactions in %fs",i, seconds))
	return nil
}

func (node *nodeConfig) postInputTransaction(_receiverId int, _message string) error {

	receiverAddress := node.idArray[_receiverId]
	tx := node.NewMessageTransaction(receiverAddress, _message)
	err := node.sendTransaction(&tx)
	if err != nil {
		node.logTransaction("Failed sending transaction:", &tx)
		return err
	}
	node.logTransaction("Transaction sent", &tx)
	
	if node.defaultStake < 0 {
		return nil
	} 

	tx = node.NewTransferTransaction("0", node.defaultStake)
	err = node.sendTransaction(&tx)
	if err != nil {
		node.logTransaction("Failed sending transaction:", &tx)
		return err
	}
	node.logTransaction("Stake sent", &tx)
	return nil
}
