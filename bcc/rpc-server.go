package bcc

import (
	"fmt"
	"net"
	"net/rpc"
	"os"
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
	return nil
}

func (p *RPC) Stake(args *TransactionArgs, reply *error) error {
	var tx Transaction = NewTransferTransaction(node.currentPublicKey, "0", args.Amount, myNonce, node.currentPrivateKey)
	myNonce++
	sendTransaction(node.writer, tx)
	return nil
}

func (p *RPC) Stop(_ struct{}, reply *error) error {
	fmt.Println("Exiting...called by CLI")
	closeKafka()
	go os.Exit(1)
	return nil
}

func (p *RPC) Balance(_ struct{}, reply *float64) error {
	*reply = node.myDB.getBalance(node.currentPublicKey)
	return nil
}

func (p *RPC) GenerateWallet(_ struct{}, reply *WalletArgs) error {
	_public, _private := GenerateKeys()

	(*reply).PublicKey = _public
	(*reply).PrivateKey = _private

	return nil
}

func (p *RPC) UseNodeWallet(_ struct{}, reply *error) error {
	*reply = nil
	node.currentPublicKey = node.myPublicKey
	node.currentPrivateKey = node.myPrivateKey
	return nil
}

func (p *RPC) PrintWallet(_ struct{}, reply *string) error {
	*reply = node.currentPublicKey

	return nil
}

func (p *RPC) UseWallet(wallet *WalletArgs, reply *error) error {
	node.currentPublicKey = wallet.PublicKey
	node.currentPrivateKey = wallet.PrivateKey

	return nil
}

func start_rpc() {
	myrpc := new(RPC)

	rpc.Register(myrpc)

	listener, err := net.Listen(node.protocol, node.socket)
	if err != nil {
		fmt.Println("Error starting listener:", err)
		return
	}

	fmt.Println("Server listening on named socket")
	for {
		conn, err := listener.Accept()
		if err != nil {
			fmt.Println("Error accepting connection:", err)
			continue
		}
		// Start a new goroutine to handle each incoming connection
		rpc.ServeConn(conn)
	}

}
