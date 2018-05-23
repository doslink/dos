// Copyright 2016 The dos Authors
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

// Package les implements the Light Doslink Subprotocol.
package les

import (
	"fmt"
	"sync"
	"time"

	"github.com/doslink/dos/accounts"
	"github.com/doslink/dos/common"
	"github.com/doslink/dos/common/hexutil"
	"github.com/doslink/dos/consensus"
	"github.com/doslink/dos/core"
	"github.com/doslink/dos/core/bloombits"
	"github.com/doslink/dos/core/rawdb"
	"github.com/doslink/dos/core/types"
	"github.com/doslink/dos/dos"
	"github.com/doslink/dos/dos/downloader"
	"github.com/doslink/dos/dos/filters"
	"github.com/doslink/dos/dos/gasprice"
	"github.com/doslink/dos/dosdb"
	"github.com/doslink/dos/event"
	"github.com/doslink/dos/internal/dosapi"
	"github.com/doslink/dos/light"
	"github.com/doslink/dos/log"
	"github.com/doslink/dos/node"
	"github.com/doslink/dos/p2p"
	"github.com/doslink/dos/p2p/discv5"
	"github.com/doslink/dos/params"
	rpc "github.com/doslink/dos/rpc"
)

type LightDoslink struct {
	config *dos.Config

	odr         *LesOdr
	relay       *LesTxRelay
	chainConfig *params.ChainConfig
	// Channel for shutting down the service
	shutdownChan chan bool
	// Handlers
	peers           *peerSet
	txPool          *light.TxPool
	blockchain      *light.LightChain
	protocolManager *ProtocolManager
	serverPool      *serverPool
	reqDist         *requestDistributor
	retriever       *retrieveManager
	// DB interfaces
	chainDb dosdb.Database // Block chain database

	bloomRequests                              chan chan *bloombits.Retrieval // Channel receiving bloom data retrieval requests
	bloomIndexer, chtIndexer, bloomTrieIndexer *core.ChainIndexer

	ApiBackend *LesApiBackend

	eventMux       *event.TypeMux
	engine         consensus.Engine
	accountManager *accounts.Manager

	networkId     uint64
	netRPCService *dosapi.PublicNetAPI

	wg sync.WaitGroup
}

func New(ctx *node.ServiceContext, config *dos.Config) (*LightDoslink, error) {
	chainDb, err := dos.CreateDB(ctx, config, "lightchaindata")
	if err != nil {
		return nil, err
	}
	chainConfig, genesisHash, genesisErr := core.SetupGenesisBlock(chainDb, config.Genesis)
	if _, isCompat := genesisErr.(*params.ConfigCompatError); genesisErr != nil && !isCompat {
		return nil, genesisErr
	}
	log.Info("Initialised chain configuration", "config", chainConfig)

	peers := newPeerSet()
	quitSync := make(chan struct{})

	ldos := &LightDoslink{
		config:           config,
		chainConfig:      chainConfig,
		chainDb:          chainDb,
		eventMux:         ctx.EventMux,
		peers:            peers,
		reqDist:          newRequestDistributor(peers, quitSync),
		accountManager:   ctx.AccountManager,
		engine:           dos.CreateConsensusEngine(ctx, &config.Dosash, chainConfig, chainDb),
		shutdownChan:     make(chan bool),
		networkId:        config.NetworkId,
		bloomRequests:    make(chan chan *bloombits.Retrieval),
		bloomIndexer:     dos.NewBloomIndexer(chainDb, light.BloomTrieFrequency),
		chtIndexer:       light.NewChtIndexer(chainDb, true),
		bloomTrieIndexer: light.NewBloomTrieIndexer(chainDb, true),
	}

	ldos.relay = NewLesTxRelay(peers, ldos.reqDist)
	ldos.serverPool = newServerPool(chainDb, quitSync, &ldos.wg)
	ldos.retriever = newRetrieveManager(peers, ldos.reqDist, ldos.serverPool)
	ldos.odr = NewLesOdr(chainDb, ldos.chtIndexer, ldos.bloomTrieIndexer, ldos.bloomIndexer, ldos.retriever)
	if ldos.blockchain, err = light.NewLightChain(ldos.odr, ldos.chainConfig, ldos.engine); err != nil {
		return nil, err
	}
	ldos.bloomIndexer.Start(ldos.blockchain)
	// Rewind the chain in case of an incompatible config upgrade.
	if compat, ok := genesisErr.(*params.ConfigCompatError); ok {
		log.Warn("Rewinding chain to upgrade configuration", "err", compat)
		ldos.blockchain.SetHead(compat.RewindTo)
		rawdb.WriteChainConfig(chainDb, genesisHash, chainConfig)
	}

	ldos.txPool = light.NewTxPool(ldos.chainConfig, ldos.blockchain, ldos.relay)
	if ldos.protocolManager, err = NewProtocolManager(ldos.chainConfig, true, ClientProtocolVersions, config.NetworkId, ldos.eventMux, ldos.engine, ldos.peers, ldos.blockchain, nil, chainDb, ldos.odr, ldos.relay, quitSync, &ldos.wg); err != nil {
		return nil, err
	}
	ldos.ApiBackend = &LesApiBackend{ldos, nil}
	gpoParams := config.GPO
	if gpoParams.Default == nil {
		gpoParams.Default = config.GasPrice
	}
	ldos.ApiBackend.gpo = gasprice.NewOracle(ldos.ApiBackend, gpoParams)
	return ldos, nil
}

func lesTopic(genesisHash common.Hash, protocolVersion uint) discv5.Topic {
	var name string
	switch protocolVersion {
	case lpv1:
		name = "LES"
	case lpv2:
		name = "LES2"
	default:
		panic(nil)
	}
	return discv5.Topic(name + "@" + common.Bytes2Hex(genesisHash.Bytes()[0:8]))
}

type LightDummyAPI struct{}

// Doserbase is the address that mining rewards will be send to
func (s *LightDummyAPI) Doserbase() (common.Address, error) {
	return common.Address{}, fmt.Errorf("not supported")
}

// Coinbase is the address that mining rewards will be send to (alias for Doserbase)
func (s *LightDummyAPI) Coinbase() (common.Address, error) {
	return common.Address{}, fmt.Errorf("not supported")
}

// Hashrate returns the POW hashrate
func (s *LightDummyAPI) Hashrate() hexutil.Uint {
	return 0
}

// Mining returns an indication if this node is currently mining.
func (s *LightDummyAPI) Mining() bool {
	return false
}

// APIs returns the collection of RPC services the doslink package offers.
// NOTE, some of these services probably need to be moved to somewhere else.
func (s *LightDoslink) APIs() []rpc.API {
	return append(dosapi.GetAPIs(s.ApiBackend), []rpc.API{
		{
			Namespace: "dos",
			Version:   "1.0",
			Service:   &LightDummyAPI{},
			Public:    true,
		}, {
			Namespace: "dos",
			Version:   "1.0",
			Service:   downloader.NewPublicDownloaderAPI(s.protocolManager.downloader, s.eventMux),
			Public:    true,
		}, {
			Namespace: "dos",
			Version:   "1.0",
			Service:   filters.NewPublicFilterAPI(s.ApiBackend, true),
			Public:    true,
		}, {
			Namespace: "net",
			Version:   "1.0",
			Service:   s.netRPCService,
			Public:    true,
		},
	}...)
}

func (s *LightDoslink) ResetWithGenesisBlock(gb *types.Block) {
	s.blockchain.ResetWithGenesisBlock(gb)
}

func (s *LightDoslink) BlockChain() *light.LightChain      { return s.blockchain }
func (s *LightDoslink) TxPool() *light.TxPool              { return s.txPool }
func (s *LightDoslink) Engine() consensus.Engine           { return s.engine }
func (s *LightDoslink) LesVersion() int                    { return int(s.protocolManager.SubProtocols[0].Version) }
func (s *LightDoslink) Downloader() *downloader.Downloader { return s.protocolManager.downloader }
func (s *LightDoslink) EventMux() *event.TypeMux           { return s.eventMux }

// Protocols implements node.Service, returning all the currently configured
// network protocols to start.
func (s *LightDoslink) Protocols() []p2p.Protocol {
	return s.protocolManager.SubProtocols
}

// Start implements node.Service, starting all internal goroutines needed by the
// Doslink protocol implementation.
func (s *LightDoslink) Start(srvr *p2p.Server) error {
	s.startBloomHandlers()
	log.Warn("Light client mode is an experimental feature")
	s.netRPCService = dosapi.NewPublicNetAPI(srvr, s.networkId)
	// clients are searching for the first advertised protocol in the list
	protocolVersion := AdvertiseProtocolVersions[0]
	s.serverPool.start(srvr, lesTopic(s.blockchain.Genesis().Hash(), protocolVersion))
	s.protocolManager.Start(s.config.LightPeers)
	return nil
}

// Stop implements node.Service, terminating all internal goroutines used by the
// Doslink protocol.
func (s *LightDoslink) Stop() error {
	s.odr.Stop()
	if s.bloomIndexer != nil {
		s.bloomIndexer.Close()
	}
	if s.chtIndexer != nil {
		s.chtIndexer.Close()
	}
	if s.bloomTrieIndexer != nil {
		s.bloomTrieIndexer.Close()
	}
	s.blockchain.Stop()
	s.protocolManager.Stop()
	s.txPool.Stop()

	s.eventMux.Stop()

	time.Sleep(time.Millisecond * 200)
	s.chainDb.Close()
	close(s.shutdownChan)

	return nil
}
