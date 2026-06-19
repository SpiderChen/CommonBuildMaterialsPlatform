package appliance

import (
	"crypto/aes"
	"crypto/cipher"
	"crypto/rand"
	"crypto/sha256"
	"encoding/base64"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"sync"
	"time"
)

const vaultMagic = "CBMP1\n"

type DataStore interface {
	Load() error
	Snapshot() (AppData, error)
	Mutate(func(*AppData) error) error
}

type Store struct {
	mu   sync.RWMutex
	path string
	key  [32]byte
	data AppData
}

func NewStore(path string, passphrase string) *Store {
	if passphrase == "" {
		passphrase = "change-me-common-build-materials-platform-local-key"
	}
	return &Store{path: path, key: sha256.Sum256([]byte(passphrase))}
}

func (s *Store) Load() error {
	s.mu.Lock()
	defer s.mu.Unlock()

	if err := os.MkdirAll(filepath.Dir(s.path), 0755); err != nil {
		return err
	}
	raw, err := os.ReadFile(s.path)
	if errors.Is(err, os.ErrNotExist) {
		s.data = SeedData()
		return s.saveLocked()
	}
	if err != nil {
		return err
	}
	plain, err := s.decrypt(raw)
	if err != nil {
		return err
	}
	if err := json.Unmarshal(plain, &s.data); err != nil {
		return err
	}
	if s.data.Next == nil {
		s.data.Next = map[string]int64{}
	}
	changed := ensureSeedCredentials(&s.data)
	if ensureEnterpriseDefaults(&s.data) {
		changed = true
	}
	if changed {
		return s.saveLocked()
	}
	return nil
}

func (s *Store) Snapshot() (AppData, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()
	return cloneData(s.data)
}

func cloneData(data AppData) (AppData, error) {
	raw, err := json.Marshal(data)
	if err != nil {
		return AppData{}, err
	}
	var copied AppData
	if err := json.Unmarshal(raw, &copied); err != nil {
		return AppData{}, err
	}
	return copied, nil
}

func (s *Store) Mutate(fn func(*AppData) error) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	if err := fn(&s.data); err != nil {
		return err
	}
	return s.saveLocked()
}

func (s *Store) saveLocked() error {
	raw, err := json.MarshalIndent(s.data, "", "  ")
	if err != nil {
		return err
	}
	encrypted, err := s.encrypt(raw)
	if err != nil {
		return err
	}
	tmp := s.path + ".tmp"
	if err := os.WriteFile(tmp, encrypted, 0600); err != nil {
		return err
	}
	return os.Rename(tmp, s.path)
}

func (s *Store) encrypt(plain []byte) ([]byte, error) {
	block, err := aes.NewCipher(s.key[:])
	if err != nil {
		return nil, err
	}
	gcm, err := cipher.NewGCM(block)
	if err != nil {
		return nil, err
	}
	nonce := make([]byte, gcm.NonceSize())
	if _, err := io.ReadFull(rand.Reader, nonce); err != nil {
		return nil, err
	}
	ciphertext := gcm.Seal(nil, nonce, plain, nil)
	out := append([]byte(vaultMagic), nonce...)
	out = append(out, ciphertext...)
	return out, nil
}

func (s *Store) decrypt(raw []byte) ([]byte, error) {
	if len(raw) < len(vaultMagic) || string(raw[:len(vaultMagic)]) != vaultMagic {
		return nil, fmt.Errorf("invalid data vault header")
	}
	block, err := aes.NewCipher(s.key[:])
	if err != nil {
		return nil, err
	}
	gcm, err := cipher.NewGCM(block)
	if err != nil {
		return nil, err
	}
	payload := raw[len(vaultMagic):]
	if len(payload) < gcm.NonceSize() {
		return nil, fmt.Errorf("invalid data vault payload")
	}
	nonce := payload[:gcm.NonceSize()]
	ciphertext := payload[gcm.NonceSize():]
	return gcm.Open(nil, nonce, ciphertext, nil)
}

func nextID(data *AppData, key string) int64 {
	if data.Next == nil {
		data.Next = map[string]int64{}
	}
	data.Next[key]++
	return data.Next[key]
}

func nowString() string {
	return time.Now().Format("2006-01-02 15:04:05")
}

func todayString() string {
	return time.Now().Format("2006-01-02")
}

func periodString() string {
	return time.Now().Format("2006-01")
}

func publicUser(u User) User {
	u.TenantID = 0
	u.PasswordHash = ""
	u.PasswordSalt = ""
	u.MFASecret = ""
	u.MFALastUsedStep = 0
	return u
}

func publicUsers(items []User) []User {
	out := make([]User, 0, len(items))
	for _, item := range items {
		out = append(out, publicUser(item))
	}
	return out
}

func publicCompany(item Company) Company {
	item.TenantID = 0
	return item
}

func publicCompanies(items []Company) []Company {
	out := make([]Company, 0, len(items))
	for _, item := range items {
		out = append(out, publicCompany(item))
	}
	return out
}

func ensureSeedCredentials(data *AppData) bool {
	defaults := map[string]string{
		"admin":      "admin123",
		"dispatcher": "dispatch123",
		"driver":     "driver123",
		"customer":   "customer123",
		"quality":    "quality123",
	}
	changed := false
	for i := range data.Users {
		if data.Users[i].PasswordHash != "" && data.Users[i].PasswordSalt != "" {
			continue
		}
		password, ok := defaults[data.Users[i].Username]
		if !ok {
			continue
		}
		salt, hash := makePassword(password)
		data.Users[i].PasswordSalt = salt
		data.Users[i].PasswordHash = hash
		changed = true
	}
	return changed
}

func makePassword(password string) (salt string, hash string) {
	seed := make([]byte, 16)
	if _, err := io.ReadFull(rand.Reader, seed); err != nil {
		copy(seed, []byte(time.Now().Format(time.RFC3339Nano)))
	}
	salt = base64.RawStdEncoding.EncodeToString(seed)
	return salt, hashPassword(password, salt)
}

func hashPassword(password, salt string) string {
	sum := sha256.Sum256([]byte(salt + ":" + password))
	buf := sum[:]
	for i := 0; i < 60000; i++ {
		next := sha256.Sum256(append(buf, []byte(salt)...))
		buf = next[:]
	}
	return base64.RawStdEncoding.EncodeToString(buf)
}

func verifyPassword(password string, u User) bool {
	return hashPassword(password, u.PasswordSalt) == u.PasswordHash
}

func addAudit(data *AppData, user, action, resource string, resourceID int64, detail, ip string) {
	data.AuditLogs = append(data.AuditLogs, AuditLog{
		ID:         nextID(data, "audit"),
		User:       user,
		Action:     action,
		Resource:   resource,
		ResourceID: resourceID,
		Detail:     detail,
		IP:         ip,
		CreatedAt:  nowString(),
	})
	if len(data.AuditLogs) > 500 {
		data.AuditLogs = data.AuditLogs[len(data.AuditLogs)-500:]
	}
}
