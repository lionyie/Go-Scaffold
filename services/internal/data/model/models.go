package model

import (
	"time"

	"gorm.io/gorm"
)

// UserPO is the persisted account row.
type UserPO struct {
	ID                 string `gorm:"primaryKey;size:36"`
	Username           string `gorm:"uniqueIndex;size:64;not null"`
	DisplayName        string `gorm:"size:128"`
	Email              string `gorm:"size:255"`
	PasswordHash       string `gorm:"size:255;not null"`
	IsAdmin            bool   `gorm:"not null;default:false"`
	MustChangePassword bool   `gorm:"not null;default:false"`
	Status             string `gorm:"size:32;not null;default:active"`
	CreatedAt          time.Time
	UpdatedAt          time.Time
	DeletedAt          gorm.DeletedAt `gorm:"index"`
}

func (UserPO) TableName() string { return "users" }

// RolePO is a named permission bundle.
type RolePO struct {
	ID          string `gorm:"primaryKey;size:36"`
	Name        string `gorm:"uniqueIndex;size:64;not null"`
	DisplayName string `gorm:"size:128"`
	Description string `gorm:"size:512"`
	IsSystem    bool   `gorm:"not null;default:false"`
	CreatedAt   time.Time
	UpdatedAt   time.Time
}

func (RolePO) TableName() string { return "roles" }

// PermissionPO is a single capability in the catalog.
type PermissionPO struct {
	Code     string `gorm:"primaryKey;size:128"`
	Name     string `gorm:"size:128"`
	Resource string `gorm:"size:64"`
	Action   string `gorm:"size:64"`
}

func (PermissionPO) TableName() string { return "permissions" }

// RolePermissionPO joins roles to permissions.
type RolePermissionPO struct {
	RoleID         string `gorm:"primaryKey;size:36"`
	PermissionCode string `gorm:"primaryKey;size:128"`
}

func (RolePermissionPO) TableName() string { return "role_permissions" }

// UserRolePO assigns a role to a user.
type UserRolePO struct {
	UserID string `gorm:"primaryKey;size:36"`
	RoleID string `gorm:"primaryKey;size:36"`
}

func (UserRolePO) TableName() string { return "user_roles" }

// RefreshTokenPO stores only the hash of a refresh token.
type RefreshTokenPO struct {
	ID        string `gorm:"primaryKey;size:36"`
	UserID    string `gorm:"index;size:36;not null"`
	TokenHash string `gorm:"uniqueIndex;size:64;not null"`
	SessionID string `gorm:"index;size:36;not null"`
	ExpiresAt time.Time
	RevokedAt *time.Time
	CreatedAt time.Time
}

func (RefreshTokenPO) TableName() string { return "refresh_tokens" }

// OAuthAccountPO links a third-party identity to a local user.
type OAuthAccountPO struct {
	ID             string `gorm:"primaryKey;size:36"`
	UserID         string `gorm:"index;size:36;not null"`
	Provider       string `gorm:"uniqueIndex:idx_oauth_provider_subject;size:32;not null"`
	ProviderUserID string `gorm:"uniqueIndex:idx_oauth_provider_subject;size:191;not null"`
	Email          string `gorm:"size:255"`
	EmailVerified  bool   `gorm:"not null;default:false"`
	Name           string `gorm:"size:128"`
	Avatar         string `gorm:"size:512"`
	CreatedAt      time.Time
	UpdatedAt      time.Time
}

func (OAuthAccountPO) TableName() string { return "oauth_accounts" }
