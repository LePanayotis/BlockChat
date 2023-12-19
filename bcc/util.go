package bcc

import (
	"crypto/rand"
	"crypto/rsa"
	"crypto/x509"
	"encoding/hex"
)



func GenerateKeys() (string, string) {
	key, _ := rsa.GenerateKey(rand.Reader, KEY_LENGTH)

	priv := hex.EncodeToString(x509.MarshalPKCS1PrivateKey(key))
	pub := hex.EncodeToString(x509.MarshalPKCS1PublicKey(&key.PublicKey))

	return pub, priv
}

	