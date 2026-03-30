package store

import "sync"

type Store interface {
	Save(code string, originalURL string)
	Get(code string) (string, bool)
	GetUrlMap() *map[string]string
}

type MemoryStore struct {
	mu   sync.RWMutex
	urls map[string]string
}

func NewMemoryStore() *MemoryStore {
	return &MemoryStore{
		urls: make(map[string]string),
	}
}

func (s *MemoryStore) Save(code string, originalURL string) {
	s.mu.Lock()
	defer s.mu.Unlock()

	s.urls[code] = originalURL
}

func (s *MemoryStore) Get(code string) (string, bool) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	originalURL, ok := s.urls[code]

	return originalURL, ok
}

func (s *MemoryStore) GetUrlMap() *map[string]string {
	return &s.urls
}
