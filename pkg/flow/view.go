// Package flow: view rendering helpers.
//
// ViewManager is a small template loader/cacher used by the framework to
// render templates according to conventions. It is intentionally minimal
// for the prototype: templates are looked up by name relative to a root
// directory and parsed on first use.
package flow

import (
	"fmt"
	"html/template"
	"os"
	"path/filepath"
	"sync"
)

// ViewManager holds template loading configuration and a simple cache.
type ViewManager struct {
	TemplateDir string
	mu          sync.RWMutex
	cache       map[string]*template.Template
}

// NewViewManager constructs a ViewManager which will look for templates in
// templateDir (relative to the working directory).
func NewViewManager(templateDir string) *ViewManager {
	return &ViewManager{TemplateDir: templateDir, cache: make(map[string]*template.Template)}
}

// Render loads (or retrieves from cache) the named template and executes it
// with the provided data into the context's ResponseWriter. Template names
// are file paths relative to TemplateDir without extension, e.g. "users/show".
func (v *ViewManager) Render(name string, data interface{}, ctx *Context) error {
	if v == nil {
		return fmt.Errorf("view manager: nil")
	}
	tpl, err := v.loadTemplate(name)
	if err != nil {
		return err
	}
	// Prefer executing a "content" template (common pattern where views
	// define {{ define "content" }}...{{ end }} and layouts render that
	// via {{ template "content" . }}). If no "content" template exists,
	// fall back to executing the parsed file's base name (e.g. "show.html").
	execName := "content"
	if tpl.Lookup(execName) == nil {
		execName = filepath.Base(name) + ".html"
	}
	return ctx.RenderTemplate(tpl, execName, data)
}

func (v *ViewManager) loadTemplate(name string) (*template.Template, error) {
	v.mu.RLock()
	t, ok := v.cache[name]
	v.mu.RUnlock()
	if ok {
		return t, nil
	}

	// build list of candidate files: layouts, partials, shared, then the view
	var files []string

	// collect layouts (prefer application/layout order)
	layoutGlob := filepath.Join(v.TemplateDir, "layouts", "*.html")
	if lays, _ := filepath.Glob(layoutGlob); len(lays) > 0 {
		files = append(files, lays...)
	}

	// collect partials
	partialGlob := filepath.Join(v.TemplateDir, "partials", "*.html")
	if parts, _ := filepath.Glob(partialGlob); len(parts) > 0 {
		files = append(files, parts...)
	}

	// collect shared helpers (optional)
	sharedGlob := filepath.Join(v.TemplateDir, "shared", "*.html")
	if sh, _ := filepath.Glob(sharedGlob); len(sh) > 0 {
		files = append(files, sh...)
	}

	// finally add the view file itself
	viewPath := filepath.Join(v.TemplateDir, name+".html")
	if _, err := os.Stat(viewPath); err != nil {
		return nil, fmt.Errorf("view file not found: %s", viewPath)
	}
	files = append(files, viewPath)

	// parse template set
	tpl := template.New(filepath.Base(viewPath))
	parsed, err := tpl.ParseFiles(files...)
	if err != nil {
		return nil, fmt.Errorf("parse templates %v: %w", files, err)
	}

	v.mu.Lock()
	v.cache[name] = parsed
	v.mu.Unlock()
	return parsed, nil
}
