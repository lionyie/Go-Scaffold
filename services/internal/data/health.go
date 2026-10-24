package data

import (
	"context"
	"time"

	"github.com/example/scaffold/services/internal/biz"
	"github.com/example/scaffold/services/internal/data/database"
)

type healthChecker struct {
	db *database.Data
}

// NewHealthChecker reports database readiness.
func NewHealthChecker(db *database.Data) biz.HealthChecker {
	return &healthChecker{db: db}
}

func (h *healthChecker) Check(ctx context.Context) (map[string]string, error) {
	ctx, cancel := context.WithTimeout(ctx, 5*time.Second)
	defer cancel()

	checks := map[string]string{}
	var firstErr error

	sqlDB, err := h.db.DB.DB()
	if err != nil {
		checks["database"] = "error: " + err.Error()
		firstErr = err
	} else if err := sqlDB.PingContext(ctx); err != nil {
		checks["database"] = "error: " + err.Error()
		firstErr = err
	} else {
		checks["database"] = "ok"
	}

	return checks, firstErr
}
