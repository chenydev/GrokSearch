package grok

import (
	"net/http"
	"testing"
	"time"
)

func TestParseRetryAfterSeconds(t *testing.T) {
	d, ok := parseRetryAfter("3")
	if !ok || d != 3*time.Second {
		t.Fatalf("got %v %v", d, ok)
	}
}

func TestParseRetryAfterDate(t *testing.T) {
	value := time.Now().Add(2 * time.Second).UTC().Format(http.TimeFormat)
	d, ok := parseRetryAfter(value)
	if !ok || d <= 0 {
		t.Fatalf("got %v %v", d, ok)
	}
}
