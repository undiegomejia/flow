package generator

import (
	"os"
	"strings"
	"testing"
)

func TestParseFieldSpecSimple(t *testing.T) {
	fs, err := ParseFieldSpec("title:string")
	if err != nil {
		t.Fatal(err)
	}
	if fs.Name != "title" {
		t.Fatalf("expected name title, got %s", fs.Name)
	}
	if fs.GoName != "Title" {
		t.Fatalf("expected GoName Title, got %s", fs.GoName)
	}
	if fs.GoType != "string" {
		t.Fatalf("expected GoType string, got %s", fs.GoType)
	}
	if fs.SQLType != "TEXT" {
		t.Fatalf("expected SQLType TEXT, got %s", fs.SQLType)
	}
}

func TestParseFieldSpecNullableDefault(t *testing.T) {
	fs, err := ParseFieldSpec("price:decimal(10,2),default=0,nullable")
	if err != nil {
		t.Fatal(err)
	}
	if fs.Name != "price" {
		t.Fatalf("expected name price, got %s", fs.Name)
	}
	if fs.Nullable != true {
		t.Fatalf("expected nullable true")
	}
	if fs.Default == nil || *fs.Default != "0" {
		t.Fatalf("expected default 0, got %v", fs.Default)
	}
}

func TestGenerateScaffoldWithFields(t *testing.T) {
	td := t.TempDir()
	name := "product"
	fields := []string{"title:string", "price:decimal(10,2),default=0,nullable", "stock:int"}
	created, err := GenerateScaffold(td, name, fields...)
	if err != nil {
		t.Fatalf("GenerateScaffold error: %v", err)
	}
	// verify migration files exist in created list
	foundUp := ""
	for _, p := range created {
		if strings.HasSuffix(p, ".up.sql") {
			foundUp = p
			break
		}
	}
	if foundUp == "" {
		t.Fatalf("no up migration created, created: %v", created)
	}
	b, err := os.ReadFile(foundUp)
	if err != nil {
		t.Fatalf("read migration failed: %v", err)
	}
	content := string(b)
	// check columns present
	if !strings.Contains(content, "title TEXT") {
		t.Fatalf("migration missing title column: %s", content)
	}
	if !strings.Contains(content, "price DECIMAL(10,2)") && !strings.Contains(content, "price DECIMAL") {
		t.Fatalf("migration missing price column: %s", content)
	}
	if !strings.Contains(content, "stock INTEGER") {
		t.Fatalf("migration missing stock column: %s", content)
	}
}
