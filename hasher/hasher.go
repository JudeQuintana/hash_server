package hasher

import (
	"crypto/sha512"
	"encoding/base64"
	"fmt"
	"sync"
	"time"
)

// GenerateHash calculate SHA512 hash of a string
// encode the resulting hash string as Base64
func GenerateHash(s string) string {
	sha512Sum := sha512.Sum512([]byte(s))
	base64Str := base64.StdEncoding.EncodeToString(sha512Sum[:])
	return base64Str
}

// HashStats is used to store caclulated total and average stats
// according to the state of Hasher's internal stats slice
// only used for json marshalling
type HashStats struct {
	Total   int64 `json:"total"`
	Average int64 `json:"average"`
}

// Hasher is what keeps track of id mappings to hash strings
// uses RWMutex for concurrency safety and uses a counter for id generation
// stats is a slice of time.Durations so that you can do conversion
// on read.
// Use HashStats to set calculated stats and json marshalling
type Hasher struct {
	mu        sync.RWMutex
	idCounter int
	hashes    map[int]string
	stats     []time.Duration
}

// NextId Increment counter for next id
func (h *Hasher) NextId() int {
	h.mu.Lock()
	defer h.mu.Unlock()

	h.idCounter++
	return h.idCounter
}

// Add
// Makes sense to store id and hash into map and startTime into a separate
// time.Duration slice for stats calculation later
func (h *Hasher) Add(id int, password string, startTime time.Time) {
	h.mu.Lock()
	defer h.mu.Unlock()

	h.hashes[id] = GenerateHash(password)
	timeDiff := time.Now().Sub(startTime)
	h.stats = append(h.stats, timeDiff)
}

// Get
// wrapper for pulling hashes out the map
// went with returning error instead of return ok
func (h *Hasher) Get(id int) (string, error) {
	h.mu.RLock()
	defer h.mu.RUnlock()

	value, ok := h.hashes[id]

	if !ok {
		return "", fmt.Errorf("id not found")
	}

	return value, nil
}

// GenerateStats on the fly
func (h *Hasher) GenerateStats() HashStats {
	h.mu.RLock()
	defer h.mu.RUnlock()

	var zero, total, sum, milliseconds int64

	total = int64(len(h.stats))
	milliseconds = int64(time.Millisecond)

	// divide by zero, world go boom
	if total == zero {
		return HashStats{Total: zero, Average: zero}
	}

	// Convert time.Duration to milliseconds:
	// divide nanoseconds by int64 milliseconds constant
	for _, t := range h.stats {
		sum += t.Nanoseconds() / milliseconds
	}

	average := sum / total

	return HashStats{Total: total, Average: average}
}

// NewHasher
// life begins explicitly on allocation
func NewHasher() *Hasher {
	return &Hasher{
		mu:        sync.RWMutex{},
		idCounter: 0,
		hashes:    map[int]string{},
		stats:     []time.Duration{},
	}
}
