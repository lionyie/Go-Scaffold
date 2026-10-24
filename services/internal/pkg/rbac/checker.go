package rbac

import (
	"context"
	"fmt"
	"sync"
	"time"
)

// RoleLookup is implemented by the data layer. It returns the union of
// permissions granted by the user's roles.
type RoleLookup interface {
	PermissionsForUser(ctx context.Context, userID string) ([]string, error)
}

// Checker evaluates authorization decisions.
type Checker interface {
	HasGlobal(ctx context.Context, userID, code string) (bool, error)
	Permissions(ctx context.Context, userID string) ([]string, error)
	Invalidate(userID string)
}

type cacheEntry struct {
	permissions []string
	expiresAt   time.Time
}

type checker struct {
	lookup RoleLookup
	ttl    time.Duration
	mu     sync.RWMutex
	cache  map[string]cacheEntry
}

// NewChecker returns a Checker backed by a short-lived in-memory cache.
func NewChecker(lookup RoleLookup) Checker {
	return &checker{
		lookup: lookup,
		ttl:    30 * time.Second,
		cache:  make(map[string]cacheEntry),
	}
}

func (c *checker) Permissions(ctx context.Context, userID string) ([]string, error) {
	c.mu.RLock()
	entry, ok := c.cache[userID]
	c.mu.RUnlock()
	if ok && time.Now().Before(entry.expiresAt) {
		return entry.permissions, nil
	}

	permissions, err := c.lookup.PermissionsForUser(ctx, userID)
	if err != nil {
		return nil, fmt.Errorf("lookup permissions: %w", err)
	}
	c.mu.Lock()
	c.cache[userID] = cacheEntry{permissions: permissions, expiresAt: time.Now().Add(c.ttl)}
	c.mu.Unlock()
	return permissions, nil
}

func (c *checker) HasGlobal(ctx context.Context, userID, code string) (bool, error) {
	permissions, err := c.Permissions(ctx, userID)
	if err != nil {
		return false, err
	}
	for _, p := range permissions {
		if p == code {
			return true, nil
		}
	}
	return false, nil
}

func (c *checker) Invalidate(userID string) {
	c.mu.Lock()
	defer c.mu.Unlock()
	delete(c.cache, userID)
}
