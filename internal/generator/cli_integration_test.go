package generator

import (
    "os"
    "os/exec"
    "path/filepath"
    "strings"
    "testing"
)

// findRepoRoot walks up from cwd until it finds a go.mod file.
func findRepoRoot() string {
    dir, err := os.Getwd()
    if err != nil {
        return "."
    }
    for {
        if _, err := os.Stat(filepath.Join(dir, "go.mod")); err == nil {
            return dir
        }
        parent := filepath.Dir(dir)
        if parent == dir {
            break
        }
        dir = parent
    }
    return "."
}

func TestCLI_GenerateModel_WritesBunTaggedModel(t *testing.T) {
    repo := findRepoRoot()
    tmp := t.TempDir()

    // build the CLI binary into the temp dir then run it to avoid go run
    bin := filepath.Join(tmp, "flow-cli")
    build := exec.Command("go", "build", "-o", bin, "./cmd/flow")
    build.Dir = repo
    if bout, err := build.CombinedOutput(); err != nil {
        t.Fatalf("build cli failed: %v\noutput: %s", err, string(bout))
    }

    // run generated binary: generate model into tmp target
    cmd := exec.Command(bin, "generate", "model", "Post", "title:string", "--target", tmp)
    cmd.Dir = repo
    out, err := cmd.CombinedOutput()
    t.Logf("cmd output: %s", string(out))
    if err != nil {
        t.Fatalf("go run cmd/flow failed: %v", err)
    }

    // check file exists and contains bun tag
    gotPath := filepath.Join(tmp, "app", "models", "post.go")
    b, err := os.ReadFile(gotPath)
    if err != nil {
        t.Fatalf("read generated model: %v", err)
    }
    s := string(b)
    if !strings.Contains(s, `bun:"title"`) {
        t.Fatalf("generated model missing bun tag for title: %s", s)
    }
    if !strings.Contains(s, "type Post struct") {
        t.Fatalf("generated model missing struct declaration: %s", s)
    }
}
