package authz

import (
	"context"
	"strings"
	"sync"
	"time"

	"github.com/example/scaffold/services/internal/biz"
	"github.com/example/scaffold/services/internal/pkg/operations"
)

const policyCacheTTL = 30 * time.Second

// OperationView is a read model for the operation catalog page.
type OperationView struct {
	Operation        string
	Service          string
	Method           string
	Path             string
	Permission       string
	StaticPermission string
	Source           string
	Locked           bool
}

type policyCacheEntry struct {
	code      string
	found     bool
	expiresAt time.Time
}

// Manager resolves the effective requirement for an operation: the compiled
// mapping first, then a runtime binding stored in the database. Public and
// authenticated-only operations are locked and cannot be rebound.
type Manager struct {
	repo    biz.PolicyRepo
	catalog map[string]operations.Operation

	mu    sync.RWMutex
	cache map[string]policyCacheEntry
}

// NewManager builds the policy manager over the generated operation catalog.
func NewManager(repo biz.PolicyRepo) *Manager {
	catalog := make(map[string]operations.Operation, len(operations.All))
	for _, op := range operations.All {
		catalog[op.Name] = op
	}
	return &Manager{
		repo:    repo,
		catalog: catalog,
		cache:   make(map[string]policyCacheEntry),
	}
}

// Resolve returns the requirement for an operation. The boolean is false when
// the operation is unknown, which the middleware treats as fail closed.
func (m *Manager) Resolve(ctx context.Context, operation string) (Requirement, bool, error) {
	static, known := Requirements[operation]
	if !known {
		return Requirement{}, false, nil
	}
	if static.Public || static.Code == "" {
		return static, true, nil
	}
	code, found, err := m.operationPolicy(ctx, operation)
	if err != nil {
		return Requirement{}, false, err
	}
	if found {
		return Requirement{Code: code}, true, nil
	}
	return static, true, nil
}

// List returns every generated operation with its effective binding.
func (m *Manager) List(ctx context.Context) ([]OperationView, error) {
	policies, err := m.repo.ListOperationPolicies(ctx)
	if err != nil {
		return nil, err
	}
	views := make([]OperationView, 0, len(operations.All))
	for _, op := range operations.All {
		static, known := Requirements[op.Name]
		code, found := policies[op.Name]
		views = append(views, buildView(op, static, known, code, found))
	}
	return views, nil
}

// Set binds an operation to a permission code.
func (m *Manager) Set(ctx context.Context, operation, permission string) (OperationView, error) {
	op, static, err := m.rebindable(operation)
	if err != nil {
		return OperationView{}, err
	}
	permission = strings.TrimSpace(permission)
	if permission == "" {
		return OperationView{}, biz.ErrInvalidArgument("permission is required")
	}
	exists, err := m.repo.PermissionExists(ctx, permission)
	if err != nil {
		return OperationView{}, err
	}
	if !exists {
		return OperationView{}, biz.ErrPermissionNotFound()
	}
	if err := m.repo.SetOperationPolicy(ctx, operation, permission); err != nil {
		return OperationView{}, err
	}
	m.Invalidate(operation)
	return buildView(op, static, true, permission, true), nil
}

// Reset removes a runtime binding and restores the compiled mapping.
func (m *Manager) Reset(ctx context.Context, operation string) (OperationView, error) {
	op, static, err := m.rebindable(operation)
	if err != nil {
		return OperationView{}, err
	}
	if err := m.repo.DeleteOperationPolicy(ctx, operation); err != nil {
		return OperationView{}, err
	}
	m.Invalidate(operation)
	return buildView(op, static, true, "", false), nil
}

// Invalidate drops the cached binding for one operation.
func (m *Manager) Invalidate(operation string) {
	m.mu.Lock()
	delete(m.cache, operation)
	m.mu.Unlock()
}

// InvalidateAll drops every cached binding.
func (m *Manager) InvalidateAll() {
	m.mu.Lock()
	m.cache = make(map[string]policyCacheEntry)
	m.mu.Unlock()
}

func (m *Manager) rebindable(operation string) (operations.Operation, Requirement, error) {
	op, ok := m.catalog[operation]
	if !ok {
		return operations.Operation{}, Requirement{}, biz.ErrInvalidArgument("unknown operation")
	}
	static, known := Requirements[operation]
	if !known {
		return operations.Operation{}, Requirement{}, biz.ErrInvalidArgument("operation has no compiled requirement")
	}
	if static.Public || static.Code == "" {
		return operations.Operation{}, Requirement{}, biz.ErrForbidden()
	}
	return op, static, nil
}

func (m *Manager) operationPolicy(ctx context.Context, operation string) (string, bool, error) {
	m.mu.RLock()
	entry, ok := m.cache[operation]
	m.mu.RUnlock()
	if ok && time.Now().Before(entry.expiresAt) {
		return entry.code, entry.found, nil
	}
	code, found, err := m.repo.GetOperationPolicy(ctx, operation)
	if err != nil {
		return "", false, err
	}
	m.mu.Lock()
	m.cache[operation] = policyCacheEntry{code: code, found: found, expiresAt: time.Now().Add(policyCacheTTL)}
	m.mu.Unlock()
	return code, found, nil
}

func buildView(op operations.Operation, static Requirement, known bool, dbCode string, dbFound bool) OperationView {
	view := OperationView{
		Operation: op.Name,
		Service:   op.Service,
		Method:    op.Method,
		Path:      op.Path,
	}
	switch {
	case !known:
		view.Source = "none"
		view.Locked = true
	case static.Public:
		view.Source = "public"
		view.Locked = true
	case static.Code == "":
		view.Source = "authenticated"
		view.Locked = true
	default:
		view.StaticPermission = static.Code
		view.Permission = static.Code
		view.Source = "static"
	}
	if !view.Locked && dbFound {
		view.Permission = dbCode
		view.Source = "database"
	}
	return view
}
