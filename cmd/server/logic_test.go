package main

import "testing"

func TestParseBanCommand(t *testing.T) {
	tests := []struct {
		name    string
		input   string
		login   string
		success bool
	}{
		{name: "valid command", input: "/ban alice", login: "alice", success: true},
		{name: "extra spaces", input: "/ban   bob", login: "bob", success: true},
		{name: "missing login", input: "/ban", login: "", success: false},
		{name: "unknown command", input: "/kick alice", login: "", success: false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			gotLogin, gotSuccess := parseBanCommand(tt.input)
			if gotSuccess != tt.success {
				t.Fatalf("success mismatch: got %v want %v", gotSuccess, tt.success)
			}
			if gotLogin != tt.login {
				t.Fatalf("login mismatch: got %q want %q", gotLogin, tt.login)
			}
		})
	}
}

func TestIsBan(t *testing.T) {
	words := []string{"bad", " rude ", ""}

	if !isBan(words, "this is a BAD message") {
		t.Fatal("expected banned word detection to be case-insensitive")
	}

	if !isBan(words, "very rude statement") {
		t.Fatal("expected trimmed banned word to be detected")
	}

	if isBan(words, "clean message") {
		t.Fatal("did not expect banned word in clean message")
	}
}
