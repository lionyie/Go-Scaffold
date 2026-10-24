package auth

import (
	"errors"
	"time"

	"github.com/golang-jwt/jwt/v5"

	"github.com/example/scaffold/services/internal/conf"
)

var (
	// ErrTokenInvalid means the token is malformed or the signature is wrong.
	ErrTokenInvalid = errors.New("token invalid")
	// ErrTokenExpired means the token is no longer valid.
	ErrTokenExpired = errors.New("token expired")
	// ErrTokenType means an access token was used where a refresh token is required,
	// or vice versa.
	ErrTokenType = errors.New("token type mismatch")
	// ErrTokenIssuedAt means the token has no usable claims.
	ErrTokenIssuedAt = errors.New("token claims invalid")
)

// Claims is the payload carried by both access and refresh tokens.
type Claims struct {
	UserID             string `json:"uid"`
	Username           string `json:"usr"`
	SessionID          string `json:"sid"`
	TokenType          string `json:"typ"`
	Admin              bool   `json:"adm,omitempty"`
	MustChangePassword bool   `json:"mcp,omitempty"`
	jwt.RegisteredClaims
}

// TokenManager issues and verifies JWTs.
type TokenManager struct {
	secret     []byte
	issuer     string
	accessTTL  time.Duration
	refreshTTL time.Duration
}

// NewTokenManager builds a manager from the application config.
func NewTokenManager(bc *conf.Bootstrap) *TokenManager {
	cfg := bc.GetAuth().GetJwt()
	accessTTL := cfg.GetAccessTtl().AsDuration()
	if accessTTL <= 0 {
		accessTTL = 15 * time.Minute
	}
	refreshTTL := cfg.GetRefreshTtl().AsDuration()
	if refreshTTL <= 0 {
		refreshTTL = 7 * 24 * time.Hour
	}
	return &TokenManager{
		secret:     []byte(cfg.GetSecret()),
		issuer:     cfg.GetIssuer(),
		accessTTL:  accessTTL,
		refreshTTL: refreshTTL,
	}
}

// Issue creates an access/refresh token pair for the given identity.
func (m *TokenManager) Issue(userID, username, sessionID string, admin, mustChangePassword bool) (string, string, int32, error) {
	now := time.Now()
	access, err := m.sign(Claims{
		UserID:             userID,
		Username:           username,
		SessionID:          sessionID,
		TokenType:          TokenTypeAccess,
		Admin:              admin,
		MustChangePassword: mustChangePassword,
		RegisteredClaims: jwt.RegisteredClaims{
			Issuer:    m.issuer,
			Subject:   userID,
			IssuedAt:  jwt.NewNumericDate(now),
			NotBefore: jwt.NewNumericDate(now),
			ExpiresAt: jwt.NewNumericDate(now.Add(m.accessTTL)),
			ID:        sessionID,
		},
	})
	if err != nil {
		return "", "", 0, err
	}
	refresh, err := m.sign(Claims{
		UserID:             userID,
		Username:           username,
		SessionID:          sessionID,
		TokenType:          TokenTypeRefresh,
		Admin:              admin,
		MustChangePassword: mustChangePassword,
		RegisteredClaims: jwt.RegisteredClaims{
			Issuer:    m.issuer,
			Subject:   userID,
			IssuedAt:  jwt.NewNumericDate(now),
			NotBefore: jwt.NewNumericDate(now),
			ExpiresAt: jwt.NewNumericDate(now.Add(m.refreshTTL)),
			ID:        sessionID,
		},
	})
	if err != nil {
		return "", "", 0, err
	}
	return access, refresh, int32(m.accessTTL.Seconds()), nil
}

func (m *TokenManager) sign(claims Claims) (string, error) {
	return jwt.NewWithClaims(jwt.SigningMethodHS256, claims).SignedString(m.secret)
}

// ParseAccess verifies an access token.
func (m *TokenManager) ParseAccess(token string) (*Claims, error) {
	return m.parse(token, TokenTypeAccess)
}

// ParseRefresh verifies a refresh token.
func (m *TokenManager) ParseRefresh(token string) (*Claims, error) {
	return m.parse(token, TokenTypeRefresh)
}

func (m *TokenManager) parse(token, expectedType string) (*Claims, error) {
	claims := &Claims{}
	parsed, err := jwt.ParseWithClaims(token, claims, func(t *jwt.Token) (any, error) {
		if _, ok := t.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, ErrTokenInvalid
		}
		return m.secret, nil
	})
	if err != nil {
		if errors.Is(err, jwt.ErrTokenExpired) {
			return nil, ErrTokenExpired
		}
		return nil, ErrTokenInvalid
	}
	if !parsed.Valid || claims.TokenType != expectedType {
		if parsed != nil && parsed.Valid {
			return nil, ErrTokenType
		}
		return nil, ErrTokenInvalid
	}
	if claims.UserID == "" || claims.SessionID == "" {
		return nil, ErrTokenIssuedAt
	}
	return claims, nil
}

// RefreshTTL exposes the refresh token lifetime.
func (m *TokenManager) RefreshTTL() time.Duration { return m.refreshTTL }
