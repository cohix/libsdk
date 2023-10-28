package fabric

type Replier func(msg any)                  // function for sending message replies
type Generator func() any                   // function to generate objects for message unmarshalling
type Receiver func(msg any)                 // function for receiving messages
type Handler func(msg any, replier Replier) // function for receiving messages and sending replies

type Fabric interface {
	// Messenger is for async request/reply messaging with other services over the fabric.
	// Message receiving when the recipient is not connected is best-effort and not guaranteed.
	Messenger(service string) (MsgConnection, error)

	// Create a 'replayer', i.e. a pub/sub connection with ordered messages that durably persist in the fabric.
	// If 'beginning' is true, messages will be replayed from the "beginning of time".
	// If false, messages will only be played from the current time. If only publishing, pass false.
	Replayer(subject string, beginning bool) (ReplayConnection, error)
}

type MsgConnection interface {
	SendAndRecv(msg any, receiver Receiver) error
	RecvAndReply(handler Handler)
}

type ReplayConnection interface {
	Publish(msg any) error
	Replay(gen Generator, receiver Receiver) (chan bool, error)
}
