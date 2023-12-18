package models

const (
	MessageTypeConnect    = "__USER__CONNECTED"
	MessageTypeDisconnect = "__USER__DISCONNECTED"
	MessageTypeListUsers  = "__LIST__USERS"
)

type Message struct {
	Message  string
	UserName string
	Target   string
	Type     string
}
