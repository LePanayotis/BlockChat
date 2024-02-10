package blockchat

import (
	"net"
	"net/rpc"
	"os"
	"time"
)

type RPC struct{}

type TransactionArgs struct {
	Receiver_address string
	Amount           float64
	Message          string
	IsMessage        bool
	ToNode           int
}

type WalletArgs struct {
	PublicKey  string
	PrivateKey string
}

func (p *RPC) Create_transaction(args *TransactionArgs, reply *error) error{
	logger.Info("Create_transaction RPC called")
	var tx Transaction
	var receiver_address string
	if args.ToNode > -1 && args.ToNode < node.nodes {
		receiver_address = node.nodeIdArray[args.ToNode]
	} else {
		receiver_address = args.Receiver_address
	}
	if args.IsMessage {
		tx = NewMessageTransaction(node.currentPublicKey, receiver_address, args.Message, myNonce, node.currentPrivateKey)
	} else {
		tx = NewTransferTransaction(node.currentPublicKey, receiver_address, args.Amount, myNonce, node.currentPrivateKey)
	}
	myNonce++

	*reply = sendTransaction(node.writer, tx)
	if *reply != nil {
		logger.Error("Failed to send transaction",*reply)
	}
	return *reply
}

func (p *RPC) Stake(args *TransactionArgs, reply *error) error {
	logger.Info("Stake RPC called")
	var tx Transaction = NewTransferTransaction(node.currentPublicKey, "0", args.Amount, myNonce, node.currentPrivateKey)
	myNonce++
	*reply = sendTransaction(node.writer, tx)
	if *reply != nil {
		logger.Error("Failed to send transaction",*reply)
	}
	return *reply
}

func (p *RPC) Stop(_ struct{}, reply *error) error {
	logger.Info("Stop RPC called")
	
	go func() {
		logger.Info("Node will stop in 500ms")
		time.Sleep(time.Millisecond*500)
		logger.Info("Node is stopping")
		closeKafka()
		os.Exit(1)
	}()
	return nil
}

func (p *RPC) Balance(_ struct{}, reply *float64) error {
	logger.Info("Balance RPC called")
	*reply = node.myDB.getBalance(node.currentPublicKey)
	return nil
}

func (p *RPC) GenerateWallet(_ struct{}, reply *WalletArgs) error {
	logger.Info("GenerateWallet RPC called")
	_public, _private := GenerateKeys()

	(*reply).PublicKey = _public
	(*reply).PrivateKey = _private

	return nil
}

func (p *RPC) UseNodeWallet(_ struct{}, reply *error) error {

	logger.Info("UseNodeWallet RPC called")
	*reply = nil
	node.currentPublicKey = node.myPublicKey
	node.currentPrivateKey = node.myPrivateKey
	return nil
}

func (p *RPC) PrintWallet(_ struct{}, reply *string) error {
	logger.Info("PrintWallet RPC called")
	*reply = node.currentPublicKey

	return nil
}

func (p *RPC) UseWallet(wallet *WalletArgs, reply *error) error {
	logger.Info("UseWallet RPC called")
	node.currentPublicKey = wallet.PublicKey
	node.currentPrivateKey = wallet.PrivateKey

	return nil
}

func start_rpc() error{
	myrpc := new(RPC)

	rpc.Register(myrpc)

	listener, err := net.Listen(node.protocol, node.socket)
	if err != nil {
		logger.Error("Error starting listener:", err)
		return err
	}

	logger.Info("Server listening protocol: "+node.protocol+" at socket: "+node.socket)
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
