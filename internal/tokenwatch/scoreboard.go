package tokenwatch

import (
	"fmt"
	"io"
	"sort"
	"sync"
	"time"

	"github.com/vaultwatch/internal/alert"
)

// ScoreEntry holds the aggregated risk score for a single token.
type ScoreEntry struct {
	TokenID   string
	Score     int
	TopReason string
	UpdatedAt time.Time
}

// Scoreboard tracks per-token risk scores derived from alert severity.
type Scoreboard struct {
	mu      sync.RWMutex
	entries map[string]*ScoreEntry
}

// NewScoreboard returns an empty Scoreboard.
func NewScoreboard() *Scoreboard {
	return &Scoreboard{entries: make(map[string]*ScoreEntry)}
}

// Record updates the score for the token referenced in the alert.
func (s *Scoreboard) Record(a alert.Alert) {
	points := scoreForLevel(a.Level)
	if points == 0 {
		return
	}
	s.mu.Lock()
	defer s.mu.Unlock()
	e, ok := s.entries[a.LeaseID]
	if !ok {
		e = &ScoreEntry{TokenID: a.LeaseID}
		s.entries[a.LeaseID] = e
	}
	e.Score += points
	e.TopReason = a.Message
	e.UpdatedAt = time.Now()
}

// Top returns the n highest-scoring entries, sorted descending.
func (s *Scoreboard) Top(n int) []ScoreEntry {
	s.mu.RLock()
	defer s.mu.RUnlock()
	all := make([]ScoreEntry, 0, len(s.entries))
	for _, e := range s.entries {
		all = append(all, *e)
	}
	sort.Slice(all, func(i, j int) bool { return all[i].Score > all[j].Score })
	if n > len(all) {
		n = len(all)
	}
	return all[:n]
}

// Reset clears all scores.
func (s *Scoreboard) Reset() {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.entries = make(map[string]*ScoreEntry)
}

// Print writes a human-readable table to w.
func (s *Scoreboard) Print(w io.Writer, n int) {
	top := s.Top(n)
	if len(top) == 0 {
		fmt.Fprintln(w, "scoreboard: no entries")
		return
	}
	fmt.Fprintf(w, "%-40s %6s  %s\n", "TOKEN", "SCORE", "TOP REASON")
	for _, e := range top {
		fmt.Fprintf(w, "%-40s %6d  %s\n", e.TokenID, e.Score, e.TopReason)
	}
}

func scoreForLevel(level alert.Level) int {
	switch level {
	case alert.LevelCritical:
		return 10
	case alert.LevelWarning:
		return 3
	case alert.LevelInfo:
		return 1
	default:
		return 0
	}
}
