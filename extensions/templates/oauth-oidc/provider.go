package providers

import (
	"context"
	"crypto/rsa"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"io"
	"math/big"
	"net/http"
	"net/url"
	"os"
	"strings"
	"sync"
	"time"

	"github.com/golang-jwt/jwt/v5"

	"github.com/example/scaffold/services/internal/auth/oauth"
)

type oidcDiscovery struct {
	AuthorizationEndpoint string `json:"authorization_endpoint"`
	TokenEndpoint         string `json:"token_endpoint"`
	UserinfoEndpoint      string `json:"userinfo_endpoint"`
	JwksURI               string `json:"jwks_uri"`
	Issuer                string `json:"issuer"`
}

type oidcProvider struct {
	client *http.Client

	mu        sync.Mutex
	discovery *oidcDiscovery
	fetchedAt time.Time
}

func init() { oauth.Register(&oidcProvider{client: &http.Client{Timeout: 15 * time.Second}}) }

func (p *oidcProvider) Name() string        { return "oidc" }
func (p *oidcProvider) DisplayName() string { return "SSO" }

func (p *oidcProvider) Enabled() bool {
	return os.Getenv("SCAFFOLD_OAUTH_OIDC_ISSUER") != "" &&
		os.Getenv("SCAFFOLD_OAUTH_OIDC_CLIENT_ID") != "" &&
		os.Getenv("SCAFFOLD_OAUTH_OIDC_CLIENT_SECRET") != ""
}

func (p *oidcProvider) scopes() string {
	if scopes := strings.TrimSpace(os.Getenv("SCAFFOLD_OAUTH_OIDC_SCOPES")); scopes != "" {
		return scopes
	}
	return "openid email profile"
}

func (p *oidcProvider) AuthCodeURL(ctx context.Context, state, redirectURL, nonce string, pkce *oauth.PKCE) string {
	meta, err := p.metadata(ctx)
	if err != nil {
		return ""
	}
	query := url.Values{}
	query.Set("client_id", os.Getenv("SCAFFOLD_OAUTH_OIDC_CLIENT_ID"))
	query.Set("redirect_uri", redirectURL)
	query.Set("response_type", "code")
	query.Set("scope", p.scopes())
	query.Set("state", state)
	query.Set("nonce", nonce)
	if pkce != nil {
		query.Set("code_challenge", pkce.Challenge)
		query.Set("code_challenge_method", pkce.Method)
	}
	return meta.AuthorizationEndpoint + "?" + query.Encode()
}

func (p *oidcProvider) Exchange(ctx context.Context, code, redirectURL string, pkce *oauth.PKCE) (*oauth.Token, error) {
	meta, err := p.metadata(ctx)
	if err != nil {
		return nil, err
	}
	form := url.Values{}
	form.Set("grant_type", "authorization_code")
	form.Set("client_id", os.Getenv("SCAFFOLD_OAUTH_OIDC_CLIENT_ID"))
	form.Set("client_secret", os.Getenv("SCAFFOLD_OAUTH_OIDC_CLIENT_SECRET"))
	form.Set("code", code)
	form.Set("redirect_uri", redirectURL)
	if pkce != nil {
		form.Set("code_verifier", pkce.Verifier)
	}
	req, err := http.NewRequestWithContext(ctx, http.MethodPost, meta.TokenEndpoint, strings.NewReader(form.Encode()))
	if err != nil {
		return nil, err
	}
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	resp, err := p.client.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()
	body, err := io.ReadAll(io.LimitReader(resp.Body, 1<<20))
	if err != nil {
		return nil, err
	}
	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("oidc: token endpoint returned %d: %s", resp.StatusCode, string(body))
	}
	var payload struct {
		AccessToken string `json:"access_token"`
		TokenType   string `json:"token_type"`
		IDToken     string `json:"id_token"`
	}
	if err := json.Unmarshal(body, &payload); err != nil {
		return nil, err
	}
	if payload.IDToken == "" {
		return nil, fmt.Errorf("oidc: token response has no id_token")
	}
	return &oauth.Token{
		AccessToken: payload.AccessToken,
		TokenType:   payload.TokenType,
		Raw:         map[string]string{"id_token": payload.IDToken},
	}, nil
}

func (p *oidcProvider) UserInfo(ctx context.Context, token *oauth.Token) (*oauth.OAuthUser, error) {
	meta, err := p.metadata(ctx)
	if err != nil {
		return nil, err
	}
	claims, err := p.verifyIDToken(ctx, token.Raw["id_token"], meta)
	if err != nil {
		return nil, err
	}
	if claims.Subject == "" {
		return nil, fmt.Errorf("oidc: id_token has no subject")
	}
	return &oauth.OAuthUser{
		ProviderUserID: claims.Subject,
		Email:          claims.Email,
		EmailVerified:  claims.EmailVerified,
		Name:           firstNonEmpty(claims.Name, claims.PreferredUsername),
		Avatar:         claims.Picture,
	}, nil
}

type oidcClaims struct {
	Subject           string `json:"sub"`
	Email             string `json:"email"`
	EmailVerified     bool   `json:"email_verified"`
	Name              string `json:"name"`
	PreferredUsername string `json:"preferred_username"`
	Picture           string `json:"picture"`
	Nonce             string `json:"nonce"`
	jwt.RegisteredClaims
}

func (p *oidcProvider) verifyIDToken(ctx context.Context, raw string, meta *oidcDiscovery) (*oidcClaims, error) {
	if raw == "" {
		return nil, fmt.Errorf("oidc: missing id_token")
	}
	jwks, err := p.jwks(ctx, meta.JwksURI)
	if err != nil {
		return nil, err
	}
	claims := &oidcClaims{}
	parsed, err := jwt.ParseWithClaims(raw, claims, func(token *jwt.Token) (any, error) {
		kid, _ := token.Header["kid"].(string)
		key, ok := jwks[kid]
		if !ok {
			return nil, fmt.Errorf("oidc: unknown signing key %q", kid)
		}
		return key, nil
	},
		jwt.WithValidMethods([]string{"RS256", "RS384", "RS512"}),
		jwt.WithIssuer(meta.Issuer),
		jwt.WithAudience(os.Getenv("SCAFFOLD_OAUTH_OIDC_CLIENT_ID")),
		jwt.WithExpirationRequired(),
	)
	if err != nil {
		return nil, fmt.Errorf("oidc: id_token verification failed: %w", err)
	}
	if !parsed.Valid {
		return nil, fmt.Errorf("oidc: id_token is invalid")
	}
	return claims, nil
}

func (p *oidcProvider) metadata(ctx context.Context) (*oidcDiscovery, error) {
	p.mu.Lock()
	defer p.mu.Unlock()
	if p.discovery != nil && time.Since(p.fetchedAt) < 10*time.Minute {
		return p.discovery, nil
	}
	issuer := strings.TrimRight(os.Getenv("SCAFFOLD_OAUTH_OIDC_ISSUER"), "/")
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, issuer+"/.well-known/openid-configuration", nil)
	if err != nil {
		return nil, err
	}
	resp, err := p.client.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("oidc: discovery returned %d", resp.StatusCode)
	}
	var meta oidcDiscovery
	if err := json.NewDecoder(io.LimitReader(resp.Body, 1<<20)).Decode(&meta); err != nil {
		return nil, err
	}
	if meta.Issuer == "" {
		meta.Issuer = issuer
	}
	p.discovery = &meta
	p.fetchedAt = time.Now()
	return p.discovery, nil
}

func (p *oidcProvider) jwks(ctx context.Context, uri string) (map[string]*rsa.PublicKey, error) {
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, uri, nil)
	if err != nil {
		return nil, err
	}
	resp, err := p.client.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("oidc: jwks returned %d", resp.StatusCode)
	}
	var set struct {
		Keys []struct {
			Kid string `json:"kid"`
			Kty string `json:"kty"`
			N   string `json:"n"`
			E   string `json:"e"`
		} `json:"keys"`
	}
	if err := json.NewDecoder(io.LimitReader(resp.Body, 1<<20)).Decode(&set); err != nil {
		return nil, err
	}
	out := make(map[string]*rsa.PublicKey, len(set.Keys))
	for _, key := range set.Keys {
		if key.Kty != "RSA" || key.N == "" || key.E == "" {
			continue
		}
		nBytes, err := base64.RawURLEncoding.DecodeString(key.N)
		if err != nil {
			continue
		}
		eBytes, err := base64.RawURLEncoding.DecodeString(key.E)
		if err != nil {
			continue
		}
		e := int(new(big.Int).SetBytes(eBytes).Int64())
		out[key.Kid] = &rsa.PublicKey{N: new(big.Int).SetBytes(nBytes), E: e}
	}
	if len(out) == 0 {
		return nil, fmt.Errorf("oidc: jwks has no usable RSA keys")
	}
	return out, nil
}
