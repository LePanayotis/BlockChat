package blockchat

import (
	"crypto/rand"
	"crypto/rsa"
	"crypto/x509"
	"encoding/hex"
)



func generateKeys() (string, string) {
	key, _ := rsa.GenerateKey(rand.Reader, node.keyLength)
	priv := hex.EncodeToString(x509.MarshalPKCS1PrivateKey(key))
	pub := hex.EncodeToString(x509.MarshalPKCS1PublicKey(&key.PublicKey))
	return pub, priv
}

func (node * nodeConfig) setPublicKey(_key string, _node int){
	node.nodeMap[_key] = _node
	node.idArray[_node] = _key
}

func isValidPublicKey(_key string) bool {
	return len(_key) == node.keyLength/4 || _key == "0"
}

func (node * nodeConfig) generateKeysUpdate() (string, string) {
	pub, priv := generateKeys()
	node.publicKey = pub
	node.privateKey = priv
	node.setPublicKey(pub, node.id)
	return pub, priv
}

func closeKafka() {
	node.writer.Close()
	node.txConsumer.Close()
	node.blockConsumer.Close()
}