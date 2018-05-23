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

package params

// MainnetBootnodes are the enode URLs of the P2P bootstrap nodes running on
// the main Doslink network.
var MainnetBootnodes = []string{
	"enode://a1108e8162a0b977754026f1fc42aa189361fe12185b6fb67b745cf81bc11267a3db1cd654b9a212060a5fc7b27f79f0e7fb093f3e46628e8acfacd3f86ab851@69.172.85.245:30605",
	"enode://dfa8d41f69b2698c1f27cd499b3037da7a5f55625188efd26849bd923790123b8581283f4c8a8bf1d6718cac1ba289d2d7ad9ce05cb6ab57be291f4b31a7ed85@69.172.85.246:30605",
}

// TestnetBootnodes are the enode URLs of the P2P bootstrap nodes running on the
// Ropsten test network.
var TestnetBootnodes = []string{
}

// RinkebyBootnodes are the enode URLs of the P2P bootstrap nodes running on the
// Rinkeby test network.
var RinkebyBootnodes = []string{
}

// DiscoveryV5Bootnodes are the enode URLs of the P2P bootstrap nodes for the
// experimental RLPx v5 topic-discovery network.
var DiscoveryV5Bootnodes = []string{
}
