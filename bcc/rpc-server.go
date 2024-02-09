package bcc

import (
	"fmt"
	"net"
	"net/rpc"
	"os"
)

type RPC struct{}

func (p *RPC) Create_transaction(args *TransactionArgs, reply *error) error{
	var tx Transaction
	if args.IsMessage {
		tx = NewMessageTransaction(userPubKey, args.Receiver_address, args.Message, myNonce, userPrivKey)
	} else {
		tx = NewTransferTransaction(userPubKey, args.Receiver_address, args.Amount, myNonce, userPrivKey)
	}
	myNonce++

	*reply = SendTransaction(Writer, tx)

	return nil

}

func (p *RPC) Stake(args *TransactionArgs, reply *error) error {
	var tx Transaction = NewTransferTransaction(userPubKey, "0", args.Amount, myNonce, userPrivKey)
	myNonce++
	SendTransaction(Writer, tx)
	return nil
}

func (p *RPC) Stop(_ struct{},reply * error) error {
	fmt.Println("Exiting...called by CLI")
	closeKafka()
	os.Exit(1)
	return nil
}



func (p *RPC) Balance(_ struct{}, reply *float64) error{
	*reply = ValidDB.GetBalance(userPubKey)
	return nil
}

func (p *RPC) GenerateWallet(_ struct{}, reply *WalletArgs ) error{ 
	_public, _private := GenerateKeys()
	
	(*reply).PublicKey = _public
	(*reply).PrivateKey = _private
	
	return nil
}

func (p *RPC) UseNodeWallet(_ struct{}, reply * error) error {
	*reply = nil
	userPubKey = MyPublicKey
	userPrivKey = MyPrivateKey
	return nil
}

func (p *RPC) PrintWallet(_ struct{}, reply *string) error {
	*reply = userPubKey

	return nil
}

func (p *RPC) UseWallet(wallet * WalletArgs, reply * error) error {
	userPubKey = wallet.PublicKey
	userPrivKey = wallet.PrivateKey

	return nil
}


type TransactionArgs struct {
	Receiver_address string
	Amount           float64
	Message          string
	IsMessage        bool
}

type WalletArgs struct {
	PublicKey string
	PrivateKey string
}

func Start_rpc() {
	myrpc := new(RPC)

	rpc.Register(myrpc)

	listener, err := net.Listen(PROTOCOL, SOCKET)
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
