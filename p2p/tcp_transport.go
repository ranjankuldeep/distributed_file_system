package p2p

import (
	"errors"
	"fmt"
	"net"

	"github.com/ranjankuldeep/distributed_file_system/logs"
)

type TCPTransportOpts struct {
	ListenAddr    string // Holds the address where a peer is listening.
	HandshakeFunc HandshakeFunc
	Decoder       Decoder // Should provide its own decoder for different protocols on how to decode the message
	OnPeer        func(Peer) error
}

type TCPTransport struct {
	TCPTransportOpts
	listener net.Listener
	rpcch    chan RPC // used in consume method
}

func NewTCPTransport(opts TCPTransportOpts) *TCPTransport {
	return &TCPTransport{
		TCPTransportOpts: opts,
		rpcch:            make(chan RPC, 1024),
	}
}

// Addr implements the Transport interface return the address
// the transport is accepting connections.
func (t *TCPTransport) Addr() string {
	return t.ListenAddr
}

// Consume implements the Tranport interface, which will return read-only channel
// for reading the incoming messages received from another peer in the network.
func (t *TCPTransport) Consume() <-chan RPC {
	return t.rpcch
}

// Close implements the Transport interface.
func (t *TCPTransport) Close() error {
	return t.listener.Close()
}

// Dial implements the Transport interface.
func (t *TCPTransport) Dial(addr string) error {
	conn, err := net.Dial("tcp", addr)
	if err != nil {
		return err
	}
	go t.handleConn(conn, true) // Since you dial and recive the connection, make this as true for the outbound rule.
	return nil
}

func (t *TCPTransport) ListenAndAccept() error {
	var err error
	t.listener, err = net.Listen("tcp", t.ListenAddr)
	if err != nil {
		return err
	}
	go t.startAcceptLoop()
	logs.Logger.Infof("TCP transport listening on port: %s\n", t.ListenAddr)
	return nil
}

// A blocking loop.
func (t *TCPTransport) startAcceptLoop() {
	for {
		conn, err := t.listener.Accept()
		if errors.Is(err, net.ErrClosed) {
			return
		}
		if err != nil {
			fmt.Printf("TCP accept error: %s\n", err)
		}
		go t.handleConn(conn, false)
	}
}

// Spinned up for every request.
func (t *TCPTransport) handleConn(conn net.Conn, outbound bool) {
	var err error
	defer func() {
		logs.Logger.Infof("dropping peer connection: %s", err)
		conn.Close()
	}()

	peer := NewTCPPeer(conn, outbound) // outbound represents that request for connecton is sent by the client.
	if err = t.HandshakeFunc(peer); err != nil {
		return
	}
	// Function to be called on the peer.
	// make sure any data structure inside the fucntion
	// is race protected as multiple gorouitne will be acessing this.
	if t.OnPeer != nil {
		if err = t.OnPeer(peer); err != nil {
			return
		}
	}
	//This read loop here will run indefinitely.
	for {
		rpc := RPC{}
		// This will populate the rpc struct.
		err = t.Decoder.Decode(conn, &rpc)
		if err != nil {
			return
		}
		rpc.From = conn.RemoteAddr().String()

		// Do not handle here, stream data could be very huge, resulting in full memory blockage.
		if rpc.Stream {
			peer.wg.Add(1)
			logs.Logger.Infof("[%s] incoming stream, waiting...\n", conn.RemoteAddr())
			peer.wg.Wait()
			logs.Logger.Infof("[%s] stream closed, resuming read loop\n", conn.RemoteAddr())
			continue
		}
		t.rpcch <- rpc
	}
}

// func handleStream(conn net.Conn) error {
// 	buffer := make([]byte, 1024)
// 	for {
// 		n, err := conn.Read(buffer)
// 		if err != nil {
// 			if err == io.EOF {
// 				break
// 			}
// 			return fmt.Errorf("error reading from stream: %w", err)
// 		}
// 		// Process the stream data
// 		fmt.Printf("Stream data: %s\n", string(buffer[:n]))
// 	}
// 	return nil
// }
