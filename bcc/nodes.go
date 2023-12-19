package bcc

import ()



var NodeID int = 0
var BlockIndex int = 0
var Last_hash string = GENESIS_HASH




var NodeMap map[string]int = make(map[string]int)
var NodeIDArray [NODES]string
var NodeStakes [NODES]float64

func SetPublicKey(_key string, _node int){
	NodeMap[_key] = _node
	NodeIDArray[_node] = _key
}

func IsValidPublicKey(_key string) bool {
	return len(_key) == KEY_LENGTH || _key == "0"
}

func GenerateKeysUpdate() (string, string) {
	pub, priv := GenerateKeys()
	SetPublicKey(pub, NodeID)
	return pub, priv
}

//func CalcValidator()
