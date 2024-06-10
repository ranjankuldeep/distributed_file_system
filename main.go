package main

import (
	"bytes"
	"io/ioutil"
	"time"

	"github.com/labstack/gommon/log"
	"github.com/ranjankuldeep/distributed_file_system/encrypt"
	"github.com/ranjankuldeep/distributed_file_system/fileserver"
	"github.com/ranjankuldeep/distributed_file_system/logs"
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

	time.Sleep(1 * time.Second)
	go s2.Start()
	time.Sleep(1 * time.Second)

	key := "myfuckingkey"
	data := bytes.NewReader([]byte("My Big Data File here!"))
	s2.Store(key, data)

	// DELETE THE DATA LOCALLY
	if err := s2.FsStore.Delete("1234", key); err != nil {
		logs.Logger.Errorf("Unable to delete key localy %s", err)
		return
	}
	r, err := s2.Get(key)
	if err != nil {
		logs.Logger.Fatalf(err.Error())
	}
	b, err := ioutil.ReadAll(r)
	if err != nil {
		log.Fatal(err)
	}

	logs.Logger.Info(string(b))
	select {}
}

func makeServer(listenAddr string, nodes ...string) *fileserver.FileServer {
	tcptransportOpts := p2p.TCPTransportOpts{
		ListenAddr:    listenAddr,
		HandshakeFunc: p2p.NOPHandshakeFunc,
		Decoder:       p2p.DefaultDecoder{},
	}
	tcpTransport := p2p.NewTCPTransport(tcptransportOpts)

	fileServerOpts := fileserver.FileServerOpts{
		EncKey:            encrypt.NewEncryptionKey(),
		StorageRoot:       listenAddr + "_network",
		PathTransformFunc: store.CASPathTransformFunc,
		Transport:         tcpTransport,
		BootStrapNodes:    nodes,
	}

	s := fileserver.NewFileServer(fileServerOpts)
	tcpTransport.OnPeer = s.OnPeer
	return s
}
