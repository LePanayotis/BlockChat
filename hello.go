package main

import (
	"fmt"
	"ntua-ds/bcc/bcc"
)
func main() {

	pub, priv := bcc.GenerateKeysUpdate()

	b := bcc.GenesisBlock(pub, priv)
	hashp := b.Current_hash
	var bc bcc.Blockchain = bcc.Blockchain{b}
	b = bcc.NewBlock(1,hashp)

	t := bcc.NewTransferTransaction(pub, pub, 1000, 1, priv)
	b.AddTransaction(t)
	t = bcc.NewMessageTransaction(pub, pub, "Geia sou",0,priv)
	b.AddTransaction(t)
	
	b.CalcHash()
	fmt.Println(b.CalcValidator())
	bc = append(bc, b)
	bc.WriteBlockchain()
	// fmt.Println(bc.IsValid())
	new_db, _ := bc.MakeDB()
	new_db.WriteDB()

	// bstring, _ := b.JSONify()
	// fmt.Println(bstring)
	// new_block, _ := bcc.ParseBlockJSON(bstring)
	// fmt.Printf("%+v\n", new_block)
	// fmt.Println(new_block.IsValid())
	for key := range new_db {
		fmt.Println(key)
	}
	//bcc.Server(&new_db)

}