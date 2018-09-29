package main

import (
	"fmt"
	"github.com/paradigm-network/paradigm/accounts/keystore"
	"github.com/paradigm-network/paradigm/common/crypto"
	"github.com/paradigm-network/paradigm/common/log"
	"github.com/paradigm-network/paradigm/config"
	"github.com/paradigm-network/paradigm/core"
	"github.com/paradigm-network/paradigm/network/peer"
	"github.com/paradigm-network/paradigm/network/tcp"
	"github.com/paradigm-network/paradigm/proxy"
	"github.com/paradigm-network/paradigm/storage"
	"github.com/paradigm-network/paradigm/version"
	"gopkg.in/urfave/cli.v1"
	"io/ioutil"
	"os"
	"os/user"
	"path/filepath"
	"runtime"
	"sort"
	"time"
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
	SequentiaAddress = cli.StringFlag{
		Name:  "seq_address",
		Usage: "IP:Port to bind Senqutia module",
		Value: "127.0.0.1:8090",
	}
	OnlyAccretion = cli.BoolFlag{
		Name:  "only_accretion",
		Usage: "Only if join accretion network",
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
	KeyStorePathFlag = cli.StringFlag{
		Name:  "keystore_path",
		Usage: "File containing the store keyfile",
		Value: defaultKeyStoreDir(),
	}
	PwdFilePathFlag = cli.StringFlag{
		Name:  "pwd_path",
		Usage: "File containing the store password file",
		Value: defaultPwdPath(),
	}
	RpcAddr = cli.StringFlag{
		Name:  "rpc_addr",
		Usage: "RPC host address",
		Value: "127.0.0.1:7000",
	}
)

func main() {
	app := cli.NewApp()
	app.Name = "paradigm"
	app.Usage = "Paradigm Network"
	app.HideVersion = true //there is a special command to print the version
	app.Commands = []cli.Command{
		//{
		//	Name:   "keygen",
		//	Usage:  "Dump new key pair",
		//	Action: keygen,
		//},
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
				ServiceAddressFlag,
				LogLevelFlag,
				HeartbeatFlag,
				MaxPoolFlag,
				TcpTimeoutFlag,
				CacheSizeFlag,
				SyncLimitFlag,
				StoreFlag,
				StorePathFlag,
				KeyStorePathFlag,
				PwdFilePathFlag,
				SequentiaAddress,
				RpcAddr,
			},
		},
		{
			Name:   "version",
			Usage:  "Show version info",
			Action: printVersion,
		},
		{
			Name:   "initAccount",
			Usage:  "Init Account",
			Action: createAccount,
		},
	}
	app.Run(os.Args)
}

//func keygen(c *cli.Context) error {
//	pemDump, err := crypto.GeneratePemKey()
//	if err != nil {
//		fmt.Println("Error generating PemDump")
//		os.Exit(2)
//	}
//
//	fmt.Println("PublicKey:")
//	fmt.Println(pemDump.PublicKey)
//	fmt.Println("PrivateKey:")
//	fmt.Println(pemDump.PrivateKey)
//
//	return nil
//}

func printVersion(c *cli.Context) error {
	fmt.Println(version.Version)
	return nil
}

func createAccount(c *cli.Context) {
	scryptN := keystore.StandardScryptN
	scryptP := keystore.StandardScryptP
	path := filepath.Join(c.String(KeyStorePathFlag.Name), config.PemKeyPath)

	passwordFile := c.String(PwdFilePathFlag.Name)
	pwd, _ := ioutil.ReadFile(passwordFile)
	fmt.Println(string(path))
	fmt.Println(string(passwordFile))
	fmt.Println(string(pwd))

	//Generate key and store it in the give directory.
	address, _ := keystore.StoreKey(path, string(pwd), scryptN, scryptP)
	fmt.Printf("Address: {%x}\n", address)
}

func run(c *cli.Context) error {
	fmt.Println("Paradigm Starting...")
	onlyAccretion := c.Bool(OnlyAccretion.Name)
	datadir := c.String(DataDirFlag.Name)
	addr := c.String(NodeAddressFlag.Name)
	gw2Address := c.String(Gw2AddressFlag.Name)
	fn2Address := c.String(Fn2AddressFlag.Name)
	serviceAddress := c.String(ServiceAddressFlag.Name)
	heartbeat := c.Int(HeartbeatFlag.Name)
	maxPool := c.Int(MaxPoolFlag.Name)
	tcpTimeout := c.Int(TcpTimeoutFlag.Name)
	cacheSize := c.Int(CacheSizeFlag.Name)
	syncLimit := c.Int(SyncLimitFlag.Name)
	storePath := c.String(StorePathFlag.Name)
	sequentiaAddress := c.String(SequentiaAddress.Name)
	keyStoreDir := c.String(KeyStorePathFlag.Name)
	pwdFilePath := c.String(PwdFilePathFlag.Name)
	rpcAddr := c.String(RpcAddr.Name)

	log.InitRotateWriter(datadir + "/paradigm.log")
	logger := log.GetLogger("Main")
	logger.Info().Interface(
		"only_accretion", onlyAccretion).Interface(
		"datadir", datadir).Interface(
		"gw2_addr", gw2Address).Interface(
		"fn2_addr", fn2Address).Interface(
		"node_addr", addr).Interface(
		"service_addr", serviceAddress).Interface(
		"heartbeat", heartbeat).Interface(
		"max_pool", maxPool).Interface(
		"tcp_timeout", tcpTimeout).Interface(
		"cache_size", cacheSize).Interface(
		"store_path", storePath).Interface(
		"rpcAddr", rpcAddr).Msg("Running Args")

	conf := config.NewConfig(onlyAccretion, time.Duration(heartbeat)*time.Millisecond,
		time.Duration(tcpTimeout)*time.Millisecond,
		cacheSize, syncLimit, storePath, gw2Address, fn2Address, sequentiaAddress, keyStoreDir, pwdFilePath,nil,nil, rpcAddr)

	//===============================================================================================================
	//// Create the PEM key
	//pemKey := crypto.NewPemKey(datadir)
	//
	//// Try a read
	//key, err := pemKey.ReadKey()
	//if err != nil {
	//	return cli.NewExitError(err, 1)
	//}
	//===============================================================================================================
	jsonPrivKey,err:=ioutil.ReadFile(filepath.Join(c.String(KeyStorePathFlag.Name), config.PemKeyPath))
	if err != nil {
		logger.Error().Msg("Key reading error")
	}
	pwd, _ := ioutil.ReadFile(c.String(PwdFilePathFlag.Name))
	kk,err := keystore.DecryptKey(jsonPrivKey,string(pwd))
	if err == nil {
		logger.Error().Msg("key decrypt error.")
	}
	//===============================================================================================================

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

	//Find the ID --common.Address-- of this node
	//Raw punlic key ,[]byte
	nodePub := fmt.Sprintf("0x%X", crypto.FromECDSAPub(&kk.PrivateKey.PublicKey))
	nodeID := pmap[nodePub]

	logger.Info().Interface("participantMap", pmap).Int("nodeID", nodeID).Msg("PARTICIPANTS")

	//Instantiate the Store (badger)
	var store storage.Store
	var needBootstrap bool
	if _, err := os.Stat(conf.StorePath); err == nil {
		logger.Info().Msg("Loading badger store from existing database")
		store, err = storage.LoadBadgerStore(conf.CacheSize, conf.StorePath)
		if err != nil {
			return cli.NewExitError(
				fmt.Sprintf("failed to load BadgerStore from existing file: %s", err),
				1)
		}
		needBootstrap = true
	} else {
		//Otherwise create a new one
		logger.Info().Msg("Creating new badger store from fresh database")
		store, err = storage.NewBadgerStore(pmap, conf.CacheSize, conf.StorePath)
		if err != nil {
			return cli.NewExitError(
				fmt.Sprintf("failed to create new BadgerStore: %s", err),
				1)
		}
	}

	trans, err := tcp.NewTCPTransport(addr,
		nil, maxPool, conf.TCPTimeout)
	if err != nil {
		return cli.NewExitError(err, 1)
	}

	proxy := proxy.NewInmemAppProxy(conf, store)

	//todo impl. if no_client
	node := core.NewNode(conf, nodeID, kk.PrivateKey, peers, store, trans, proxy)
	if err := node.Init(needBootstrap); err != nil {
		return cli.NewExitError(
			fmt.Sprintf("failed to initialize node: %s", err),
			1)
	}

	//serviceServer := service.NewService(node)

	//start rpc server
	//exitCh := make(chan interface{}, 0)
	//go func() {
	//	err = jsonrpc.StartRPCServer(conf, serviceServer)
	//	close(exitCh)
	//}()

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
func defaultKeyStoreDir() string {
	dataDir := defaultDataDir()
	if dataDir != "" {
		return filepath.Join(dataDir, "key_store")
	}
	return ""
}
func defaultPwdPath() string {
	dataDir := defaultDataDir()
	if dataDir != "" {
		return filepath.Join(dataDir, "pwd")
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
