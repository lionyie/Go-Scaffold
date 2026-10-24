package server

import (
	"encoding/json"
	"errors"
	"net/http"
	"net/url"
	"strings"
	"time"

	"github.com/gorilla/mux"

	"github.com/example/scaffold/services/internal/auth/oauth"
	_ "github.com/example/scaffold/services/internal/auth/oauthproviders_gen"
	"github.com/example/scaffold/services/internal/biz"
	"github.com/example/scaffold/services/internal/conf"
	authpkg "github.com/example/scaffold/services/internal/pkg/auth"
)

const (
	oauthStateTTL  = 10 * time.Minute
	oauthTicketTTL = 60 * time.Second
)

// OAuthHandler serves the provider-agnostic sign-in flow. These routes live
// outside the Kratos middleware chain because the callback is a browser
// redirect from a third party.
type OAuthHandler struct {
	bc     *conf.Bootstrap
	store  oauth.Store
	uc     *biz.OAuthUsecase
	tokens *authpkg.TokenManager
	signer *oauth.Signer
	router *mux.Router
}

// NewOAuthHandler builds the OAuth HTTP handler.
func NewOAuthHandler(bc *conf.Bootstrap, store oauth.Store, uc *biz.OAuthUsecase, tokens *authpkg.TokenManager) *OAuthHandler {
	h := &OAuthHandler{
		bc:     bc,
		store:  store,
		uc:     uc,
		tokens: tokens,
		signer: oauth.NewSigner(bc.GetAuth().GetOauth().GetStateSecret()),
	}
	r := mux.NewRouter()
	r.HandleFunc("/api/v1/auth/oauth/providers", h.listProviders).Methods(http.MethodGet)
	r.HandleFunc("/api/v1/auth/oauth/exchange", h.exchange).Methods(http.MethodPost)
	r.HandleFunc("/api/v1/auth/oauth/accounts", h.listAccounts).Methods(http.MethodGet)
	r.HandleFunc("/api/v1/auth/oauth/{provider}/start", h.start).Methods(http.MethodPost)
	r.HandleFunc("/api/v1/auth/oauth/{provider}/callback", h.callback).Methods(http.MethodGet)
	r.HandleFunc("/api/v1/auth/oauth/{provider}", h.unlink).Methods(http.MethodDelete)
	h.router = r
	return h
}

func (h *OAuthHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	h.router.ServeHTTP(w, r)
}

func (h *OAuthHandler) listProviders(w http.ResponseWriter, _ *http.Request) {
	type item struct {
		Name        string `json:"name"`
		DisplayName string `json:"displayName"`
	}
	providers := oauth.Enabled()
	items := make([]item, 0, len(providers))
	for _, p := range providers {
		items = append(items, item{Name: p.Name(), DisplayName: p.DisplayName()})
	}
	writeJSON(w, http.StatusOK, map[string]any{"providers": items})
}

func (h *OAuthHandler) start(w http.ResponseWriter, r *http.Request) {
	if !h.signer.Enabled() {
		writeError(w, http.StatusServiceUnavailable, "oauth state secret is not configured")
		return
	}
	providerName := mux.Vars(r)["provider"]
	provider, ok := oauth.Get(providerName)
	if !ok || !provider.Enabled() {
		writeError(w, http.StatusNotFound, "oauth provider is not enabled")
		return
	}

	redirectURL, err := h.callbackURL(r, providerName)
	if err != nil {
		writeError(w, http.StatusBadRequest, err.Error())
		return
	}

	state, err := oauth.RandomToken(24)
	if err != nil {
		writeError(w, http.StatusInternalServerError, "failed to create state")
		return
	}
	pkce, err := oauth.NewPKCE()
	if err != nil {
		writeError(w, http.StatusInternalServerError, "failed to create PKCE")
		return
	}
	nonce, err := oauth.RandomToken(24)
	if err != nil {
		writeError(w, http.StatusInternalServerError, "failed to create nonce")
		return
	}
	if err := h.store.SaveState(r.Context(), state, oauth.StateData{
		Provider:      providerName,
		PKCEVerifier:  pkce.Verifier,
		Nonce:         nonce,
		RedirectAfter: "",
	}, oauthStateTTL); err != nil {
		writeError(w, http.StatusInternalServerError, "failed to persist state")
		return
	}

	authURL := provider.AuthCodeURL(r.Context(), h.signer.Sign(state), redirectURL, nonce, pkce)
	if authURL == "" {
		writeError(w, http.StatusInternalServerError, "provider is not ready")
		return
	}
	writeJSON(w, http.StatusOK, map[string]string{"authUrl": authURL})
}

func (h *OAuthHandler) callback(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Referrer-Policy", "no-referrer")
	providerName := mux.Vars(r)["provider"]
	provider, ok := oauth.Get(providerName)
	if !ok || !provider.Enabled() {
		h.redirectError(w, r, "oauth provider is not enabled")
		return
	}
	if errParam := r.URL.Query().Get("error"); errParam != "" {
		h.redirectError(w, r, "provider returned "+errParam)
		return
	}

	rawState, err := h.signer.Verify(r.URL.Query().Get("state"))
	if err != nil {
		h.redirectError(w, r, "invalid oauth state")
		return
	}
	state, err := h.store.ConsumeState(r.Context(), rawState)
	if err != nil {
		h.redirectError(w, r, "oauth state expired or already used")
		return
	}
	if state.Provider != providerName {
		h.redirectError(w, r, "oauth provider mismatch")
		return
	}

	redirectURL, err := h.callbackURL(r, providerName)
	if err != nil {
		h.redirectError(w, r, err.Error())
		return
	}
	var pkce *oauth.PKCE
	if state.PKCEVerifier != "" {
		pkce = &oauth.PKCE{Verifier: state.PKCEVerifier, Method: "S256"}
	}
	token, err := provider.Exchange(r.Context(), r.URL.Query().Get("code"), redirectURL, pkce)
	if err != nil {
		h.redirectError(w, r, "token exchange failed")
		return
	}
	user, err := provider.UserInfo(r.Context(), token)
	if err != nil {
		h.redirectError(w, r, "failed to load provider profile")
		return
	}
	result, err := h.uc.LoginOrRegister(r.Context(), providerName, user)
	if errors.Is(err, oauth.ErrNotLinked) {
		h.redirectError(w, r, "account is not linked")
		return
	}
	if err != nil {
		h.redirectError(w, r, "sign-in failed")
		return
	}

	ticket, err := oauth.RandomToken(32)
	if err != nil {
		h.redirectError(w, r, "failed to create login ticket")
		return
	}
	if err := h.store.SaveTicket(r.Context(), ticket, oauth.TicketData{
		AccessToken:  result.AccessToken,
		RefreshToken: result.RefreshToken,
	}, oauthTicketTTL); err != nil {
		h.redirectError(w, r, "failed to persist login ticket")
		return
	}
	target := h.adminRedirectURL(r)
	parsed, err := url.Parse(target)
	if err != nil {
		h.redirectError(w, r, "invalid admin redirect url")
		return
	}
	query := parsed.Query()
	query.Set("ticket", ticket)
	parsed.RawQuery = query.Encode()
	http.Redirect(w, r, parsed.String(), http.StatusFound)
}

func (h *OAuthHandler) exchange(w http.ResponseWriter, r *http.Request) {
	var body struct {
		Ticket string `json:"ticket"`
	}
	if err := json.NewDecoder(http.MaxBytesReader(w, r.Body, 8<<10)).Decode(&body); err != nil {
		writeError(w, http.StatusBadRequest, "invalid request body")
		return
	}
	ticket, err := h.store.ConsumeTicket(r.Context(), body.Ticket)
	if err != nil {
		writeError(w, http.StatusBadRequest, "login ticket is invalid or expired")
		return
	}
	writeJSON(w, http.StatusOK, map[string]string{
		"accessToken":  ticket.AccessToken,
		"refreshToken": ticket.RefreshToken,
		"tokenType":    "Bearer",
	})
}

func (h *OAuthHandler) listAccounts(w http.ResponseWriter, r *http.Request) {
	principal, ok := h.principal(r)
	if !ok {
		writeError(w, http.StatusUnauthorized, "authentication required")
		return
	}
	accounts, err := h.uc.ListAccounts(r.Context(), principal.UserID)
	if err != nil {
		writeError(w, http.StatusInternalServerError, "failed to list accounts")
		return
	}
	type item struct {
		Provider string `json:"provider"`
		Email    string `json:"email"`
		Name     string `json:"name"`
		Avatar   string `json:"avatar"`
	}
	items := make([]item, 0, len(accounts))
	for _, account := range accounts {
		items = append(items, item{
			Provider: account.Provider,
			Email:    account.Email,
			Name:     account.Name,
			Avatar:   account.Avatar,
		})
	}
	writeJSON(w, http.StatusOK, map[string]any{"accounts": items})
}

func (h *OAuthHandler) unlink(w http.ResponseWriter, r *http.Request) {
	principal, ok := h.principal(r)
	if !ok {
		writeError(w, http.StatusUnauthorized, "authentication required")
		return
	}
	providerName := mux.Vars(r)["provider"]
	if err := h.uc.Unlink(r.Context(), principal.UserID, providerName); err != nil {
		writeError(w, http.StatusInternalServerError, "failed to unlink account")
		return
	}
	w.WriteHeader(http.StatusNoContent)
}

func (h *OAuthHandler) principal(r *http.Request) (*authpkg.Principal, bool) {
	token := authpkg.Bearer(r.Header.Get("Authorization"))
	if token == "" {
		return nil, false
	}
	claims, err := h.tokens.ParseAccess(token)
	if err != nil {
		return nil, false
	}
	return &authpkg.Principal{
		UserID:             claims.UserID,
		Username:           claims.Username,
		SessionID:          claims.SessionID,
		IsAdmin:            claims.Admin,
		MustChangePassword: claims.MustChangePassword,
	}, true
}

func (h *OAuthHandler) callbackURL(r *http.Request, provider string) (string, error) {
	scheme := "http"
	if r.TLS != nil {
		scheme = "https"
	}
	if forwarded := r.Header.Get("X-Forwarded-Proto"); forwarded != "" {
		scheme = strings.Split(forwarded, ",")[0]
	}
	host := r.Host
	if forwarded := r.Header.Get("X-Forwarded-Host"); forwarded != "" {
		host = strings.Split(forwarded, ",")[0]
	}
	raw := scheme + "://" + host + "/api/v1/auth/oauth/" + url.PathEscape(provider) + "/callback"
	return oauth.ValidateRedirectURL(raw, h.bc.GetAuth().GetOauth().GetAllowedRedirectHosts())
}

func (h *OAuthHandler) adminRedirectURL(r *http.Request) string {
	if target := h.bc.GetAuth().GetOauth().GetAdminRedirectUrl(); target != "" {
		return target
	}
	scheme := "http"
	if r.TLS != nil {
		scheme = "https"
	}
	if forwarded := r.Header.Get("X-Forwarded-Proto"); forwarded != "" {
		scheme = strings.Split(forwarded, ",")[0]
	}
	host := r.Host
	if forwarded := r.Header.Get("X-Forwarded-Host"); forwarded != "" {
		host = strings.Split(forwarded, ",")[0]
	}
	return scheme + "://" + host + "/admin/oauth/callback"
}

func (h *OAuthHandler) redirectError(w http.ResponseWriter, r *http.Request, message string) {
	parsed, err := url.Parse(h.adminRedirectURL(r))
	if err != nil {
		writeError(w, http.StatusBadRequest, message)
		return
	}
	query := parsed.Query()
	query.Set("error", message)
	parsed.RawQuery = query.Encode()
	http.Redirect(w, r, parsed.String(), http.StatusFound)
}

func writeJSON(w http.ResponseWriter, status int, payload any) {
	w.Header().Set("Content-Type", "application/json; charset=utf-8")
	w.WriteHeader(status)
	_ = json.NewEncoder(w).Encode(payload)
}

func writeError(w http.ResponseWriter, status int, message string) {
	writeJSON(w, status, map[string]string{"error": message})
}
