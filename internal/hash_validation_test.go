package internal

import (
	"testing"

	"golang.org/x/crypto/bcrypt"
)

func TestCheckPasswordValidAndMalformedLines(t *testing.T) {
	hash, err := bcrypt.GenerateFromPassword([]byte("pass123"), 12)
	if err != nil {
		t.Fatalf("failed to generate hash: %v", err)
	}

	lines := []string{
		"broken-line-without-separators",
		"alice;admin;" + string(hash),
	}

	ok, role := CheckPassword(lines, "alice", "pass123")
	if !ok {
		t.Fatal("expected valid credentials to pass")
	}
	if role != "admin" {
		t.Fatalf("expected role admin, got %q", role)
	}
}

func TestCheckPasswordMissingUser(t *testing.T) {
	lines := []string{"broken-line", "bob;user;hash"}

	ok, role := CheckPassword(lines, "alice", "pass123")
	if ok {
		t.Fatal("expected missing user authentication to fail")
	}
	if role != "" {
		t.Fatalf("expected empty role, got %q", role)
	}
}
