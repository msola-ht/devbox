package systeminfo

import (
	"context"
	"log"
	"sync"
	"time"
)

// Service coordinates system information collection and snapshot persistence.
type Service struct {
	mu     sync.Mutex
	latest *SystemSnapshot
}

// GetSnapshot returns the latest in-memory or saved snapshot.
// It returns null when no compatible snapshot exists.
func (s *Service) GetSnapshot() (*SystemSnapshot, error) {
	s.mu.Lock()
	defer s.mu.Unlock()
	if s.latest != nil {
		return s.latest, nil
	}
	return loadSnapshot()
}

// Refresh collects fresh information within five seconds and remembers it.
// Saving the snapshot to disk is best effort.
func (s *Service) Refresh(ctx context.Context) (SystemSnapshot, error) {
	ctx, cancel := context.WithTimeout(ctx, 5*time.Second)
	defer cancel()

	info, err := collect(ctx)
	if err != nil {
		return SystemSnapshot{}, err
	}
	snapshot := SystemSnapshot{Info: info, CollectedAt: time.Now().UTC()}

	s.mu.Lock()
	defer s.mu.Unlock()
	if err := ctx.Err(); err != nil {
		return SystemSnapshot{}, err
	}
	s.latest = &snapshot
	if err := saveSnapshot(snapshot); err != nil {
		log.Printf("systeminfo: save snapshot: %v", err)
	}
	return snapshot, nil
}
