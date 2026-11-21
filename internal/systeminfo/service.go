package systeminfo

import (
	"context"
	"log"
	"time"

	"devbox/internal/readcache"
)

// Service coordinates system information collection and snapshot persistence.
type Service struct {
	snapshots readcache.Cache[SystemSnapshot]
}

// Get returns a remembered snapshot or collects fresh system information.
func (s *Service) Get(ctx context.Context, refresh bool) (SystemSnapshot, error) {
	return s.snapshots.Read(ctx, refresh, 0, func(ctx context.Context) (SystemSnapshot, error) {
		if !refresh {
			snapshot, err := loadSnapshot()
			if canceled := ctx.Err(); canceled != nil {
				return SystemSnapshot{}, canceled
			}
			if err != nil {
				log.Printf("systeminfo: load snapshot: %v", err)
			}
			if snapshot != nil {
				return *snapshot, nil
			}
		}
		return collectSnapshot(ctx)
	})
}

func collectSnapshot(ctx context.Context) (SystemSnapshot, error) {
	collectCtx, cancel := context.WithTimeout(ctx, 5*time.Second)
	defer cancel()

	info, err := collect(collectCtx)
	if err != nil {
		return SystemSnapshot{}, err
	}
	snapshot := SystemSnapshot{Info: info, CollectedAt: time.Now().UTC()}

	if err := ctx.Err(); err != nil {
		return SystemSnapshot{}, err
	}
	if err := saveSnapshot(snapshot); err != nil {
		log.Printf("systeminfo: save snapshot: %v", err)
	}
	return snapshot, nil
}
