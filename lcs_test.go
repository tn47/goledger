package main

import "fmt"
import "testing"

var _ = fmt.Sprintf("dummy")

func TestLCS2(t *testing.T) {
	str := lcs("hello", "world")
	if str != "l" {
		t.Errorf("expected %q, got %q", "l", str)
	}
	str = lcs("hello world", "world")
	if str != "world" {
		t.Errorf("expected %q, got %q", "world", str)
	}
	str = lcs("world hello", "world")
	if str != "world" {
		t.Errorf("expected %q, got %q", "world", str)
	}
}
