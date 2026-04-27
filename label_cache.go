package main

import (
	"errors"
	"strconv"
	"strings"
	"sync"
	"time"

	"go.mau.fi/whatsmeow/types/events"
)

// maxWhatsappLabels is the upper bound enforced by WhatsApp Business clients.
// We refuse to create label IDs above this number to keep parity with the
// official UI.
const maxWhatsappLabels = 20

// labelInfo is the cached representation of a WhatsApp label, fed exclusively
// from *events.LabelEdit (server echoes) and from optimistic inserts performed
// right after a successful BuildLabelEdit SendAppState call.
type labelInfo struct {
	ID        string    `json:"id"`
	Name      string    `json:"name"`
	Color     int32     `json:"color"`
	Deleted   bool      `json:"deleted"`
	Timestamp time.Time `json:"timestamp"`
}

var (
	// labelCache holds labels per userID. Map: userID -> labelID -> labelInfo.
	// Reads/writes are guarded by the embedded RWMutex.
	labelCache = struct {
		sync.RWMutex
		m map[string]map[string]labelInfo
	}{m: make(map[string]map[string]labelInfo)}

	// labelCreationLocks serialises label creation per userID. The outer
	// Mutex protects lazy initialisation of inner Mutexes; once obtained,
	// the inner lock is what HTTP handlers rely on to make name-based
	// auto-create idempotent against concurrent requests of the same client.
	labelCreationLocks = struct {
		sync.Mutex
		m map[string]*sync.Mutex
	}{m: make(map[string]*sync.Mutex)}
)

// labelCreationLockFor returns (lazy-initialising) the per-user creation lock.
func labelCreationLockFor(userID string) *sync.Mutex {
	labelCreationLocks.Lock()
	defer labelCreationLocks.Unlock()
	if mu, ok := labelCreationLocks.m[userID]; ok {
		return mu
	}
	mu := &sync.Mutex{}
	labelCreationLocks.m[userID] = mu
	return mu
}

// normalizeLabelName produces the canonical form used to compare labels by
// name (case-insensitive, trimmed).
func normalizeLabelName(name string) string {
	return strings.ToLower(strings.TrimSpace(name))
}

// labelCacheUpsertFromEdit ingests a LabelEdit event into the cache.
// Safe to call from whatsmeow goroutines (event handler).
func labelCacheUpsertFromEdit(userID string, evt *events.LabelEdit) {
	if evt == nil || evt.LabelID == "" {
		return
	}
	labelCache.Lock()
	defer labelCache.Unlock()
	if labelCache.m[userID] == nil {
		labelCache.m[userID] = make(map[string]labelInfo)
	}
	info := labelInfo{
		ID:        evt.LabelID,
		Timestamp: evt.Timestamp,
	}
	if evt.Action != nil {
		info.Name = evt.Action.GetName()
		info.Color = evt.Action.GetColor()
		info.Deleted = evt.Action.GetDeleted()
	}
	labelCache.m[userID][evt.LabelID] = info
}

// labelCacheGetAll returns a snapshot copy of every label cached for the user.
// Includes deleted labels so callers can render them faithfully if desired.
func labelCacheGetAll(userID string) []labelInfo {
	labelCache.RLock()
	defer labelCache.RUnlock()
	src, ok := labelCache.m[userID]
	if !ok {
		return []labelInfo{}
	}
	out := make([]labelInfo, 0, len(src))
	for _, v := range src {
		out = append(out, v)
	}
	return out
}

// labelCacheFindByNormalizedName returns a copy of the first non-deleted label
// whose normalised name matches, or nil. Idempotency relies on this lookup
// being performed under the per-user creation lock.
func labelCacheFindByNormalizedName(userID, normalized string) *labelInfo {
	if normalized == "" {
		return nil
	}
	labelCache.RLock()
	defer labelCache.RUnlock()
	src, ok := labelCache.m[userID]
	if !ok {
		return nil
	}
	for _, v := range src {
		if v.Deleted {
			continue
		}
		if normalizeLabelName(v.Name) == normalized {
			cp := v
			return &cp
		}
	}
	return nil
}

// labelCacheNextID picks the smallest free ID in [1..maxWhatsappLabels].
// IDs occupied by deleted labels are reused (we overwrite them via BuildLabelEdit).
// Returns an error when no slot is free.
func labelCacheNextID(userID string) (string, error) {
	labelCache.RLock()
	defer labelCache.RUnlock()
	src := labelCache.m[userID]
	used := make(map[int]bool, len(src))
	activeCount := 0
	for id, v := range src {
		if !v.Deleted {
			activeCount++
		} else {
			continue
		}
		if n, err := strconv.Atoi(id); err == nil {
			used[n] = true
		}
	}
	if activeCount >= maxWhatsappLabels {
		return "", errors.New("WhatsApp Business label limit reached (20)")
	}
	for i := 1; i <= maxWhatsappLabels; i++ {
		if !used[i] {
			return strconv.Itoa(i), nil
		}
	}
	return "", errors.New("no free label ID available")
}

// labelCacheOptimisticInsert seeds the cache with a freshly created label so
// that concurrent in-flight HTTP calls observe it before the server echoes the
// LabelEdit event back. Always called under the per-user creation lock.
func labelCacheOptimisticInsert(userID string, info labelInfo) {
	if info.ID == "" {
		return
	}
	labelCache.Lock()
	defer labelCache.Unlock()
	if labelCache.m[userID] == nil {
		labelCache.m[userID] = make(map[string]labelInfo)
	}
	labelCache.m[userID][info.ID] = info
}

// labelCacheClear removes all cached labels for a user (called on logout).
// Also drops the per-user creation lock to avoid stale locks accumulating.
func labelCacheClear(userID string) {
	labelCache.Lock()
	delete(labelCache.m, userID)
	labelCache.Unlock()

	labelCreationLocks.Lock()
	delete(labelCreationLocks.m, userID)
	labelCreationLocks.Unlock()
}
