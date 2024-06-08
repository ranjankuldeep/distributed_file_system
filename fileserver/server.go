package fileserver

import (
	"github.com/ranjankuldeep/distributed_file_system/p2p"
	"github.com/ranjankuldeep/distributed_file_system/store"
)

type FileServerOpts struct {
	ListenAddr        string
	StorageRoot       string
	PathTransformFunc store.PathTransformFunc
	Tranport          p2p.Transport
}
type FileServer struct {
	FileServerOpts
	store *store.Store
}

func NewFileServer(opts FileServerOpts) *FileServer {
	storeOpts := store.StoreOpts{
		Root:              opts.StorageRoot,
		PathTransformFunc: opts.PathTransformFunc,
	}
	return &FileServer{
		FileServerOpts: opts,
		store:          store.NewStore(storeOpts),
	}
}
