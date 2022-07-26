package state

import (
	"sync"
)

type (
	State struct {
		lock sync.Mutex
		data map[string]*PolicyState
	}
)

func New() *State {
	return &State{
		data: make(map[string]*PolicyState),
	}
}

func (s *State) AddOrUpdate(key string, info *PolicyState) {
	s.lock.Lock()
	defer s.lock.Unlock()

	s.data[key] = info
}

func (s *State) Get(key string) (*PolicyState, bool) {
	s.lock.Lock()
	defer s.lock.Unlock()

	v, exists := s.data[key]
	return v, exists
}

func (s *State) Delete(key string) {
	s.lock.Lock()
	defer s.lock.Unlock()

	delete(s.data, key)
}
