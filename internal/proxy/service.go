package proxy

import (
	"context"
	"errors"
	"sync"
	"time"

	"devbox/internal/readcache"
)

// Service owns the proxy configuration and exposes the settings operations.
type Service struct {
	mu       sync.Mutex
	settings *Settings
	system   readcache.Cache[SystemProxyInfo]
}

// Get returns the current settings, or null if the file needs to be reset.
func (s *Service) Get(ctx context.Context) (*Settings, error) {
	s.mu.Lock()
	defer s.mu.Unlock()
	if err := ctx.Err(); err != nil {
		return nil, err
	}
	settings, err := s.load()
	if errors.Is(err, errInvalidSettings) {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}
	return &settings, nil
}

// GetSystemInfo returns the system proxy settings and current routes.
func (s *Service) GetSystemInfo(ctx context.Context) (SystemProxyInfo, error) {
	return s.system.Read(ctx, false, 30*time.Second, systemInfo)
}

// Save validates and persists settings before making them effective.
func (s *Service) Save(input Settings) (Settings, error) {
	s.mu.Lock()
	defer s.mu.Unlock()
	next, err := s.load()
	if err != nil {
		return Settings{}, err
	}
	if !validMode(input.Mode) {
		return Settings{}, errInvalidMode
	}
	next.Mode = input.Mode
	if input.Mode == "manual" {
		next.Manual, err = normalizeManual(input.Manual)
		if err != nil {
			return Settings{}, err
		}
	}
	return s.persist(next)
}

// Reset replaces the saved settings with the defaults.
func (s *Service) Reset() (Settings, error) {
	s.mu.Lock()
	defer s.mu.Unlock()
	return s.persist(defaultSettings())
}

// Check checks a draft without saving it.
func (*Service) Check(ctx context.Context, input Settings) error {
	return checkConnection(ctx, input)
}

func (s *Service) load() (Settings, error) {
	if s.settings != nil {
		return *s.settings, nil
	}
	settings, err := loadSettings()
	if err != nil {
		return Settings{}, err
	}
	s.settings = &settings
	return settings, nil
}

func (s *Service) persist(settings Settings) (Settings, error) {
	if err := saveSettings(settings); err != nil {
		return Settings{}, err
	}
	s.settings = &settings
	return settings, nil
}
