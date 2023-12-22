package bcc

import (
	"crypto/rand"
	"crypto/rsa"
	"crypto/x509"
	"encoding/hex"
	"os"
	"strconv"

	"github.com/segmentio/kafka-go"
)


func StartEnv() {

	FEE_PERCENTAGE, _ = strconv.ParseFloat(os.Getenv("FEE_PERCENTAGE"),64)
	COST_PER_CHAR, _ = strconv.Atoi(os.Getenv("COST_PER_CHAR"))
	BROKER_URL = os.Getenv("BROKER_URL")
	NODES, _ = strconv.Atoi(os.Getenv("NODES"))
	CAPACITY, _ = strconv.Atoi(os.Getenv("CAPACITY"))


	ns, b := os.LookupEnv("NODE_ID")
	NodeIDString = ns
	if !b {
		NodeIDString = os.Args[1]
	}
	NodeID, _ = strconv.Atoi(NodeIDString)
	BLOCKCHAIN_PATH = `blockchain`+NodeIDString+`.json`
	DB_PATH = `db`+NodeIDString+`.json`
	BlockIndex = 0
	Last_hash = GENESIS_HASH
	NodeIDArray = make([]string, NODES)

	MyPublicKey, MyPrivateKey  = GenerateKeysUpdate()

	MyHeaders = []kafka.Header{
		{
			Key:   "NodeId",
			Value: []byte(NodeIDString),
		},
		{
			Key:   "NodeWallet",
			Value: []byte(MyPublicKey),
		},
	}
}

func GenerateKeys() (string, string) {
	key, _ := rsa.GenerateKey(rand.Reader, KEY_LENGTH)
	priv := hex.EncodeToString(x509.MarshalPKCS1PrivateKey(key))
	pub := hex.EncodeToString(x509.MarshalPKCS1PublicKey(&key.PublicKey))
	return pub, priv
}

func SetPublicKey(_key string, _node int){
	NodeMap[_key] = _node
	NodeIDArray[_node] = _key
}

func IsValidPublicKey(_key string) bool {
	return len(_key) == KEY_LENGTH/4 || _key == "0"
}

func GenerateKeysUpdate() (string, string) {
	pub, priv := GenerateKeys()
	SetPublicKey(pub, NodeID)
	return pub, priv
}
