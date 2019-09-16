/*
Copyright SecureKey Technologies Inc. All Rights Reserved.
SPDX-License-Identifier: Apache-2.0
*/

package storage

import (
	"sync"

	"github.com/hyperledger/aries-framework-go/pkg/storage"
)

// MockStoreProvider mock store provider.
type MockStoreProvider struct {
	Store             MockStore
	ErrGetStoreHandle error
}

// NewMockStoreProvider new store provider instance.
func NewMockStoreProvider() *MockStoreProvider {
	return &MockStoreProvider{Store: MockStore{
		Store: make(map[string][]byte),
	}}
}

// GetStoreHandle returns a store.
func (s *MockStoreProvider) GetStoreHandle() (storage.Store, error) {
	return &s.Store, s.ErrGetStoreHandle
}

// Close closes the store provider.
func (s *MockStoreProvider) Close() error {
	return nil
}

// MockStore mock store.
type MockStore struct {
	Store  map[string][]byte
	lock   sync.RWMutex
	ErrPut error
	ErrGet error
}

// Put stores the key and the record
func (s *MockStore) Put(k string, v []byte) error {
	s.lock.Lock()
	s.Store[k] = v
	s.lock.Unlock()

	return s.ErrPut
}

// Get fetches the record based on key
func (s *MockStore) Get(k string) ([]byte, error) {
	s.lock.RLock()
	defer s.lock.RUnlock()

	val, ok := s.Store[k]
	if !ok {
		return nil, storage.ErrDataNotFound
	}

	return val, s.ErrGet
}