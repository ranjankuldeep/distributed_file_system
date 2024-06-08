package fileserver

import (
	"fmt"

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
