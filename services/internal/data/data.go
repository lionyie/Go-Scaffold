package data

import (
	"github.com/google/wire"

	"github.com/example/scaffold/services/internal/data/database"
)

// ProviderSet is the infrastructure layer provider set.
var ProviderSet = wire.NewSet(
	database.NewData,
	database.NewUserRepo,
	database.NewRoleRepo,
	database.NewRefreshTokenRepo,
	database.NewRBACRepo,
	database.NewOAuthAccountRepo,
	NewHealthChecker,
)
