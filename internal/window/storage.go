package window

import (
	"sync"
)

// StorageEstimate represents the estimated storage quota and usage.
type StorageEstimate struct {
	// The total amount of storage available in bytes.
	Quota uint64 `json:"quota"`
	// The amount of storage used by the origin in bytes.
	Usage uint64 `json:"usage"`
	// The amount of data stored that is unique to this origin in bytes.
	UsageDetails struct {
		// Database storage usage
		Database uint64 `json:"database"`
		// Cache storage usage
		Cache uint64 `json:"cache"`
		// Shared storage usage
		Shared uint64 `json:"shared"`
	} `json:"usageDetails"`
}

// StorageManager represents the StorageManager API (navigator.storage).
// Provides access to the Storage API for checking and requesting storage quotas.
type StorageManager struct {
	mu sync.RWMutex
	// persisted indicates whether the origin has persistent storage permission.
	persisted bool
	// estimate stores the last estimated quota/usage.
	estimate *StorageEstimate
	// persistPermission indicates if the origin can request persistent storage.
	persistPermission bool
	// onchange handlers for estimate changes
	onchange []func(*StorageEstimate)
}

// NewStorageManager creates a new StorageManager instance.
func NewStorageManager() *StorageManager {
	return &StorageManager{
		persisted:        false,
		persistPermission: true, // Default to true for now
		estimate: &StorageEstimate{
			Quota: 1024 * 1024 * 1024, // 1 GB default quota
			Usage: 0,
		},
	}
}

// Estimate returns the estimated storage quota and current usage for the origin.
// This implements navigator.storage.estimate().
func (sm *StorageManager) Estimate() (*StorageEstimate, error) {
	sm.mu.Lock()
	defer sm.mu.Unlock()

	// If we already have an estimate, return it
	if sm.estimate != nil {
		return sm.estimate, nil
	}

	// In a real implementation, this would query the OS/filesystem for actual usage.
	// For ViBrowsing, we return reasonable defaults.

	est := &StorageEstimate{
		Quota: 1024 * 1024 * 1024, // 1 GB
		Usage: 0,
	}
	est.UsageDetails.Database = 0
	est.UsageDetails.Cache = 0
	est.UsageDetails.Shared = 0

	sm.estimate = est
	return est, nil
}

// Persisted returns true if the origin has persistent storage permission.
func (sm *StorageManager) Persisted() bool {
	sm.mu.RLock()
	defer sm.mu.RUnlock()
	return sm.persisted
}

// Persist returns a promise that resolves to true if the caller was able to
// establish persistent storage permission.
// In this implementation, it returns immediately with the current permission state.
func (sm *StorageManager) Persist() (bool, error) {
	sm.mu.Lock()
	defer sm.mu.Unlock()

	if sm.persistPermission {
		sm.persisted = true
		return true, nil
	}
	return false, nil
}

// PersistPermission returns whether the origin has permission to use persistent storage.
func (sm *StorageManager) PersistPermission() bool {
	sm.mu.RLock()
	defer sm.mu.RUnlock()
	return sm.persistPermission
}

// SetPersistPermission sets whether the origin has permission for persistent storage.
// This is primarily for testing purposes.
func (sm *StorageManager) SetPersistPermission(allowed bool) {
	sm.mu.Lock()
	defer sm.mu.Unlock()
	sm.persistPermission = allowed
}

// SetPersisted sets the persisted state.
// This is primarily for testing purposes.
func (sm *StorageManager) SetPersisted(persisted bool) {
	sm.mu.Lock()
	defer sm.mu.Unlock()
	sm.persisted = persisted
}

// SetQuota sets the storage quota for testing purposes.
func (sm *StorageManager) SetQuota(quota uint64) {
	sm.mu.Lock()
	defer sm.mu.Unlock()
	if sm.estimate == nil {
		sm.estimate = &StorageEstimate{}
	}
	sm.estimate.Quota = quota
}

// SetUsage sets the storage usage for testing purposes.
func (sm *StorageManager) SetUsage(usage uint64) {
	sm.mu.Lock()
	defer sm.mu.Unlock()
	if sm.estimate == nil {
		sm.estimate = &StorageEstimate{}
	}
	sm.estimate.Usage = usage
}

// OnChange registers a handler to be called when the storage estimate changes.
func (sm *StorageManager) OnChange(handler func(*StorageEstimate)) {
	sm.mu.Lock()
	defer sm.mu.Unlock()
	sm.onchange = append(sm.onchange, handler)
}

// NotifyEstimateChange notifies all onchange handlers of an estimate change.
func (sm *StorageManager) NotifyEstimateChange() {
	sm.mu.RLock()
	handlers := make([]func(*StorageEstimate), len(sm.onchange))
	copy(handlers, sm.onchange)
	est := sm.estimate
	sm.mu.RUnlock()

	for _, handler := range handlers {
		handler(est)
	}
}
