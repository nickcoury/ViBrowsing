package window

import (
	"testing"
)

func TestNewStorageManager(t *testing.T) {
	sm := NewStorageManager()
	if sm == nil {
		t.Fatal("NewStorageManager returned nil")
	}
}

func TestStorageManagerEstimate(t *testing.T) {
	sm := NewStorageManager()

	est, err := sm.Estimate()
	if err != nil {
		t.Errorf("Estimate() returned error: %v", err)
	}
	if est == nil {
		t.Fatal("Estimate() returned nil")
	}
	if est.Quota == 0 {
		t.Error("Quota should not be zero")
	}
	if est.Usage > est.Quota {
		t.Errorf("Usage (%d) should not exceed Quota (%d)", est.Usage, est.Quota)
	}
}

func TestStorageManagerPersisted(t *testing.T) {
	sm := NewStorageManager()

	if sm.Persisted() {
		t.Error("Persisted() should be false initially")
	}

	sm.SetPersisted(true)
	if !sm.Persisted() {
		t.Error("Persisted() should be true after SetPersisted(true)")
	}
}

func TestStorageManagerPersist(t *testing.T) {
	sm := NewStorageManager()

	// With permission granted
	sm.SetPersistPermission(true)
	ok, err := sm.Persist()
	if err != nil {
		t.Errorf("Persist() returned error: %v", err)
	}
	if !ok {
		t.Error("Persist() should return true when permission is granted")
	}
	if !sm.Persisted() {
		t.Error("Persisted() should be true after successful Persist()")
	}

	// Without permission
	sm2 := NewStorageManager()
	sm2.SetPersistPermission(false)
	ok, err = sm2.Persist()
	if err != nil {
		t.Errorf("Persist() returned error: %v", err)
	}
	if ok {
		t.Error("Persist() should return false when permission is denied")
	}
}

func TestStorageManagerPersistPermission(t *testing.T) {
	sm := NewStorageManager()

	if !sm.PersistPermission() {
		t.Error("PersistPermission() should be true by default")
	}

	sm.SetPersistPermission(false)
	if sm.PersistPermission() {
		t.Error("PersistPermission() should be false after SetPersistPermission(false)")
	}
}

func TestStorageManagerSetQuota(t *testing.T) {
	sm := NewStorageManager()

	testQuota := uint64(500 * 1024 * 1024) // 500 MB
	sm.SetQuota(testQuota)

	est, _ := sm.Estimate()
	if est.Quota != testQuota {
		t.Errorf("Quota = %d, want %d", est.Quota, testQuota)
	}
}

func TestStorageManagerSetUsage(t *testing.T) {
	sm := NewStorageManager()

	testUsage := uint64(100 * 1024 * 1024) // 100 MB
	sm.SetUsage(testUsage)

	est, _ := sm.Estimate()
	if est.Usage != testUsage {
		t.Errorf("Usage = %d, want %d", est.Usage, testUsage)
	}
}

func TestStorageManagerOnChange(t *testing.T) {
	sm := NewStorageManager()

	callCount := 0
	sm.OnChange(func(est *StorageEstimate) {
		callCount++
	})

	// Estimate() should initialize estimate if nil
	sm.Estimate()
	// Trigger a change notification
	sm.NotifyEstimateChange()

	if callCount != 1 {
		t.Errorf("OnChange handler called %d times, want 1", callCount)
	}
}

func TestStorageManagerMultipleOnChange(t *testing.T) {
	sm := NewStorageManager()

	callCount1 := 0
	callCount2 := 0

	sm.OnChange(func(est *StorageEstimate) {
		callCount1++
	})
	sm.OnChange(func(est *StorageEstimate) {
		callCount2++
	})

	// Initialize estimate
	sm.Estimate()
	sm.NotifyEstimateChange()

	if callCount1 != 1 {
		t.Errorf("First OnChange handler called %d times, want 1", callCount1)
	}
	if callCount2 != 1 {
		t.Errorf("Second OnChange handler called %d times, want 1", callCount2)
	}
}
