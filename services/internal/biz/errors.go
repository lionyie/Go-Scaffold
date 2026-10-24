package biz

import (
	kerrors "github.com/go-kratos/kratos/v3/errors"

	commonv1 "github.com/example/scaffold/services/api/common/v1"
)

// Error constructors keep reason codes stable across transports and make the
// data layer independent of proto details.

func ErrUnauthorized() error {
	return kerrors.Unauthorized(commonv1.ErrorReason_UNAUTHORIZED.String(), "authentication required")
}

func ErrForbidden() error {
	return kerrors.Forbidden(commonv1.ErrorReason_FORBIDDEN.String(), "permission denied")
}

func ErrInvalidCredentials() error {
	return kerrors.Unauthorized(commonv1.ErrorReason_INVALID_CREDENTIALS.String(), "invalid username or password")
}

func ErrTokenInvalid() error {
	return kerrors.Unauthorized(commonv1.ErrorReason_TOKEN_INVALID.String(), "token is invalid")
}

func ErrTokenExpired() error {
	return kerrors.Unauthorized(commonv1.ErrorReason_TOKEN_EXPIRED.String(), "token has expired")
}

func ErrUserNotFound() error {
	return kerrors.NotFound(commonv1.ErrorReason_USER_NOT_FOUND.String(), "user not found")
}

func ErrUserAlreadyExists() error {
	return kerrors.Conflict(commonv1.ErrorReason_USER_ALREADY_EXISTS.String(), "user already exists")
}

func ErrRoleNotFound() error {
	return kerrors.NotFound(commonv1.ErrorReason_ROLE_NOT_FOUND.String(), "role not found")
}

func ErrRoleAlreadyExists() error {
	return kerrors.Conflict(commonv1.ErrorReason_ROLE_ALREADY_EXISTS.String(), "role already exists")
}

func ErrInvalidArgument(message string) error {
	return kerrors.BadRequest(commonv1.ErrorReason_INVALID_ARGUMENT.String(), message)
}

func ErrMustChangePassword() error {
	return kerrors.Forbidden(commonv1.ErrorReason_MUST_CHANGE_PASSWORD.String(), "password change required")
}

func ErrPluginNotFound() error {
	return kerrors.NotFound(commonv1.ErrorReason_PLUGIN_NOT_FOUND.String(), "plugin not found")
}

func ErrStorageUnavailable() error {
	return kerrors.ServiceUnavailable(commonv1.ErrorReason_STORAGE_UNAVAILABLE.String(), "storage backend is unavailable")
}

func ErrStorageForbidden() error {
	return kerrors.Forbidden(commonv1.ErrorReason_STORAGE_FORBIDDEN.String(), "storage backend denied the request")
}

func ErrInternal(message string) error {
	return kerrors.InternalServer(commonv1.ErrorReason_INTERNAL_ERROR.String(), message)
}
