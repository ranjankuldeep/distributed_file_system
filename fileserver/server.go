package fileserver

import (
	"fmt"
	"sync"

	"github.com/ranjankuldeep/distributed_file_system/logs"
	"github.com/ranjankuldeep/distributed_file_system/p2p"
	"github.com/ranjankuldeep/distributed_file_system/store"
)

type FileServerOpts struct {
	StorageRoot       string
	PathTransformFunc store.PathTransformFunc
	Transport         p2p.Transport
	BootStrapNodes    []string
}
type FileServer struct {
	FileServerOpts
	store  *store.Store
	quitch chan struct{}

	peerLock sync.Mutex
	peers    map[string]p2p.Peer
}

func NewFileServer(opts FileServerOpts) *FileServer {
	storeOpts := store.StoreOpts{
		Root:              opts.StorageRoot,
		PathTransformFunc: opts.PathTransformFunc,
	}
	return &FileServer{
		FileServerOpts: opts,
		store:          store.NewStore(storeOpts),
		quitch:         make(chan struct{}),
		peers:          make(map[string]p2p.Peer),
		peerLock:       sync.Mutex{},
	}
}

func (fs *FileServer) Start() error {
	if err := fs.Transport.ListenAndAccept(); err != nil {
		return err
	}
	fs.bootStrapNetwork()
	fs.loop() // blocks for the reading of data
	return nil
}

func (fs *FileServer) Stop() error {
	fs.quitch <- struct{}{}
	return nil
}

// Make sure that only a single go routine can change the
// peers map at a time
// map read is optimized for concurrent read but not map write.
func (s *FileServer) OnPeer(p p2p.Peer) error {
	s.peerLock.Lock()
	defer s.peerLock.Unlock()

	s.peers[p.RemoteAddr().String()] = p
	logs.Logger.Infof("connected with remote %s", p.RemoteAddr().String())
	return nil
}

func (fs *FileServer) loop() {
	// Keeps on looping for ever unitl quit. Blockin in nature.
	// Unless select it will again keeps on listenitng even if a channel has been hadled once.
	defer func() {
		logs.Logger.Info("File Server Stopped due to user quit action")
	}()
	for {
		select {
		case data := <-fs.Transport.Consume():
			fmt.Println(data)
		case <-fs.quitch:
			return
		}
	}
}

func (fs *FileServer) bootStrapNetwork() error {
	for _, addr := range fs.BootStrapNodes {
		if len(addr) == 0 {
			continue
		}
		go func(addr string) {
			logs.Logger.Infof("attemting to connect with remote:%s", addr)
			if err := fs.Transport.Dial(addr); err != nil {
				logs.Logger.Errorf("Error BootStraping Network %v", err)
			}
		}(addr)
	}
	return nil
}
