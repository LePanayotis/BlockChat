package bcc

import (
	"bufio"
	"errors"
	"fmt"
	"os"
	"strconv"
	"strings"
	//"github.com/spf13/cobra"
)

var userPubKey, userPrivKey string

func StartCLI() {

	userPubKey, userPrivKey = MyPublicKey, MyPrivateKey
	scanner := bufio.NewScanner(os.Stdin)

	fmt.Println("Welcome to the CLI, type exit to exit")
	fmt.Printf("This node's public key is: %s\n", MyPublicKey)
	for {
		fmt.Print("> ")
		scanner.Scan()
		command := scanner.Text()
		parts := strings.Fields(command)
		length := len(parts)
		if length == 0 {
			continue
		}
		switch parts[0] {
		case "q":
			fmt.Println("Terminating node...")
			closeKafka()
			os.Exit(0)
		case "exit":
			fmt.Println("Terminating node...")
			closeKafka()
			os.Exit(0)
		case "help":
			fmt.Println("MAY GOD HELP YOU")
		case "use-wallet":
			fmt.Print("Insert Public Key: ")
			fmt.Scanln(&userPubKey)
			fmt.Print("Insert Private Key: ")
			fmt.Scanln(&userPrivKey)
		case "use-node-wallet":
			userPubKey = MyPublicKey
			userPrivKey = MyPrivateKey
			fmt.Printf("Using node's wallet,\nPublic key: %s\n", MyPublicKey)
		case "generate-wallet":
			pub, priv := GenerateKeys()
			fmt.Printf("Key pair generated:\nPublic key: %s\nPrivate key: %s\n", pub, priv)
			if parts[length-1] == "-u" {
				fmt.Println("Using these keys")
				userPrivKey = priv
				userPubKey = pub
			}
		case "balance":
			account := userPubKey
			if length == 2 {
				account = parts[1]
			}
			fmt.Printf("Account balance: %f\n", ValidDB.GetBalance(account))
		case "stake":
			MakeStake(parts, MyPublicKey, MyPrivateKey)
		case "t":
			err := InterpretTransaction(parts, userPubKey, userPrivKey)
			if err != nil {
				fmt.Println(err)
			}
		case "view":
			fmt.Println(MyBlockchain[len(MyBlockchain)-1].JSONify())
		default:
			fmt.Println("Provide a valid command, type 'help' for more information")
		}
	}
}

func InterpretTransaction(_command []string, _sender_pub_key string, _sender_priv_key string) error {
	var tx Transaction
	var _receiver string
	var _amount float64
	var _message string = ""
	
	length := len(_command)
	if length < 3 {
		return errors.New("Wrong command format")
	}
	if _command[1] == "-n" {
		if length < 4 {
			return errors.New("Wrong command format")
		}
		receiver_no, err := strconv.Atoi(_command[2])
		_receiver = NodeIDArray[receiver_no]
		if err != nil || receiver_no > NODES-1 {
			return errors.New("NodeID must be an integer from 0 to NODES-1")
		}
		if _command[3] == "-m" {
			for _, v := range _command[4:] {
				_message += (v + " ")
			}
			_message = strings.Trim(_message, " ")
		} else {
			_amount, err = strconv.ParseFloat(_command[3], 64)
			if err != nil {
				return errors.New("Amount must be float")
			}
		}
	} else {
		var err error
		_receiver = _command[1]
		if _command[2] == "-m" {
			for _, v := range _command[3:] {
				_message += (v + " ")
			}
			_message = strings.Trim(_message, " ")
		} else {
			_amount, err = strconv.ParseFloat(_command[3], 64)
			if err != nil  {
				return err
			}
		}
	}
	
	if _message == "" {
		tx = NewTransferTransaction(_sender_pub_key, _receiver, _amount, myNonce, _sender_priv_key)
	} else {
		tx = NewMessageTransaction(_sender_pub_key, _receiver,_message, myNonce, _sender_priv_key)
	}
	if !ValidDB.IsTransactionPossible(&tx) {
		return errors.New("Transaction not possible")
	}
	myNonce++
	SendTransaction(Writer, tx)
	return nil
}

func MakeStake(_command []string, _sender_pub_key string, _sender_priv_key string) error {
	var tx Transaction
	if len(_command) == 2 {
		_amount, err := strconv.ParseFloat(_command[1], 64)
		if err != nil {
			fmt.Println(err)
			return err
		}
		tx = NewTransferTransaction(_sender_pub_key, "0", _amount, myNonce, _sender_priv_key)
		if !ValidDB.IsTransactionPossible(&tx) {
			fmt.Println("Transaction not possible")
			return errors.New("Transaction not possible")

		}
	} else {
		fmt.Println("Transaction command incorrect format")
		return nil
	}
	myNonce++
	SendTransaction(Writer, tx)
	return nil
}
