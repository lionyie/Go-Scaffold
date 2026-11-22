package providers

import (
	"context"
	"crypto/rand"
	"crypto/rsa"
	"encoding/base64"
	"encoding/json"
	"math/big"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/golang-jwt/jwt/v5"

	"github.com/example/scaffold/services/internal/auth/oauth"
)

func TestOIDCNonceVerification(t *testing.T) {
	key, err := rsa.GenerateKey(rand.Reader, 2048)
	if err != nil {
		t.Fatalf("generate key: %v", err)
	}

	var server *httptest.Server
	mux := http.NewServeMux()
	mux.HandleFunc("/.well-known/openid-configuration", func(w http.ResponseWriter, _ *http.Request) {
		_ = json.NewEncoder(w).Encode(map[string]string{
			"issuer":                 server.URL,
			"authorization_endpoint": server.URL + "/auth",
			"token_endpoint":         server.URL + "/token",
			"userinfo_endpoint":      server.URL + "/userinfo",
			"jwks_uri":               server.URL + "/jwks",
		})
	})
	mux.HandleFunc("/jwks", func(w http.ResponseWriter, _ *http.Request) {
		_ = json.NewEncoder(w).Encode(map[string]any{
			"keys": []map[string]string{{
				"kid": "test",
				"kty": "RSA",
				"n":   base64.RawURLEncoding.EncodeToString(key.N.Bytes()),
				"e":   base64.RawURLEncoding.EncodeToString(big.NewInt(int64(key.E)).Bytes()),
			}},
		})
	})
	server = httptest.NewServer(mux)
	defer server.Close()

	t.Setenv("SCAFFOLD_OAUTH_OIDC_ISSUER", server.URL)
	t.Setenv("SCAFFOLD_OAUTH_OIDC_CLIENT_ID", "client")
	t.Setenv("SCAFFOLD_OAUTH_OIDC_CLIENT_SECRET", "secret")

	provider := &oidcProvider{client: server.Client()}
	token := signIDToken(t, key, server.URL, "client", "nonce-123")

	user, err := provider.UserInfoWithNonce(context.Background(), token, "nonce-123")
	if err != nil {
		t.Fatalf("matching nonce rejected: %v", err)
	}
	if user.ProviderUserID != "user-1" || user.Email != "admin@example.com" {
		t.Fatalf("unexpected user: %+v", user)
	}
	if len(user.Roles) != 2 || user.Roles[0] != "user:list" || user.Roles[1] != "viewer" {
		t.Fatalf("realm roles = %v", user.Roles)
	}

	if _, err := provider.UserInfoWithNonce(context.Background(), token, "wrong-nonce"); err == nil {
		t.Fatal("mismatched nonce was accepted")
	}

	if _, err := provider.UserInfo(context.Background(), token); err != nil {
		t.Fatalf("plain UserInfo should not require a nonce: %v", err)
	}
}

func signIDToken(t *testing.T, key *rsa.PrivateKey, issuer, audience, nonce string) *oauth.Token {
	t.Helper()
	claims := jwt.MapClaims{
		"iss":            issuer,
		"sub":            "user-1",
		"aud":            audience,
		"exp":            time.Now().Add(time.Hour).Unix(),
		"email":          "admin@example.com",
		"email_verified": true,
		"nonce":          nonce,
		"realm_access":   map[string]any{"roles": []string{"user:list", "viewer"}},
	}
	token := jwt.NewWithClaims(jwt.SigningMethodRS256, claims)
	token.Header["kid"] = "test"
	signed, err := token.SignedString(key)
	if err != nil {
		t.Fatalf("sign token: %v", err)
	}
	return &oauth.Token{Raw: map[string]string{"id_token": signed}}
}
