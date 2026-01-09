// Package flow: lightweight model helpers and DB accessor.
//
// This file provides a thin abstraction over database/sql so application
// code can access the DB from the App. It intentionally does not depend on
// a specific ORM — projects can plug bun/gorm on top of the provided
// *sql.DB. A small Model struct defines common fields used by generators.
package flow

import (
	"database/sql"
	"time"
)

// Model is a small embedding struct that generated models can include to
// obtain standard fields. It deliberately uses primitive types to avoid
// forcing an ORM implementation.
type Model struct {
	ID        int64        `db:"id" json:"id"`
	CreatedAt time.Time    `db:"created_at" json:"created_at"`
	UpdatedAt time.Time    `db:"updated_at" json:"updated_at"`
	DeletedAt sql.NullTime `db:"deleted_at" json:"deleted_at,omitempty"`
}

// WithDB sets a *sql.DB on the App during construction. Use this option to
// inject a database connection into an App.
func WithDB(db *sql.DB) Option {
	return func(a *App) { a.SetDB(db) }
}

// SetDB attaches a database connection to the App.
func (a *App) SetDB(db *sql.DB) {
	a.db = db
}

// DB returns the attached *sql.DB or nil if none was set.
func (a *App) DB() *sql.DB { return a.db }
