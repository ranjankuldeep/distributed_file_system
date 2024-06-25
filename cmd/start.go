package cmd

import (
	"sync"

	"github.com/google/uuid"
	"github.com/ranjankuldeep/distributed_file_system/encrypt"
	"github.com/ranjankuldeep/distributed_file_system/fileserver"
	"github.com/ranjankuldeep/distributed_file_system/logs"
	"github.com/ranjankuldeep/distributed_file_system/p2p"
	"github.com/ranjankuldeep/distributed_file_system/store"
	"github.com/spf13/cobra"
)

// start cmd needs two flag
var randomUserName = uuid.New().String()

var (
	UserName   string
	ListenPort string

	DefaultUserName = randomUserName
	server          *fileserver.FileServer
	serverMu        sync.Mutex
	serverReady     = make(chan struct{})
	stopServer      = make(chan struct{})
)
var (
	startCmd = &cobra.Command{
		Use:   "start",
		Short: "Start the server",
		Long:  "Spin Up the Server on the Specified Port",
		Args:  cobra.ArbitraryArgs,
		RunE: func(cmd *cobra.Command, args []string) error {
			serverMu.Lock()
			defer serverMu.Unlock()

			// name, _ := cmd.Flags().GetString("name")
			// port, _ := cmd.Flags().GetString("port")

			server = makeServer(UserName, ListenPort, args...)
			go server.StartServer()
			close(serverReady) // Signal that the server is ready
			logs.Logger.Infof("Server started successfully, Username:%s, ListenPort:%s", UserName, ListenPort)

			<-stopServer
			logs.Logger.Info("Stopping server...")
			return nil
		},
	}
)

func makeServer(userId string, listenAddr string, nodes ...string) *fileserver.FileServer {
	tcptransportOpts := p2p.TCPTransportOpts{
		ListenAddr:    listenAddr,
		HandshakeFunc: p2p.NOPHandshakeFunc,
		Decoder:       p2p.DefaultDecoder{},
	}
	tcpTransport := p2p.NewTCPTransport(tcptransportOpts)

	fileServerOpts := fileserver.FileServerOpts{
		ID:                userId,
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

func init() {
	startCmd.Flags().StringVarP(&UserName, "name", "n", UserName, "Your userName")
	startCmd.Flags().StringVarP(&ListenPort, "port", "p", ":4000", "Specify Start Server Port (default :4000)")
}
