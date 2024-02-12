package blockchat

const (
	//Fee on the amount of each transaction
	//Will be credited to block validator
	feePercentage = float64(0.03)

	//Cost per message character
	//The fee of message transaction is the length of the message*costPerChar
	costPerChar = 1

	//Length of RSA public key
	keyLength = 512

	//Time format
	timeFormat = "02-01-2006 15:04:05.000"

	//Hash of the Genesis Block of the Blockchain
	genesisHash = "1"

	//Initial BCCs credited to each node
	initialBCC = float64(1000)
)
