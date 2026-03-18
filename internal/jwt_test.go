package internal

import "testing"

func TestGenerateAndValidateJWT(t *testing.T) {
	t.Setenv("JWT_SECRET", "test-secret")

	token, err := GenerateJWT("alice", "admin")
	if err != nil {
		t.Fatalf("GenerateJWT failed: %v", err)
	}
	if token == "" {
		t.Fatal("expected non-empty token")
	}

	claims, err := ValidateToken(token)
	if err != nil {
		t.Fatalf("ValidateToken failed: %v", err)
	}
	if claims.Login != "alice" {
		t.Fatalf("expected login alice, got %q", claims.Login)
	}
	if claims.Role != "admin" {
		t.Fatalf("expected role admin, got %q", claims.Role)
	}
}

func TestGenerateJWTWithoutSecret(t *testing.T) {
	t.Setenv("JWT_SECRET", "")

	_, err := GenerateJWT("alice", "admin")
	if err == nil {
		t.Fatal("expected error when JWT_SECRET is not set")
	}
}
