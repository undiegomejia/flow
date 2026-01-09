package router

import (
	"io"
	"net/http"
	"net/http/httptest"
	"testing"
)

// testCtrl is a simple ResourceController used by tests.
type testCtrl struct{}

func (c *testCtrl) Index(w http.ResponseWriter, r *http.Request)   { _, _ = w.Write([]byte("i")) }
func (c *testCtrl) New(w http.ResponseWriter, r *http.Request)     { _, _ = w.Write([]byte("n")) }
func (c *testCtrl) Create(w http.ResponseWriter, r *http.Request)  { _, _ = w.Write([]byte("c")) }
func (c *testCtrl) Show(w http.ResponseWriter, r *http.Request)    { _, _ = w.Write([]byte("s")) }
func (c *testCtrl) Edit(w http.ResponseWriter, r *http.Request)    { _, _ = w.Write([]byte("e")) }
func (c *testCtrl) Update(w http.ResponseWriter, r *http.Request)  { _, _ = w.Write([]byte("u")) }
func (c *testCtrl) Destroy(w http.ResponseWriter, r *http.Request) { _, _ = w.Write([]byte("d")) }

func TestRouterBasicMatching(t *testing.T) {
	t.Run("static route", func(t *testing.T) {
		r := New()
		r.Get("/", func(w http.ResponseWriter, req *http.Request) {
			_, _ = w.Write([]byte("ok"))
		})

		rr := httptest.NewRecorder()
		req := httptest.NewRequest("GET", "/", nil)
		r.ServeHTTP(rr, req)

		if rr.Code != http.StatusOK {
			t.Fatalf("expected status 200, got %d", rr.Code)
		}
		body, _ := io.ReadAll(rr.Body)
		if string(body) != "ok" {
			t.Fatalf("unexpected body: %s", string(body))
		}
	})

	t.Run("param route", func(t *testing.T) {
		r := New()
		r.Get("/users/:id", func(w http.ResponseWriter, req *http.Request) {
			id := Param(req, "id")
			_, _ = w.Write([]byte(id))
		})

		rr := httptest.NewRecorder()
		req := httptest.NewRequest("GET", "/users/42", nil)
		r.ServeHTTP(rr, req)

		if rr.Code != http.StatusOK {
			t.Fatalf("expected status 200, got %d", rr.Code)
		}
		body, _ := io.ReadAll(rr.Body)
		if string(body) != "42" {
			t.Fatalf("expected param 42, got %s", string(body))
		}
	})

	t.Run("method not allowed", func(t *testing.T) {
		r := New()
		r.Get("/onlyget", func(w http.ResponseWriter, req *http.Request) {
			_, _ = w.Write([]byte("get"))
		})

		rr := httptest.NewRecorder()
		req := httptest.NewRequest("POST", "/onlyget", nil)
		r.ServeHTTP(rr, req)

		if rr.Code != http.StatusMethodNotAllowed {
			t.Fatalf("expected 405, got %d", rr.Code)
		}
	})

	t.Run("not found", func(t *testing.T) {
		r := New()

		rr := httptest.NewRecorder()
		req := httptest.NewRequest("GET", "/nope", nil)
		r.ServeHTTP(rr, req)

		if rr.Code != http.StatusNotFound {
			t.Fatalf("expected 404, got %d", rr.Code)
		}
	})

	t.Run("trailing slash equivalence", func(t *testing.T) {
		r := New()
		r.Get("/users", func(w http.ResponseWriter, req *http.Request) {
			_, _ = w.Write([]byte("users"))
		})

		rr := httptest.NewRecorder()
		req := httptest.NewRequest("GET", "/users/", nil)
		r.ServeHTTP(rr, req)

		if rr.Code != http.StatusOK {
			t.Fatalf("expected 200 for /users/, got %d", rr.Code)
		}
	})

	t.Run("multiple params", func(t *testing.T) {
		r := New()
		r.Get("/orgs/:org_id/users/:id", func(w http.ResponseWriter, req *http.Request) {
			org := Param(req, "org_id")
			id := Param(req, "id")
			_, _ = w.Write([]byte(org + ":" + id))
		})

		rr := httptest.NewRecorder()
		req := httptest.NewRequest("GET", "/orgs/7/users/99", nil)
		r.ServeHTTP(rr, req)

		if rr.Code != http.StatusOK {
			t.Fatalf("expected 200, got %d", rr.Code)
		}
		body, _ := io.ReadAll(rr.Body)
		if string(body) != "7:99" {
			t.Fatalf("expected 7:99, got %s", string(body))
		}
	})
}

func TestNamedRoutesAndMiddleware(t *testing.T) {
	t.Run("named route url generation", func(t *testing.T) {
		r := New()
		r.HandleNamed("post_show", "GET", "/posts/:id", func(w http.ResponseWriter, req *http.Request) {
			_, _ = w.Write([]byte("show"))
		})

		path, err := r.URL("post_show", map[string]string{"id": "42"})
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if path != "/posts/42" {
			t.Fatalf("expected /posts/42 got %s", path)
		}
	})

	t.Run("missing param in url generation", func(t *testing.T) {
		r := New()
		r.HandleNamed("post_show", "GET", "/posts/:id", func(w http.ResponseWriter, req *http.Request) {})
		_, err := r.URL("post_show", map[string]string{})
		if err == nil {
			t.Fatalf("expected error for missing param")
		}
	})

	t.Run("per-route middleware execution", func(t *testing.T) {
		r := New()
		called := ""
		mw := func(next http.Handler) http.Handler {
			return http.HandlerFunc(func(w http.ResponseWriter, req *http.Request) {
				called += "m"
				next.ServeHTTP(w, req)
			})
		}
		r.HandleWith("GET", "/ok", func(w http.ResponseWriter, req *http.Request) {
			called += "h"
			_, _ = w.Write([]byte("ok"))
		}, mw)

		rr := httptest.NewRecorder()
		req := httptest.NewRequest("GET", "/ok", nil)
		r.ServeHTTP(rr, req)
		if rr.Code != http.StatusOK {
			t.Fatalf("expected 200, got %d", rr.Code)
		}
		if called != "mh" {
			t.Fatalf("expected middleware then handler called, got %s", called)
		}
	})
}

func TestResourcesRegisterNames(t *testing.T) {
	r := New()
	c := &testCtrl{}
	if err := r.Resources("users", c); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	// ensure named URL generation works for show
	p, err := r.URL("users_show", map[string]string{"id": "7"})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if p != "/users/7" {
		t.Fatalf("expected /users/7 got %s", p)
	}
}
