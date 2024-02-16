package blockchat

import (
	"bufio"
	"encoding/csv"
	"fmt"
	"log/slog"
	"net/rpc"
	"os"
	"strconv"
	"strings"
	"github.com/spf13/cobra"
)

// CLI command to create new transaction via RPC
var transactionCmd = &cobra.Command{
	Use:     "transaction",
	Aliases: []string{"t"},
	Short:   "Sends new transaction",
	Long:    "Sends new transaction",
	Args:    cobra.RangeArgs(1, 2),
	Run: func(_cmd *cobra.Command, _args []string) {

		// Prepare the TransactionArgs for RPC
		var amount float64 = 0
		// Get receiver node id from flags
		receiverId, err := strconv.Atoi(_args[0])
		if err != nil {
			fmt.Println("Error parsing recipient node id", err)
			return
		}
		// Transaction is message when payload is not empty string

		messagePayload, _ := _cmd.Flags().GetString("message")
		isMessage := messagePayload != ""
		// If not message, _args array has to have length of 2
		if !isMessage && len(_args) == 2 {
			// Parses amount from args
			amount, err = strconv.ParseFloat(_args[1], 64)
			if err != nil {
				fmt.Println("Error parsing amount", err)
				return
			}
		} else {
			// Case _args length is not 2
			fmt.Println("Provide receiver node id and amount")
			return
		}

		// Creates struct
		tx := &TransactionArgs{
			ReceiverNode: receiverId,
			Amount:       amount,
			Message:      messagePayload,
			IsMessage:    isMessage,
		}

		// Dials socket
		protocol, _ := _cmd.Flags().GetString("protocol")
		socket, _ := _cmd.Flags().GetString("socket")
		client, err := rpc.Dial(protocol, socket)
		if err != nil {
			fmt.Println("Error connecting to server:", err)
			return
		}
		defer client.Close()

		// Calls RPC
		var reply error
		err = client.Call("RPC.Create_transaction", tx, &reply)
		if err != nil {
			fmt.Println("Error sending transaction:", err)
			return
		}
		fmt.Println("Transaction sent, acceptance not guaranteed")
	},
}

// CLI command to return node's wallet balance
var balanceCmd = &cobra.Command{
	Use:     "balance",
	Aliases: []string{"b"},
	Args:    cobra.NoArgs,
	Short:   "Returns the current balance in BlockChatCoins",
	Long:    "Returns the current balance of the specified account in BlockChatCoins Default account is the current node",
	Run: func(_cmd *cobra.Command, _args []string) {
		// Dials socket
		protocol, _ := _cmd.Flags().GetString("protocol")
		socket, _ := _cmd.Flags().GetString("socket")
		client, err := rpc.Dial(protocol, socket)
		if err != nil {
			fmt.Println("Error connecting to server:", err)
			return
		}
		defer client.Close()

		// Calls RPC
		var reply float64
		err = client.Call("RPC.Balance", struct{}{}, &reply)
		if err != nil {
			fmt.Println("Error calling balance:", err)
			return
		}
		// Prints balance
		fmt.Printf("Balance: %.2f BCCs\n", reply)
	},
}

// Command to get node's recorded nonce
var getNonce = &cobra.Command{
	Use:     "nonce",
	Aliases: []string{"n"},
	Short:   "Returns the current nonce stored in the node's database",
	Run: func(_cmd *cobra.Command, _args []string) {
		protocol, _ := _cmd.Flags().GetString("protocol")
		socket, _ := _cmd.Flags().GetString("socket")
		client, err := rpc.Dial(protocol, socket)
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
		fmt.Printf("The current nonce is: %d\n", reply)
	},
}

// Command to get node's wallet address (public key)
var printWalletCmd = &cobra.Command{
	Use:     "print-wallet",
	Aliases: []string{"pw"},
	Short:   "Returns the wallet address of the node, its public key",
	Run: func(_cmd *cobra.Command, _args []string) {
		// Dials socket
		protocol, _ := _cmd.Flags().GetString("protocol")
		socket, _ := _cmd.Flags().GetString("socket")
		client, err := rpc.Dial(protocol, socket)
		if err != nil {
			fmt.Println("Error connecting to server:", err)
			return
		}
		defer client.Close()

		// Gets wallet address
		var reply string
		err = client.Call("RPC.PrintWallet", struct{}{}, &reply)
		if err != nil {
			fmt.Println("Error executing command:", err)
			return
		}
		// Prints address
		fmt.Println("The current wallet is:\n", reply)

	},
}

var stakeCmd = &cobra.Command{
	Use:     "stake",
	Aliases: []string{"s"},
	Args:    cobra.ExactArgs(1),
	Short:   "Stakes ammount",
	Run: func(_cmd *cobra.Command, _args []string) {
		protocol, _ := _cmd.Flags().GetString("protocol")
		socket, _ := _cmd.Flags().GetString("socket")
		client, err := rpc.Dial(protocol, socket)
		if err != nil {
			fmt.Println("Error connecting to server:", err)
			return
		}
		defer client.Close()
		var amount float64 = 0
		amount, err = strconv.ParseFloat(_args[0], 64)
		if err != nil {
			fmt.Println("Provide float64 amount. Decimal delimiter is '.'")
		}
		tx := &TransactionArgs{
			ReceiverNode: -1,
			Amount:       amount,
			Message:      "",
			IsMessage:    false,
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

var stopCmd = &cobra.Command{
	Use:   "stop",
	Args:  cobra.NoArgs,
	Short: "Stopping running process",
	Run: func(_cmd *cobra.Command, _args []string) {
		protocol, _ := _cmd.Flags().GetString("protocol")
		socket, _ := _cmd.Flags().GetString("socket")
		client, err := rpc.Dial(protocol, socket)
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
	Run: func(_cmd *cobra.Command, _args []string) {

		var node *nodeConfig = DefaultNodeConfig()
		node.EnvironmentConfig()

		node.dbPath, _ = _cmd.Flags().GetString("database-path")
		node.blockchainPath, _ = _cmd.Flags().GetString("blockchain-path")
		node.brokerURL, _ = _cmd.Flags().GetString("broker-url")

		node.id, _ = _cmd.Flags().GetInt("node-id")
		node.capacity, _ = _cmd.Flags().GetInt("capacity")
		node.nodes, _ = _cmd.Flags().GetInt("nodes")

		node.protocol, _ = _cmd.Flags().GetString("protocol")
		node.socket, _ = _cmd.Flags().GetString("socket")

		node.useCLI, _ = _cmd.Flags().GetBool("cli")

		node.inputPath, _ = _cmd.Flags().GetString("input-path")
		node.logPath, _ = _cmd.Flags().GetString("log-path")

		if node.logPath != "" {
			logfile, err := os.Create(node.logPath)
			if err != nil {
				return
			}
			writer := &TeeWriter{
				stdout: os.Stderr,
				file:   logfile,
			}

			logger = slog.New(slog.NewTextHandler(writer, &slog.HandlerOptions{
				Level: slog.LevelInfo,
			}))
		}

		_cmd.Parent().RemoveCommand(_cmd)
		if node.useCLI {
			fmt.Println(banner)
		}
		node.Start()
	},
}

var rootCmd = &cobra.Command{
	Use:     "blockchat",
	Short:   "BlockChat is a simple CLI application",
	Long:    `BlockChat is a simple CLI application built using Cobra.`,
	Version: "1.0.0",
	Run: func(_cmd *cobra.Command, _args []string) {
		fmt.Println(banner)
	},
}

func ConfigCommands() *cobra.Command {

	var tempConfig *nodeConfig = DefaultNodeConfig()
	tempConfig.EnvironmentConfig()

	var commandSet []*cobra.Command = []*cobra.Command{
		startCmd,
		balanceCmd,
		stopCmd,
		stakeCmd,
		transactionCmd,
		printWalletCmd,
		getNonce,
	}

	// Defaults are the one specified in the environment
	startCmd.Flags().IntP("node-id", "n", tempConfig.id, "The node id")
	startCmd.Flags().IntP("capacity", "c", tempConfig.capacity, "The block capacity")
	startCmd.Flags().StringP("blockchain-path", "b", tempConfig.blockchainPath, "The path of the blockchain's json file")
	startCmd.Flags().StringP("database-path", "d", tempConfig.dbPath, "The path of the blockchain's json file")
	startCmd.Flags().StringP("broker-url", "k", tempConfig.brokerURL, "The adress and port of the kafka broker")
	startCmd.Flags().IntP("nodes", "N", tempConfig.nodes, "The number of nodes")
	startCmd.Flags().BoolP("cli", "i", tempConfig.useCLI, "If present, logs are silenced and interactive CLI activates")
	startCmd.Flags().StringP("input-path", "f", tempConfig.inputPath, "Sets the input file of transactions")
	startCmd.Flags().StringP("log-path", "l", tempConfig.logPath, "Redirects stderr to this file")

	startCmd.MarkFlagFilename("log-path")
	startCmd.MarkFlagFilename("input-path")
	startCmd.MarkFlagFilename("blockchain-path")
	startCmd.MarkFlagFilename("databases-path")

	transactionCmd.Flags().StringP("message", "m", "", "If this flag exist, the transaction is a message")

	rootCmd.Flags().Bool("license", false, "Prints the license")
	rootCmd.SetVersionTemplate(fmt.Sprintf("BlockChat Version: %s", version))
	for _, cmd := range commandSet {
		cmd.Flags().StringP("protocol", "p", tempConfig.protocol, "The socket protocol: unix, tcp, udp")
		cmd.Flags().StringP("socket", "s", tempConfig.socket, "The socket to connect to, use : before port number, example :1500")
		rootCmd.AddCommand(cmd)
	}

	return rootCmd

}

func (node *nodeConfig) startCLI() error {

	scanner := bufio.NewScanner(os.Stdin)

	prompt := fmt.Sprintf("[NODE %s]> ", node.idString)

	var commandSet []*cobra.Command = []*cobra.Command{
		balanceCmd,
		stopCmd,
		stakeCmd,
		transactionCmd,
		printWalletCmd,
		getNonce,
	}
	for _, cmd := range commandSet {
		cmd.Flags().Lookup("protocol").Value.Set(node.protocol)
		cmd.Flags().Lookup("socket").Value.Set(node.socket)
	}

	for {

		fmt.Print(prompt)
		scanner.Scan()
		input := scanner.Text()
		input = strings.TrimSpace(input)
		r := csv.NewReader(strings.NewReader(input))
		r.Comma = ' '
		cmdArgs, err := r.Read()
		if err != nil {
			continue
		}

		rootCmd.SetArgs(cmdArgs[:])
		if err := rootCmd.Execute(); err != nil {
			fmt.Println("Error executing command:", err)
		}

	}
}
