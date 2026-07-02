package appliance

import (
	"context"
	"crypto/aes"
	"crypto/cipher"
	"crypto/rand"
	"crypto/sha256"
	"encoding/json"
	"fmt"
	"io"
	"os"
	"strings"
	"sync"
	"time"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

type PostgresStore struct {
	mu   sync.RWMutex
	dsn  string
	key  [32]byte
	pool *pgxpool.Pool
	data AppData
}

func NewPostgresStore(dsn string, passphrase string) *PostgresStore {
	if passphrase == "" {
		passphrase = "change-me-common-build-materials-platform-local-key"
	}
	return &PostgresStore{dsn: dsn, key: sha256.Sum256([]byte(passphrase))}
}

func (s *PostgresStore) Load() error {
	s.mu.Lock()
	defer s.mu.Unlock()

	ctx, cancel := context.WithTimeout(context.Background(), 15*time.Second)
	defer cancel()
	pool, err := pgxpool.New(ctx, s.dsn)
	if err != nil {
		return err
	}
	if err := pool.Ping(ctx); err != nil {
		pool.Close()
		return err
	}
	s.pool = pool
	if err := s.ensureSchema(ctx); err != nil {
		return err
	}
	if postgresLoadFromDomainRows() {
		if data, ok, err := s.loadDomainRows(ctx); err != nil {
			return err
		} else if ok {
			s.data = data
			if s.data.Next == nil {
				s.data.Next = map[string]int64{}
			}
			changed := ensureSeedCredentials(&s.data)
			if ensureEnterpriseDefaults(&s.data) {
				changed = true
			}
			if ensureWorkflowDefaults(&s.data) {
				changed = true
			}
			if changed {
				return s.saveLocked(ctx)
			}
			return nil
		}
	}
	var encrypted []byte
	err = s.pool.QueryRow(ctx, `select payload from cbmp_snapshot where id = 'default'`).Scan(&encrypted)
	if err == pgx.ErrNoRows {
		if data, ok, err := s.loadDomainRows(ctx); err != nil {
			return err
		} else if ok {
			s.data = data
			if s.data.Next == nil {
				s.data.Next = map[string]int64{}
			}
			changed := ensureSeedCredentials(&s.data)
			if ensureEnterpriseDefaults(&s.data) {
				changed = true
			}
			if ensureWorkflowDefaults(&s.data) {
				changed = true
			}
			if changed {
				return s.saveLocked(ctx)
			}
			return nil
		}
		s.data = initialStoreData()
		ensureEnterpriseDefaults(&s.data)
		ensureWorkflowDefaults(&s.data)
		return s.saveLocked(ctx)
	}
	if err != nil {
		return err
	}
	plain, err := s.decrypt(encrypted)
	if err != nil {
		if data, ok, domainErr := s.loadDomainRows(ctx); domainErr == nil && ok {
			s.data = data
			return s.saveLocked(ctx)
		}
		return err
	}
	if err := json.Unmarshal(plain, &s.data); err != nil {
		if data, ok, domainErr := s.loadDomainRows(ctx); domainErr == nil && ok {
			s.data = data
			return s.saveLocked(ctx)
		}
		return err
	}
	if s.data.Next == nil {
		s.data.Next = map[string]int64{}
	}
	changed := ensureSeedCredentials(&s.data)
	if ensureEnterpriseDefaults(&s.data) {
		changed = true
	}
	if ensureWorkflowDefaults(&s.data) {
		changed = true
	}
	if changed {
		return s.saveLocked(ctx)
	}
	return nil
}

func (s *PostgresStore) Snapshot() (AppData, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()
	return cloneData(s.data)
}

func (s *PostgresStore) Mutate(fn func(*AppData) error) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	if err := fn(&s.data); err != nil {
		return err
	}
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	return s.saveLocked(ctx)
}

func (s *PostgresStore) ensureSchema(ctx context.Context) error {
	_, err := s.pool.Exec(ctx, `
		create table if not exists cbmp_snapshot (
			id text primary key,
			payload bytea not null,
			checksum text not null,
			updated_at timestamptz not null default now()
		);
		create table if not exists cbmp_outbox (
			id bigserial primary key,
			topic text not null,
			payload jsonb not null,
			status text not null default 'pending',
			created_at timestamptz not null default now()
		);
	`+postgresDomainSchemaSQL()+postgresProjectionSchemaSQL())
	return err
}

func (s *PostgresStore) saveLocked(ctx context.Context) error {
	raw, err := json.MarshalIndent(s.data, "", "  ")
	if err != nil {
		return err
	}
	encrypted, err := s.encrypt(raw)
	if err != nil {
		return err
	}
	checksum := sha256.Sum256(raw)
	checksumText := fmt.Sprintf("%x", checksum[:])
	tx, err := s.pool.Begin(ctx)
	if err != nil {
		return err
	}
	committed := false
	defer func() {
		if !committed {
			_ = tx.Rollback(ctx)
		}
	}()
	_, err = tx.Exec(ctx, `
		insert into cbmp_snapshot (id, payload, checksum, updated_at)
		values ('default', $1, $2, now())
		on conflict (id)
		do update set payload = excluded.payload, checksum = excluded.checksum, updated_at = now()
	`, encrypted, checksumText)
	if err != nil {
		return err
	}
	if err := s.refreshDomainRows(ctx, tx, s.data, checksumText); err != nil {
		return err
	}
	if err := s.refreshBusinessProjections(ctx, tx, s.data, checksumText); err != nil {
		return err
	}
	if err := tx.Commit(ctx); err != nil {
		return err
	}
	committed = true
	return nil
}

func (s *PostgresStore) encrypt(plain []byte) ([]byte, error) {
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

func (s *PostgresStore) decrypt(raw []byte) ([]byte, error) {
	if len(raw) < len(vaultMagic) || string(raw[:len(vaultMagic)]) != vaultMagic {
		return nil, fmt.Errorf("invalid postgres vault header")
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
		return nil, fmt.Errorf("invalid postgres vault payload")
	}
	return gcm.Open(nil, payload[:gcm.NonceSize()], payload[gcm.NonceSize():], nil)
}

func postgresLoadFromDomainRows() bool {
	value := strings.TrimSpace(os.Getenv("CBMP_POSTGRES_LOAD_FROM_DOMAIN"))
	return value == "1" || strings.EqualFold(value, "true") || strings.EqualFold(value, "yes")
}
