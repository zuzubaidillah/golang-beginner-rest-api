// File: /users_store.go
package main

import (
	"sync"
	"time"
)

type User struct {
	ID        int       `json:"id"`
	Name      string    `json:"name"`
	CreatedAt time.Time `json:"createdAt"`
}

type UserStore struct {
	mu     sync.RWMutex
	nextID int
	items  map[int]User
}

func NewUserStore() *UserStore {
	return &UserStore{
		nextID: 1,
		items:  make(map[int]User),
	}
}

func (s *UserStore) Create(name string) User {
	s.mu.Lock()
	defer s.mu.Unlock()

	u := User{
		ID:        s.nextID,
		Name:      name,
		CreatedAt: time.Now().UTC(),
	}
	s.items[u.ID] = u
	s.nextID++
	return u
}

func (s *UserStore) Get(id int) (User, bool) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	u, ok := s.items[id]
	return u, ok
}

func (s *UserStore) Delete(id int) bool {
	s.mu.Lock()
	defer s.mu.Unlock()

	if _, ok := s.items[id]; !ok {
		return false
	}
	delete(s.items, id)
	return true
}

func (s *UserStore) List() []User {
	s.mu.RLock()
	defer s.mu.RUnlock()

	out := make([]User, 0, len(s.items))
	for _, u := range s.items {
		out = append(out, u)
	}
	return out
}
