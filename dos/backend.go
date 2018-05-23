// Copyright 2014 The dos Authors
// This file is part of the dos library.
//
// The dos library is free software: you can redistribute it and/or modify
// it under the terms of the GNU Lesser General Public License as published by
// the Free Software Foundation, either version 3 of the License, or
// (at your option) any later version.
//
// The dos library is distributed in the hope that it will be useful,
// but WITHOUT ANY WARRANTY; without even the implied warranty of
// MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE. See the
// GNU Lesser General Public License for more details.
//
// You should have received a copy of the GNU Lesser General Public License
// along with the dos library. If not, see <http://www.gnu.org/licenses/>.

// Package dos implements the Doslink protocol.
package dos

import (
	"errors"
	"fmt"
	"math/big"
	"runtime"
	"sync"
	"sync/atomic"

	"github.com/doslink/dos/accounts"
	"github.com/doslink/dos/common"
	"github.com/doslink/dos/common/hexutil"
	"github.com/doslink/dos/consensus"
	"github.com/doslink/dos/consensus/clique"
	"github.com/doslink/dos/consensus/dosash"
	"github.com/doslink/dos/core"
	"github.com/doslink/dos/core/bloombits"
	"github.com/doslink/dos/core/rawdb"
	"github.com/doslink/dos/core/types"
	"github.com/doslink/dos/core/vm"
	"github.com/doslink/dos/dos/downloader"
	"github.com/doslink/dos/dos/filters"
	"github.com/doslink/dos/dos/gasprice"
	"github.com/doslink/dos/dosdb"
	"github.com/doslink/dos/event"
	"github.com/doslink/dos/internal/dosapi"
	"github.com/doslink/dos/log"
	"github.com/doslink/dos/miner"
	"github.com/doslink/dos/node"
	"github.com/doslink/dos/p2p"
	"github.com/doslink/dos/params"
	"github.com/doslink/dos/rlp"
	"github.com/doslink/dos/rpc"
)

type LesServer interface {
	Start(srvr *p2p.Server)
	Stop()
	Protocols() []p2p.Protocol
	SetBloomBitsIndexer(bbIndexer *core.ChainIndexer)
}

// Doslink implements the Doslink full node service.
type Doslink struct {
	config      *Config
	chainConfig *params.ChainConfig

	// Channel for shutting down the service
	shutdownChan chan bool // Channel for shutting down the Doslink

	// Handlers
	txPool          *core.TxPool
	blockchain      *core.BlockChain
	protocolManager *ProtocolManager
	lesServer       LesServer

	// DB interfaces
	chainDb dosdb.Database // Block chain database

	eventMux       *event.TypeMux
	engine         consensus.Engine
	accountManager *accounts.Manager

	bloomRequests chan chan *bloombits.Retrieval // Channel receiving bloom data retrieval requests
	bloomIndexer  *core.ChainIndexer             // Bloom indexer operating during block imports

	APIBackend *DosAPIBackend

	miner     *miner.Miner
	gasPrice  *big.Int
	doserbase common.Address

	networkId     uint64
	netRPCService *dosapi.PublicNetAPI

	lock sync.RWMutex // Protects the variadic fields (e.g. gas price and doserbase)
}

func (s *Doslink) AddLesServer(ls LesServer) {
	s.lesServer = ls
	ls.SetBloomBitsIndexer(s.bloomIndexer)
}

// New creates a new Doslink object (including the
// initialisation of the common Doslink object)
func New(ctx *node.ServiceContext, config *Config) (*Doslink, error) {
	if config.SyncMode == downloader.LightSync {
		return nil, errors.New("can't run dos.Doslink in light sync mode, use les.LightDoslink")
	}
	if !config.SyncMode.IsValid() {
		return nil, fmt.Errorf("invalid sync mode %d", config.SyncMode)
	}
	chainDb, err := CreateDB(ctx, config, "chaindata")
	if err != nil {
		return nil, err
	}
	chainConfig, genesisHash, genesisErr := core.SetupGenesisBlock(chainDb, config.Genesis)
	if _, ok := genesisErr.(*params.ConfigCompatError); genesisErr != nil && !ok {
		return nil, genesisErr
	}
	log.Info("Initialised chain configuration", "config", chainConfig)

	dos := &Doslink{
		config:         config,
		chainDb:        chainDb,
		chainConfig:    chainConfig,
		eventMux:       ctx.EventMux,
		accountManager: ctx.AccountManager,
		engine:         CreateConsensusEngine(ctx, &config.Dosash, chainConfig, chainDb),
		shutdownChan:   make(chan bool),
		networkId:      config.NetworkId,
		gasPrice:       config.GasPrice,
		doserbase:      config.Doserbase,
		bloomRequests:  make(chan chan *bloombits.Retrieval),
		bloomIndexer:   NewBloomIndexer(chainDb, params.BloomBitsBlocks),
	}

	log.Info("Initialising Doslink protocol", "versions", ProtocolVersions, "network", config.NetworkId)

	if !config.SkipBcVersionCheck {
		bcVersion := rawdb.ReadDatabaseVersion(chainDb)
		if bcVersion != core.BlockChainVersion && bcVersion != 0 {
			return nil, fmt.Errorf("Blockchain DB version mismatch (%d / %d). Run gdos upgradedb.\n", bcVersion, core.BlockChainVersion)
		}
		rawdb.WriteDatabaseVersion(chainDb, core.BlockChainVersion)
	}
	var (
		vmConfig    = vm.Config{EnablePreimageRecording: config.EnablePreimageRecording}
		cacheConfig = &core.CacheConfig{Disabled: config.NoPruning, TrieNodeLimit: config.TrieCache, TrieTimeLimit: config.TrieTimeout}
	)
	dos.blockchain, err = core.NewBlockChain(chainDb, cacheConfig, dos.chainConfig, dos.engine, vmConfig)
	if err != nil {
		return nil, err
	}
	// Rewind the chain in case of an incompatible config upgrade.
	if compat, ok := genesisErr.(*params.ConfigCompatError); ok {
		log.Warn("Rewinding chain to upgrade configuration", "err", compat)
		dos.blockchain.SetHead(compat.RewindTo)
		rawdb.WriteChainConfig(chainDb, genesisHash, chainConfig)
	}
	dos.bloomIndexer.Start(dos.blockchain)

	if config.TxPool.Journal != "" {
		config.TxPool.Journal = ctx.ResolvePath(config.TxPool.Journal)
	}
	dos.txPool = core.NewTxPool(config.TxPool, dos.chainConfig, dos.blockchain)

	if dos.protocolManager, err = NewProtocolManager(dos.chainConfig, config.SyncMode, config.NetworkId, dos.eventMux, dos.txPool, dos.engine, dos.blockchain, chainDb); err != nil {
		return nil, err
	}
	dos.miner = miner.New(dos, dos.chainConfig, dos.EventMux(), dos.engine)
	dos.miner.SetExtra(makeExtraData(config.ExtraData))

	dos.APIBackend = &DosAPIBackend{dos, nil}
	gpoParams := config.GPO
	if gpoParams.Default == nil {
		gpoParams.Default = config.GasPrice
	}
	dos.APIBackend.gpo = gasprice.NewOracle(dos.APIBackend, gpoParams)

	return dos, nil
}

func makeExtraData(extra []byte) []byte {
	if len(extra) == 0 {
		// create default extradata
		extra, _ = rlp.EncodeToBytes([]interface{}{
			uint(params.VersionMajor<<16 | params.VersionMinor<<8 | params.VersionPatch),
			"gdos",
			runtime.Version(),
			runtime.GOOS,
		})
	}
	if uint64(len(extra)) > params.MaximumExtraDataSize {
		log.Warn("Miner extra data exceed limit", "extra", hexutil.Bytes(extra), "limit", params.MaximumExtraDataSize)
		extra = nil
	}
	return extra
}

// CreateDB creates the chain database.
func CreateDB(ctx *node.ServiceContext, config *Config, name string) (dosdb.Database, error) {
	db, err := ctx.OpenDatabase(name, config.DatabaseCache, config.DatabaseHandles)
	if err != nil {
		return nil, err
	}
	if db, ok := db.(*dosdb.LDBDatabase); ok {
		db.Meter("dos/db/chaindata/")
	}
	return db, nil
}

// CreateConsensusEngine creates the required type of consensus engine instance for an Doslink service
func CreateConsensusEngine(ctx *node.ServiceContext, config *dosash.Config, chainConfig *params.ChainConfig, db dosdb.Database) consensus.Engine {
	// If proof-of-authority is requested, set it up
	if chainConfig.Clique != nil {
		return clique.New(chainConfig.Clique, db)
	}
	// Otherwise assume proof-of-work
	switch {
	case config.PowMode == dosash.ModeFake:
		log.Warn("Dosash used in fake mode")
		return dosash.NewFaker()
	case config.PowMode == dosash.ModeTest:
		log.Warn("Dosash used in test mode")
		return dosash.NewTester()
	case config.PowMode == dosash.ModeShared:
		log.Warn("Dosash used in shared mode")
		return dosash.NewShared()
	default:
		engine := dosash.New(dosash.Config{
			CacheDir:       ctx.ResolvePath(config.CacheDir),
			CachesInMem:    config.CachesInMem,
			CachesOnDisk:   config.CachesOnDisk,
			DatasetDir:     config.DatasetDir,
			DatasetsInMem:  config.DatasetsInMem,
			DatasetsOnDisk: config.DatasetsOnDisk,
		})
		engine.SetThreads(-1) // Disable CPU mining
		return engine
	}
}

// APIs returns the collection of RPC services the doslink package offers.
// NOTE, some of these services probably need to be moved to somewhere else.
func (s *Doslink) APIs() []rpc.API {
	apis := dosapi.GetAPIs(s.APIBackend)

	// Append any APIs exposed explicitly by the consensus engine
	apis = append(apis, s.engine.APIs(s.BlockChain())...)

	// Append all the local APIs and return
	return append(apis, []rpc.API{
		{
			Namespace: "dos",
			Version:   "1.0",
			Service:   NewPublicDoslinkAPI(s),
			Public:    true,
		}, {
			Namespace: "dos",
			Version:   "1.0",
			Service:   NewPublicMinerAPI(s),
			Public:    true,
		}, {
			Namespace: "dos",
			Version:   "1.0",
			Service:   downloader.NewPublicDownloaderAPI(s.protocolManager.downloader, s.eventMux),
			Public:    true,
		}, {
			Namespace: "miner",
			Version:   "1.0",
			Service:   NewPrivateMinerAPI(s),
			Public:    false,
		}, {
			Namespace: "dos",
			Version:   "1.0",
			Service:   filters.NewPublicFilterAPI(s.APIBackend, false),
			Public:    true,
		}, {
			Namespace: "admin",
			Version:   "1.0",
			Service:   NewPrivateAdminAPI(s),
		}, {
			Namespace: "debug",
			Version:   "1.0",
			Service:   NewPublicDebugAPI(s),
			Public:    true,
		}, {
			Namespace: "debug",
			Version:   "1.0",
			Service:   NewPrivateDebugAPI(s.chainConfig, s),
		}, {
			Namespace: "net",
			Version:   "1.0",
			Service:   s.netRPCService,
			Public:    true,
		},
	}...)
}

func (s *Doslink) ResetWithGenesisBlock(gb *types.Block) {
	s.blockchain.ResetWithGenesisBlock(gb)
}

func (s *Doslink) Doserbase() (eb common.Address, err error) {
	s.lock.RLock()
	doserbase := s.doserbase
	s.lock.RUnlock()

	if doserbase != (common.Address{}) {
		return doserbase, nil
	}
	if wallets := s.AccountManager().Wallets(); len(wallets) > 0 {
		if accounts := wallets[0].Accounts(); len(accounts) > 0 {
			doserbase := accounts[0].Address

			s.lock.Lock()
			s.doserbase = doserbase
			s.lock.Unlock()

			log.Info("Doserbase automatically configured", "address", doserbase)
			return doserbase, nil
		}
	}
	return common.Address{}, fmt.Errorf("doserbase must be explicitly specified")
}

// SetDoserbase sets the mining reward address.
func (s *Doslink) SetDoserbase(doserbase common.Address) {
	s.lock.Lock()
	s.doserbase = doserbase
	s.lock.Unlock()

	s.miner.SetDoserbase(doserbase)
}

func (s *Doslink) StartMining(local bool) error {
	eb, err := s.Doserbase()
	if err != nil {
		log.Error("Cannot start mining without doserbase", "err", err)
		return fmt.Errorf("doserbase missing: %v", err)
	}
	if clique, ok := s.engine.(*clique.Clique); ok {
		wallet, err := s.accountManager.Find(accounts.Account{Address: eb})
		if wallet == nil || err != nil {
			log.Error("Doserbase account unavailable locally", "err", err)
			return fmt.Errorf("signer missing: %v", err)
		}
		clique.Authorize(eb, wallet.SignHash)
	}
	if local {
		// If local (CPU) mining is started, we can disable the transaction rejection
		// mechanism introduced to speed sync times. CPU mining on mainnet is ludicrous
		// so none will ever hit this path, whereas marking sync done on CPU mining
		// will ensure that private networks work in single miner mode too.
		atomic.StoreUint32(&s.protocolManager.acceptTxs, 1)
	}
	go s.miner.Start(eb)
	return nil
}

func (s *Doslink) StopMining()         { s.miner.Stop() }
func (s *Doslink) IsMining() bool      { return s.miner.Mining() }
func (s *Doslink) Miner() *miner.Miner { return s.miner }

func (s *Doslink) AccountManager() *accounts.Manager  { return s.accountManager }
func (s *Doslink) BlockChain() *core.BlockChain       { return s.blockchain }
func (s *Doslink) TxPool() *core.TxPool               { return s.txPool }
func (s *Doslink) EventMux() *event.TypeMux           { return s.eventMux }
func (s *Doslink) Engine() consensus.Engine           { return s.engine }
func (s *Doslink) ChainDb() dosdb.Database            { return s.chainDb }
func (s *Doslink) IsListening() bool                  { return true } // Always listening
func (s *Doslink) DosVersion() int                    { return int(s.protocolManager.SubProtocols[0].Version) }
func (s *Doslink) NetVersion() uint64                 { return s.networkId }
func (s *Doslink) Downloader() *downloader.Downloader { return s.protocolManager.downloader }

// Protocols implements node.Service, returning all the currently configured
// network protocols to start.
func (s *Doslink) Protocols() []p2p.Protocol {
	if s.lesServer == nil {
		return s.protocolManager.SubProtocols
	}
	return append(s.protocolManager.SubProtocols, s.lesServer.Protocols()...)
}

// Start implements node.Service, starting all internal goroutines needed by the
// Doslink protocol implementation.
func (s *Doslink) Start(srvr *p2p.Server) error {
	// Start the bloom bits servicing goroutines
	s.startBloomHandlers()

	// Start the RPC service
	s.netRPCService = dosapi.NewPublicNetAPI(srvr, s.NetVersion())

	// Figure out a max peers count based on the server limits
	maxPeers := srvr.MaxPeers
	if s.config.LightServ > 0 {
		if s.config.LightPeers >= srvr.MaxPeers {
			return fmt.Errorf("invalid peer config: light peer count (%d) >= total peer count (%d)", s.config.LightPeers, srvr.MaxPeers)
		}
		maxPeers -= s.config.LightPeers
	}
	// Start the networking layer and the light server if requested
	s.protocolManager.Start(maxPeers)
	if s.lesServer != nil {
		s.lesServer.Start(srvr)
	}
	return nil
}

// Stop implements node.Service, terminating all internal goroutines used by the
// Doslink protocol.
func (s *Doslink) Stop() error {
	s.bloomIndexer.Close()
	s.blockchain.Stop()
	s.protocolManager.Stop()
	if s.lesServer != nil {
		s.lesServer.Stop()
	}
	s.txPool.Stop()
	s.miner.Stop()
	s.eventMux.Stop()

	s.chainDb.Close()
	close(s.shutdownChan)

	return nil
}
