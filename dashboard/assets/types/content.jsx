// @flow

// Copyright 2017 The dos Authors
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

export type Content = {
	general: General,
	home: Home,
	chain: Chain,
	txpool: TxPool,
	network: Network,
	system: System,
	logs: Logs,
};

export type ChartEntries = Array<ChartEntry>;

export type ChartEntry = {
	time: Date,
	value: number,
};

export type General = {
    version: ?string,
    commit: ?string,
};

export type Home = {
	/* TODO (kurkomisi) */
};

export type Chain = {
	/* TODO (kurkomisi) */
};

export type TxPool = {
	/* TODO (kurkomisi) */
};

export type Network = {
	/* TODO (kurkomisi) */
};

export type System = {
    activeMemory: ChartEntries,
    virtualMemory: ChartEntries,
    networkIngress: ChartEntries,
    networkEgress: ChartEntries,
    processCPU: ChartEntries,
    systemCPU: ChartEntries,
    diskRead: ChartEntries,
    diskWrite: ChartEntries,
};

export type Logs = {
	log: Array<string>,
};