package fileserver

import (
	"bytes"
	"encoding/gob"
	"io"
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
	fs.bootStrapNetwork() // Non Blocking
	fs.ReadLoop()         // Blocking
	return nil
}
func (fs *FileServer) Store(id string, key string, r io.Reader) error {
	buf := new(bytes.Buffer)
	tee := io.TeeReader(r, buf)
	// 1. SAVE THE FILE TO THIS DISK.
	if _, err := fs.store.Write(id, key, tee); err != nil {
		return err
	}
	// 2. BROADCAST THE FILE TO ALL KNONW PEERS IN THE NETWORK.
	p := &DataMessage{
		Key:  key,
		Data: buf.Bytes(),
	}

	logs.Logger.Infof("Stored %v", buf.String())
	return fs.BroadCast(&Message{
		From:    "todo",
		Payload: p,
	})
}

type Message struct {
	From    string
	Payload any
}

type DataMessage struct {
	Key  string
	Data []byte
}

func (fs *FileServer) BroadCast(msg *Message) error {
	peers := []io.Writer{}

	for _, peer := range fs.peers {
		peers = append(peers, peer)
	}
	mw := io.MultiWriter(peers...)
	return gob.NewEncoder(mw).Encode(msg)
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

func (fs *FileServer) ReadLoop() {
	// Keeps on looping for ever unitl quit. Blockin in nature.
	// Unless select it will again keeps on listenitng even if a channel has been hadled once.
	defer func() {
		logs.Logger.Info("File Server Stopped due to user quit action")
	}()
	for {
		select {
		case msg := <-fs.Transport.Consume():
			logs.Logger.Info(msg.Payload)
			var m Message
			if err := gob.NewDecoder(bytes.NewReader(msg.Payload)).Decode(&m); err != nil {
				logs.Logger.Error(err)
			}
			if err := fs.handleMessage(&m); err != nil {
				logs.Logger.Error(err)
			}
		case <-fs.quitch:
			return
		}
	}
}

func (fs *FileServer) handleMessage(msg *Message) error {
	switch v := msg.Payload.(type) {
	case *DataMessage:
		logs.Logger.Infof("received data %+v\n", v)
	}
	return nil
}

// Non blocking
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
