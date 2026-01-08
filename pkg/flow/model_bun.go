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
