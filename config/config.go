package config

import (
	"io/ioutil"
	"time"
)

type Config struct {
	OnlyAccretionNetwork bool //if true node will only join the accretion network. false will try to join sequentia network.
	HeartbeatTimeout     time.Duration
	TCPTimeout           time.Duration
	CacheSize            int
	SyncLimit            int
	StorePath            string

	Gw2Address       string // api gate-way address
	Fn2Address       string // function execute engine address
	SequentiaAddress string // sequentia address
	KeyStoreDir      string //keyfile dir
	PwdFile          string //password  file
	//TODO add QCP config here
}

func NewConfig(
	onlyAccretion bool,
	heartbeat time.Duration,
	timeout time.Duration,
	cacheSize int,
	syncLimit int,
	storePath string,
	gw2Address, fn2Address, SequentiaAddress, KeyStoreDir, PwdFile string,
) *Config {
	return &Config{
		OnlyAccretionNetwork: onlyAccretion,
		HeartbeatTimeout:     heartbeat,
		TCPTimeout:           timeout,
		CacheSize:            cacheSize,
		SyncLimit:            syncLimit,
		StorePath:            storePath,
		Gw2Address:           gw2Address,
		Fn2Address:           fn2Address,
		SequentiaAddress:     SequentiaAddress,
		KeyStoreDir:          KeyStoreDir,
		PwdFile:              PwdFile,
	}
}

func DefaultConfig() *Config {
	storePath, _ := ioutil.TempDir("", "pdm_badger_store")
	return &Config{
		OnlyAccretionNetwork: false,
		HeartbeatTimeout:     1000 * time.Millisecond,
		TCPTimeout:           1000 * time.Millisecond,
		CacheSize:            500,
		SyncLimit:            100,
		StorePath:            storePath,
		Gw2Address:           "127.0.0.1:9000",
		Fn2Address:           "127.0.0.1:8000",
		SequentiaAddress:     "127.0.0.1:8090",
		KeyStoreDir:          storePath,
		PwdFile:              storePath + "/pwd",
	}
}
