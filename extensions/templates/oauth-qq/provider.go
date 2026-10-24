package providers

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"os"
	"time"

	"github.com/example/scaffold/services/internal/auth/oauth"
)

type qqProvider struct {
	client *http.Client
}

func init() { oauth.Register(&qqProvider{client: &http.Client{Timeout: 15 * time.Second}}) }

func (p *qqProvider) Name() string        { return "qq" }
func (p *qqProvider) DisplayName() string { return "QQ" }

func (p *qqProvider) Enabled() bool {
	return os.Getenv("SCAFFOLD_OAUTH_QQ_APP_ID") != "" &&
		os.Getenv("SCAFFOLD_OAUTH_QQ_APP_KEY") != ""
}

func (p *qqProvider) AuthCodeURL(_ context.Context, state, redirectURL, _ string, _ *oauth.PKCE) string {
	query := url.Values{}
	query.Set("response_type", "code")
	query.Set("client_id", os.Getenv("SCAFFOLD_OAUTH_QQ_APP_ID"))
	query.Set("redirect_uri", redirectURL)
	query.Set("state", state)
	return "https://graph.qq.com/oauth2.0/authorize?" + query.Encode()
}

func (p *qqProvider) Exchange(ctx context.Context, code, redirectURL string, _ *oauth.PKCE) (*oauth.Token, error) {
	query := url.Values{}
	query.Set("grant_type", "authorization_code")
	query.Set("client_id", os.Getenv("SCAFFOLD_OAUTH_QQ_APP_ID"))
	query.Set("client_secret", os.Getenv("SCAFFOLD_OAUTH_QQ_APP_KEY"))
	query.Set("code", code)
	query.Set("redirect_uri", redirectURL)
	query.Set("fmt", "json")
	req, err := http.NewRequestWithContext(ctx, http.MethodGet,
		"https://graph.qq.com/oauth2.0/token?"+query.Encode(), nil)
	if err != nil {
		return nil, err
	}
	resp, err := p.client.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()
	var payload struct {
		AccessToken string `json:"access_token"`
		Error       int    `json:"error"`
		ErrorDesc   string `json:"error_description"`
	}
	if err := json.NewDecoder(io.LimitReader(resp.Body, 1<<20)).Decode(&payload); err != nil {
		return nil, err
	}
	if payload.Error != 0 || payload.AccessToken == "" {
		return nil, fmt.Errorf("qq: token exchange failed: %d %s", payload.Error, payload.ErrorDesc)
	}
	return &oauth.Token{AccessToken: payload.AccessToken}, nil
}

func (p *qqProvider) UserInfo(ctx context.Context, token *oauth.Token) (*oauth.OAuthUser, error) {
	openID, err := p.openID(ctx, token.AccessToken)
	if err != nil {
		return nil, err
	}
	query := url.Values{}
	query.Set("access_token", token.AccessToken)
	query.Set("oauth_consumer_key", os.Getenv("SCAFFOLD_OAUTH_QQ_APP_ID"))
	query.Set("openid", openID)
	req, err := http.NewRequestWithContext(ctx, http.MethodGet,
		"https://graph.qq.com/user/get_user_info?"+query.Encode(), nil)
	if err != nil {
		return nil, err
	}
	resp, err := p.client.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()
	var profile struct {
		Ret        int    `json:"ret"`
		Msg        string `json:"msg"`
		Nickname   string `json:"nickname"`
		AvatarURL  string `json:"figureurl_qq_2"`
		AvatarURL1 string `json:"figureurl_qq_1"`
	}
	if err := json.NewDecoder(io.LimitReader(resp.Body, 1<<20)).Decode(&profile); err != nil {
		return nil, err
	}
	if profile.Ret != 0 {
		return nil, fmt.Errorf("qq: userinfo failed: %d %s", profile.Ret, profile.Msg)
	}
	avatar := profile.AvatarURL
	if avatar == "" {
		avatar = profile.AvatarURL1
	}
	// QQ Connect does not return an email address.
	return &oauth.OAuthUser{
		ProviderUserID: openID,
		Name:           profile.Nickname,
		Avatar:         avatar,
		EmailVerified:  false,
	}, nil
}

func (p *qqProvider) openID(ctx context.Context, accessToken string) (string, error) {
	query := url.Values{}
	query.Set("access_token", accessToken)
	query.Set("fmt", "json")
	req, err := http.NewRequestWithContext(ctx, http.MethodGet,
		"https://graph.qq.com/oauth2.0/me?"+query.Encode(), nil)
	if err != nil {
		return "", err
	}
	resp, err := p.client.Do(req)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()
	var payload struct {
		OpenID  string `json:"openid"`
		Error   int    `json:"error"`
		Message string `json:"error_description"`
	}
	if err := json.NewDecoder(io.LimitReader(resp.Body, 1<<20)).Decode(&payload); err != nil {
		return "", err
	}
	if payload.Error != 0 || payload.OpenID == "" {
		return "", fmt.Errorf("qq: openid lookup failed: %d %s", payload.Error, payload.Message)
	}
	return payload.OpenID, nil
}
