package agent

import (
	"testing"
)

func TestTruncate_ShortString(t *testing.T) {
	got := truncate("hello", 10)
	if got != "hello" {
		t.Errorf("truncate short = %q, want %q", got, "hello")
	}
}

func TestTruncate_ExactLength(t *testing.T) {
	got := truncate("hello", 5)
	if got != "hello" {
		t.Errorf("truncate exact = %q, want %q", got, "hello")
	}
}

func TestTruncate_LongString(t *testing.T) {
	got := truncate("hello world", 5)
	want := "hello..."
	if got != want {
		t.Errorf("truncate long = %q, want %q", got, want)
	}
}

func TestTruncate_EmptyString(t *testing.T) {
	got := truncate("", 10)
	if got != "" {
		t.Errorf("truncate empty = %q, want %q", got, "")
	}
}

func TestTruncate_ZeroN(t *testing.T) {
	got := truncate("hello", 0)
	want := "..."
	if got != want {
		t.Errorf("truncate zero = %q, want %q", got, want)
	}
}

func TestTruncate_SingleChar(t *testing.T) {
	got := truncate("a", 1)
	if got != "a" {
		t.Errorf("truncate single = %q, want %q", got, "a")
	}
}

func TestTruncate_TruncatedAtBoundary(t *testing.T) {
	s := "abcdefghij"
	got := truncate(s, 10)
	if got != s {
		t.Errorf("truncate boundary = %q, want %q", got, s)
	}

	got = truncate(s, 9)
	want := "abcdefghi..."
	if got != want {
		t.Errorf("truncate just over = %q, want %q", got, want)
	}
}
