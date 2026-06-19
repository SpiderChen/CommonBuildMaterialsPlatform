package appliance

import (
	"os"
	"path/filepath"
	"testing"
)

func TestPasswordHash(t *testing.T) {
	salt, hash := makePassword("secret")
	user := User{PasswordSalt: salt, PasswordHash: hash}
	if !verifyPassword("secret", user) {
		t.Fatalf("expected password to verify")
	}
	if verifyPassword("wrong", user) {
		t.Fatalf("expected wrong password to fail")
	}
}

func TestEncryptedStoreRoundTrip(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "app.vault")
	store := NewStore(path, "unit-test-key")
	if err := store.Load(); err != nil {
		t.Fatalf("load seed: %v", err)
	}
	raw, err := os.ReadFile(path)
	if err != nil {
		t.Fatalf("read vault: %v", err)
	}
	if len(raw) == 0 || string(raw[:len(vaultMagic)]) != vaultMagic {
		t.Fatalf("vault header missing")
	}
	if string(raw) == "{" {
		t.Fatalf("vault should not be plain json")
	}
	store2 := NewStore(path, "unit-test-key")
	if err := store2.Load(); err != nil {
		t.Fatalf("reload vault: %v", err)
	}
	data, _ := store2.Snapshot()
	if len(data.Orders) == 0 || len(data.Modules) == 0 {
		t.Fatalf("expected seeded data")
	}
}

func TestSeedUsersCanLogin(t *testing.T) {
	data := SeedData()
	cases := map[string]string{
		"admin":      "admin123",
		"dispatcher": "dispatch123",
		"driver":     "driver123",
		"customer":   "customer123",
		"quality":    "quality123",
	}
	for username, password := range cases {
		found := false
		for _, user := range data.Users {
			if user.Username == username {
				found = true
				if !verifyPassword(password, user) {
					t.Fatalf("seed password failed for %s", username)
				}
			}
		}
		if !found {
			t.Fatalf("missing seed user %s", username)
		}
	}
}
