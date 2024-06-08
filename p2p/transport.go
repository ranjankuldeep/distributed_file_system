package p2p

// Transport is anything that handles the communication
// between the nodes in the network. This can be of the
// form (TCP, UDP, websockets, ...)
type Transport interface {
	Addr() string           // Should give an listening ddress
	Dial(string) error      // Should able to dial to any node in the network
	ListenAndAccept() error // Should be conitnously listening for any incoming connection and accept the connection with handshaking
	Consume() <-chan RPC    // A cosume channel so that we can listen for any message.
	Close() error           // Close the connection once done.
}
