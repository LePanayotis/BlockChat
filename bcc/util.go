package bcc

import (
	"crypto/rand"
	"crypto/rsa"
	"crypto/x509"
	"encoding/hex"
	"fmt"
	"github.com/spf13/cobra"
)

var StartCmd = &cobra.Command{
	Use:   "start",
	Short: "Starts the node",
	Args:  cobra.NoArgs,
	Long:  "Starts the node based on the configuration at the .env file",
	Run: func(cmd *cobra.Command, args []string) {
		StartNode()		
	},
}

func SetStartFlags() {
	StartCmd.Flags().BoolVarP(&node.detached,"detached","d",false,"Run or no the CLI")
	StartCmd.Flags().IntVarP(&node.id,"node-id","n",0,"The node id")
	StartCmd.Flags().StringVarP(&node.socket,"socket","s",":1500","The tcp socket to connect to")
	StartCmd.Flags().StringVarP(&node.protocol,"protocol","p","tcp","The socket protocol")
	StartCmd.Flags().IntVarP(&node.capacity,"capacity","c",3,"The block capacity")
	StartCmd.Flags().IntVar(&node.costPerChar,"cost-per-char",1,"The cost per character of messages")
	StartCmd.Flags().Float64VarP(&node.feePercentage,"fee","f",0.03,"The fee percentage written like 0.03")
	StartCmd.Flags().StringVar(&node.blockchainPath,"blockchain-path","blockchain.json","The path of the blockchain's json file")
	StartCmd.Flags().StringVar(&node.dbPath,"database-path","db.json","The path of the blockchain's json file")
	StartCmd.Flags().StringVar(&node.genesisHash,"genesis-hash","1","The hash of the Genesis Block")
	StartCmd.Flags().Float64VarP(&node.initialBCC,"initial-bcc","b",1000,"The initial BCC per node")
	StartCmd.Flags().StringVarP(&node.brokerURL,"broker-url","k","localhost:9093","The adress and port of the kafka broker")
	StartCmd.Flags().IntVarP(&node.nodes,"nodes","N",1,"The number of nodes")
	StartCmd.MarkFlagRequired("node-id")

}

func GenerateKeys() (string, string) {
	fmt.Println("Key length is:",node.keyLength)
	fmt.Println(node)
	key, _ := rsa.GenerateKey(rand.Reader, node.keyLength)
	priv := hex.EncodeToString(x509.MarshalPKCS1PrivateKey(key))
	pub := hex.EncodeToString(x509.MarshalPKCS1PublicKey(&key.PublicKey))
	return pub, priv
}

func (n * NodeConfig)setPublicKey(_key string, _node int){
	n.nodeMap[_key] = _node
	n.nodeIdArray[_node] = _key
}

func IsValidPublicKey(_key string) bool {
	return len(_key) == node.keyLength/4 || _key == "0"
}

func (n* NodeConfig) generateKeysUpdate() (string, string) {
	pub, priv := GenerateKeys()
	n.myPublicKey = pub
	n.myPrivateKey = priv
	n.currentPublicKey = pub
	n.currentPrivateKey = priv
	n.setPublicKey(pub, n.id)
	return pub, priv
}

func closeKafka() {
	node.writer.Close()
	node.txConsumer.Close()
	node.blockConsumer.Close()
}