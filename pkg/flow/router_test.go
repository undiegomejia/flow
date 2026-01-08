package flow

import (
    "net/http/httptest"
    "testing"

    routerpkg "github.com/dministrator/flow/internal/router"
)

// dummy resource that writes the id param back
type usersResource struct{}

func (u *usersResource) Index(ctx *Context)   { _ = ctx.JSON(200, map[string]string{"action": "index"}) }
func (u *usersResource) New(ctx *Context)     { _ = ctx.JSON(200, map[string]string{"action": "new"}) }
func (u *usersResource) Create(ctx *Context)  { _ = ctx.JSON(201, map[string]string{"action": "create"}) }
func (u *usersResource) Show(ctx *Context)    { _ = ctx.JSON(200, map[string]string{"id": ctx.Param("id")}) }
func (u *usersResource) Edit(ctx *Context)    { _ = ctx.JSON(200, map[string]string{"action": "edit"}) }
func (u *usersResource) Update(ctx *Context)  { _ = ctx.JSON(200, map[string]string{"action": "update"}) }
func (u *usersResource) Destroy(ctx *Context) { _ = ctx.JSON(200, map[string]string{"action": "destroy"}) }

func TestPublicRouterResources(t *testing.T) {
    app := New("test")
    // register via MakeResourceAdapter directly on inner
    rr := routerpkg.New()
    if err := rr.Resources("users", MakeResourceAdapter(app, &usersResource{})); err != nil {
        t.Fatalf("failed to register resources: %v", err)
    }

    // smoke test: GET /users/7 -> returns JSON with id
    w := httptest.NewRecorder()
    req := httptest.NewRequest("GET", "/users/7", nil)
    rr.ServeHTTP(w, req)
    if w.Code != 200 {
        t.Fatalf("expected 200, got %d", w.Code)
    }
}
