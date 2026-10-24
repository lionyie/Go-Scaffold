package oauth

import (
	"context"
	"errors"
	"sync"
	"time"
)

// ErrStateNotFound is returned when a state is missing or expired.
var ErrStateNotFound = errors.New("oauth: state not found")

// ErrTicketNotFound is returned when a ticket is missing, expired, or reused.
var ErrTicketNotFound = errors.New("oauth: ticket not found")

// StateData is stored server-side between start and callback.
type StateData struct {
	Provider      string
	PKCEVerifier  string
	Nonce         string
	RedirectAfter string
}

// TicketData is stored server-side between callback and exchange.
type TicketData struct {
	AccessToken  string
	RefreshToken string
	ExpiresAt    time.Time
}

// Store persists short-lived OAuth state and one-time login tickets. The
// in-memory implementation is sufficient for a single replica; multi-replica
// deployments must provide a Redis or database implementation.
type Store interface {
	SaveState(ctx context.Context, state string, v StateData, ttl time.Duration) error
	ConsumeState(ctx context.Context, state string) (StateData, error)
	SaveTicket(ctx context.Context, ticket string, v TicketData, ttl time.Duration) error
	ConsumeTicket(ctx context.Context, ticket string) (TicketData, error)
}

type stateEntry struct {
	data      StateData
	expiresAt time.Time
}

type MemoryStore struct {
	mu      sync.Mutex
	states  map[string]stateEntry
	tickets map[string]TicketData
}

// NewMemoryStore returns a process-local OAuth store.
func NewMemoryStore() *MemoryStore {
	return &MemoryStore{
		states:  make(map[string]stateEntry),
		tickets: make(map[string]TicketData),
	}
}

func (m *MemoryStore) SaveState(_ context.Context, state string, v StateData, ttl time.Duration) error {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.states[state] = stateEntry{data: v, expiresAt: time.Now().Add(ttl)}
	return nil
}

func (m *MemoryStore) ConsumeState(_ context.Context, state string) (StateData, error) {
	m.mu.Lock()
	defer m.mu.Unlock()
	entry, ok := m.states[state]
	if !ok {
		return StateData{}, ErrStateNotFound
	}
	delete(m.states, state)
	if time.Now().After(entry.expiresAt) {
		return StateData{}, ErrStateNotFound
	}
	return entry.data, nil
}

func (m *MemoryStore) SaveTicket(_ context.Context, ticket string, v TicketData, ttl time.Duration) error {
	m.mu.Lock()
	defer m.mu.Unlock()
	v.ExpiresAt = time.Now().Add(ttl)
	m.tickets[ticket] = v
	return nil
}

func (m *MemoryStore) ConsumeTicket(_ context.Context, ticket string) (TicketData, error) {
	m.mu.Lock()
	defer m.mu.Unlock()
	v, ok := m.tickets[ticket]
	if !ok {
		return TicketData{}, ErrTicketNotFound
	}
	delete(m.tickets, ticket)
	if time.Now().After(v.ExpiresAt) {
		return TicketData{}, ErrTicketNotFound
	}
	return v, nil
}
