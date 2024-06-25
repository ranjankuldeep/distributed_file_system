package fileserver

import (
	"bytes"
	"encoding/binary"
	"encoding/gob"
	"fmt"
	"io"
	"sync"
	"time"

	"github.com/ranjankuldeep/distributed_file_system/encrypt"
	"github.com/ranjankuldeep/distributed_file_system/logs"
	"github.com/ranjankuldeep/distributed_file_system/p2p"
	"github.com/ranjankuldeep/distributed_file_system/store"
)

type FileServerOpts struct {
	EncKey []byte
	// ID of the owner of the storage, which will be used to store all the files and folders at the location
	// so we can sync all the files if needed.
	ID                string
	StorageRoot       string
	PathTransformFunc store.PathTransformFunc
	Transport         p2p.Transport
	BootStrapNodes    []string
}
type FileServer struct {
	FileServerOpts
	FsStore *store.Store
	Quitch  chan struct{}

	PeerLock sync.Mutex
	Peers    map[string]p2p.Peer
}

// Message that is wired over.
type Message struct {
	Payload any
}

// Idenifier that payload will be of to store files.
type MessageStoreFile struct {
	ID   string
	Key  string
	Size int64
}

type MessageGetFile struct {
	ID  string
	Key string
}

type MessageDeleteFile struct {
	ID  string
	Key string
}

func NewFileServer(opts FileServerOpts) *FileServer {
	storeOpts := store.StoreOpts{
		Root:              opts.StorageRoot,
		PathTransformFunc: opts.PathTransformFunc,
	}
	if len(opts.ID) == 0 {
		opts.ID = encrypt.GenerateID()
	}
	return &FileServer{
		FileServerOpts: opts,
		FsStore:        store.NewStore(storeOpts),
		Quitch:         make(chan struct{}),
		Peers:          make(map[string]p2p.Peer),
		PeerLock:       sync.Mutex{},
	}
}

func (fs *FileServer) StartServer() error {
	if err := fs.Transport.ListenAndAccept(); err != nil {
		logs.Logger.Errorf("Failed to Listen")
		return err
	}
	fs.bootStrapNetwork() // Non Blocking
	fs.ReadLoop()         // Blocking
	return nil
}

func (fs *FileServer) StopServer() error {
	close(fs.Quitch)
	if err := fs.Transport.Close(); err != nil {
		logs.Logger.Error("Failed to stop the Server")
		return err
	}
	logs.Logger.Info("Quiting the File Server")
	return nil
}

func (fs *FileServer) Get(key string) (io.Reader, error) {
	if fs.FsStore.Has(fs.ID, key) {
		logs.Logger.Infof("[%s] serving file (%s) from local disk\n", fs.Transport.Addr(), key)
		_, r, err := fs.FsStore.Read(fs.ID, key)
		return r, err
	}

	logs.Logger.Infof("[%s] dont have file (%s) locally, fetching from network...\n", fs.Transport.Addr(), key)
	msg := Message{
		Payload: MessageGetFile{
			ID:  fs.ID,
			Key: key,
		},
	}

	if err := fs.BroadCast(&msg); err != nil {
		return nil, err
	}
	time.Sleep(time.Millisecond * 500)

	// Any peer over the network will start streaming the data.
	for _, peer := range fs.Peers {
		// First read the file size so we can limit the amount of bytes that we read
		// from the connection, so it will not keep hanging.
		var fileSize int64
		binary.Read(peer, binary.LittleEndian, &fileSize)
		_, err := fs.FsStore.WriteDecrypt(fs.EncKey, fs.ID, key, io.LimitReader(peer, fileSize))
		if err != nil {
			logs.Logger.Errorf("Unable to Write the Data Fetched by over the Network.")
		}
		logs.Logger.Infof("[%s] received (%d) bytes over the network from (%s)", fs.Transport.Addr(), fileSize, peer.RemoteAddr())
		peer.CloseStream()
	}

	_, r, err := fs.FsStore.Read(fs.ID, key)
	if err != nil {
		logs.Logger.Errorf("Cannot read from the store %s", key)
	}
	return r, err
}

func (fs *FileServer) Store(key string, r io.Reader) error {
	var (
		fileBuffer = new(bytes.Buffer)
		tee        = io.TeeReader(r, fileBuffer)
	)
	// 1. SAVE THE FILE TO THIS DISK and get the size of the file written locally (important for EOF on the network)
	size, err := fs.FsStore.Write(fs.ID, key, tee)
	if err != nil {
		return err
	}
	logs.Logger.Infof("[%s] Stored (%d) bytes to disk\n", fs.Transport.Addr(), size)
	logs.Logger.Info("Broadcasting to other Peers")

	msg := Message{
		// Payload if of message store file hinting remote server to store the data.
		Payload: MessageStoreFile{
			ID:  fs.ID,
			Key: key,
			// Hashed key will be stored on network as we don't want the other server to guess about the file by its name.
			// Specify the data size. (important)

			// Since we are first encrypting the data, it cost additional 16 byte of blockSize.
			Size: size + 16,
		},
	}
	// 2. BROADCAST THE FILE TO ALL KNOWN PEERS IN THE NETWORK.
	// 3. Broadcast the Key over the network
	if err := fs.BroadCast(&msg); err != nil {
		return err
	}

	time.Sleep(time.Millisecond * 1000)

	peers := []io.Writer{}
	for _, peer := range fs.Peers {
		peers = append(peers, peer)
	}
	mw := io.MultiWriter(peers...)
	mw.Write([]byte{p2p.IncomingStream})
	// Send the encrypted message data over the network.
	if _, err := encrypt.CopyEncrypt(fs.EncKey, fileBuffer, mw); err != nil {
		logs.Logger.Errorf("Failed to stream data %v", err)
		return err
	}
	return nil
}

// Todo: IMplement it.
// 1. Delete it Locally and through out the network.
func (fs *FileServer) Delete(key string) error {
	if err := fs.FsStore.Delete(fs.ID, key); err != nil {
		logs.Logger.Errorf("Error Deleting Key Locally %s", key)
	}
	msg := Message{
		Payload: MessageDeleteFile{
			ID:  fs.ID,
			Key: key,
		},
	}
	logs.Logger.Infof("BroadCasting the Delete Request over the network %s", key)
	if err := fs.BroadCast(&msg); err != nil {
		logs.Logger.Errorf("Error Broadcasting the Delete Request: %+v", err)
		return err
	}
	logs.Logger.Info("Deleting Data from Network")
	return nil
}

// Broadcasting the message.
func (fs *FileServer) BroadCast(msg *Message) error {
	buf := new(bytes.Buffer)
	if err := gob.NewEncoder(buf).Encode(msg); err != nil {
		return err
	}
	for _, peer := range fs.Peers {
		if err := peer.Send([]byte{p2p.IncomingMessage}); err != nil { // First send the incoming message after encoding.
			logs.Logger.Error(err)
			return err
		}
		// Then send the message to the peer.
		if err := peer.Send(buf.Bytes()); err != nil {
			logs.Logger.Error(err)
			return err
		}
	}
	return nil
}

func (fs *FileServer) Stop() error {
	fs.Quitch <- struct{}{}
	return nil
}

// Make sure that only a single go routine can change the
// peers map at a time
// map read is optimized for concurrent read but not map write.
func (s *FileServer) OnPeer(p p2p.Peer) error {
	s.PeerLock.Lock()
	defer s.PeerLock.Unlock()

	s.Peers[p.RemoteAddr().String()] = p
	logs.Logger.Infof("connected with remote %s", p.RemoteAddr().String())
	return nil
}

func (fs *FileServer) ReadLoop() {
	// Keeps on looping for ever unitl quit. Blockin in nature.
	// Unless select it will again keeps on listenitng even if a channel has been hadled once.
	defer func() {
		logs.Logger.Info("File Server Stopped")
		fs.Transport.Close()
	}()
	for {
		select {
		case rpc := <-fs.Transport.Consume():
			var m Message // This is what recived over the wire.
			if err := gob.NewDecoder(bytes.NewReader(rpc.Payload)).Decode(&m); err != nil {
				logs.Logger.Errorf("Decoding Error %+v", err)
			}
			if err := fs.handleMessage(rpc.From, &m); err != nil {
				logs.Logger.Error(err)
			}
		case <-fs.Quitch:
			logs.Logger.Info("User Quit Action")
			return
		}
	}
}

func (fs *FileServer) handleMessage(from string, msg *Message) error {
	switch v := msg.Payload.(type) {
	case MessageStoreFile:
		logs.Logger.Infof("Received key for Storing %+v\n from %s", v, from)
		return fs.handleMessageStoreFile(from, &v)
	case MessageGetFile:
		return fs.handleMessageGetFile(from, v)
	case MessageDeleteFile:
		return fs.handleMessageDeleteFile(from, v)
	}
	return nil
}

func (fs *FileServer) handleMessageStoreFile(from string, msg *MessageStoreFile) error {
	// Secuirty check.
	peer, ok := fs.Peers[from]
	if !ok {
		return fmt.Errorf("peer (%s) could not be found in the peer list", from)
	}
	// TODO: Add decryptor

	// A limit reader is necassary as over the network
	// when reading from the connection directly it will not send the EOF.
	// Which results in keep waiting until EOF.
	n, err := fs.FsStore.Write(msg.ID, msg.Key, io.LimitReader(peer, msg.Size))
	if err != nil {
		logs.Logger.Errorf(err.Error())
		return err
	}

	logs.Logger.Infof("[%s] written %d bytes to disk\n", fs.Transport.Addr(), n)
	peer.CloseStream() // Will trigger the read loop again for the connection.
	return nil
}

func (s *FileServer) handleMessageGetFile(from string, msg MessageGetFile) error {
	peer, ok := s.Peers[from]
	if !ok {
		return fmt.Errorf("peer %s not in map", from)
	}
	if !s.FsStore.Has(msg.ID, msg.Key) {
		return fmt.Errorf("[%s] need to serve file (%s) but it does not exist on disk", s.Transport.Addr(), msg.Key)
	}
	fmt.Printf("[%s] serving file (%s) over the network\n", s.Transport.Addr(), msg.Key)
	fileSize, r, err := s.FsStore.Read(msg.ID, msg.Key)
	if err != nil {
		return err
	}
	if rc, ok := r.(io.ReadCloser); ok {
		fmt.Println("closing readCloser")
		defer rc.Close()
	}

	// 1. Send the "incomingStream" byte to the peer and then
	// 2. Send the file size as an int64.
	// 3. Stream the data over the network.
	peer.Send([]byte{p2p.IncomingStream})
	binary.Write(peer, binary.LittleEndian, fileSize)
	n, err := io.Copy(peer, r)
	if err != nil {
		return err
	}
	fmt.Printf("[%s] written (%d) bytes over the network to %s\n", s.Transport.Addr(), n, from)
	return nil
}

func (fs *FileServer) handleMessageDeleteFile(from string, msg MessageDeleteFile) error {
	logs.Logger.Infof("Recived Delete Request from %+v", from)
	// Security Check
	_, ok := fs.Peers[from]
	if !ok {
		return fmt.Errorf("peer (%s) could not be found in the peer list", from)
	}
	return fs.FsStore.Delete(msg.ID, msg.Key)
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

func init() {
	gob.Register(MessageStoreFile{})
	gob.Register(MessageGetFile{})
	gob.Register(MessageDeleteFile{})
}
