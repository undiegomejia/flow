package router

import (
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestStaticAndParamRoutes(t *testing.T) {
	r := New()
	r.Get("/", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(200)
		_, _ = w.Write([]byte("root"))
	})

	r.Get("/users/:id", func(w http.ResponseWriter, r *http.Request) {
		id := Param(r, "id")
		w.WriteHeader(200)
		_, _ = w.Write([]byte("user:" + id))
	})

	// root
	rr := httptest.NewRecorder()
	req := httptest.NewRequest("GET", "/", nil)
	r.ServeHTTP(rr, req)
	if rr.Code != 200 || rr.Body.String() != "root" {
		t.Fatalf("unexpected root response: %d %q", rr.Code, rr.Body.String())
	}

	// user show
	rr = httptest.NewRecorder()
	req = httptest.NewRequest("GET", "/users/42", nil)
	r.ServeHTTP(rr, req)
	if rr.Code != 200 || rr.Body.String() != "user:42" {
		t.Fatalf("unexpected user response: %d %q", rr.Code, rr.Body.String())
	}
}

func TestMethodNotAllowed(t *testing.T) {
	r := New()
	r.Get("/items", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(200)
	})

	rr := httptest.NewRecorder()
	req := httptest.NewRequest("POST", "/items", nil)
	r.ServeHTTP(rr, req)
	if rr.Code != http.StatusMethodNotAllowed {
		t.Fatalf("expected 405 for method mismatch, got %d", rr.Code)
	}
}
