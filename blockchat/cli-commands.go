package blockchat

import (
	"fmt"
	"github.com/spf13/cobra"
	"net/rpc"
	"strconv"
)

var isMessage bool
var messagePayload string

var transactionCmd = &cobra.Command{
	Use:     "transaction",
	Aliases: []string{"t"},
	Short:   "Sends new transaction",
	Long:    "Sends new transaction",
	Args:    cobra.RangeArgs(1,2),
	Run: func(cmd *cobra.Command, args []string) {
		client, err := rpc.Dial(node.protocol, node.socket)
		if err != nil {
			fmt.Println("Error connecting to server:", err)
			return
		}
		defer client.Close()
		//var params int = len(args)
		var amount float64 = 0

		receiverId, err := strconv.Atoi(args[0])
		if err != nil {
			fmt.Println("Error parsing recipient node id", err)
			return
		}
		isMessage = messagePayload != ""
		if !isMessage && len(args) == 2{
			amount, err = strconv.ParseFloat(args[1],64)
			if err != nil {
				fmt.Println("Error parsing amount", err)
				return
			}
		}
		tx := &TransactionArgs{
			Receiver_node: receiverId,
			Amount:           amount,
			Message:          messagePayload,
			IsMessage:        isMessage,
		}
		var reply error
		err = client.Call("RPC.Create_transaction", tx, &reply)
		if err != nil {
			fmt.Println("Error calling balance:", err)
			return
		}

		fmt.Println("Transaction send")
	},
}

var balanceCmd = &cobra.Command{
	Use:   "balance",
	Args:  cobra.NoArgs,
	Short: "Returns the current balance in BlockChatCoins",
	Long:  "Returns the current balance of the specified account in BlockChatCoins Default account is the current node",
	Run: func(cmd *cobra.Command, args []string) {
		client, err := rpc.Dial(node.protocol, node.socket)
		if err != nil {
			fmt.Println("Error connecting to server:", err)
			return
		}
		defer client.Close()
		var reply float64
		err = client.Call("RPC.Balance", struct{}{}, &reply)
		if err != nil {
			fmt.Println("Error calling balance:", err)
			return
		}

		fmt.Printf("Balance: %f BCCs\\\\n", reply)
	},
}



var printWalletCmd = &cobra.Command{
	Use:     "print-wallet",
	Aliases: []string{"pcw"},
	Short:   "Returns the current wallet used by the daemon",
	Run: func(cmd *cobra.Command, args []string) {
		client, err := rpc.Dial(node.protocol, node.socket)
		if err != nil {
			fmt.Println("Error connecting to server:", err)
			return
		}
		defer client.Close()
		var reply string
		err = client.Call("RPC.PrintWallet", struct{}{}, &reply)
		if err != nil {
			fmt.Println("Error executing command:", err)
			return
		}
		fmt.Println("The current wallet is:\\\\n", reply)

	},
}

var getNonce = &cobra.Command{
	Use:     "nonce",
	Short:   "Returns the nonce wallet used by the daemon",
	Run: func(cmd *cobra.Command, args []string) {
		client, err := rpc.Dial(node.protocol, node.socket)
		if err != nil {
			fmt.Println("Error connecting to server:", err)
			return
		}
		defer client.Close()
		var reply uint
		err = client.Call("RPC.GetNonce", struct{}{}, &reply)
		if err != nil {
			fmt.Println("Error executing command:", err)
			return
		}
		fmt.Printf("The current nonce is: %d\\\\n", reply)

	},
}


var stakeCmd = &cobra.Command{
	Use:   "stake",
	Args:  cobra.ExactArgs(1),
	Short: "Stakes ammount",
	Run: func(cmd *cobra.Command, args []string) {
		client, err := rpc.Dial(node.protocol, node.socket)
		if err != nil {
			fmt.Println("Error connecting to server:", err)
			return
		}
		defer client.Close()
		var amount float64 = 0
		if !isMessage {
			amount, err = strconv.ParseFloat(args[0], 64)
			if err != nil {
				fmt.Println("Provide float64 amount. Decimal delimiter is '.'")
			}
		}
		tx := &TransactionArgs{
			Receiver_node: -1,
			Amount:           amount,
			Message:          "",
			IsMessage:        false,
		}
		var reply error
		err = client.Call("RPC.Stake", &tx, &reply)
		if err != nil || reply != nil {
			fmt.Println("Error calling balance:", err)
			return
		}

		fmt.Println("Stake send")
	},
}

// to add
var showBlockchain = &cobra.Command{
	Use:     "show-blockchain",
	Args:    cobra.NoArgs,
	Aliases: []string{"blockchain", "bc"},
	Short:   "Shows blockchain in json format",
	Run: func(cmd *cobra.Command, args []string) {
		fmt.Println("Your blockchain")
	},
}

var stopCmd = &cobra.Command{
	Use:   "stop",
	Args:  cobra.NoArgs,
	Short: "Stopping running process",
	Run: func(cmd *cobra.Command, args []string) {
		client, err := rpc.Dial(node.protocol, node.socket)
		if err != nil {
			fmt.Println("Error connecting to server:", err)
			return
		}
		defer client.Close()
		var reply error
		err = client.Call("RPC.Stop", struct{}{}, &reply)
		if err != nil {
			fmt.Println("Error executing command:", err)
			return
		}
		fmt.Println("Node stopped successfully")
	},
}

var startCmd = &cobra.Command{
	Use:   "start",
	Short: "Starts the node",
	Args:  cobra.NoArgs,
	Long:  "Starts the node based on the configuration at the .env file",
	Run: func(cmd *cobra.Command, args []string) {
		StartNode()
	},
}


var commandSet []*cobra.Command = []*cobra.Command{
	startCmd,
	balanceCmd,
	showBlockchain,
	stopCmd,
	stakeCmd,
	transactionCmd,
	printWalletCmd,
	getNonce,
}

var RootCmd = &cobra.Command{
	Use:   "blockchat",
	Short: "BlockChat is a simple CLI application",
	Long:  `BlockChat is a simple CLI application built using Cobra.`,
	Run: func(cmd *cobra.Command, args []string) {
		fmt.Println(" ______  __      ______  ______  __  __   ______  __  __  ______  ______  \n/\\  == \\/\\ \\    /\\  __ \\/\\  ___\\/\\ \\/ /  /\\  ___\\/\\ \\_\\ \\/\\  __ \\/\\__  _\\ \n\\ \\  __<\\ \\ \\___\\ \\ \\/\\ \\ \\ \\___\\ \\  _\"-.\\ \\ \\___\\ \\  __ \\ \\  __ \\/_/\\ \\/ \n \\ \\_____\\ \\_____\\ \\_____\\ \\_____\\ \\_\\ \\_\\\\ \\_____\\ \\_\\ \\_\\ \\_\\ \\_\\ \\ \\_\\ \n  \\/_____/\\/_____/\\/_____/\\/_____/\\/_/\\/_/ \\/_____/\\/_/\\/_/\\/_/\\/_/  \\/_/")
	},
}

func ConfigApp() {
	node.EnvironmentConfig()

	startCmd.Flags().IntVarP(&node.id, "node-id","n", node.id, "The node id")
	startCmd.Flags().IntVarP(&node.capacity, "capacity", "c", node.capacity, "The block capacity")
	startCmd.Flags().StringVarP(&node.blockchainPath, "blockchain-path","b", node.blockchainPath, "The path of the blockchain's json file")
	startCmd.Flags().StringVarP(&node.dbPath, "database-path","d", node.dbPath, "The path of the blockchain's json file")
	
	startCmd.Flags().StringVarP(&node.brokerURL, "broker-url", "k", node.brokerURL, "The adress and port of the kafka broker")
	startCmd.Flags().IntVarP(&node.nodes, "nodes", "N", node.nodes, "The number of nodes")

	transactionCmd.Flags().StringVarP(&messagePayload, "message", "m", "", "If this flag exist, the transaction is a message")

	for _, cmd := range commandSet {
		cmd.Flags().StringVarP(&node.socket, "socket", "s", node.socket, "The tcp socket to connect to")
		cmd.Flags().StringVarP(&node.protocol, "protocol","p", node.protocol, "The socket protocol")

		RootCmd.AddCommand(cmd)
	}

}