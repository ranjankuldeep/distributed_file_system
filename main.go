package main

import (
	"github.com/google/uuid"
	"github.com/ranjankuldeep/distributed_file_system/cmd"
	"github.com/ranjankuldeep/distributed_file_system/encrypt"
	"github.com/ranjankuldeep/distributed_file_system/fileserver"
	"github.com/ranjankuldeep/distributed_file_system/p2p"
	"github.com/ranjankuldeep/distributed_file_system/store"
)

func main() {
	cmd.Execute()
	// // needed port to run and the peers port
	// user1Id := uuid.New()
	// user2Id := uuid.New()

	// s1 := makeServer(user2Id, ":3000", "")
	// s2 := makeServer(user1Id, ":4000", ":3000")

	// // START THE FILE SERVER AND DIAL THE BOOT STRAP NODES.
	// go func() {
	// 	log.Fatal(s1.StartServer())
	// }()

	// time.Sleep(1 * time.Second)
	// go s2.StartServer()
	// // 1. Dial all the network.
	// // 2. Start the read loop listening from the channel.
	// time.Sleep(1 * time.Second)

	// // need key and data to tbe stored
	// path := "./big.txt"
	// key, err := util.GetFileName(path)
	// if err != nil {
	// 	logs.Logger.Errorf("Error reading file stat from the specified path %s", path)
	// 	log.Fatal(err)
	// }
	// // read the file
	// file, err := util.ReadFile(path)
	// if err != nil {
	// 	logs.Logger.Errorf("Error reading file from the specified path %s", path)
	// 	file.Close()
	// 	log.Fatal(err)

	// }
	// // data := bytes.NewReader([]byte("My Big Data File here!"))
	// s2.Store(key, file)
	// file.Close()
	// DELETE THE DATA LOCALLY directly from the store
	// if err := s2.FsStore.Delete(user1Id.String(), key); err != nil {
	// 	logs.Logger.Errorf("deleted")
	// }
	// if err := s2.Delete(key); err != nil {
	// 	logs.Logger.Errorf("Error deleting")
	// }
	// time.Sleep(1 * time.Millisecond)
	// r, err := s2.Get(key) // s2 will now fetch now from the network
	// if err != nil {
	// 	logs.Logger.Fatalf(err.Error())
	// }
	// b, err := ioutil.ReadAll(r)
	// if err != nil {
	// 	log.Fatal(err)
	// }
	// logs.Logger.Info(string(b))
	// select {}
}

func makeServer(userId uuid.UUID, listenAddr string, nodes ...string) *fileserver.FileServer {
	tcptransportOpts := p2p.TCPTransportOpts{
		ListenAddr:    listenAddr,
		HandshakeFunc: p2p.NOPHandshakeFunc,
		Decoder:       p2p.DefaultDecoder{},
	}
	tcpTransport := p2p.NewTCPTransport(tcptransportOpts)

	fileServerOpts := fileserver.FileServerOpts{
		ID:                userId.String(),
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
