package providers

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"os"
	"strings"
	"time"

	"github.com/example/scaffold/services/internal/auth/oauth"
)

type googleProvider struct {
	client *http.Client
}

func init() { oauth.Register(&googleProvider{client: &http.Client{Timeout: 15 * time.Second}}) }

func (p *googleProvider) Name() string        { return "google" }
func (p *googleProvider) DisplayName() string { return "Google" }

func (p *googleProvider) Enabled() bool {
	return os.Getenv("SCAFFOLD_OAUTH_GOOGLE_CLIENT_ID") != "" &&
		os.Getenv("SCAFFOLD_OAUTH_GOOGLE_CLIENT_SECRET") != ""
}

func (p *googleProvider) AuthCodeURL(_ context.Context, state, redirectURL, nonce string, pkce *oauth.PKCE) string {
	query := url.Values{}
	query.Set("client_id", os.Getenv("SCAFFOLD_OAUTH_GOOGLE_CLIENT_ID"))
	query.Set("redirect_uri", redirectURL)
	query.Set("response_type", "code")
	query.Set("scope", "openid email profile")
	query.Set("state", state)
	query.Set("nonce", nonce)
	query.Set("prompt", "select_account")
	if pkce != nil {
		query.Set("code_challenge", pkce.Challenge)
		query.Set("code_challenge_method", pkce.Method)
	}
	return "https://accounts.google.com/o/oauth2/v2/auth?" + query.Encode()
}

func (p *googleProvider) Exchange(ctx context.Context, code, redirectURL string, pkce *oauth.PKCE) (*oauth.Token, error) {
	form := url.Values{}
	form.Set("client_id", os.Getenv("SCAFFOLD_OAUTH_GOOGLE_CLIENT_ID"))
	form.Set("client_secret", os.Getenv("SCAFFOLD_OAUTH_GOOGLE_CLIENT_SECRET"))
	form.Set("code", code)
	form.Set("redirect_uri", redirectURL)
	form.Set("grant_type", "authorization_code")
	if pkce != nil {
		form.Set("code_verifier", pkce.Verifier)
	}
	req, err := http.NewRequestWithContext(ctx, http.MethodPost,
		"https://oauth2.googleapis.com/token", strings.NewReader(form.Encode()))
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
		return nil, fmt.Errorf("google: token endpoint returned %d: %s", resp.StatusCode, string(body))
	}
	var payload struct {
		AccessToken string `json:"access_token"`
		TokenType   string `json:"token_type"`
	}
	if err := json.Unmarshal(body, &payload); err != nil {
		return nil, err
	}
	if payload.AccessToken == "" {
		return nil, fmt.Errorf("google: token exchange returned no access token")
	}
	return &oauth.Token{AccessToken: payload.AccessToken, TokenType: payload.TokenType}, nil
}

func (p *googleProvider) UserInfo(ctx context.Context, token *oauth.Token) (*oauth.OAuthUser, error) {
	req, err := http.NewRequestWithContext(ctx, http.MethodGet,
		"https://openidconnect.googleapis.com/v1/userinfo", nil)
	if err != nil {
		return nil, err
	}
	req.Header.Set("Authorization", "Bearer "+token.AccessToken)
	resp, err := p.client.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("google: userinfo endpoint returned %d", resp.StatusCode)
	}
	var profile struct {
		Sub           string `json:"sub"`
		Email         string `json:"email"`
		EmailVerified bool   `json:"email_verified"`
		Name          string `json:"name"`
		Picture       string `json:"picture"`
	}
	if err := json.NewDecoder(io.LimitReader(resp.Body, 1<<20)).Decode(&profile); err != nil {
		return nil, err
	}
	if profile.Sub == "" {
		return nil, fmt.Errorf("google: userinfo returned no subject")
	}
	return &oauth.OAuthUser{
		ProviderUserID: profile.Sub,
		Email:          profile.Email,
		EmailVerified:  profile.EmailVerified,
		Name:           profile.Name,
		Avatar:         profile.Picture,
	}, nil
}
