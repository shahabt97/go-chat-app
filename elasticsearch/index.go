package elasticsearch

import ()

type UserIndex struct {
	Id       string `json:"id"`
	Username string `json:"username"`
	Email    string `json:"email"`
}

type PubMessageIndex struct {
	Id      string `json:"id"`
	Message string `json:"message"`
}
type PvMessageIndex struct {
	Id       string `json:"id"`
	Message  string `json:"message"`
	Sender   string `json:"sender"`
	Receiver string `json:"receiver"`
}
