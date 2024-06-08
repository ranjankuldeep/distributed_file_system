package main

import (
	"time"

	"github.com/ranjankuldeep/distributed_file_system/fileserver"
	"github.com/ranjankuldeep/distributed_file_system/logs"
	"github.com/ranjankuldeep/distributed_file_system/p2p"
	"github.com/ranjankuldeep/distributed_file_system/store"
)

func main() {
	listenAddr := ":3000"
	tcptransportOpts := p2p.TCPTransportOpts{
		ListenAddr:    listenAddr,
		HandshakeFunc: p2p.NOPHandshakeFunc,
		Decoder:       p2p.DefaultDecoder{},
		// Todo onpeer function
	}
	tcpTransport := p2p.NewTCPTransport(tcptransportOpts)

	fileServerOpts := fileserver.FileServerOpts{
		StorageRoot:       listenAddr + "_network",
		PathTransformFunc: store.CASPathTransformFunc,
		Transport:         tcpTransport,
	}

	s := fileserver.NewFileServer(fileServerOpts)

	go func() {
		time.Sleep(3 * time.Second)
		s.Stop()
	}()

	if err := s.Start(); err != nil {
		logs.Logger.Fatalf("Failed to start the file server %v", err)
	}

}
