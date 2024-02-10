package bcc

import (
	"crypto/rand"
	"crypto/rsa"
	"crypto/x509"
	"encoding/hex"
)



func GenerateKeys() (string, string) {
	key, _ := rsa.GenerateKey(rand.Reader, node.keyLength)
	priv := hex.EncodeToString(x509.MarshalPKCS1PrivateKey(key))
	pub := hex.EncodeToString(x509.MarshalPKCS1PublicKey(&key.PublicKey))
	return pub, priv
}

func (node * NodeConfig) setPublicKey(_key string, _node int){
	node.nodeMap[_key] = _node
	node.nodeIdArray[_node] = _key
}

func IsValidPublicKey(_key string) bool {
	return len(_key) == node.keyLength/4 || _key == "0"
}

func (node * NodeConfig) generateKeysUpdate() (string, string) {
	pub, priv := GenerateKeys()
	node.myPublicKey = pub
	node.myPrivateKey = priv
	node.currentPublicKey = pub
	node.currentPrivateKey = priv
	node.setPublicKey(pub, node.id)
	return pub, priv
}

func closeKafka() {
	node.writer.Close()
	node.txConsumer.Close()
	node.blockConsumer.Close()
}