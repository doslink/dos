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

package tests

import (
	"fmt"
	"math/big"

	"github.com/doslink/dos/params"
)

// Forks table defines supported forks and their chain config.
var Forks = map[string]*params.ChainConfig{
	"Frontier": {
		ChainId: big.NewInt(605),
	},
	"Homestead": {
		ChainId:        big.NewInt(605),
		HomesteadBlock: big.NewInt(0),
	},
	"EIP150": {
		ChainId:        big.NewInt(605),
		HomesteadBlock: big.NewInt(0),
		EIP150Block:    big.NewInt(0),
	},
	"EIP158": {
		ChainId:        big.NewInt(605),
		HomesteadBlock: big.NewInt(0),
		EIP150Block:    big.NewInt(0),
		EIP155Block:    big.NewInt(0),
		EIP158Block:    big.NewInt(0),
	},
	"Byzantium": {
		ChainId:        big.NewInt(605),
		HomesteadBlock: big.NewInt(0),
		EIP150Block:    big.NewInt(0),
		EIP155Block:    big.NewInt(0),
		EIP158Block:    big.NewInt(0),
		DAOForkBlock:   big.NewInt(0),
		ByzantiumBlock: big.NewInt(0),
	},
	"FrontierToHomesteadAt5": {
		ChainId:        big.NewInt(605),
		HomesteadBlock: big.NewInt(5),
	},
	"HomesteadToEIP150At5": {
		ChainId:        big.NewInt(605),
		HomesteadBlock: big.NewInt(0),
		EIP150Block:    big.NewInt(5),
	},
	"HomesteadToDaoAt5": {
		ChainId:        big.NewInt(605),
		HomesteadBlock: big.NewInt(0),
		DAOForkBlock:   big.NewInt(5),
		DAOForkSupport: false,
	},
	"EIP158ToByzantiumAt5": {
		ChainId:        big.NewInt(605),
		HomesteadBlock: big.NewInt(0),
		EIP150Block:    big.NewInt(0),
		EIP155Block:    big.NewInt(0),
		EIP158Block:    big.NewInt(0),
		ByzantiumBlock: big.NewInt(5),
	},
}

// UnsupportedForkError is returned when a test requests a fork that isn't implemented.
type UnsupportedForkError struct {
	Name string
}

func (e UnsupportedForkError) Error() string {
	return fmt.Sprintf("unsupported fork %q", e.Name)
}
