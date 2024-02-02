package appcontext

import "context"

type contextKey string

func (c contextKey) String() string {
	return string(c)
}

var (
	ContextJWTToken = contextKey("jwtToken")
)

func WithJWTToken(ctx context.Context, token string) context.Context {
	return context.WithValue(ctx, ContextJWTToken, token)
}

func GetJWTToken(ctx context.Context) (string, bool) {
	token, ok := ctx.Value(ContextJWTToken).(string)
	return token, ok
}
