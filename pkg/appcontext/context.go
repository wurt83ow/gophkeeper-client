// Package appcontext provides utility functions for working with context in the application.

package appcontext

import "context"

type contextKey string

// String returns the string representation of the context key.
func (c contextKey) String() string {
	return string(c)
}

// ContextJWTToken represents the context key for JWT token.
var (
	ContextJWTToken = contextKey("jwtToken")
)

// WithJWTToken returns a new context with the provided JWT token.
func WithJWTToken(ctx context.Context, token string) context.Context {
	return context.WithValue(ctx, ContextJWTToken, token)
}

// GetJWTToken retrieves the JWT token from the context.
func GetJWTToken(ctx context.Context) (string, bool) {
	token, ok := ctx.Value(ContextJWTToken).(string)
	return token, ok
}
