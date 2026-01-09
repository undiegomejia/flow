package generator

import (
    "bytes"
    "fmt"
    "os"
    "path/filepath"
    "strings"
    "text/template"
    "time"
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
func GenerateModel(projectRoot, name string, fields ...string) (string, error) {
    mname := strings.Title(name)
    dst := filepath.Join(projectRoot, "app", "models", strings.ToLower(name)+".go")

    // parse fields and build struct lines and migration columns using FieldSpec
    var fieldsCodeLines []string
    var columnsLines []string
    needTime := false
    specs, err := ParseFields(fields)
    if err != nil {
        return dst, err
    }
    for _, fs := range specs {
        if strings.Contains(fs.GoType, "time.Time") || strings.Contains(fs.GoType, "*time.Time") {
            needTime = true
        }
        // struct tag: bun and json; use omitempty for nullable
        jsonTag := fs.Name
        if fs.Nullable {
            jsonTag = jsonTag + ",omitempty"
        }
        tag := fmt.Sprintf("`bun:\"%s\" json:\"%s\"`", fs.Name, jsonTag)
        fieldsCodeLines = append(fieldsCodeLines, fmt.Sprintf("    %s %s %s", fs.GoName, fs.GoType, tag))

        // column SQL line (skip id/created/updated handled separately)
        notnull := ""
        if !fs.Nullable {
            notnull = " NOT NULL"
        }
        colLine := fmt.Sprintf("    %s %s%s", fs.Name, fs.SQLType, notnull)
        if fs.Default != nil {
            colLine = colLine + " DEFAULT " + *fs.Default
        }
        if fs.Unique {
            colLine = colLine + " UNIQUE"
        }
        columnsLines = append(columnsLines, colLine)
    }

    fieldsCode := ""
    if len(fieldsCodeLines) > 0 {
        fieldsCode = strings.Join(fieldsCodeLines, "\n") + "\n"
    }
    cols := ""
    if len(columnsLines) > 0 {
        cols = ",\n" + strings.Join(columnsLines, ",\n")
    }

    extraImports := ""
    if needTime {
        extraImports = "\n    \"time\""
    }

    data := map[string]string{
        "Package":    "models",
        "Model":      mname,
        "FieldsCode": fieldsCode,
        "Columns":    cols,
        "ExtraImports": extraImports,
    }

    return dst, generateFile(bunModelTmpl, data, dst, false)
}

// GenerateScaffold generates controller + model + basic views.
func GenerateScaffold(projectRoot, name string, fields ...string) ([]string, error) {
    var created []string
    cpath, err := GenerateController(projectRoot, name)
    if err != nil {
        return created, err
    }
    created = append(created, cpath)
    mpath, err := GenerateModel(projectRoot, name, fields...)
    if err != nil {
        return created, err
    }
    created = append(created, mpath)
    // views: index, show, new, edit
    viewsDir := filepath.Join(projectRoot, "app", "views", name)
    if err := os.MkdirAll(viewsDir, 0o755); err != nil {
        return created, err
    }
    idxPath := filepath.Join(viewsDir, "index.html")
    showPath := filepath.Join(viewsDir, "show.html")
    newPath := filepath.Join(viewsDir, "new.html")
    editPath := filepath.Join(viewsDir, "edit.html")
    // write using templates
    _ = os.WriteFile(idxPath, []byte(viewIndexTmpl), 0o644)
    _ = os.WriteFile(showPath, []byte(viewShowTmpl), 0o644)
    _ = os.WriteFile(newPath, []byte(viewNewTmpl), 0o644)
    _ = os.WriteFile(editPath, []byte(viewEditTmpl), 0o644)
    created = append(created, idxPath, showPath, newPath, editPath)

    // migrations: create timestamped up/down SQL files under db/migrate
    migDir := filepath.Join(projectRoot, "db", "migrate")
    if err := os.MkdirAll(migDir, 0o755); err != nil {
        return created, err
    }
    ts := TimestampNow()
    table := TableName(name)
    upName := fmt.Sprintf("%s_create_%s.up.sql", ts, table)
    downName := fmt.Sprintf("%s_create_%s.down.sql", ts, table)
    upPath := filepath.Join(migDir, upName)
    downPath := filepath.Join(migDir, downName)
    // compute columns SQL for migration based on fields (use ParseFields)
    var columnsLines []string
    specs2, err := ParseFields(fields)
    if err != nil {
        return created, err
    }
    for _, fs := range specs2 {
        notnull := ""
        if !fs.Nullable {
            notnull = " NOT NULL"
        }
        col := fmt.Sprintf("    %s %s%s", fs.Name, fs.SQLType, notnull)
        if fs.Default != nil {
            col = col + " DEFAULT " + *fs.Default
        }
        if fs.Unique {
            col = col + " UNIQUE"
        }
        columnsLines = append(columnsLines, col)
    }
    cols := ""
    if len(columnsLines) > 0 {
        cols = ",\n" + strings.Join(columnsLines, ",\n")
    }

    // build extras: indexes (CREATE INDEX) and corresponding DROP INDEX for down
    var extrasUpLines []string
    var extrasDownLines []string
    for _, fs := range specs2 {
        if fs.Index {
            idxName := fmt.Sprintf("idx_%s_%s", table, fs.Name)
            extrasUpLines = append(extrasUpLines, fmt.Sprintf("CREATE INDEX IF NOT EXISTS %s ON %s(%s);", idxName, table, fs.Name))
            extrasDownLines = append(extrasDownLines, fmt.Sprintf("DROP INDEX IF EXISTS %s;", idxName))
        }
    }
    extrasUp := ""
    if len(extrasUpLines) > 0 {
        extrasUp = strings.Join(extrasUpLines, "\n") + "\n"
    }
    extrasDown := ""
    if len(extrasDownLines) > 0 {
        extrasDown = strings.Join(extrasDownLines, "\n") + "\n"
    }

    // render migration templates (include extras for indexes)
    upData := map[string]string{"Timestamp": ts, "Table": table, "Columns": cols, "ExtrasUp": extrasUp}
    downData := map[string]string{"Timestamp": ts, "Table": table, "ExtrasDown": extrasDown}
    if err := generateFile(migrationUpTmpl, upData, upPath, false); err != nil {
        return created, err
    }
    if err := generateFile(migrationDownTmpl, downData, downPath, false); err != nil {
        return created, err
    }
    created = append(created, upPath, downPath)

    // small delay to avoid duplicate timestamps when called rapidly
    time.Sleep(1 * time.Second)
    return created, nil
}
