package config

import (
	"io/ioutil"
	"time"
)


const (
	DEFAULT_GEN_BLOCK_TIME   = 6
	DBFT_MIN_NODE_NUM        = 4 //min node number of dbft consensus
	SOLO_MIN_NODE_NUM        = 1 //min node number of solo consensus
	VBFT_MIN_NODE_NUM        = 4 //min node number of vbft consensus
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
	RpcAddr          string
	//TODO add QCP config here
	P2PNodeConfig   *P2PNodeConfig
	ConsensusConfig *ConsensusConfig
	GenesisConfig   *GenesisConfig
}

type P2PRsvConfig struct {
	ReservedPeers []string `json:"reserved"`
	MaskPeers     []string `json:"mask"`
}

type P2PNodeConfig struct {
	ReservedPeersOnly         bool
	ReservedCfg               *P2PRsvConfig
	NetworkMagic              uint32
	NetworkId                 uint32
	NetworkName               string
	NodePort                  uint
	NodeConsensusPort         uint
	DualPortSupport           bool
	IsTLS                     bool
	CertPath                  string
	KeyPath                   string
	CAPath                    string
	HttpInfoPort              uint
	MaxHdrSyncReqs            uint
	MaxConnInBound            uint
	MaxConnOutBound           uint
	MaxConnInBoundForSingleIP uint
}

type ConsensusConfig struct {
	EnableConsensus bool
	MaxTxInBlock    uint
}

type GenesisConfig struct {
	SeedList      []string
	ConsensusType string
	VBFT          *VBFTConfig
	DBFT          *DBFTConfig
	SOLO          *SOLOConfig
}

//
// VBFT genesis config, from local config file
//
type VBFTConfig struct {
	N                    uint32               `json:"n"` // network size
	C                    uint32               `json:"c"` // consensus quorum
	K                    uint32               `json:"k"`
	L                    uint32               `json:"l"`
	BlockMsgDelay        uint32               `json:"block_msg_delay"`
	HashMsgDelay         uint32               `json:"hash_msg_delay"`
	PeerHandshakeTimeout uint32               `json:"peer_handshake_timeout"`
	MaxBlockChangeView   uint32               `json:"max_block_change_view"`
	MinInitStake         uint32               `json:"min_init_stake"`
	AdminOntID           string               `json:"admin_ont_id"`
	VrfValue             string               `json:"vrf_value"`
	VrfProof             string               `json:"vrf_proof"`
	Peers                []*VBFTPeerStakeInfo `json:"peers"`
}

type VBFTPeerStakeInfo struct {
	Index      uint32 `json:"index"`
	PeerPubkey string `json:"peerPubkey"`
	Address    string `json:"address"`
	InitPos    uint64 `json:"initPos"`
}

type DBFTConfig struct {
	GenBlockTime uint
	Bookkeepers  []string
}

type SOLOConfig struct {
	GenBlockTime uint
	Bookkeepers  []string
}

var DefConfig = DefaultConfig()

func NewConfig(
	onlyAccretion bool,
	heartbeat time.Duration,
	timeout time.Duration,
	cacheSize int,
	syncLimit int,
	storePath string,
	gw2Address, fn2Address, SequentiaAddress, KeyStoreDir, PwdFile string, P2PNodeConfig *P2PNodeConfig,
	ConsensusConfig *ConsensusConfig,
	RpcAddr string,
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
		P2PNodeConfig:        P2PNodeConfig,
		ConsensusConfig:      ConsensusConfig,
		RpcAddr:              RpcAddr,
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
		P2PNodeConfig:        nil,
		ConsensusConfig:      nil,
		RpcAddr:              "127.0.0.1:7000",
	}
}
