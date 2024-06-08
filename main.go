package main

import (
	"time"

	"github.com/labstack/gommon/log"
	"github.com/ranjankuldeep/distributed_file_system/fileserver"
	"github.com/ranjankuldeep/distributed_file_system/p2p"
	"github.com/ranjankuldeep/distributed_file_system/store"
)

func main() {
	s1 := makeServer(":3000", "")
	s2 := makeServer(":4000", ":3000")

	// START THE FILE SERVER AND DIAL THE BOOT STRAP NODES.
	go func() {
		log.Fatal(s1.Start())
	}()
	time.Sleep(500 * time.Millisecond)
	s2.Start()
}

func makeServer(listenAddr string, nodes ...string) *fileserver.FileServer {
	tcptransportOpts := p2p.TCPTransportOpts{
		ListenAddr:    listenAddr,
		HandshakeFunc: p2p.NOPHandshakeFunc,
		Decoder:       p2p.DefaultDecoder{},
	}
	tcpTransport := p2p.NewTCPTransport(tcptransportOpts)

	fileServerOpts := fileserver.FileServerOpts{
		StorageRoot:       listenAddr + "_network",
		PathTransformFunc: store.CASPathTransformFunc,
		Transport:         tcpTransport,
		BootStrapNodes:    nodes,
	}

	s := fileserver.NewFileServer(fileServerOpts)
	tcpTransport.OnPeer = s.OnPeer
	return s
}
