package pagination

import (
	"encoding/base64"
	"strconv"
)

const (
	// DefaultPageSize is used when the caller does not specify a page size.
	DefaultPageSize = 20
	// MaxPageSize caps how many items a single request may return.
	MaxPageSize = 100
)

// Normalize clamps the requested page size into the supported range.
func Normalize(size int) int {
	switch {
	case size <= 0:
		return DefaultPageSize
	case size > MaxPageSize:
		return MaxPageSize
	default:
		return size
	}
}

// EncodeOffset turns a numeric offset into an opaque page token.
func EncodeOffset(offset int) string {
	if offset <= 0 {
		return ""
	}
	return base64.RawURLEncoding.EncodeToString([]byte(strconv.Itoa(offset)))
}

// DecodeOffset parses an opaque page token back into an offset.
func DecodeOffset(token string) (int, error) {
	if token == "" {
		return 0, nil
	}
	raw, err := base64.RawURLEncoding.DecodeString(token)
	if err != nil {
		return 0, err
	}
	offset, err := strconv.Atoi(string(raw))
	if err != nil {
		return 0, err
	}
	if offset < 0 {
		return 0, strconv.ErrRange
	}
	return offset, nil
}
