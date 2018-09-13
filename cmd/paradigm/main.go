package main

import (
	"fmt"
	"gopkg.in/urfave/cli.v1"
	"os"
	"time"
	"sort"
	"os/user"
	"path/filepath"
	"runtime"
	"github.com/sirupsen/logrus"
	"github.com/paradigm-network/paradigm/core"
	"github.com/paradigm-network/paradigm/storage"
	"github.com/paradigm-network/paradigm/network/tcp"
	"github.com/paradigm-network/paradigm/proxy"
	"github.com/paradigm-network/paradigm/network/peer"
	"github.com/paradigm-network/paradigm/common/crypto"
	"github.com/paradigm-network/paradigm/version"
	"github.com/paradigm-network/paradigm/common/log"
)

var (
	DataDirFlag = cli.StringFlag{
		Name:  "datadir",
		Usage: "Directory for the configuration",
		Value: defaultDataDir(),
	}
	NodeAddressFlag = cli.StringFlag{
		Name:  "node_addr",
		Usage: "IP:Port to bind Paradigm",
		Value: "127.0.0.1:1337",
	}
	Gw2AddressFlag = cli.StringFlag{
		Name:  "gw2_addr",
		Usage: "IP:Port to bind Gw2 module",
		Value: "127.0.0.1:9000",
	}
	Fn2AddressFlag = cli.StringFlag{
		Name:  "fn2_address",
		Usage: "IP:Port to bind Fn2 module",
		Value: "127.0.0.1:8000",
	}
	OnlyAccretion = cli.BoolFlag{
		Name:  "only_accretion",
		Usage: "Only if join accretion network",
	}
	NoClientFlag = cli.BoolFlag{
		Name:  "no_client",
		Usage: "Run Paradigm with dummy in-memory App client",
	}
	ProxyAddressFlag = cli.StringFlag{
		Name:  "proxy_addr",
		Usage: "IP:Port to bind Proxy Server",
		Value: "127.0.0.1:1338",
	}
	ClientAddressFlag = cli.StringFlag{
		Name:  "client_addr",
		Usage: "IP:Port of Client App",
		Value: "127.0.0.1:1339",
	}
	ServiceAddressFlag = cli.StringFlag{
		Name:  "service_addr",
		Usage: "IP:Port of HTTP Service",
		Value: "127.0.0.1:8000",
	}
	LogLevelFlag = cli.StringFlag{
		Name:  "log_level",
		Usage: "debug, info, warn, error, fatal, panic",
		Value: "debug",
	}
	HeartbeatFlag = cli.IntFlag{
		Name:  "heartbeat",
		Usage: "Heartbeat timer milliseconds (time between gossips)",
		Value: 1000,
	}
	MaxPoolFlag = cli.IntFlag{
		Name:  "max_pool",
		Usage: "Max number of pooled connections",
		Value: 2,
	}
	TcpTimeoutFlag = cli.IntFlag{
		Name:  "tcp_timeout",
		Usage: "TCP timeout milliseconds",
		Value: 1000,
	}
	CacheSizeFlag = cli.IntFlag{
		Name:  "cache_size",
		Usage: "Number of items in LRU caches",
		Value: 500,
	}
	SyncLimitFlag = cli.IntFlag{
		Name:  "sync_limit",
		Usage: "Max number of events for sync",
		Value: 1000,
	}
	StoreFlag = cli.StringFlag{
		Name:  "store",
		Usage: "badger",
		Value: "badger",
	}
	StorePathFlag = cli.StringFlag{
		Name:  "store_path",
		Usage: "File containing the store database",
		Value: defaultBadgerDir(),
	}
)

func main() {
	app := cli.NewApp()
	app.Name = "paradigm"
	app.Usage = "Paradigm Network"
	app.HideVersion = true //there is a special command to print the version
	app.Commands = []cli.Command{
		{
			Name:   "keygen",
			Usage:  "Dump new key pair",
			Action: keygen,
		},
		{
			Name:   "run",
			Usage:  "Run paradigm",
			Action: run,
			Flags: []cli.Flag{
				OnlyAccretion,
				DataDirFlag,
				NodeAddressFlag,
				Gw2AddressFlag,
				Fn2AddressFlag,
				NoClientFlag,
				ProxyAddressFlag,
				ClientAddressFlag,
				ServiceAddressFlag,
				LogLevelFlag,
				HeartbeatFlag,
				MaxPoolFlag,
				TcpTimeoutFlag,
				CacheSizeFlag,
				SyncLimitFlag,
				StoreFlag,
				StorePathFlag,
			},
		},
		{
			Name:   "version",
			Usage:  "Show version info",
			Action: printVersion,
		},
	}
	app.Run(os.Args)
}

func keygen(c *cli.Context) error {
	log.InitRotateWriter("./")
	logger := log.GetLogger("main")
	pemDump, err := crypto.GeneratePemKey()
	if err != nil {
		logger.Error().Msg("Error generating PemDump")
		os.Exit(2)
	}

	//fmt.Println("PublicKey:")
	logger.Info().Str("PublicKey",pemDump.PublicKey)
	//fmt.Println(pemDump.PublicKey)
	//fmt.Println("PrivateKey:")
	//fmt.Println(pemDump.PrivateKey)
	logger.Info().Str("PrivateKey",pemDump.PrivateKey)

	return nil
}
func printVersion(c *cli.Context) error {
	fmt.Println(version.Version)
	return nil
}

func run(c *cli.Context) error {
	logger := logrus.New()
	logger.SetLevel(logrus.DebugLevel)
	onlyAccretion := c.Bool(OnlyAccretion.Name)
	datadir := c.String(DataDirFlag.Name)
	addr := c.String(NodeAddressFlag.Name)
	gw2Address := c.String(Gw2AddressFlag.Name)
	fn2Address := c.String(Fn2AddressFlag.Name)
	noclient := c.Bool(NoClientFlag.Name)
	proxyAddress := c.String(ProxyAddressFlag.Name)
	clientAddress := c.String(ClientAddressFlag.Name)
	serviceAddress := c.String(ServiceAddressFlag.Name)
	heartbeat := c.Int(HeartbeatFlag.Name)
	maxPool := c.Int(MaxPoolFlag.Name)
	tcpTimeout := c.Int(TcpTimeoutFlag.Name)
	cacheSize := c.Int(CacheSizeFlag.Name)
	syncLimit := c.Int(SyncLimitFlag.Name)
	storeType := c.String(StoreFlag.Name)
	storePath := c.String(StorePathFlag.Name)

	logger.WithFields(logrus.Fields{
		"only_accretion": onlyAccretion,
		"datadir":        datadir,
		"gw2_addr":       gw2Address,
		"fn2_addr":       fn2Address,
		"node_addr":      addr,
		"no_client":      noclient,
		"proxy_addr":     proxyAddress,
		"client_addr":    clientAddress,
		"service_addr":   serviceAddress,
		"heartbeat":      heartbeat,
		"max_pool":       maxPool,
		"tcp_timeout":    tcpTimeout,
		"cache_size":     cacheSize,
		"store":          storeType,
		"store_path":     storePath,
	}).Debug("RUN")

	conf := core.NewConfig(onlyAccretion, time.Duration(heartbeat)*time.Millisecond,
		time.Duration(tcpTimeout)*time.Millisecond,
		cacheSize, syncLimit, storeType, storePath, logger, gw2Address, fn2Address)

	// Create the PEM key
	pemKey := crypto.NewPemKey(datadir)

	// Try a read
	key, err := pemKey.ReadKey()
	if err != nil {
		return cli.NewExitError(err, 1)
	}

	// Create the peer store
	peerStore := peer.NewJSONPeers(datadir)

	// Try a read
	peers, err := peerStore.Peers()
	if err != nil {
		return cli.NewExitError(err, 1)
	}

	// There should be at least two peers
	if len(peers) < 2 {
		return cli.NewExitError("participants.json should define at least two peers", 1)
	}

	//Sort peers by public key and assign them an int ID
	//Every participant in the network will run this and assign the same IDs
	sort.Sort(peer.ByPubKey(peers))
	pmap := make(map[string]int)
	for i, p := range peers {
		pmap[p.PubKeyHex] = i
	}

	//Find the ID of this node
	nodePub := fmt.Sprintf("0x%X", crypto.FromECDSAPub(&key.PublicKey))
	nodeID := pmap[nodePub]

	logger.WithFields(logrus.Fields{
		"pmap": pmap,
		"id":   nodeID,
	}).Debug("PARTICIPANTS")

	//Instantiate the Store (badger)
	var store storage.Store
	var needBootstrap bool
	if _, err := os.Stat(conf.StorePath); err == nil {
		logger.Debug("loading badger store from existing database")
		store, err = storage.LoadBadgerStore(conf.CacheSize, conf.StorePath)
		if err != nil {
			return cli.NewExitError(
				fmt.Sprintf("failed to load BadgerStore from existing file: %s", err),
				1)
		}
		needBootstrap = true
	} else {
		//Otherwise create a new one
		logger.Debug("creating new badger store from fresh database")
		store, err = storage.NewBadgerStore(pmap, conf.CacheSize, conf.StorePath)
		if err != nil {
			return cli.NewExitError(
				fmt.Sprintf("failed to create new BadgerStore: %s", err),
				1)
		}
	}

	trans, err := tcp.NewTCPTransport(addr,
		nil, maxPool, conf.TCPTimeout, logger)
	if err != nil {
		return cli.NewExitError(err, 1)
	}

	var prox proxy.AppProxy
	prox = proxy.NewInmemAppProxy(logger)
	//todo impl. if no_client

	node := core.NewNode(conf, nodeID, key, peers, store, trans, prox)
	if err := node.Init(needBootstrap); err != nil {
		return cli.NewExitError(
			fmt.Sprintf("failed to initialize node: %s", err),
			1)
	}

	//serviceServer := service.NewService(serviceAddress, node, logger)
	//go serviceServer.Serve()

	node.Run(true)

	return nil
}

func defaultBadgerDir() string {
	dataDir := defaultDataDir()
	if dataDir != "" {
		return filepath.Join(dataDir, "badger_db")
	}
	return ""
}

func defaultDataDir() string {
	// Try to place the data folder in the user's home dir
	home := homeDir()
	if home != "" {
		if runtime.GOOS == "darwin" {
			return filepath.Join(home, ".paradigm")
		} else if runtime.GOOS == "windows" {
			return filepath.Join(home, "AppData", "Roaming", "PARADIGM")
		} else {
			return filepath.Join(home, ".paradigm")
		}
	}
	// As we cannot guess a stable location, return empty and handle later
	return ""
}

func homeDir() string {
	if home := os.Getenv("HOME"); home != "" {
		return home
	}
	if usr, err := user.Current(); err == nil {
		return usr.HomeDir
	}
	return ""
}

func logLevel(l string) logrus.Level {
	switch l {
	case "debug":
		return logrus.DebugLevel
	case "info":
		return logrus.InfoLevel
	case "warn":
		return logrus.WarnLevel
	case "error":
		return logrus.ErrorLevel
	case "fatal":
		return logrus.FatalLevel
	case "panic":
		return logrus.PanicLevel
	default:
		return logrus.DebugLevel
	}
}
