package blockchat

import (
	"fmt"
	"net"
	"net/rpc"
	"os"
	"time"
)

type RPC struct{}

type TransactionArgs struct {
	Receiver_node int
	Amount        float64
	Message       string
	IsMessage     bool
}

type WalletArgs struct {
	PublicKey  string
	PrivateKey string
}

func (p *RPC) Create_transaction(args *TransactionArgs, reply *error) error {
	logger.Info("Create_transaction RPC called")
	var tx Transaction
	var receiver_address string
	if args.Receiver_node > -1 && args.Receiver_node < node.nodes {
		receiver_address = node.idArray[args.Receiver_node]
	} else {
		*reply = fmt.Errorf("node must be between 0 and %d", node.nodes-1)
		return *reply
	}
	node.outboundNonce++
	if args.IsMessage {
		tx = NewMessageTransaction(node.publicKey, receiver_address, args.Message, node.outboundNonce, node.privateKey)
	} else {
		tx = NewTransferTransaction(node.publicKey, receiver_address, args.Amount, node.outboundNonce, node.privateKey)
	}
	*reply = node.sendTransaction(tx)
	if *reply != nil {
		logger.Error("Failed to send transaction", *reply)
		node.outboundNonce--
	} 
	return *reply
}

func (p *RPC) Stake(args *TransactionArgs, reply *error) error {
	logger.Info("Stake RPC called")
	node.outboundNonce++
	var tx Transaction = NewTransferTransaction(node.publicKey, "0", args.Amount, node.outboundNonce, node.privateKey)
	*reply = node.sendTransaction(tx)
	if *reply != nil {
		logger.Error("Failed to send transaction", *reply)
		node.outboundNonce--
	}
	return *reply
}

func (p *RPC) Stop(_ struct{}, reply *error) error {
	logger.Info("Stop RPC called")

	go func() {
		logger.Info("Node will stop in 500ms")
		time.Sleep(time.Millisecond * 500)
		logger.Info("Node is stopping")
		closeKafka()
		os.Exit(1)
	}()
	return nil
}

func (p *RPC) Balance(_ struct{}, reply *float64) error {
	logger.Info("Balance RPC called")
	*reply = node.myDB.getBalance(node.id)
	return nil
}

func (p *RPC) PrintWallet(_ struct{}, reply *string) error {
	logger.Info("PrintWallet RPC called")
	*reply = node.publicKey

	return nil
}

func (p *RPC) GetNonce(_ struct{}, reply *uint) error {
	logger.Info("UseWallet RPC called")

	var nonce uint = node.myDB.getNonce(node.id)
	*reply = nonce
	return nil
}

func startRPC() error {
	myrpc := new(RPC)

	rpc.Register(myrpc)

	listener, err := net.Listen(node.protocol, node.socket)
	if err != nil {
		logger.Error("Error starting listener:", err)
		return err
	}

	logger.Info("Server listening protocol: " + node.protocol + " at socket: " + node.socket)
	for {
		conn, err := listener.Accept()
		if err != nil {
			logger.Error("Error accepting connection:", err)
			continue
		}
		// Start a new goroutine to handle each incoming connection
		go rpc.ServeConn(conn)
	}

}
