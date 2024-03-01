package blockchat

import (
	"crypto/rand"
	"crypto/rsa"
	"crypto/x509"
	"encoding/hex"
	"errors"
)

// Returns RSA public and private keys in string format
func generateKeys() (string, string) {
	key, err := rsa.GenerateKey(rand.Reader, keyLength)
	if err != nil { 
		logger.Error("Failed to generate RSA key", "error",err)
		return "", ""
	}
	priv := hex.EncodeToString(x509.MarshalPKCS1PrivateKey(key))
	pub := hex.EncodeToString(x509.MarshalPKCS1PublicKey(&key.PublicKey))
	return pub, priv
}

// Updates the node's node map (maps public keys to node ids)
// and id array (idArray[id]  has the public key of node id)
func (node * nodeConfig) setPublicKey(_key string, _node int){
	node.nodeMap[_key] = _node
	node.idArray[_node] = _key
}

// Generates public-private keys and updates map and array
func (node * nodeConfig) generateKeysUpdate() (string, string) {
	pub, priv := generateKeys()
	node.publicKey = pub
	node.privateKey = priv
	node.setPublicKey(pub, node.id)
	return pub, priv
}

// Util function to close open connections of kafka producers and consumers
func (node *nodeConfig) closeKafka() error{
	return errors.Join(
		node.writer.Close(),
	 	node.txConsumer.Close(),
	 	node.blockConsumer.Close())
}