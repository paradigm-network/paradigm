package core

import (
	"github.com/sirupsen/logrus"
	"time"
	"io/ioutil"
)

type Config struct {
	OnlyAccretionNetwork bool //if true node will only join the accretion network. false will try to join sequentia network.
	HeartbeatTimeout time.Duration
	TCPTimeout       time.Duration
	CacheSize        int
	SyncLimit        int
	StoreType        string
	StorePath        string

	Gw2Address		string // api gate-way address
	Fn2Address		string // function execute engine address

	//TODO add QCP config here

	Logger           *logrus.Logger
}

func NewConfig(heartbeat time.Duration,
	timeout time.Duration,
	cacheSize int,
	syncLimit int,
	storeType string,
	storePath string,
	logger *logrus.Logger) *Config {
	return &Config{
		HeartbeatTimeout: heartbeat,
		TCPTimeout:       timeout,
		CacheSize:        cacheSize,
		SyncLimit:        syncLimit,
		StoreType:        storeType,
		StorePath:        storePath,
		Logger:           logger,
	}
}

func DefaultConfig() *Config {
	logger := logrus.New()
	logger.Level = logrus.DebugLevel
	storeType := "badger"
	storePath, _ := ioutil.TempDir("", "pdm_badger_store")
	return &Config{
		HeartbeatTimeout: 1000 * time.Millisecond,
		TCPTimeout:       1000 * time.Millisecond,
		CacheSize:        500,
		SyncLimit:        100,
		StoreType:        storeType,
		StorePath:        storePath,
		Logger:           logger,
	}
}
