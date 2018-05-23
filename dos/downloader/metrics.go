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

// Contains the metrics collected by the downloader.

package downloader

import (
	"github.com/doslink/dos/metrics"
)

var (
	headerInMeter      = metrics.NewRegisteredMeter("dos/downloader/headers/in", nil)
	headerReqTimer     = metrics.NewRegisteredTimer("dos/downloader/headers/req", nil)
	headerDropMeter    = metrics.NewRegisteredMeter("dos/downloader/headers/drop", nil)
	headerTimeoutMeter = metrics.NewRegisteredMeter("dos/downloader/headers/timeout", nil)

	bodyInMeter      = metrics.NewRegisteredMeter("dos/downloader/bodies/in", nil)
	bodyReqTimer     = metrics.NewRegisteredTimer("dos/downloader/bodies/req", nil)
	bodyDropMeter    = metrics.NewRegisteredMeter("dos/downloader/bodies/drop", nil)
	bodyTimeoutMeter = metrics.NewRegisteredMeter("dos/downloader/bodies/timeout", nil)

	receiptInMeter      = metrics.NewRegisteredMeter("dos/downloader/receipts/in", nil)
	receiptReqTimer     = metrics.NewRegisteredTimer("dos/downloader/receipts/req", nil)
	receiptDropMeter    = metrics.NewRegisteredMeter("dos/downloader/receipts/drop", nil)
	receiptTimeoutMeter = metrics.NewRegisteredMeter("dos/downloader/receipts/timeout", nil)

	stateInMeter   = metrics.NewRegisteredMeter("dos/downloader/states/in", nil)
	stateDropMeter = metrics.NewRegisteredMeter("dos/downloader/states/drop", nil)
)
