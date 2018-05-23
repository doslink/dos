// Copyright 2015 The dos Authors
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

package dos

import (
	"context"
	"math/big"

	"github.com/doslink/dos/accounts"
	"github.com/doslink/dos/common"
	"github.com/doslink/dos/common/math"
	"github.com/doslink/dos/core"
	"github.com/doslink/dos/core/bloombits"
	"github.com/doslink/dos/core/rawdb"
	"github.com/doslink/dos/core/state"
	"github.com/doslink/dos/core/types"
	"github.com/doslink/dos/core/vm"
	"github.com/doslink/dos/dos/downloader"
	"github.com/doslink/dos/dos/gasprice"
	"github.com/doslink/dos/dosdb"
	"github.com/doslink/dos/event"
	"github.com/doslink/dos/params"
	"github.com/doslink/dos/rpc"
)

// DosAPIBackend implements dosapi.Backend for full nodes
type DosAPIBackend struct {
	dos *Doslink
	gpo *gasprice.Oracle
}

func (b *DosAPIBackend) ChainConfig() *params.ChainConfig {
	return b.dos.chainConfig
}

func (b *DosAPIBackend) CurrentBlock() *types.Block {
	return b.dos.blockchain.CurrentBlock()
}

func (b *DosAPIBackend) SetHead(number uint64) {
	b.dos.protocolManager.downloader.Cancel()
	b.dos.blockchain.SetHead(number)
}

func (b *DosAPIBackend) HeaderByNumber(ctx context.Context, blockNr rpc.BlockNumber) (*types.Header, error) {
	// Pending block is only known by the miner
	if blockNr == rpc.PendingBlockNumber {
		block := b.dos.miner.PendingBlock()
		return block.Header(), nil
	}
	// Otherwise resolve and return the block
	if blockNr == rpc.LatestBlockNumber {
		return b.dos.blockchain.CurrentBlock().Header(), nil
	}
	return b.dos.blockchain.GetHeaderByNumber(uint64(blockNr)), nil
}

func (b *DosAPIBackend) BlockByNumber(ctx context.Context, blockNr rpc.BlockNumber) (*types.Block, error) {
	// Pending block is only known by the miner
	if blockNr == rpc.PendingBlockNumber {
		block := b.dos.miner.PendingBlock()
		return block, nil
	}
	// Otherwise resolve and return the block
	if blockNr == rpc.LatestBlockNumber {
		return b.dos.blockchain.CurrentBlock(), nil
	}
	return b.dos.blockchain.GetBlockByNumber(uint64(blockNr)), nil
}

func (b *DosAPIBackend) StateAndHeaderByNumber(ctx context.Context, blockNr rpc.BlockNumber) (*state.StateDB, *types.Header, error) {
	// Pending state is only known by the miner
	if blockNr == rpc.PendingBlockNumber {
		block, state := b.dos.miner.Pending()
		return state, block.Header(), nil
	}
	// Otherwise resolve the block number and return its state
	header, err := b.HeaderByNumber(ctx, blockNr)
	if header == nil || err != nil {
		return nil, nil, err
	}
	stateDb, err := b.dos.BlockChain().StateAt(header.Root)
	return stateDb, header, err
}

func (b *DosAPIBackend) GetBlock(ctx context.Context, hash common.Hash) (*types.Block, error) {
	return b.dos.blockchain.GetBlockByHash(hash), nil
}

func (b *DosAPIBackend) GetReceipts(ctx context.Context, hash common.Hash) (types.Receipts, error) {
	if number := rawdb.ReadHeaderNumber(b.dos.chainDb, hash); number != nil {
		return rawdb.ReadReceipts(b.dos.chainDb, hash, *number), nil
	}
	return nil, nil
}

func (b *DosAPIBackend) GetLogs(ctx context.Context, hash common.Hash) ([][]*types.Log, error) {
	number := rawdb.ReadHeaderNumber(b.dos.chainDb, hash)
	if number == nil {
		return nil, nil
	}
	receipts := rawdb.ReadReceipts(b.dos.chainDb, hash, *number)
	if receipts == nil {
		return nil, nil
	}
	logs := make([][]*types.Log, len(receipts))
	for i, receipt := range receipts {
		logs[i] = receipt.Logs
	}
	return logs, nil
}

func (b *DosAPIBackend) GetTd(blockHash common.Hash) *big.Int {
	return b.dos.blockchain.GetTdByHash(blockHash)
}

func (b *DosAPIBackend) GetEVM(ctx context.Context, msg core.Message, state *state.StateDB, header *types.Header, vmCfg vm.Config) (*vm.EVM, func() error, error) {
	state.SetBalance(msg.From(), math.MaxBig256)
	vmError := func() error { return nil }

	context := core.NewEVMContext(msg, header, b.dos.BlockChain(), nil)
	return vm.NewEVM(context, state, b.dos.chainConfig, vmCfg), vmError, nil
}

func (b *DosAPIBackend) SubscribeRemovedLogsEvent(ch chan<- core.RemovedLogsEvent) event.Subscription {
	return b.dos.BlockChain().SubscribeRemovedLogsEvent(ch)
}

func (b *DosAPIBackend) SubscribeChainEvent(ch chan<- core.ChainEvent) event.Subscription {
	return b.dos.BlockChain().SubscribeChainEvent(ch)
}

func (b *DosAPIBackend) SubscribeChainHeadEvent(ch chan<- core.ChainHeadEvent) event.Subscription {
	return b.dos.BlockChain().SubscribeChainHeadEvent(ch)
}

func (b *DosAPIBackend) SubscribeChainSideEvent(ch chan<- core.ChainSideEvent) event.Subscription {
	return b.dos.BlockChain().SubscribeChainSideEvent(ch)
}

func (b *DosAPIBackend) SubscribeLogsEvent(ch chan<- []*types.Log) event.Subscription {
	return b.dos.BlockChain().SubscribeLogsEvent(ch)
}

func (b *DosAPIBackend) SendTx(ctx context.Context, signedTx *types.Transaction) error {
	return b.dos.txPool.AddLocal(signedTx)
}

func (b *DosAPIBackend) GetPoolTransactions() (types.Transactions, error) {
	pending, err := b.dos.txPool.Pending()
	if err != nil {
		return nil, err
	}
	var txs types.Transactions
	for _, batch := range pending {
		txs = append(txs, batch...)
	}
	return txs, nil
}

func (b *DosAPIBackend) GetPoolTransaction(hash common.Hash) *types.Transaction {
	return b.dos.txPool.Get(hash)
}

func (b *DosAPIBackend) GetPoolNonce(ctx context.Context, addr common.Address) (uint64, error) {
	return b.dos.txPool.State().GetNonce(addr), nil
}

func (b *DosAPIBackend) Stats() (pending int, queued int) {
	return b.dos.txPool.Stats()
}

func (b *DosAPIBackend) TxPoolContent() (map[common.Address]types.Transactions, map[common.Address]types.Transactions) {
	return b.dos.TxPool().Content()
}

func (b *DosAPIBackend) SubscribeTxPreEvent(ch chan<- core.TxPreEvent) event.Subscription {
	return b.dos.TxPool().SubscribeTxPreEvent(ch)
}

func (b *DosAPIBackend) Downloader() *downloader.Downloader {
	return b.dos.Downloader()
}

func (b *DosAPIBackend) ProtocolVersion() int {
	return b.dos.DosVersion()
}

func (b *DosAPIBackend) SuggestPrice(ctx context.Context) (*big.Int, error) {
	return b.gpo.SuggestPrice(ctx)
}

func (b *DosAPIBackend) ChainDb() dosdb.Database {
	return b.dos.ChainDb()
}

func (b *DosAPIBackend) EventMux() *event.TypeMux {
	return b.dos.EventMux()
}

func (b *DosAPIBackend) AccountManager() *accounts.Manager {
	return b.dos.AccountManager()
}

func (b *DosAPIBackend) BloomStatus() (uint64, uint64) {
	sections, _, _ := b.dos.bloomIndexer.Sections()
	return params.BloomBitsBlocks, sections
}

func (b *DosAPIBackend) ServiceFilter(ctx context.Context, session *bloombits.MatcherSession) {
	for i := 0; i < bloomFilterThreads; i++ {
		go session.Multiplex(bloomRetrievalBatch, bloomRetrievalWait, b.dos.bloomRequests)
	}
}
