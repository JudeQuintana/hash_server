package hasher

import (
	"testing"
	"time"
)

// test that we get correctly generated hashes
func TestHasherGenerateHash(t *testing.T) {
	hashCases := []struct {
		password string
		hash     string
	}{
		{"angryMonkey", "ZEHhWB65gUlzdVwtDQArEyx+KVLzp/aTaRaPlBzYRIFj6vjFdqEb0Q5B8zVKCZ0vKbZPZklJz0Fd7su2A+gf7Q=="},
		{"baloo1", "74YskF/4/vdsObP1fBbGzZHVzIWKtBkpz0QotXEvGW5087z1MxwbwOc7dsxLZ98TBPAP1LIIdCOb0AsyESuQhw=="},
	}

	for _, c := range hashCases {
		got := GenerateHash(c.password)

		if got != c.hash {
			t.Errorf("Hash(%q): expected: %q, got: %q", c.password, c.hash, got)
		}
	}
}

// tests Hasher's NextID(), Add() and Get() as a whole
func TestHasherNextIDAddGet(t *testing.T) {
	h := NewHasher()
	password1 := "angryMonkey"
	password2 := "baloo1"
	timeNow := time.Now()

	hash, err := h.Get(1)

	if err == nil {
		t.Errorf("Get(1): returned nil error when error should have occurred")
	}
	if hash != "" {
		t.Errorf("Get(1): returned %s when empty", hash)
	}

	id := h.NextId()

	if id != 1 {
		t.Errorf("NextId(): expected: %d, got: returned %d as first id", 1, id)
	}

	h.Add(id, password1, timeNow)

	if length := len(h.hashes); length != 1 {
		t.Errorf("Add(%d, %s, %s): expected: underlying hash map of length %d, got: %d", id, password1, timeNow, 1, length)
	}
	if length := len(h.stats); length != 1 {
		t.Errorf("Add(%d, %s, %s): expected: underlying stats slice of length %d, got: %d", id, password1, timeNow, 1, length)
	}

	hash, err = h.Get(1)

	expectedHash := "ZEHhWB65gUlzdVwtDQArEyx+KVLzp/aTaRaPlBzYRIFj6vjFdqEb0Q5B8zVKCZ0vKbZPZklJz0Fd7su2A+gf7Q=="
	if hash != expectedHash || err != nil {
		t.Errorf("Get(2): expected: %s, got: %s", expectedHash, hash)
	}

	id = h.NextId()

	if id != 2 {
		t.Errorf("NextId(): expected: %d, got: returned %d as second id", 2, id)
	}

	h.Add(id, password2, timeNow)

	if length := len(h.hashes); length != 2 {
		t.Errorf("Add(%d, %s, %s): expected: underlying hash map of length %d, got: %d", id, password2, timeNow, 2, length)
	}
	if length := len(h.stats); length != 2 {
		t.Errorf("Add(%d, %s, %s): expected: underlying stats slice of length %d, got: %d", id, password2, timeNow, 2, length)
	}

	hash, err = h.Get(2)

	expectedHash = "74YskF/4/vdsObP1fBbGzZHVzIWKtBkpz0QotXEvGW5087z1MxwbwOc7dsxLZ98TBPAP1LIIdCOb0AsyESuQhw=="
	if hash != expectedHash || err != nil {
		t.Errorf("Get(2): expected: %s, got: %s", expectedHash, hash)
	}
}

// Test Hasher's GeneratedStats() returns correct values
func TestHasherGenerateStats(t *testing.T) {
	h := NewHasher()
	zero := int64(0)

	timeTwoSecondsAgo := time.Now().Add(-(2 * time.Second))
	timeSixSecondsAgo := time.Now().Add(-(6 * time.Second))
	twoThousandMilliseconds := (2 * time.Second).Nanoseconds() / int64(time.Millisecond)
	fourThousandMilliseconds := (4 * time.Second).Nanoseconds() / int64(time.Millisecond)

	stats := h.GenerateStats()

	if total := stats.Total; total != zero {
		t.Errorf("GenerateStats(): Stats.Total: expected: %d, got: %d", zero, total)
	}
	if average := stats.Average; average != zero {
		t.Errorf("GenerateStats(): Stats.Average: expected: %d, got: %d", zero, average)
	}

	h.Add(1, "angryMonkey", timeTwoSecondsAgo)
	stats = h.GenerateStats()

	if total := stats.Total; total != int64(1) {
		t.Errorf("GenerateStats(): Stats.Total: expected: %d, got: %d", int64(1), total)
	}
	if average := stats.Average; average != twoThousandMilliseconds {
		t.Errorf("GenerateStats(): Stats.Average: expected: %d, got: %d", twoThousandMilliseconds, average)
	}

	h.Add(2, "baloo1", timeSixSecondsAgo)
	stats = h.GenerateStats()

	if total := stats.Total; total != int64(2) {
		t.Errorf("GenerateStats(): Stats.Total: expected: %d,  got: %d", int64(2), total)
	}
	if average := stats.Average; average != fourThousandMilliseconds {
		t.Errorf("GenerateStats(): Stats.Average: expected: %d,  got: %d", fourThousandMilliseconds, average)
	}
}
