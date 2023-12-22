package bcc

import (
	"bufio"
	"errors"
	"fmt"
	"os"
	"strconv"
	"strings"
)

func StartCLI() {

	userPubKey, userPrivKey := MyPublicKey, MyPrivateKey
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
		case "exit":
			fmt.Println("Terminating node...")
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
		case "t":
			InterpretTransaction(parts, userPubKey, userPrivKey)
		case "view":
			fmt.Println(MyBlockchain[len(MyBlockchain)-1].JSONify())
		default:
			fmt.Println("Provide a valid command")
		}
	}
}

func InterpretTransaction(_command []string, _sender_pub_key string, _sender_priv_key string) error {
	var tx Transaction
	if len(_command) == 3 {
		_amount, err := strconv.ParseFloat(_command[2], 64)
		if err != nil {
			fmt.Println(err)
			return err
		}
		receiver := _command[1]
		sender_nonce := TempDB.IncreaseNonce(_sender_pub_key)
		tx = NewTransferTransaction(_sender_pub_key, receiver, _amount, sender_nonce, _sender_priv_key)
		if !TempDB.IsTransactionPossible(&tx) {
			fmt.Println("Transaction not possible")
			return errors.New("Transaction not possible")

		}
	} else if len(_command) > 3 && _command[2] == "-m" {
		s := ""
		for _, v := range _command[3:] {
			s += v
		}
		s = strings.Trim(s, " ")
		receiver := _command[1]
		sender_nonce := TempDB.IncreaseNonce(_sender_pub_key)
		tx = NewMessageTransaction(_sender_pub_key, receiver, s, sender_nonce, _sender_priv_key)
		if !TempDB.IsTransactionPossible(&tx) {
			fmt.Println("Transaction not possible")
			return errors.New("Transaction not possible")

		}

	} else {
		fmt.Println("Transaction command incorrect format")
		return nil
	}
	SendTransaction(Writer, tx)
	return nil
}
