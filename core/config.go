package core

import (
	"time"
	"io/ioutil"
)

type Config struct {
	OnlyAccretionNetwork bool //if true node will only join the accretion network. false will try to join sequentia network.
	HeartbeatTimeout     time.Duration
	TCPTimeout           time.Duration
	CacheSize            int
	SyncLimit            int
	StoreType            string
	StorePath            string

	Gw2Address string // api gate-way address
	Fn2Address string // function execute engine address

	//TODO add QCP config here
}

func NewConfig(
	onlyAccretion bool,
	heartbeat time.Duration,
	timeout time.Duration,
	cacheSize int,
	syncLimit int,
	storeType string,
	storePath string,
	gw2Address, fn2Address string,
) *Config {
	return &Config{
		OnlyAccretionNetwork: onlyAccretion,
		HeartbeatTimeout:     heartbeat,
		TCPTimeout:           timeout,
		CacheSize:            cacheSize,
		SyncLimit:            syncLimit,
		StoreType:            storeType,
		StorePath:            storePath,
		Gw2Address:           gw2Address,
		Fn2Address:           fn2Address,
	}
}

func DefaultConfig() *Config {
	storeType := "badger"
	storePath, _ := ioutil.TempDir("", "pdm_badger_store")
	return &Config{
		OnlyAccretionNetwork: false,
		HeartbeatTimeout:     1000 * time.Millisecond,
		TCPTimeout:           1000 * time.Millisecond,
		CacheSize:            500,
		SyncLimit:            100,
		StoreType:            storeType,
		StorePath:            storePath,
		Gw2Address:           "127.0.0.1:9000",
		Fn2Address:           "127.0.0.1:8000",
	}
}
