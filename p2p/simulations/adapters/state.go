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

package adapters

type SimStateStore struct {
	m map[string][]byte
}

func (st *SimStateStore) Load(s string) ([]byte, error) {
	return st.m[s], nil
}

func (st *SimStateStore) Save(s string, data []byte) error {
	st.m[s] = data
	return nil
}

func NewSimStateStore() *SimStateStore {
	return &SimStateStore{
		make(map[string][]byte),
	}
}