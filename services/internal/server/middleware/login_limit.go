package middleware

import (
	"context"
	"net"
	"strings"
	"sync"
	"time"

	"github.com/go-kratos/kratos/v3/errors"
	"github.com/go-kratos/kratos/v3/middleware"
	"github.com/go-kratos/kratos/v3/transport"
	khttp "github.com/go-kratos/kratos/v3/transport/http"

	identityv1 "github.com/example/scaffold/services/api/identity/v1"
)

const loginFailureWindow = time.Minute

type loginLimiter struct {
	mu       sync.Mutex
	attempts map[string][]time.Time
}

func newLoginLimiter() *loginLimiter {
	return &loginLimiter{attempts: make(map[string][]time.Time)}
}

func (l *loginLimiter) blocked(key string) bool {
	l.mu.Lock()
	defer l.mu.Unlock()
	l.prune(key)
	return len(l.attempts[key]) >= 5
}

func (l *loginLimiter) fail(key string) {
	l.mu.Lock()
	defer l.mu.Unlock()
	l.attempts[key] = append(l.attempts[key], time.Now())
}

func (l *loginLimiter) reset(key string) {
	l.mu.Lock()
	defer l.mu.Unlock()
	delete(l.attempts, key)
}

func (l *loginLimiter) prune(key string) {
	cutoff := time.Now().Add(-loginFailureWindow)
	items := l.attempts[key]
	kept := items[:0]
	for _, item := range items {
		if item.After(cutoff) {
			kept = append(kept, item)
		}
	}
	if len(kept) == 0 {
		delete(l.attempts, key)
		return
	}
	l.attempts[key] = kept
}

func loginKey(ctx context.Context, req any) string {
	username := ""
	if login, ok := req.(interface{ GetUsername() string }); ok {
		username = strings.ToLower(strings.TrimSpace(login.GetUsername()))
	}
	return loginClientIP(ctx) + "|" + username
}

func loginClientIP(ctx context.Context) string {
	if rq, ok := khttp.RequestFromServerContext(ctx); ok {
		if forwarded := rq.Header.Get("X-Forwarded-For"); forwarded != "" {
			return strings.TrimSpace(strings.Split(forwarded, ",")[0])
		}
		if host, _, err := net.SplitHostPort(rq.RemoteAddr); err == nil {
			return host
		}
		return rq.RemoteAddr
	}
	if tr, ok := transport.FromServerContext(ctx); ok {
		if forwarded := tr.RequestHeader().Get("X-Forwarded-For"); forwarded != "" {
			return strings.TrimSpace(strings.Split(forwarded, ",")[0])
		}
	}
	return "unknown"
}

// LoginRateLimit blocks a username/IP pair after five failures in a minute.
func LoginRateLimit() middleware.Middleware {
	limiter := newLoginLimiter()
	return func(handler middleware.Handler) middleware.Handler {
		return func(ctx context.Context, req any) (any, error) {
			if contextOperation(ctx) != identityv1.OperationAuthServiceLogin {
				return handler(ctx, req)
			}
			key := loginKey(ctx, req)
			if limiter.blocked(key) {
				return nil, errors.New(429, "TOO_MANY_REQUESTS", "too many login attempts; try again later")
			}
			resp, err := handler(ctx, req)
			if err != nil {
				limiter.fail(key)
			} else {
				limiter.reset(key)
			}
			return resp, err
		}
	}
}
