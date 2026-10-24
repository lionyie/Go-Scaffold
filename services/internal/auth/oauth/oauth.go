// Package oauth defines the provider abstraction and registry used by
// first-party sign-in extensions. The concrete providers are generated into
// separate packages and register themselves from init().
package oauth

import (
	"context"
	"crypto/hmac"
	"crypto/rand"
	"crypto/sha256"
	"encoding/base64"
	"errors"
	"fmt"
	"net/url"
	"sort"
	"strings"
	"sync"
)

// Token is the normalized token response from a provider.
type Token struct {
	AccessToken  string
	RefreshToken string
	TokenType    string
	Raw          map[string]string
}

// OAuthUser is the normalized profile returned by a provider.
type OAuthUser struct {
	ProviderUserID string
	Email          string
	EmailVerified  bool
	Name           string
	Avatar         string
}

// PKCE carries the proof key for providers that support it.
type PKCE struct {
	Verifier  string
	Challenge string
	Method    string
}

// Provider is implemented by each sign-in extension.
type Provider interface {
	Name() string
	DisplayName() string
	Enabled() bool
	AuthCodeURL(ctx context.Context, state, redirectURL, nonce string, pkce *PKCE) string
	Exchange(ctx context.Context, code, redirectURL string, pkce *PKCE) (*Token, error)
	UserInfo(ctx context.Context, token *Token) (*OAuthUser, error)
}

var (
	registryMu sync.RWMutex
	registry   = map[string]Provider{}
)

// Register installs a provider. It is called from provider init() functions.
func Register(p Provider) {
	if p == nil || p.Name() == "" {
		return
	}
	registryMu.Lock()
	defer registryMu.Unlock()
	registry[p.Name()] = p
}

// Get returns a registered provider by name.
func Get(name string) (Provider, bool) {
	registryMu.RLock()
	defer registryMu.RUnlock()
	p, ok := registry[name]
	return p, ok
}

// Enabled returns the registered providers that are configured.
func Enabled() []Provider {
	registryMu.RLock()
	defer registryMu.RUnlock()
	out := make([]Provider, 0, len(registry))
	for _, p := range registry {
		if p.Enabled() {
			out = append(out, p)
		}
	}
	sort.Slice(out, func(i, j int) bool { return out[i].Name() < out[j].Name() })
	return out
}

// ErrNotLinked means the provider identity is valid but no local user is
// linked and auto-registration is disabled.
var ErrNotLinked = errors.New("oauth account is not linked")

// RandomToken returns a URL-safe random string.
func RandomToken(bytes int) (string, error) {
	buf := make([]byte, bytes)
	if _, err := rand.Read(buf); err != nil {
		return "", err
	}
	return base64.RawURLEncoding.EncodeToString(buf), nil
}

// NewPKCE creates an S256 verifier/challenge pair.
func NewPKCE() (*PKCE, error) {
	verifier, err := RandomToken(32)
	if err != nil {
		return nil, err
	}
	sum := sha256.Sum256([]byte(verifier))
	return &PKCE{
		Verifier:  verifier,
		Challenge: base64.RawURLEncoding.EncodeToString(sum[:]),
		Method:    "S256",
	}, nil
}

// Matches reports whether a verifier produces this challenge.
func (p *PKCE) Matches(verifier string) bool {
	if p == nil || p.Method != "S256" {
		return false
	}
	sum := sha256.Sum256([]byte(verifier))
	return hmac.Equal([]byte(p.Challenge), []byte(base64.RawURLEncoding.EncodeToString(sum[:])))
}

// CanLinkByEmail is true only when the provider returned a verified address.
func CanLinkByEmail(u *OAuthUser) bool {
	return u != nil && u.EmailVerified && strings.TrimSpace(u.Email) != ""
}

// Signer signs and verifies the state parameter with a dedicated secret.
type Signer struct {
	secret []byte
}

// NewSigner returns a signer for the OAuth state parameter.
func NewSigner(secret string) *Signer {
	return &Signer{secret: []byte(secret)}
}

// Enabled reports whether the signer has a secret.
func (s *Signer) Enabled() bool { return len(s.secret) > 0 }

// Sign returns "<state>.<hmac>".
func (s *Signer) Sign(state string) string {
	return state + "." + s.mac(state)
}

// Verify checks the signature and returns the raw state.
func (s *Signer) Verify(signed string) (string, error) {
	idx := strings.LastIndexByte(signed, '.')
	if idx <= 0 {
		return "", errors.New("oauth: malformed state")
	}
	state, sig := signed[:idx], signed[idx+1:]
	if !hmac.Equal([]byte(sig), []byte(s.mac(state))) {
		return "", errors.New("oauth: state signature mismatch")
	}
	return state, nil
}

func (s *Signer) mac(state string) string {
	mac := hmac.New(sha256.New, s.secret)
	_, _ = mac.Write([]byte(state))
	return base64.RawURLEncoding.EncodeToString(mac.Sum(nil))
}

// ValidateRedirectURL rejects redirect hosts outside the allow-list.
func ValidateRedirectURL(raw string, allowedHosts []string) (string, error) {
	parsed, err := url.Parse(raw)
	if err != nil {
		return "", fmt.Errorf("oauth: invalid redirect url: %w", err)
	}
	if parsed.Scheme != "http" && parsed.Scheme != "https" {
		return "", errors.New("oauth: redirect url must be http or https")
	}
	host := parsed.Hostname()
	for _, allowed := range allowedHosts {
		if strings.EqualFold(host, allowed) {
			return parsed.String(), nil
		}
	}
	return "", fmt.Errorf("oauth: redirect host %q is not allowed", host)
}
