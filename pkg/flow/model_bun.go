// Package flow: bun model helpers
//
// This file provides small helpers to work with bun from application code
// via the App's Bun() accessor. It is intentionally minimal — a starting
// point for generator integrations and migrations.
package flow

import (
    "context"
    "fmt"

    "github.com/uptrace/bun"
)

// AutoMigrate creates tables for the provided models using bun's CreateTable
// helpers. It is a convenience for development and tests; production apps
// may prefer explicit migrations.
func AutoMigrate(ctx context.Context, app *App, models ...interface{}) error {
    if app == nil {
        return fmt.Errorf("app is nil")
    }
    db := app.Bun()
    if db == nil {
        return fmt.Errorf("bun DB not configured on app")
    }

    for _, m := range models {
        _, err := db.NewCreateTable().Model(m).IfNotExists().Exec(ctx)
        if err != nil {
            return fmt.Errorf("create table: %w", err)
        }
    }
    return nil
}

// DB returns the underlying *bun.DB or nil.
func DB(app *App) *bun.DB {
    if app == nil {
        return nil
    }
    return app.Bun()
}

// BeginTx starts a new transaction using the App's Bun DB.
func BeginTx(ctx context.Context, app *App) (*bun.Tx, error) {
    db := DB(app)
    if db == nil {
        return nil, fmt.Errorf("bun DB not configured on app")
    }
    txVal, err := db.BeginTx(ctx, nil)
    if err != nil {
        return nil, fmt.Errorf("begin tx: %w", err)
    }
    // db.BeginTx returns a value type in some bun versions; take its address
    return &txVal, nil
}

// RunInTx runs fn inside a transaction. If fn returns an error the
// transaction is rolled back; otherwise it is committed.
func RunInTx(ctx context.Context, app *App, fn func(ctx context.Context, tx *bun.Tx) error) error {
    tx, err := BeginTx(ctx, app)
    if err != nil {
        return err
    }
    // ensure rollback on panic
    defer func() {
        if r := recover(); r != nil {
            _ = tx.Rollback()
            panic(r)
        }
    }()

    if err := fn(ctx, tx); err != nil {
        _ = tx.Rollback()
        return err
    }
    if err := tx.Commit(); err != nil {
        return fmt.Errorf("commit tx: %w", err)
    }
    return nil
}

// Insert inserts the provided model using bun.
func Insert(ctx context.Context, app *App, model interface{}) error {
    db := DB(app)
    if db == nil {
        return fmt.Errorf("bun DB not configured on app")
    }
    if _, err := db.NewInsert().Model(model).Exec(ctx); err != nil {
        return err
    }
    return nil
}

// Update updates the provided model using its primary key.
func Update(ctx context.Context, app *App, model interface{}) error {
    db := DB(app)
    if db == nil {
        return fmt.Errorf("bun DB not configured on app")
    }
    if _, err := db.NewUpdate().Model(model).WherePK().Exec(ctx); err != nil {
        return err
    }
    return nil
}

// Delete removes the provided model using its primary key.
func Delete(ctx context.Context, app *App, model interface{}) error {
    db := DB(app)
    if db == nil {
        return fmt.Errorf("bun DB not configured on app")
    }
    if _, err := db.NewDelete().Model(model).WherePK().Exec(ctx); err != nil {
        return err
    }
    return nil
}

// FindByPK loads a model by primary key into dest.
func FindByPK(ctx context.Context, app *App, dest interface{}, pk interface{}) error {
    db := DB(app)
    if db == nil {
        return fmt.Errorf("bun DB not configured on app")
    }
    if err := db.NewSelect().Model(dest).Where("id = ?", pk).Scan(ctx); err != nil {
        return err
    }
    return nil
}
