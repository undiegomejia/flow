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
	return ctx.RenderTemplate(tpl, filepath.Base(name)+".html", data)
}

func (v *ViewManager) loadTemplate(name string) (*template.Template, error) {
	v.mu.RLock()
	t, ok := v.cache[name]
	v.mu.RUnlock()
	if ok {
		return t, nil
	}

	// parse template from filesystem
	path := filepath.Join(v.TemplateDir, name+".html")
	tpl, err := template.ParseFiles(path)
	if err != nil {
		return nil, fmt.Errorf("parse template %s: %w", path, err)
	}

	v.mu.Lock()
	v.cache[name] = tpl
	v.mu.Unlock()
	return tpl, nil
}
