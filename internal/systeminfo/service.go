package systeminfo

import (
	"context"
	"log"
	"time"
)

// Service coordinates system information collection and snapshot persistence.
type Service struct{}

// GetSnapshot reads the last saved snapshot without collecting new information.
// It returns null when no compatible snapshot exists.
func (Service) GetSnapshot() (*SystemSnapshot, error) {
	return loadSnapshot()
}

// Refresh collects fresh information within five seconds and returns it.
// Snapshot persistence is best effort: a storage failure is logged and does not
// discard freshly collected information.
func (Service) Refresh(ctx context.Context) (SystemSnapshot, error) {
	ctx, cancel := context.WithTimeout(ctx, 5*time.Second)
	defer cancel()

	info, err := collect(ctx)
	if err != nil {
		return SystemSnapshot{}, err
	}
	snapshot := SystemSnapshot{Info: info, CollectedAt: time.Now().UTC()}
	if err := saveSnapshot(snapshot); err != nil {
		log.Printf("systeminfo: save snapshot: %v", err)
	}
	return snapshot, nil
}
