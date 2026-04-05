package observer

import (
	"strings"
	"sync"
)

// MutationObserver monitors DOM mutations and notifies observers.
// It watches for changes to attributes, child nodes, and subtree modifications.

// MutationRecordType represents the type of mutation.
type MutationRecordType string

const (
	MutationAttributes   MutationRecordType = "attributes"
	MutationChildList    MutationRecordType = "childList"
	MutationSubtree     MutationRecordType = "subtree"
)

// MutationRecord represents a single mutation observed by the MutationObserver.
type MutationRecord struct {
	Type       MutationRecordType // "attributes", "childList", or "characterData"
	Target     interface{}       // The node that was mutated
	AttributeName string         // For attributes mutations: the attribute name
	OldValue   string            // For attributes/text mutations: the previous value
	AddedNodes []interface{}     // For childList: added nodes
	RemovedNodes []interface{}   // For childList: removed nodes
}

// MutationObserverInit provides options for configuring the MutationObserver.
type MutationObserverInit struct {
	ChildList     bool   // Observe additions/removals of child nodes
	Subtree       bool   // Observe mutations to all descendants
	Attributes    bool   // Observe attribute mutations
	AttributeOldValue string // The attribute's previous value
	CharacterData bool   // Observe mutations to text content
	CharacterDataOldValue bool // The text's previous value
	AttributeFilter []string // Only observe specific attributes
}

// MutationObserver watches for DOM mutations and fires a callback.
type MutationObserver struct {
	Callback func([]*MutationRecord) // Called when mutations are observed
	Options  *MutationObserverInit  // Configuration

	mu         sync.Mutex
	records    []*MutationRecord
	observing  bool
}

// NewMutationObserver creates a new MutationObserver with the given callback.
func NewMutationObserver(callback func([]*MutationRecord)) *MutationObserver {
	return &MutationObserver{
		Callback: callback,
		Options:  &MutationObserverInit{},
	}
}

// Observe starts observing the given target node for mutations.
// The target should implement the Node interface (or a compatible type).
// Valid options: childList, subtree, attributes, attributeOldValue,
// characterData, characterDataOldValue, attributeFilter.
func (obs *MutationObserver) Observe(target interface{}, options *MutationObserverInit) {
	obs.mu.Lock()
	defer obs.mu.Unlock()

	if options != nil {
		obs.Options = options
	}
	obs.observing = true
}

// ObserveWithConfig starts observing with explicit configuration.
func (obs *MutationObserver) ObserveWithConfig(target interface{}, childList, subtree, attributes bool, attributeFilter []string) {
	obs.mu.Lock()
	defer obs.mu.Unlock()

	obs.Options = &MutationObserverInit{
		ChildList:       childList,
		Subtree:         subtree,
		Attributes:      attributes,
		AttributeFilter: attributeFilter,
	}
	obs.observing = true
}

// Disconnect stops the MutationObserver from receiving notifications.
func (obs * MutationObserver) Disconnect() {
	obs.mu.Lock()
	defer obs.mu.Unlock()
	obs.observing = false
	obs.records = nil
}

// TakeRecords returns a copy of the current mutation records and clears them.
func (obs *MutationObserver) TakeRecords() []*MutationRecord {
	obs.mu.Lock()
	defer obs.mu.Unlock()

	records := make([]*MutationRecord, len(obs.records))
	copy(records, obs.records)
	obs.records = nil
	return records
}

// NotifyRecords is called internally when mutations should be reported.
// In a real browser, this would be called by the JavaScript engine.
// This implementation stores records and calls the callback synchronously.
func (obs *MutationObserver) NotifyRecords(records []*MutationRecord) {
	obs.mu.Lock()
	defer obs.mu.Unlock()

	if !obs.observing || obs.Callback == nil {
		return
	}

	// Call the callback with the records
	obs.Callback(records)
}

// RecordMutation records a single mutation for later reporting.
// This would be called by the DOM implementation when changes occur.
func (obs *MutationObserver) RecordMutation(target interface{}, mutationType MutationRecordType, options *MutationRecord) {
	obs.mu.Lock()
	defer obs.mu.Unlock()

	if !obs.observing {
		return
	}

	record := &MutationRecord{
		Type:   mutationType,
		Target: target,
	}

	if options != nil {
		if options.AttributeName != "" {
			record.AttributeName = options.AttributeName
		}
		if options.OldValue != "" {
			record.OldValue = options.OldValue
		}
		if options.AddedNodes != nil {
			record.AddedNodes = options.AddedNodes
		}
		if options.RemovedNodes != nil {
			record.RemovedNodes = options.RemovedNodes
		}
	}

	obs.records = append(obs.records, record)

	// In a real browser, the callback is called asynchronously (microtask)
	// For simplicity, we call it when TakeRecords is invoked
}

// RecordAttributeMutation records an attribute change.
func (obs *MutationObserver) RecordAttributeMutation(target interface{}, name, oldValue string) {
	if !obs.Options.Attributes {
		return
	}
	// Check attribute filter
	if len(obs.Options.AttributeFilter) > 0 {
		found := false
		for _, attr := range obs.Options.AttributeFilter {
			if strings.EqualFold(attr, name) {
				found = true
				break
			}
		}
		if !found {
			return
		}
	}

	obs.RecordMutation(target, MutationAttributes, &MutationRecord{
		AttributeName: name,
		OldValue:      oldValue,
	})
}

// RecordChildListMutation records child node additions/removals.
func (obs *MutationObserver) RecordChildListMutation(target interface{}, added, removed []interface{}) {
	if !obs.Options.ChildList {
		return
	}

	obs.RecordMutation(target, MutationChildList, &MutationRecord{
		AddedNodes:   added,
		RemovedNodes: removed,
	})
}

// RecordCharacterDataMutation records text content changes.
func (obs *MutationObserver) RecordCharacterDataMutation(target interface{}, oldValue string) {
	if !obs.Options.CharacterData {
		return
	}

	obs.RecordMutation(target, MutationChildList, &MutationRecord{
		OldValue: oldValue,
	})
}
