package msghandler

type handler interface {
	func HandleMessage([]byte) error
}