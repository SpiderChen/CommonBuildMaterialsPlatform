package ops

import (
	"encoding/json"
	"errors"
	"os"
	"path/filepath"
	"sync"
)

type Store struct {
	path string
	mu   sync.Mutex
}

func NewStore(path string) *Store {
	return &Store{path: path}
}

func (s *Store) Snapshot() (AppData, error) {
	s.mu.Lock()
	defer s.mu.Unlock()
	return s.loadLocked()
}

func (s *Store) Update(fn func(*AppData) error) (AppData, error) {
	s.mu.Lock()
	defer s.mu.Unlock()

	data, err := s.loadLocked()
	if err != nil {
		return AppData{}, err
	}
	if err := fn(&data); err != nil {
		return AppData{}, err
	}
	if err := s.saveLocked(data); err != nil {
		return AppData{}, err
	}
	return data, nil
}

func (s *Store) loadLocked() (AppData, error) {
	bytes, err := os.ReadFile(s.path)
	if errors.Is(err, os.ErrNotExist) {
		data := SeedData()
		if err := s.saveLocked(data); err != nil {
			return AppData{}, err
		}
		return data, nil
	}
	if err != nil {
		return AppData{}, err
	}
	var data AppData
	if err := json.Unmarshal(bytes, &data); err != nil {
		return AppData{}, err
	}
	ensureNext(&data)
	return data, nil
}

func (s *Store) saveLocked(data AppData) error {
	ensureNext(&data)
	if err := os.MkdirAll(filepath.Dir(s.path), 0755); err != nil {
		return err
	}
	bytes, err := json.MarshalIndent(data, "", "  ")
	if err != nil {
		return err
	}
	tmp := s.path + ".tmp"
	if err := os.WriteFile(tmp, bytes, 0600); err != nil {
		return err
	}
	return os.Rename(tmp, s.path)
}

func ensureNext(data *AppData) {
	if data.Next == nil {
		data.Next = map[string]int64{}
	}
	ensureNextFor(data, "customer", len(data.Customers))
	ensureNextFor(data, "renewal", len(data.Renewals))
	ensureNextFor(data, "alert", len(data.Alerts))
	ensureNextFor(data, "package", len(data.UpdatePackages))
	ensureNextFor(data, "assignment", len(data.Assignments))
	ensureNextFor(data, "audit", len(data.AuditLogs))
}

func ensureNextFor(data *AppData, key string, length int) {
	if data.Next[key] == 0 {
		data.Next[key] = int64(length)
	}
}
