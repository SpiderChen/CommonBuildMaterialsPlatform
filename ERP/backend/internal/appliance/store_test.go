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
	t.Setenv("CBMP_SEED_DEMO", "")
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
	if len(data.Orders) != 0 || len(data.Customers) != 0 || len(data.DispatchOrders) != 0 {
		t.Fatalf("new store should not seed demo business data by default: orders=%d customers=%d dispatch=%d", len(data.Orders), len(data.Customers), len(data.DispatchOrders))
	}
	if len(data.Modules) == 0 || len(data.DataDictionaries) == 0 || len(data.Users) != 1 || data.Users[0].Username != "admin" {
		t.Fatalf("new store should keep runtime defaults and admin user, got modules=%d dicts=%d users=%+v", len(data.Modules), len(data.DataDictionaries), data.Users)
	}
}

func TestEncryptedStoreSeedsDemoDataWhenEnabled(t *testing.T) {
	t.Setenv("CBMP_SEED_DEMO", "1")
	enableTestSeedPasswords(t)
	store := NewStore(filepath.Join(t.TempDir(), "app.vault"), "unit-test-key")
	if err := store.Load(); err != nil {
		t.Fatalf("load demo seed: %v", err)
	}
	data, _ := store.Snapshot()
	if len(data.Orders) == 0 || len(data.Customers) == 0 || len(data.DispatchOrders) == 0 {
		t.Fatalf("demo seed should include business sample data: orders=%d customers=%d dispatch=%d", len(data.Orders), len(data.Customers), len(data.DispatchOrders))
	}
}

func TestEncryptedStorePurgesUnmodifiedDemoBusinessSeedByDefault(t *testing.T) {
	t.Setenv("CBMP_SEED_DEMO", "1")
	enableTestSeedPasswords(t)
	path := filepath.Join(t.TempDir(), "app.vault")
	store := NewStore(path, "unit-test-key")
	if err := store.Load(); err != nil {
		t.Fatalf("load demo seed: %v", err)
	}
	seeded, _ := store.Snapshot()
	if len(seeded.Orders) == 0 || len(seeded.Customers) == 0 || len(seeded.DispatchOrders) == 0 {
		t.Fatalf("demo seed fixture did not include business data: orders=%d customers=%d dispatch=%d", len(seeded.Orders), len(seeded.Customers), len(seeded.DispatchOrders))
	}

	t.Setenv("CBMP_SEED_DEMO", "")
	store = NewStore(path, "unit-test-key")
	if err := store.Load(); err != nil {
		t.Fatalf("reload old demo seed without demo mode: %v", err)
	}
	data, _ := store.Snapshot()
	if len(data.Orders) != 0 || len(data.Customers) != 0 || len(data.DispatchOrders) != 0 || len(data.TransportSettlements) != 0 {
		t.Fatalf("old unmodified demo business data should be purged by default: orders=%d customers=%d dispatch=%d carrierSettlements=%d", len(data.Orders), len(data.Customers), len(data.DispatchOrders), len(data.TransportSettlements))
	}
	if len(data.Modules) == 0 || len(data.DataDictionaries) == 0 || len(data.Users) != 1 || data.Users[0].Username != "admin" {
		t.Fatalf("purge should keep runtime foundation and admin only, got modules=%d dicts=%d users=%+v", len(data.Modules), len(data.DataDictionaries), data.Users)
	}
	if !verifyPassword(testSeedPasswords["admin"], data.Users[0]) {
		t.Fatalf("purge should preserve admin credential")
	}
}

func TestSeedUsersCanLogin(t *testing.T) {
	enableTestSeedPasswords(t)

	data := SeedData()
	for username, password := range testSeedPasswords {
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

func TestBuiltinSuperAdminCanLoginWithoutExplicitPasswordEnv(t *testing.T) {
	clearSeedPasswordEnv(t)

	data := SeedData()
	admin, ok := userByUsername(data.Users, builtinSuperAdminUsername)
	if !ok {
		t.Fatalf("missing builtin super admin")
	}
	if admin.Status != "active" {
		t.Fatalf("builtin super admin should be active, got %q", admin.Status)
	}
	if !verifyPassword(builtinSuperAdminPassword, admin) {
		t.Fatalf("builtin super admin default password failed")
	}

	for username, defaultPassword := range testSeedPasswords {
		if username == builtinSuperAdminUsername {
			continue
		}
		user, ok := userByUsername(data.Users, username)
		if !ok {
			t.Fatalf("missing seed user %s", username)
		}
		if user.PasswordHash != "" || user.PasswordSalt != "" {
			t.Fatalf("seed user %s should not ship password material", username)
		}
		if user.Status != "pending" {
			t.Fatalf("seed user %s should require activation, got %q", username, user.Status)
		}
		if verifyPassword(defaultPassword, user) {
			t.Fatalf("seed user %s accepted default password without opt-in", username)
		}
	}
}

func TestInitialAdminPasswordActivatesOnlyAdmin(t *testing.T) {
	clearSeedPasswordEnv(t)
	t.Setenv("CBMP_INITIAL_ADMIN_PASSWORD", "customer-supplied-admin-secret")

	data := SeedData()
	admin, ok := userByUsername(data.Users, "admin")
	if !ok {
		t.Fatalf("missing admin seed user")
	}
	if admin.Status != "active" {
		t.Fatalf("admin should be active with explicit initial password, got %q", admin.Status)
	}
	if !verifyPassword("customer-supplied-admin-secret", admin) {
		t.Fatalf("admin did not accept explicit initial password")
	}
	if verifyPassword(builtinSuperAdminPassword, admin) {
		t.Fatalf("admin accepted default password despite explicit initial password")
	}

	for username, defaultPassword := range testSeedPasswords {
		if username == builtinSuperAdminUsername {
			continue
		}
		user, ok := userByUsername(data.Users, username)
		if !ok {
			t.Fatalf("missing seed user %s", username)
		}
		if user.Status != "pending" || user.PasswordHash != "" || user.PasswordSalt != "" {
			t.Fatalf("seed user %s should remain pending without a configured password, got %+v", username, user)
		}
		if verifyPassword(defaultPassword, user) {
			t.Fatalf("seed user %s accepted default password without opt-in", username)
		}
	}
}

func TestEnsureSeedCredentialsBackfillsOnlyBuiltinSuperAdminWithoutExplicitEnv(t *testing.T) {
	clearSeedPasswordEnv(t)

	data := AppData{Users: []User{
		{ID: 1, Username: builtinSuperAdminUsername, Status: "active"},
		{ID: 2, Username: "dispatcher", Status: "active"},
	}}
	if !ensureSeedCredentials(&data) {
		t.Fatalf("expected missing seed credential state to be normalized")
	}

	admin := data.Users[0]
	if admin.Status != "active" {
		t.Fatalf("builtin super admin should remain active, got %q", admin.Status)
	}
	if !verifyPassword(builtinSuperAdminPassword, admin) {
		t.Fatalf("builtin super admin should be backfilled with default password")
	}

	dispatcher := data.Users[1]
	if dispatcher.Status != "pending" {
		t.Fatalf("non-admin seed user should move to pending, got %q", dispatcher.Status)
	}
	if dispatcher.PasswordHash != "" || dispatcher.PasswordSalt != "" {
		t.Fatalf("non-admin seed user should not be backfilled without explicit env")
	}
	if verifyPassword(testSeedPasswords["dispatcher"], dispatcher) {
		t.Fatalf("non-admin seed user accepted default password after credential normalization")
	}
}
