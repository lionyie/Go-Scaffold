package identityv1

// Redact keeps credentials out of request logs. The Kratos logging middleware
// calls Redact instead of printing the payload when this interface is present.

func (x *LoginRequest) Redact() string {
	return `{"username":"<redacted>","password":"<redacted>"}`
}

func (x *RefreshRequest) Redact() string {
	return `{"refreshToken":"<redacted>"}`
}

func (x *LogoutRequest) Redact() string {
	return `{"refreshToken":"<redacted>"}`
}

func (x *CreateUserRequest) Redact() string {
	return `{"user":{...},"password":"<redacted>"}`
}

func (x *UpdateUserRequest) Redact() string {
	return `{"id":"<redacted>","user":{...},"password":"<redacted>"}`
}
