package migrations

import (
    "database/sql"
    "fmt"
    "io/fs"
    "os"
    "path/filepath"
    "sort"
    "strings"
)

// MigrationRunner runs timestamped SQL migrations stored in a directory.
// Migration files follow the naming convention:
//   20260108120000_create_users.up.sql
//   20260108120000_create_users.down.sql
// ApplyAll executes all .up.sql files in ascending timestamp order.
type MigrationRunner struct{}

// ApplyAll applies all up migrations found in dir using the provided db.
func (m *MigrationRunner) ApplyAll(dir string, db *sql.DB) error {
    ups, err := m.collect(dir, ".up.sql")
    if err != nil {
        return err
    }
    sort.Strings(ups)
    for _, p := range ups {
        if err := m.execFile(db, p); err != nil {
            return fmt.Errorf("apply %s: %w", filepath.Base(p), err)
        }
    }
    return nil
}

// RollbackLast finds the latest migration pair and executes its down SQL.
func (m *MigrationRunner) RollbackLast(dir string, db *sql.DB) error {
    downs, err := m.collect(dir, ".down.sql")
    if err != nil {
        return err
    }
    if len(downs) == 0 {
        return fmt.Errorf("no down migrations found in %s", dir)
    }
    sort.Strings(downs)
    last := downs[len(downs)-1]
    if err := m.execFile(db, last); err != nil {
        return fmt.Errorf("rollback %s: %w", filepath.Base(last), err)
    }
    return nil
}

// collect returns absolute paths of files in dir that end with suffix.
func (m *MigrationRunner) collect(dir, suffix string) ([]string, error) {
    var out []string
    entries, err := os.ReadDir(dir)
    if err != nil {
        if os.IsNotExist(err) {
            return nil, fmt.Errorf("migrations directory not found: %s", dir)
        }
        return nil, err
    }
    for _, e := range entries {
        if e.IsDir() {
            continue
        }
        name := e.Name()
        if strings.HasSuffix(name, suffix) {
            out = append(out, filepath.Join(dir, name))
        }
    }
    return out, nil
}

func (m *MigrationRunner) execFile(db *sql.DB, path string) error {
    b, err := os.ReadFile(path)
    if err != nil {
        return err
    }
    sqlText := string(b)
    // Execute in a transaction for safety
    tx, err := db.Begin()
    if err != nil {
        return err
    }
    if _, err := tx.Exec(sqlText); err != nil {
        _ = tx.Rollback()
        return err
    }
    if err := tx.Commit(); err != nil {
        return err
    }
    return nil
}

// ApplySingle runs a single migration file (convenience)
func (m *MigrationRunner) ApplySingle(path string, db *sql.DB) error {
    info, err := os.Stat(path)
    if err != nil {
        return err
    }
    if info.IsDir() {
        return fmt.Errorf("path is a directory: %s", path)
    }
    return m.execFile(db, path)
}

// ListMigrations returns file names of migrations (both up and down) in dir.
func (m *MigrationRunner) ListMigrations(dir string) ([]string, error) {
    var out []string
    err := filepath.WalkDir(dir, func(p string, d fs.DirEntry, err error) error {
        if err != nil {
            return err
        }
        if d.IsDir() {
            return nil
        }
        if strings.HasSuffix(d.Name(), ".up.sql") || strings.HasSuffix(d.Name(), ".down.sql") {
            out = append(out, p)
        }
        return nil
    })
    if err != nil {
        return nil, err
    }
    sort.Strings(out)
    return out, nil
}
