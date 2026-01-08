package flow

import (
    "context"
    "testing"
    "time"

    orm "github.com/dministrator/flow/internal/orm"
    _ "modernc.org/sqlite"
)

func TestAppSetBunAndAutoMigrate(t *testing.T) {
    adapter, err := orm.Connect("file::memory:?cache=shared")
    if err != nil {
        t.Fatalf("connect bun: %v", err)
    }
    defer adapter.Close()

    app := New("bun-test", WithBun(adapter))

    type Item struct {
        ID        int64     `bun:"id,pk,autoincrement"`
        Name      string    `bun:"name"`
        CreatedAt time.Time `bun:"created_at"`
    }

    ctx := context.Background()
    if err := AutoMigrate(ctx, app, (*Item)(nil)); err != nil {
        t.Fatalf("auto migrate: %v", err)
    }

    db := DB(app)
    if db == nil {
        t.Fatalf("expected bun DB on app")
    }

    // basic insert/select to ensure bun is usable via App
    it := &Item{Name: "alpha", CreatedAt: time.Now()}
    if _, err := db.NewInsert().Model(it).Exec(ctx); err != nil {
        t.Fatalf("insert: %v", err)
    }

    var got Item
    if err := db.NewSelect().Model(&got).Where("name = ?", "alpha").Scan(ctx); err != nil {
        t.Fatalf("select: %v", err)
    }
    if got.Name != "alpha" {
        t.Fatalf("expected alpha, got %s", got.Name)
    }
}
