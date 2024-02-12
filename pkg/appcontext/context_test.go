package appcontext

import (
	"context"
	"testing"
)

func TestContextJWTToken(t *testing.T) {
	// Создаем новый контекст с JWT токеном
	testToken := "TestToken"
	ctx := WithJWTToken(context.Background(), testToken)

	// Извлекаем JWT токен из контекста
	token, ok := GetJWTToken(ctx)

	// Проверяем, что токен был корректно извлечен
	if !ok || token != testToken {
		t.Errorf("Failed to retrieve JWT token from context. Got: %s, want: %s", token, testToken)
	}
}
