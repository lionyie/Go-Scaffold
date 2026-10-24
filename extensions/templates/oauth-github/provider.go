package providers

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"os"
	"strconv"
	"strings"
	"time"

	"github.com/example/scaffold/services/internal/auth/oauth"
)

type githubProvider struct {
	client *http.Client
}

func init() { oauth.Register(&githubProvider{client: &http.Client{Timeout: 15 * time.Second}}) }

func (p *githubProvider) Name() string        { return "github" }
func (p *githubProvider) DisplayName() string { return "GitHub" }

func (p *githubProvider) Enabled() bool {
	return os.Getenv("SCAFFOLD_OAUTH_GITHUB_CLIENT_ID") != "" &&
		os.Getenv("SCAFFOLD_OAUTH_GITHUB_CLIENT_SECRET") != ""
}

func (p *githubProvider) AuthCodeURL(_ context.Context, state, redirectURL, _ string, pkce *oauth.PKCE) string {
	query := url.Values{}
	query.Set("client_id", os.Getenv("SCAFFOLD_OAUTH_GITHUB_CLIENT_ID"))
	query.Set("redirect_uri", redirectURL)
	query.Set("scope", "read:user user:email")
	query.Set("state", state)
	if pkce != nil {
		query.Set("code_challenge", pkce.Challenge)
		query.Set("code_challenge_method", pkce.Method)
	}
	return "https://github.com/login/oauth/authorize?" + query.Encode()
}

func (p *githubProvider) Exchange(ctx context.Context, code, redirectURL string, pkce *oauth.PKCE) (*oauth.Token, error) {
	form := url.Values{}
	form.Set("client_id", os.Getenv("SCAFFOLD_OAUTH_GITHUB_CLIENT_ID"))
	form.Set("client_secret", os.Getenv("SCAFFOLD_OAUTH_GITHUB_CLIENT_SECRET"))
	form.Set("code", code)
	form.Set("redirect_uri", redirectURL)
	if pkce != nil {
		form.Set("code_verifier", pkce.Verifier)
	}
	req, err := http.NewRequestWithContext(ctx, http.MethodPost,
		"https://github.com/login/oauth/access_token", strings.NewReader(form.Encode()))
	if err != nil {
		return nil, err
	}
	req.Header.Set("Accept", "application/json")
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
		return nil, fmt.Errorf("github: token endpoint returned %d: %s", resp.StatusCode, string(body))
	}
	var payload struct {
		AccessToken string `json:"access_token"`
		TokenType   string `json:"token_type"`
		Error       string `json:"error"`
	}
	if err := json.Unmarshal(body, &payload); err != nil {
		return nil, err
	}
	if payload.Error != "" || payload.AccessToken == "" {
		return nil, fmt.Errorf("github: token exchange failed: %s", payload.Error)
	}
	return &oauth.Token{AccessToken: payload.AccessToken, TokenType: payload.TokenType}, nil
}

func (p *githubProvider) UserInfo(ctx context.Context, token *oauth.Token) (*oauth.OAuthUser, error) {
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, "https://api.github.com/user", nil)
	if err != nil {
		return nil, err
	}
	req.Header.Set("Authorization", "Bearer "+token.AccessToken)
	req.Header.Set("Accept", "application/vnd.github+json")
	resp, err := p.client.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("github: user endpoint returned %d", resp.StatusCode)
	}
	var profile struct {
		ID        int64  `json:"id"`
		Login     string `json:"login"`
		Name      string `json:"name"`
		Email     string `json:"email"`
		AvatarURL string `json:"avatar_url"`
	}
	if err := json.NewDecoder(io.LimitReader(resp.Body, 1<<20)).Decode(&profile); err != nil {
		return nil, err
	}
	user := &oauth.OAuthUser{
		ProviderUserID: strconv.FormatInt(profile.ID, 10),
		Email:          profile.Email,
		Name:           firstNonEmpty(profile.Name, profile.Login),
		Avatar:         profile.AvatarURL,
	}
	if user.Email == "" {
		email, verified := p.primaryEmail(ctx, token.AccessToken)
		user.Email = email
		user.EmailVerified = verified
	} else {
		// GitHub only exposes a verified primary address on /user/emails.
		user.EmailVerified = false
	}
	return user, nil
}

func (p *githubProvider) primaryEmail(ctx context.Context, accessToken string) (string, bool) {
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, "https://api.github.com/user/emails", nil)
	if err != nil {
		return "", false
	}
	req.Header.Set("Authorization", "Bearer "+accessToken)
	req.Header.Set("Accept", "application/vnd.github+json")
	resp, err := p.client.Do(req)
	if err != nil {
		return "", false
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		return "", false
	}
	var emails []struct {
		Email    string `json:"email"`
		Primary  bool   `json:"primary"`
		Verified bool   `json:"verified"`
	}
	if err := json.NewDecoder(io.LimitReader(resp.Body, 1<<20)).Decode(&emails); err != nil {
		return "", false
	}
	for _, item := range emails {
		if item.Primary {
			return item.Email, item.Verified
		}
	}
	return "", false
}

func firstNonEmpty(values ...string) string {
	for _, value := range values {
		if strings.TrimSpace(value) != "" {
			return value
		}
	}
	return ""
}
