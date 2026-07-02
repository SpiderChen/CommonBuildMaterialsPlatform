package appliance

import (
	"crypto/aes"
	"crypto/cipher"
	"crypto/rand"
	"crypto/sha256"
	"encoding/json"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"sort"
	"strings"
)

const backupMagic = "CBMP-BACKUP1\n"

type BackupInfo struct {
	Name      string `json:"name"`
	Path      string `json:"path"`
	Size      int64  `json:"size"`
	CreatedAt string `json:"createdAt"`
}

type BackupManager struct {
	dir string
	key [32]byte
}

func NewBackupManagerFromEnv() *BackupManager {
	dir := os.Getenv("CBMP_BACKUP_DIR")
	if dir == "" {
		dir = filepath.Join("data", "backups")
	}
	key := os.Getenv("CBMP_BACKUP_KEY")
	if key == "" {
		key = os.Getenv("CBMP_DATA_KEY")
	}
	if key == "" {
		key = "change-me-common-build-materials-platform-backup-key"
	}
	return &BackupManager{dir: dir, key: sha256.Sum256([]byte(key))}
}

func (b *BackupManager) Create(data AppData) (BackupInfo, error) {
	if err := os.MkdirAll(b.dir, 0700); err != nil {
		return BackupInfo{}, err
	}
	payload := map[string]interface{}{
		"createdAt":     nowString(),
		"schemaVersion": data.SchemaVersion,
		"data":          data,
	}
	raw, err := json.MarshalIndent(payload, "", "  ")
	if err != nil {
		return BackupInfo{}, err
	}
	encrypted, err := b.encrypt(raw)
	if err != nil {
		return BackupInfo{}, err
	}
	name := "cbmp-backup-" + strings.ReplaceAll(strings.ReplaceAll(nowString(), ":", ""), " ", "-") + ".vault"
	path := filepath.Join(b.dir, name)
	if err := os.WriteFile(path, encrypted, 0600); err != nil {
		return BackupInfo{}, err
	}
	info, err := os.Stat(path)
	if err != nil {
		return BackupInfo{}, err
	}
	return BackupInfo{Name: name, Path: path, Size: info.Size(), CreatedAt: nowString()}, nil
}

func (b *BackupManager) List() ([]BackupInfo, error) {
	entries, err := os.ReadDir(b.dir)
	if os.IsNotExist(err) {
		return []BackupInfo{}, nil
	}
	if err != nil {
		return nil, err
	}
	out := []BackupInfo{}
	for _, entry := range entries {
		if entry.IsDir() || !strings.HasSuffix(entry.Name(), ".vault") {
			continue
		}
		path := filepath.Join(b.dir, entry.Name())
		info, err := entry.Info()
		if err != nil {
			continue
		}
		out = append(out, BackupInfo{Name: entry.Name(), Path: path, Size: info.Size(), CreatedAt: info.ModTime().Format("2006-01-02 15:04:05")})
	}
	sort.Slice(out, func(i, j int) bool { return out[i].CreatedAt > out[j].CreatedAt })
	return out, nil
}

func (b *BackupManager) Restore(name string) (AppData, error) {
	if strings.Contains(name, "/") || strings.Contains(name, "\\") || name == "" {
		return AppData{}, fmt.Errorf("invalid backup name")
	}
	raw, err := os.ReadFile(filepath.Join(b.dir, name))
	if err != nil {
		return AppData{}, err
	}
	plain, err := b.decrypt(raw)
	if err != nil {
		return AppData{}, err
	}
	var payload struct {
		Data AppData `json:"data"`
	}
	if err := json.Unmarshal(plain, &payload); err != nil {
		return AppData{}, err
	}
	if payload.Data.Next == nil {
		payload.Data.Next = map[string]int64{}
	}
	ensureEnterpriseDefaults(&payload.Data)
	ensureWorkflowDefaults(&payload.Data)
	return payload.Data, nil
}

func (b *BackupManager) encrypt(plain []byte) ([]byte, error) {
	block, err := aes.NewCipher(b.key[:])
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
	out := append([]byte(backupMagic), nonce...)
	out = append(out, gcm.Seal(nil, nonce, plain, nil)...)
	return out, nil
}

func (b *BackupManager) decrypt(raw []byte) ([]byte, error) {
	if len(raw) < len(backupMagic) || string(raw[:len(backupMagic)]) != backupMagic {
		return nil, fmt.Errorf("invalid backup header")
	}
	block, err := aes.NewCipher(b.key[:])
	if err != nil {
		return nil, err
	}
	gcm, err := cipher.NewGCM(block)
	if err != nil {
		return nil, err
	}
	payload := raw[len(backupMagic):]
	if len(payload) < gcm.NonceSize() {
		return nil, fmt.Errorf("invalid backup payload")
	}
	return gcm.Open(nil, payload[:gcm.NonceSize()], payload[gcm.NonceSize():], nil)
}
