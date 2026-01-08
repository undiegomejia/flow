# Flow — A Minimal MVC Framework for Go

> Flow is an opinionated, small, and developer-friendly MVC framework for Go. It provides a tiny routing DSL, controller/context helpers, a simple view manager with layouts/partials, cookie sessions, a migrations runner, and generators to scaffold controllers, models, views and migrations. Flow's design favors explicitness, testability and a pleasant developer loop — similar in spirit to Rails, but idiomatic Go.

This README gives a concise introduction, quickstart, and reference for the main building blocks so you (or contributors) can get started quickly.

## Highlights

- Small, dependency-free internal router with RESTful `Resources` helper.
- Public `pkg/flow` API with `Controller`, `Context`, `ViewManager`, and `App` bootstrap.
- Simple template loading with layout and partial support (conventions: `views/{controller}/{action}.html`, `views/layouts/*.html`, `views/partials/*`).
- Cookie-based sessions & flash helpers (lightweight, no external deps).
- Migration runner (timestamped up/down SQL) and CLI generator scaffolding (controllers, models, migrations).
- Designed for testability: small units, adapters and good test coverage.

## Quickstart (run the example)

The repository contains a small example app under `examples/simple`. To run it from WSL or a Unix-like environment:

```bash
# from the repository root (WSL)
cd /home/dministrator/repos/flow
go run ./examples/simple
```

Open a browser or curl:

```bash
curl http://localhost:3000/users/1
```

The example demonstrates controllers and views (see `examples/simple/app/controllers` and `examples/simple/app/views`).

## Install & Tests

Make sure you have Go 1.20+ (project uses module mode). From the repository root:

```bash
# run all tests
wsl bash -lc "cd /home/dministrator/repos/flow && go test ./... -v"

# build the project
wsl bash -lc "cd /home/dministrator/repos/flow && go build ./..."
```

## Key Concepts and Files

Below is a quick map of important packages and conventions to understand Flow's internals and how to use it.

- `internal/router` — a small HTTP router used by the framework. Supports parameterized routes (`/users/:id`), `Get/Post/...` helpers and `Resources(base, controller)` for RESTful wiring.
- `pkg/flow` — the public API surface used by application code:
	- `App` (in `pkg/flow/app.go`): application bootstrap (router, middleware, server lifecycle, Views and Sessions).
	- `Context` (in `pkg/flow/context.go`): request-scoped helper passed to controller actions (render, JSON, params, sessions, flash).
	- `Controller` (in `pkg/flow/controller.go`): base type and adapter helpers.
	- `ViewManager` (in `pkg/flow/view.go`): template loader, caching, layout and partial resolution.
	- `SessionManager` (in `pkg/flow/session.go`): cookie-based sessions and flash helpers.

### Router and Controllers (example)

Flow exposes a `NewRouter(app *flow.App)` constructor. Handlers accept `func(*flow.Context)` which simplifies controller code. Example:

```go
app := flow.New("my-app")
r := flow.NewRouter(app)

// register a Context-based handler
r.Get("/hello", func(ctx *flow.Context) {
		ctx.JSON(200, map[string]string{"hello": "world"})
})

// resources (RESTful routes)
users := NewUsersController(app) // implement flow.Resource
_ = r.Resources("users", users)

app.SetRouter(r.Handler())
```

The `MakeResourceAdapter(app, res)` adapts a `flow.Resource` (methods that accept `*Context`) to the internal router.

### Views and Templates

Flow uses `html/template` and a `ViewManager` with a small convention:

- Template directory: configurable via `NewViewManager("views")`.
- View lookup: `views/{controller}/{action}.html` (use `ViewManager.Render("users/show", data, ctx)`).
- Layouts: put shared layouts in `views/layouts/*.html` (layouts can call `{{ template "content" . }}` to insert the view content).
- Partials: put reusable fragments in `views/partials/*.html` and reference them in templates.

Example controller rendering:

```go
func (u *UsersController) Show(ctx *flow.Context) {
		data := map[string]interface{}{"Title": "User", "ID": ctx.Param("id")}
		_ = u.Render(ctx, "users/show", data)
}
```

### Sessions & Flash

Flow includes a minimal cookie-based signed session manager with helpers attached to `Context`:

- `ctx.Session()` returns the session for the request.
- `ctx.AddFlash(kind, message)` adds a flash message.
- `ctx.Flashes()` reads and clears flash messages.

The implementation is intentionally small and dependency-free to keep things portable and testable.

### Migrations & Generators

There is an internal migrations runner that expects timestamped `*.up.sql` and `*.down.sql` files and runs them in a transaction. The CLI scaffolding provides generator helpers to create controllers, models, views and migrations.

Check `internal/migrations` and `internal/generator` for the implementation and templates.

## Contributing

The project is organized to be easy to contribute to:

- Write small, focused tests (packages already include tests for router, migrations, generator templates).
- Follow repository conventions: controllers under `examples/<app>/app/controllers`, views under `examples/<app>/app/views`.
- Run `go test ./...` before opening a PR.

Suggested issues to start with (examples):

- Add more `ViewManager` FuncMap helpers (url helpers, formatters).
- Add flags and options to generators (force, db dialect).
- Improve CLI UX (cobra commands, descriptive messages).

## Project Status & Roadmap

The repository implements a working prototype with router, controllers, views (layouts/partials), sessions, migrations, and generators. The next planned improvements include:

- richer view helpers and FuncMap support,
- optional ORM adapters (bun or gorm) and model helpers,
- better generator templates and flags,
- CI workflow and developer DX improvements (hot-reload integration),
- fuller documentation and examples.

## License

This project is provided under an MIT-style license. Modify as appropriate for your needs.

