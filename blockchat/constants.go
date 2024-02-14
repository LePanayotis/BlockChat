package blockchat

const (
	// Fee on the amount of each transaction
	// Will be credited to block validator
	feePercentage = float64(0.03)

	// Cost per message character
	// The fee of message transaction is the length of the message*costPerChar
	costPerChar = 1

	// Length of RSA public key
	keyLength = 512

	// Time format
	timeFormat = "02-01-2006 15:04:05.000"

	// Hash of the Genesis Block of the Blockchain
	genesisHash = "1"

	// Initial BCCs credited to each node
	initialBCC = float64(1000)

	// Banner
	banner = " ______  __      ______  ______  __  __   ______  __  __  ______  ______  \n/\\  == \\/\\ \\    /\\  __ \\/\\  ___\\/\\ \\/ /  /\\  ___\\/\\ \\_\\ \\/\\  __ \\/\\__  _\\ \n\\ \\  __<\\ \\ \\___\\ \\ \\/\\ \\ \\ \\___\\ \\  _\"-.\\ \\ \\___\\ \\  __ \\ \\  __ \\/_/\\ \\/ \n \\ \\_____\\ \\_____\\ \\_____\\ \\_____\\ \\_\\ \\_\\\\ \\_____\\ \\_\\ \\_\\ \\_\\ \\_\\ \\ \\_\\ \n  \\/_____/\\/_____/\\/_____/\\/_____/\\/_/\\/_/ \\/_____/\\/_/\\/_/\\/_/\\/_/  \\/_/"

	version = "1.0.0"
)
