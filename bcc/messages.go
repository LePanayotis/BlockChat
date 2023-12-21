package bcc

type WelcomeMessage struct{
	Bc Blockchain `json:"blockchain"`
	NodesIn []string `json:"nodesin"`	
}