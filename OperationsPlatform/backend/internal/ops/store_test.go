package ops

import (
	"path/filepath"
	"testing"
)

func TestStoreStartsEmptyByDefault(t *testing.T) {
	t.Setenv("CBM_OPS_SEED_DEMO", "")

	store := NewStore(filepath.Join(t.TempDir(), "ops.json"))
	data, err := store.Snapshot()
	if err != nil {
		t.Fatalf("snapshot: %v", err)
	}
	if len(data.Customers) != 0 || len(data.Alerts) != 0 || len(data.UpdatePackages) != 0 {
		t.Fatalf("new store should not seed demo data by default: %+v", data)
	}
}

func TestStoreSeedsDemoDataWhenEnabled(t *testing.T) {
	t.Setenv("CBM_OPS_SEED_DEMO", "1")

	store := NewStore(filepath.Join(t.TempDir(), "ops.json"))
	data, err := store.Snapshot()
	if err != nil {
		t.Fatalf("snapshot: %v", err)
	}
	if len(data.Customers) == 0 || len(data.Alerts) == 0 || len(data.UpdatePackages) == 0 {
		t.Fatalf("demo seed was not loaded: %+v", data)
	}
}
