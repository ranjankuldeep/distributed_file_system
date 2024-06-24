package p2p

import "net"

// Peer is an interface that represents the remote node.
// It implements io.multiwriter or io.reader or io.writer.
type Peer interface {
	net.Conn
	Send([]byte) error
	CloseStream()
}
