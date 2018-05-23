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

package dosclient

import "github.com/doslink/dos"

// Verify that Client implements the doslink interfaces.
var (
	_ = doslink.ChainReader(&Client{})
	_ = doslink.TransactionReader(&Client{})
	_ = doslink.ChainStateReader(&Client{})
	_ = doslink.ChainSyncReader(&Client{})
	_ = doslink.ContractCaller(&Client{})
	_ = doslink.GasEstimator(&Client{})
	_ = doslink.GasPricer(&Client{})
	_ = doslink.LogFilterer(&Client{})
	_ = doslink.PendingStateReader(&Client{})
	// _ = doslink.PendingStateEventer(&Client{})
	_ = doslink.PendingContractCaller(&Client{})
)
