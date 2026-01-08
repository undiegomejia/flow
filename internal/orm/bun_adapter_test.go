package orm

import (
    "context"
    "testing"
    "time"

    _ "modernc.org/sqlite"
)

// Basic CRUD test using bun on an in-memory sqlite.
func TestBunAdapterBasicCRUD(t *testing.T) {
    adapter, err := Connect("file::memory:?cache=shared")
    if err != nil {
        t.Fatalf("connect: %v", err)
    }
    defer adapter.Close()

    type User struct {
        ID        int64     `bun:"id,pk,autoincrement"`
        Name      string    `bun:"name"`
        CreatedAt time.Time `bun:"created_at"`
    }

    ctx := context.Background()

    // create table
    if _, err := adapter.DB.NewCreateTable().Model((*User)(nil)).IfNotExists().Exec(ctx); err != nil {
        t.Fatalf("create table: %v", err)
    }

    // insert
    u := &User{Name: "Alice", CreatedAt: time.Now()}
    if _, err := adapter.DB.NewInsert().Model(u).Exec(ctx); err != nil {
        t.Fatalf("insert: %v", err)
    }

    // select
    var got User
    if err := adapter.DB.NewSelect().Model(&got).Where("name = ?", "Alice").Scan(ctx); err != nil {
        t.Fatalf("select: %v", err)
    }
    if got.Name != "Alice" {
        t.Fatalf("expected Alice, got %s", got.Name)
    }

    // update
    got.Name = "Bob"
    if _, err := adapter.DB.NewUpdate().Model(&got).WherePK().Exec(ctx); err != nil {
        t.Fatalf("update: %v", err)
    }

    var after User
    if err := adapter.DB.NewSelect().Model(&after).Where("id = ?", got.ID).Scan(ctx); err != nil {
        t.Fatalf("select after update: %v", err)
    }
    if after.Name != "Bob" {
        t.Fatalf("expected Bob, got %s", after.Name)
    }

    // delete
    if _, err := adapter.DB.NewDelete().Model(&after).WherePK().Exec(ctx); err != nil {
        t.Fatalf("delete: %v", err)
    }

    var users []User
    if err := adapter.DB.NewSelect().Model(&users).Scan(ctx); err != nil {
        t.Fatalf("select all: %v", err)
    }
    if len(users) != 0 {
        t.Fatalf("expected 0 rows, got %d", len(users))
    }
}
