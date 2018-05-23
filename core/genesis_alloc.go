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

package core

// Constants containing the genesis allocation of built-in genesis blocks.
// Their content is an RLP-encoded list of (address, balance) tuples.
// Use mkalloc.go to create/update them.

// nolint: misspell
const mainnetAllocData = "\xe2\xe1\x940c\x92\xa7tI\xb78\xbefA\xef/\x88P{_\xd3\xdao\x8bR\xb7\xd2\xea\xa8\u00c6\x8bd\x00\x00"
const testnetAllocData = "\xe2\xe1\x940c\x92\xa7tI\xb78\xbefA\xef/\x88P{_\xd3\xdao\x8bR\xb7\xd2\xea\xa8\u00c6\x8bd\x00\x00"
const rinkebyAllocData = "\xe2\xe1\x940c\x92\xa7tI\xb78\xbefA\xef/\x88P{_\xd3\xdao\x8bR\xb7\xd2\xea\xa8\u00c6\x8bd\x00\x00"
