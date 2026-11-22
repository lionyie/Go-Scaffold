package auth

import (
	"testing"
	"time"
)

func newTestTokenManager() *TokenManager {
	return &TokenManager{
		secret:     []byte("test-secret"),
		issuer:     "test",
		accessTTL:  time.Minute,
		refreshTTL: time.Hour,
	}
}

func TestIssueCarriesRolesAndIdentityProvider(t *testing.T) {
	tm := newTestTokenManager()
	roles := []string{"user:list", "report:export"}

	access, refresh, _, err := tm.Issue("u1", "bob", "s1", false, false, roles, IdentityProviderOIDC)
	if err != nil {
		t.Fatalf("issue: %v", err)
	}

	claims, err := tm.ParseAccess(access)
	if err != nil {
		t.Fatalf("parse access: %v", err)
	}
	if claims.IdentityProvider != IdentityProviderOIDC {
		t.Fatalf("identity provider = %q", claims.IdentityProvider)
	}
	if len(claims.Roles) != 2 || claims.Roles[0] != "user:list" || claims.Roles[1] != "report:export" {
		t.Fatalf("roles = %v", claims.Roles)
	}

	refreshClaims, err := tm.ParseRefresh(refresh)
	if err != nil {
		t.Fatalf("parse refresh: %v", err)
	}
	if len(refreshClaims.Roles) != 2 {
		t.Fatalf("refresh roles = %v", refreshClaims.Roles)
	}
}

func TestPrincipalHasRole(t *testing.T) {
	p := &Principal{Roles: []string{"user:list"}}
	if !p.HasRole("user:list") {
		t.Fatal("expected user:list")
	}
	if p.HasRole("user:delete") {
		t.Fatal("did not expect user:delete")
	}
	var nilPrincipal *Principal
	if nilPrincipal.HasRole("user:list") {
		t.Fatal("nil principal should not have roles")
	}
}
