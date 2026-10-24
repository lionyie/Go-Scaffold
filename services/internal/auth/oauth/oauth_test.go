package oauth

import (
	"context"
	"errors"
	"strings"
	"testing"
	"time"
)

func TestStateExpiry(t *testing.T) {
	store := NewMemoryStore()
	ctx := context.Background()
	if err := store.SaveState(ctx, "s", StateData{Provider: "github"}, -time.Second); err != nil {
		t.Fatal(err)
	}
	if _, err := store.ConsumeState(ctx, "s"); !errors.Is(err, ErrStateNotFound) {
		t.Fatalf("expected ErrStateNotFound, got %v", err)
	}
}

func TestTicketReplay(t *testing.T) {
	store := NewMemoryStore()
	ctx := context.Background()
	if err := store.SaveTicket(ctx, "t", TicketData{AccessToken: "a", RefreshToken: "r"}, time.Minute); err != nil {
		t.Fatal(err)
	}
	if _, err := store.ConsumeTicket(ctx, "t"); err != nil {
		t.Fatalf("first consume failed: %v", err)
	}
	if _, err := store.ConsumeTicket(ctx, "t"); !errors.Is(err, ErrTicketNotFound) {
		t.Fatalf("expected replay to fail, got %v", err)
	}
}

func TestStateSignature(t *testing.T) {
	signer := NewSigner("secret")
	signed := signer.Sign("state")
	if _, err := signer.Verify(signed); err != nil {
		t.Fatalf("verify failed: %v", err)
	}
	if _, err := signer.Verify(signed + "x"); err == nil {
		t.Fatal("tampered state was accepted")
	}
	if _, err := NewSigner("other").Verify(signed); err == nil {
		t.Fatal("state signed with another secret was accepted")
	}
}

func TestPKCE(t *testing.T) {
	pkce, err := NewPKCE()
	if err != nil {
		t.Fatal(err)
	}
	if !pkce.Matches(pkce.Verifier) {
		t.Fatal("correct verifier did not match")
	}
	if pkce.Matches(pkce.Verifier + "x") {
		t.Fatal("wrong verifier matched")
	}
}

func TestCanLinkByEmail(t *testing.T) {
	if CanLinkByEmail(&OAuthUser{Email: "a@example.com", EmailVerified: false}) {
		t.Fatal("unverified email must not auto-link")
	}
	if CanLinkByEmail(&OAuthUser{Email: "", EmailVerified: true}) {
		t.Fatal("empty email must not auto-link")
	}
	if !CanLinkByEmail(&OAuthUser{Email: "a@example.com", EmailVerified: true}) {
		t.Fatal("verified email should auto-link")
	}
}

func TestWechatStyleUserWithoutEmail(t *testing.T) {
	// WeChat and QQ return no email; this must never panic and must never be
	// considered linkable by email.
	user := &OAuthUser{ProviderUserID: "openid-1", Name: "微信用户"}
	if strings.TrimSpace(user.Email) != "" {
		t.Fatal("test fixture should have no email")
	}
	if CanLinkByEmail(user) {
		t.Fatal("provider without email must not auto-link")
	}
}
