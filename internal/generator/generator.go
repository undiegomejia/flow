package generator

import (
    "bytes"
    "fmt"
    "os"
    "path/filepath"
    "strings"
    "text/template"
)

// generateFile renders tmpl with data and writes it to dstPath. It will
// create directories if necessary and will not overwrite existing files
// unless overwrite is true.
func generateFile(tmplStr string, data interface{}, dstPath string, overwrite bool) error {
    if !overwrite {
        if _, err := os.Stat(dstPath); err == nil {
            return fmt.Errorf("file exists: %s", dstPath)
        }
    }
    dir := filepath.Dir(dstPath)
    if err := os.MkdirAll(dir, 0o755); err != nil {
        return err
    }
    t, err := template.New("tpl").Funcs(template.FuncMap{
        "ToLower": strings.ToLower,
    }).Parse(tmplStr)
    if err != nil {
        return err
    }
    var buf bytes.Buffer
    if err := t.Execute(&buf, data); err != nil {
        return err
    }
    return os.WriteFile(dstPath, buf.Bytes(), 0o644)
}

// GenerateController creates a controller file at the target project path.
// name should be the base controller name (eg. "users").
func GenerateController(projectRoot, name string) (string, error) {
    cname := strings.Title(name) + "Controller"
    dst := filepath.Join(projectRoot, "app", "controllers", name+"_controller.go")
    data := map[string]string{
        "Package":    "controllers",
        "Controller": cname,
        "Name":       name,
    }
    return dst, generateFile(controllerTmpl, data, dst, false)
}

// GenerateModel creates a simple model file under app/models.
func GenerateModel(projectRoot, name string) (string, error) {
    mname := strings.Title(name)
    dst := filepath.Join(projectRoot, "app", "models", strings.ToLower(name)+".go")
    data := map[string]string{"Package": "models", "Model": mname}
    return dst, generateFile(modelTmpl, data, dst, false)
}

// GenerateScaffold generates controller + model + basic views.
func GenerateScaffold(projectRoot, name string) ([]string, error) {
    var created []string
    cpath, err := GenerateController(projectRoot, name)
    if err != nil {
        return created, err
    }
    created = append(created, cpath)
    mpath, err := GenerateModel(projectRoot, name)
    if err != nil {
        return created, err
    }
    created = append(created, mpath)
    // views: index and show
    viewsDir := filepath.Join(projectRoot, "app", "views", name)
    _ = os.MkdirAll(viewsDir, 0o755)
    idxPath := filepath.Join(viewsDir, "index.html")
    showPath := filepath.Join(viewsDir, "show.html")
    _ = os.WriteFile(idxPath, []byte("<h1>Index for "+name+"</h1>"), 0o644)
    _ = os.WriteFile(showPath, []byte("<h1>Show " + name + "</h1>"), 0o644)
    created = append(created, idxPath, showPath)
    return created, nil
}
