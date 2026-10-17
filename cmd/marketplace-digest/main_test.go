package main

import (
	"context"
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestCallEnvelopeRetriesJSONRateLimit(t *testing.T) {
	attempts := 0
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		attempts++
		w.Header().Set("Content-Type", "application/json")
		if attempts == 1 {
			w.WriteHeader(http.StatusTooManyRequests)
			fmt.Fprint(w, `{"ok":false,"error":{"code":"RATE_LIMITED","message":"retry"}}`)
			return
		}
		fmt.Fprint(w, `{"ok":true,"data":{}}`)
	}))
	defer server.Close()

	if err := callEnvelope(context.Background(), "test", http.MethodPost, server.URL, []byte(`{}`), "retry-test"); err != nil {
		t.Fatal(err)
	}
	if attempts != 2 {
		t.Fatalf("got %d attempts, want 2", attempts)
	}
}
