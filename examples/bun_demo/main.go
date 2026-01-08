package main

import (
    "context"
    "fmt"
    "log"
    "os"
    "path/filepath"
    "time"

    flow "github.com/dministrator/flow/pkg/flow"
    orm "github.com/dministrator/flow/internal/orm"
    _ "modernc.org/sqlite"
)

// Post is a simple model similar to what the generator produces. The generator
// will emit a similar struct in app/models when you run `flow generate model`.
type Post struct {
    flow.Model
    Title       string    `bun:"title" json:"title"`
    PublishedAt time.Time `bun:"published_at" json:"published_at"`
}

func main() {
    // ensure example data directory exists
    if err := os.MkdirAll(filepath.Dir("examples/bun_demo/db.sqlite"), 0o755); err != nil {
        log.Fatalf("mkdir: %v", err)
    }

    ctx := context.Background()

    // connect to sqlite file for the example
    adapter, err := orm.Connect("file:examples/bun_demo/db.sqlite?_foreign_keys=1")
    if err != nil {
        log.Fatalf("connect db: %v", err)
    }
    defer adapter.Close()

    app := flow.New("examples-bun-demo", flow.WithBun(adapter))

    // Auto-migrate the Post table (convenience for examples/tests)
    if err := flow.AutoMigrate(ctx, app, (*Post)(nil)); err != nil {
        log.Fatalf("auto migrate: %v", err)
    }

    // insert a sample record
    db := app.Bun()
    p := &Post{Title: "Hello from bun demo", PublishedAt: time.Now()}
    if _, err := db.NewInsert().Model(p).Exec(ctx); err != nil {
        log.Fatalf("insert: %v", err)
    }

    var got Post
    if err := db.NewSelect().Model(&got).Where("title = ?", "Hello from bun demo").Scan(ctx); err != nil {
        log.Fatalf("select: %v", err)
    }
    fmt.Printf("retrieved post: %#v\n", got)
}
