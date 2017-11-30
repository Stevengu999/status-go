package main

import (
	"flag"
	"fmt"
	"log"

	"github.com/status-im/status-go/geth/api"
	"github.com/status-im/status-go/geth/params"
)

var (
	prodMode    = flag.Bool("production", false, "Whether production settings should be loaded")
	nodeKeyFile = flag.String("nodekey", "", "P2P node key file (private key)")
	dataDir     = flag.String("datadir", "wnode-status-data", "Data directory for the databases and keystore")
	networkID   = flag.Int("networkid", params.RopstenNetworkID, "Network identifier (integer, 1=Homestead, 3=Ropsten, 4=Rinkeby)")
	httpEnabled = flag.Bool("http", false, "HTTP RPC enpoint enabled (default: false)")
	httpPort    = flag.Int("httpport", params.HTTPPort, "HTTP RPC server's listening port")
	ipcEnabled  = flag.Bool("ipc", false, "IPC RPC enpoint enabled")

	// wnode specific flags
	echo           = flag.Bool("echo", true, "Echo mode, prints some arguments for diagnostics")
	bootstrap      = flag.Bool("bootstrap", true, "Don't actively connect to peers, wait for incoming connections")
	notify         = flag.Bool("notify", false, "Node is capable of sending Push Notifications")
	forward        = flag.Bool("forward", false, "Only forward messages, neither send nor decrypt messages")
	mailserver     = flag.Bool("mailserver", false, "Delivers expired messages on demand")
	identity       = flag.String("identity", "", "Protocol identity file (private key used for asymmetric encryption)")
	password       = flag.String("password", "", "Password file (password is used for symmetric encryption)")
	wsport         = flag.Int("port", params.WhisperPort, "Whisper node's listening port")
	pow            = flag.Float64("pow", params.WhisperMinimumPoW, "PoW for messages to be added to queue, in float format")
	ttl            = flag.Int("ttl", params.WhisperTTL, "Time to live for messages, in seconds")
	injectAccounts = flag.Bool("injectaccounts", true, "Whether test account should be injected or not")
	firebaseAuth   = flag.String("firebaseauth", "", "FCM Authorization Key used for sending Push Notifications")
)

var backend *api.StatusBackend

func main() {
	flag.Parse()

	config, err := makeNodeConfig()
	if err != nil {
		log.Fatalf("Making config failed: %v", err)
	}

	mailNode := "enode://859bc9f0919a8bffa5a73bb8f55997aa990bf38771dd09b037db37c9efc311fc069ba146511c2b0a98bff14ed29f46982872c64b59da0f0eee0a9edaac17f599@[::]:39544?discport=0"
	//mailNode := "enode://7ef1407cccd16c90d01bfd8245b4b93c2f78e7d19769dc310cf46628d614d8aa7259005ef532d426092fa14ef0010ff7d83d5bfd108614d447b0b07499ffda78@127.0.0.1:30303"
	config.BootClusterConfig.BootNodes = append([]string{mailNode}, config.BootClusterConfig.BootNodes...)

	printHeader(config)

	if *injectAccounts {
		if err := LoadTestAccounts(config.DataDir); err != nil {
			log.Fatalf("Failed to load test accounts: %v", err)
		}
	}

	b := api.NewStatusBackend()
	started, err := b.StartNode(config)
	if err != nil {
		log.Fatalf("Node start failed: %v", err)
		return
	}

	// wait till node is started
	<-started
	if *mailserver == true {
		n1, _ := b.NodeManager().Node()
		fmt.Printf("------------------------n1.Server().NodeInfo().Enode: '%s'", n1.Server().NodeInfo().Enode)
	}

	if *injectAccounts {
		if err := InjectTestAccounts(b.NodeManager()); err != nil {
			log.Fatalf("Failed to inject accounts: %v", err)
		}
	}

	// wait till node has been stopped
	node, err := b.NodeManager().Node()
	if err != nil {
		log.Fatalf("Getting node failed: %v", err)
		return
	}

	backend = b
	log.Println("Wnode started!", config)

	node.Wait()

	log.Println("Wnode stopped!")
}

// printHeader prints command header
func printHeader(config *params.NodeConfig) {
	fmt.Println("Starting Whisper/5 node..")
	if config.WhisperConfig.EchoMode {
		fmt.Printf("Whisper Config: %s\n", config.WhisperConfig)
	}
}
