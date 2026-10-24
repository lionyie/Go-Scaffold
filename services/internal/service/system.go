package service

import (
	"context"

	v1 "github.com/example/scaffold/services/api/system/v1"
	"github.com/example/scaffold/services/internal/biz"
	kerrors "github.com/go-kratos/kratos/v3/errors"
)

// SystemService exposes liveness, readiness, and build information.
type SystemService struct {
	v1.UnimplementedSystemServiceServer
	health biz.HealthChecker
	build  *BuildInfo
}

// NewSystemService creates the system service.
func NewSystemService(health biz.HealthChecker, build *BuildInfo) *SystemService {
	return &SystemService{health: health, build: build}
}

func (s *SystemService) GetHealth(_ context.Context, _ *v1.GetHealthRequest) (*v1.GetHealthResponse, error) {
	return &v1.GetHealthResponse{Status: "ok"}, nil
}

func (s *SystemService) GetReady(ctx context.Context, _ *v1.GetReadyRequest) (*v1.GetReadyResponse, error) {
	checks, err := s.health.Check(ctx)
	if err != nil {
		return nil, kerrors.ServiceUnavailable("STORAGE_UNAVAILABLE", "storage is not ready")
	}
	return &v1.GetReadyResponse{Status: "ok", Checks: checks}, nil
}

func (s *SystemService) GetVersion(_ context.Context, _ *v1.GetVersionRequest) (*v1.GetVersionResponse, error) {
	version := "dev"
	goVersion := ""
	if s.build != nil {
		version = s.build.Version
		goVersion = s.build.GoVersion
	}
	return &v1.GetVersionResponse{Version: version, GoVersion: goVersion}, nil
}
