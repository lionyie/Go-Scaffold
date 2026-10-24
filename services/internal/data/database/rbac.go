package database

import (
	"context"

	"github.com/example/scaffold/services/internal/pkg/rbac"
)

type rbacRepo struct {
	data *Data
}

// NewRBACRepo returns the RoleLookup implementation backed by the database.
func NewRBACRepo(d *Data) rbac.RoleLookup {
	return &rbacRepo{data: d}
}

func (r *rbacRepo) PermissionsForUser(ctx context.Context, userID string) ([]string, error) {
	const query = `
SELECT DISTINCT p.code
FROM permissions AS p
JOIN role_permissions AS rp ON rp.permission_code = p.code
JOIN user_roles AS ur ON ur.role_id = rp.role_id
WHERE ur.user_id = ?
ORDER BY p.code`

	var codes []string
	if err := r.data.DB.WithContext(ctx).Raw(query, userID).Scan(&codes).Error; err != nil {
		return nil, err
	}
	return codes, nil
}
