package orm

import (
    "context"
    "database/sql"
    "fmt"

    "github.com/uptrace/bun"
    "github.com/uptrace/bun/dialect/sqlitedialect"
)

// BunAdapter is a thin wrapper around bun.DB exposing the DB instance and
// the underlying *sql.DB for lifecycle management.
type BunAdapter struct {
    DB    *bun.DB
    SQLDB *sql.DB
}

// Connect opens a database connection using the provided DSN and returns a BunAdapter.
// The caller is responsible for closing the returned adapter (adapter.Close()).
func Connect(dsn string) (*BunAdapter, error) {
    // use database/sql for driver registration (caller supplies DSN for sqlite)
    sqdb, err := sql.Open("sqlite", dsn)
    if err != nil {
        return nil, fmt.Errorf("open sql: %w", err)
    }

    db := bun.NewDB(sqdb, sqlitedialect.New())
    return &BunAdapter{DB: db, SQLDB: sqdb}, nil
}

// Close closes the underlying *sql.DB connection.
func (b *BunAdapter) Close() error {
    if b == nil || b.SQLDB == nil {
        return nil
    }
    return b.SQLDB.Close()
}

// Ping checks connectivity.
func (b *BunAdapter) Ping(ctx context.Context) error {
    if b == nil || b.SQLDB == nil {
        return fmt.Errorf("bun adapter: nil")
    }
    return b.SQLDB.PingContext(ctx)
}
