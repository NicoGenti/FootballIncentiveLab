package domain

import "testing"

func TestHello(t *testing.T) {
	if got := hello(); got != "Hello, World!" {
		t.Fatalf("hello() = %q, want %q", got, "Hello, World!")
	}
}

func hello() string { return "Hello, World!" }
