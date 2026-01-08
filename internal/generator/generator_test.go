package generator

import (
    "os"
    "path/filepath"
    "strings"
    "testing"
)

func TestGenerateScaffoldCreatesFiles(t *testing.T) {
    td := t.TempDir()
    name := "post"
    _, err := GenerateScaffold(td, name)
    if err != nil {
        t.Fatalf("GenerateScaffold error: %v", err)
    }
    // check expected files exist
    expected := []string{
        filepath.Join(td, "app", "controllers", name+"_controller.go"),
        filepath.Join(td, "app", "models", name+".go"),
        filepath.Join(td, "app", "views", name, "index.html"),
        filepath.Join(td, "app", "views", name, "show.html"),
        filepath.Join(td, "db", "migrate"),
    }
    for _, p := range expected {
        if _, err := os.Stat(p); err != nil {
            t.Fatalf("expected file/dir %s not found: %v", p, err)
        }
    }
    // ensure at least one migration .up.sql exists in the migrations directory
    migDir := filepath.Join(td, "db", "migrate")
    entries, err := os.ReadDir(migDir)
    if err != nil {
        t.Fatalf("failed reading migrations dir: %v", err)
    }
    foundUp := false
    for _, e := range entries {
        if !e.IsDir() && strings.HasSuffix(e.Name(), ".up.sql") {
            foundUp = true
            break
        }
    }
    if !foundUp {
        t.Fatalf("no .up.sql migration found in %s", migDir)
    }
}
