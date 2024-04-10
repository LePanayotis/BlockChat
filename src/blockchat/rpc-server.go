package blockchat

import (
	"fmt"
	"net"
	"net/rpc"
	"os"
	"time"
)


// Helper struct to expose remote procedure calls
type RPC struct{
	node *nodeConfig
}

// Struct to exchange transaction information between caller and RPC
type TransactionArgs struct {
	ReceiverNode int
	Amount        float64
	Message       string
	IsMessage     bool
}

// Struct to exchange wallet information between caller and RPC
type WalletArgs struct {
	PublicKey  string
	PrivateKey string
}

// Remote process call called by cli and makes service process to create and send new transaction
// according to input arguments
func (p *RPC) Create_transaction(args *TransactionArgs, reply *error) error {
	
	logger.Info("Create_transaction RPC called")
	var tx Transaction
	var receiverAddress string

	// Checks constraints for receiver and maps node id to wallet address
	if args.ReceiverNode > -1 && args.ReceiverNode < p.node.nodes {
		receiverAddress = p.node.idArray[args.ReceiverNode]
	} else {
		*reply = fmt.Errorf("node must be between 0 and %d", p.node.nodes-1)
		logger.Warn("Invalid receiver node transaction","lastNode",p.node.nodes-1, "receiver",args.ReceiverNode )
		return *reply
	}
	// Checks if transaction is message
	if args.IsMessage {
		// Creates message transaction instance
		tx = p.node.NewMessageTransaction(receiverAddress, args.Message)
	} else {
		// Creates transfer transaction instance
		tx = p.node.NewTransferTransaction(receiverAddress, args.Amount)
	}
	
	// Sends transaction
	*reply = p.node.sendTransaction(&tx)
	if *reply != nil {
		logger.Error("Failed to send transaction","error", *reply)
		// In case of error undo outbound nonce
		p.node.outboundNonce--
	} 

	logger.Info("Sent transaction","receiver",args.ReceiverNode,"amount",args.Amount, "message",args.Message)
	return *reply
}


// Remote process call to create stake transaction
func (p *RPC) Stake(args *TransactionArgs, reply *error) error {
	
	logger.Info("Stake RPC called")
	// Creates new transaction
	var tx Transaction = p.node.NewTransferTransaction("0", args.Amount)
	
	// Sends transaction
	*reply = p.node.sendTransaction(&tx)
	if *reply != nil {
		logger.Error("Failed to send transaction", "error",*reply)
		// In case of error undo outbound nonce
		p.node.outboundNonce--
	}

	logger.Warn("Stake sent","stake",args.Amount)
	return *reply
}

// RPC Called by CLI to terminate the node service process
func (p *RPC) Stop(_ struct{}, reply *error) error {
	logger.Info("Stop RPC called")
	go func() {
		logger.Info("Node will stop in 500ms")
		time.Sleep(time.Millisecond * 500)
		logger.Info("Node is stopping")
		p.node.closeKafka()
		os.Exit(1)
	}()
	return nil
}

// RPC to get node's wallet balance as set in databse
func (p *RPC) Balance(_ struct{}, reply *float64) error {
	logger.Info("Balance RPC called")
	*reply = p.node.myDB.getBalance(p.node.id)
	return nil
}

// RPC to get node's public key
func (p *RPC) PrintWallet(_ struct{}, reply *string) error {
	logger.Info("PrintWallet RPC called")
	*reply = p.node.publicKey
	return nil
}

// RPC to get node's current nonce as set in database,
// not in node.outboundNonce (this is temporary)
func (p *RPC) GetNonce(_ struct{}, reply *uint) error {
	logger.Info("GetNonce RPC called")
	*reply = p.node.myDB.getNonce(p.node.id)
	return nil
}

func (p *RPC) ViewBlock(_ struct{}, reply *Block) error {
	logger.Info("ViewBlock RPC called")
	length := len(p.node.blockchain)
	if (length >0 ) {
		*reply = p.node.blockchain[length-1]
	} else {
		*reply = Block{}
	}
	
	return nil
}

func (p *RPC) ViewDatabase(_ struct{}, reply *Database) error {
	logger.Info("ViewDatabase RPC called")
	*reply = p.node.myDB
	
	return nil
}
// Function to start and expose RPC service to defined socket
// Protocols: unix, tcp, udp, tcp4, tcp6
// Sockets: address:port in the case of internet protocols
//			/path/to/socket in the case of unix sockets
// 			You are advised to provide the address 127.0.0.1
//			Socket can be only the port, but ':' is necessary before port number
func (node *nodeConfig) startRPC() error {
	
	// Registers RPC methods
	myrpc := new(RPC)
	myrpc.node = node
	err := rpc.Register(myrpc)
	if err != nil {
		logger.Error("Could not register RPC","error",err)
		return err
	}

	// Starts RPC listener
	listener, err := net.Listen(node.protocol, node.socket)
	if err != nil {
		logger.Error("Error starting listener","error",err)
		return err
	}
	logger.Info("Server listening", "protocol",node.protocol,"socket",node.socket)

	// Loops until parent process exits
	for {

		// Accepts incoming RPC
		conn, err := listener.Accept()
		if err != nil {
			logger.Warn("Error accepting RPC connection", "error",err)
			continue
		}
		// Start a new goroutine to handle each incoming RPC
		go rpc.ServeConn(conn)
	}

}
