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

type wechatProvider struct {
	client *http.Client
}

func init() { oauth.Register(&wechatProvider{client: &http.Client{Timeout: 15 * time.Second}}) }

func (p *wechatProvider) Name() string        { return "wechat" }
func (p *wechatProvider) DisplayName() string { return "微信" }

func (p *wechatProvider) Enabled() bool {
	return os.Getenv("SCAFFOLD_OAUTH_WECHAT_APP_ID") != "" &&
		os.Getenv("SCAFFOLD_OAUTH_WECHAT_APP_SECRET") != ""
}

// WeChat Open Platform uses a QR consent page and does not support PKCE.
func (p *wechatProvider) AuthCodeURL(_ context.Context, state, redirectURL, _ string, _ *oauth.PKCE) string {
	query := url.Values{}
	query.Set("appid", os.Getenv("SCAFFOLD_OAUTH_WECHAT_APP_ID"))
	query.Set("redirect_uri", redirectURL)
	query.Set("response_type", "code")
	query.Set("scope", "snsapi_login")
	query.Set("state", state)
	return "https://open.weixin.qq.com/connect/qrconnect?" + query.Encode() + "#wechat_redirect"
}

func (p *wechatProvider) Exchange(ctx context.Context, code, _ string, _ *oauth.PKCE) (*oauth.Token, error) {
	query := url.Values{}
	query.Set("appid", os.Getenv("SCAFFOLD_OAUTH_WECHAT_APP_ID"))
	query.Set("secret", os.Getenv("SCAFFOLD_OAUTH_WECHAT_APP_SECRET"))
	query.Set("code", code)
	query.Set("grant_type", "authorization_code")
	req, err := http.NewRequestWithContext(ctx, http.MethodGet,
		"https://api.weixin.qq.com/sns/oauth2/access_token?"+query.Encode(), nil)
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
		OpenID      string `json:"openid"`
		ErrCode     int    `json:"errcode"`
		ErrMsg      string `json:"errmsg"`
	}
	if err := json.NewDecoder(io.LimitReader(resp.Body, 1<<20)).Decode(&payload); err != nil {
		return nil, err
	}
	if payload.ErrCode != 0 || payload.AccessToken == "" {
		return nil, fmt.Errorf("wechat: token exchange failed: %d %s", payload.ErrCode, payload.ErrMsg)
	}
	return &oauth.Token{
		AccessToken: payload.AccessToken,
		Raw:         map[string]string{"openid": payload.OpenID},
	}, nil
}

func (p *wechatProvider) UserInfo(ctx context.Context, token *oauth.Token) (*oauth.OAuthUser, error) {
	openID := token.Raw["openid"]
	query := url.Values{}
	query.Set("access_token", token.AccessToken)
	query.Set("openid", openID)
	query.Set("lang", "zh_CN")
	req, err := http.NewRequestWithContext(ctx, http.MethodGet,
		"https://api.weixin.qq.com/sns/userinfo?"+query.Encode(), nil)
	if err != nil {
		return nil, err
	}
	resp, err := p.client.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()
	var profile struct {
		OpenID   string `json:"openid"`
		Nickname string `json:"nickname"`
		Avatar   string `json:"headimgurl"`
		ErrCode  int    `json:"errcode"`
		ErrMsg   string `json:"errmsg"`
	}
	if err := json.NewDecoder(io.LimitReader(resp.Body, 1<<20)).Decode(&profile); err != nil {
		return nil, err
	}
	if profile.ErrCode != 0 || profile.OpenID == "" {
		return nil, fmt.Errorf("wechat: userinfo failed: %d %s", profile.ErrCode, profile.ErrMsg)
	}
	// WeChat does not return an email address. Account linking therefore never
	// relies on email matching for this provider.
	return &oauth.OAuthUser{
		ProviderUserID: profile.OpenID,
		Name:           profile.Nickname,
		Avatar:         profile.Avatar,
		EmailVerified:  false,
	}, nil
}
