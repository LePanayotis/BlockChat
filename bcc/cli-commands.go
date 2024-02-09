package bcc

import (
	"fmt"
	"net/rpc"
	"strconv"
	"github.com/spf13/cobra"
)

var _input_priv_key, _input_pub_key string
var isMessage bool
var toNode int

var TransactionCmd = &cobra.Command{
	Use:     "transaction",
	Aliases: []string{"t"},
	Short:   "Sends new transaction",
	Long:    "Sends new transaction",
	Args:    cobra.MinimumNArgs(2),
	Run: func(cmd *cobra.Command, args []string) {
		client, err := rpc.Dial(PROTOCOL, SOCKET)
		if err != nil {
			fmt.Println("Error connecting to server:", err)
			return
		}
		defer client.Close()
		var amount float64 = 0
		if !isMessage {
			amount, err = strconv.ParseFloat(args[1],64)
			if err != nil  {
				fmt.Println("Provide float64 amount. Decimal delimiter is '.'")
			}
		}
		var toAddress string = args[0]
		if toNode != -1 {
			if toNode < NODES {
				toAddress = NodeIDArray[toNode]
			} else {
				fmt.Printf("Provide an existing node id")
			}
		}
		tx := &TransactionArgs{
			Receiver_address: toAddress,
			Amount: amount,
			Message: args[1],
			IsMessage: isMessage,
		}
		var reply error
		err = client.Call("RPC.Create_transaction", &tx, &reply)
		if err != nil || reply != nil {
			fmt.Println("Error calling balance:", err)
			return
		}

		fmt.Println("Transaction send")
	},
}

var BalanceCmd = &cobra.Command{
	Use:   "balance",
	Args:  cobra.NoArgs,
	Short: "Returns the current balance in BlockChatCoins",
	Long:  "Returns the current balance of the specified account in BlockChatCoins Default account is the current node",
	Run: func(cmd *cobra.Command, args []string) {
		client, err := rpc.Dial(PROTOCOL, SOCKET)
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

		fmt.Printf("Balance: %f BCCs\n", reply)
	},
}

var GenerateWalletCmd = &cobra.Command{
	Use:     "generate-wallet",
	Aliases: []string{"gw"},
	Short:   "Returns a new pair of Public and Private Keys",
	Run: func(cmd *cobra.Command, args []string) {
		client, err := rpc.Dial(PROTOCOL, SOCKET)
		if err != nil {
			fmt.Println("Error connecting to server:", err)
			return
		}
		defer client.Close()
		var reply WalletArgs
		err = client.Call("RPC.GenerateWallet", struct{}{}, &reply)
		if err != nil {
			fmt.Println("Error calling balance:", err)
			return
		}
		fmt.Printf("Congrats, you have a new wallet!\nPublic key: %s\nPrivate key: %s\n",reply.PublicKey, reply.PrivateKey)
	},
}

var UseNodeWalletCmd = &cobra.Command{
	Use:     "use-node-wallet",
	Aliases: []string{"unw"},
	Short:   "Uses the node's wallet",
	Run: func(cmd *cobra.Command, args []string) {
		client, err := rpc.Dial(PROTOCOL, SOCKET)
		if err != nil {
			fmt.Println("Error connecting to server:", err)
			return
		}
		defer client.Close()
		var reply error
		err = client.Call("RPC.UseNodeWallet", struct{}{}, &reply)
		if err != nil {
			fmt.Println("Error executing command:", err)
			return
		}
	},
}

var PrintWalletCmd = &cobra.Command{
	Use:     "print-wallet",
	Aliases: []string{"pcw"},
	Short:   "Returns the current wallet used by the daemon",
	Run: func(cmd *cobra.Command, args []string) {
		client, err := rpc.Dial(PROTOCOL, SOCKET)
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
		fmt.Println("The current wallet is:\n",reply)

	},
}

var UseWalletCmd = &cobra.Command{
	Use:     "use-wallet",
	Args:    cobra.ExactArgs(2),
	Aliases: []string{"uw"},
	Short:   "Returns the current wallet used by the daemon",
	Run: func(cmd *cobra.Command, args []string) {
		client, err := rpc.Dial(PROTOCOL, SOCKET)
		if err != nil {
			fmt.Println("Error connecting to server:", err)
			return
		}
		defer client.Close()
		var wallet WalletArgs = WalletArgs{
			PublicKey: _input_pub_key,
			PrivateKey: _input_priv_key,
		}
		var reply error
		err = client.Call("RPC.UseWallet", &wallet, &reply)
		if err != nil {
			fmt.Println("Error executing command:", err)
			return
		}
		fmt.Println("The current wallet is:\n",_input_pub_key)

	},
}

var StakeCmd = &cobra.Command{
	Use:   "stake",
	Args:  cobra.ExactArgs(1),
	Short: "Stakes ammount",
	Run: func(cmd *cobra.Command, args []string) {
		client, err := rpc.Dial(PROTOCOL, SOCKET)
		if err != nil {
			fmt.Println("Error connecting to server:", err)
			return
		}
		defer client.Close()
		var amount float64 = 0
		if !isMessage {
			amount, err = strconv.ParseFloat(args[0],64)
			if err != nil  {
				fmt.Println("Provide float64 amount. Decimal delimiter is '.'")
			}
		}
		tx := &TransactionArgs{
			Receiver_address: "0",
			Amount: amount,
			Message: "",
			IsMessage: false,
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

//to add
var ShowBlockchain = &cobra.Command{
	Use:     "show-blockchain",
	Args:    cobra.NoArgs,
	Aliases: []string{"blockchain", "bc"},
	Short:   "Shows blockchain in json format",
	Run: func(cmd *cobra.Command, args []string) {
		fmt.Println("Your blockchain")
	},
}

var StopCmd = &cobra.Command{
	Use:   "stop",
	Args:  cobra.NoArgs,
	Short: "Stopping running process",
	Run: func(cmd *cobra.Command, args []string) {
		client, err := rpc.Dial(PROTOCOL, SOCKET)
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





func ConfigCmds() error {

	UseWalletCmd.Flags().StringVarP(&_input_pub_key, "public-key", "p", "", "Public key of the wallet")
	UseWalletCmd.Flags().StringVarP(&_input_priv_key, "private-key", "P", "", "Private key of the RSA key pair of the wallet")
	UseWalletCmd.MarkFlagRequired("public-key")
	UseWalletCmd.MarkFlagRequired("private-key")

	TransactionCmd.Flags().BoolVarP(&isMessage,"message","m",false,"If this flag exist, the transaction is a message")
	TransactionCmd.Flags().IntVarP(&toNode,"node","n", NodeID,"Transaction with receiver the wallet of the indicated node")
	return nil

}
