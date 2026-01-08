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

    // parse fields and build struct lines and migration columns
    var fieldsCodeLines []string
    var columnsLines []string
    needTime := false
    for _, f := range fields {
        // expected format: "name:type" (e.g. title:string)
        parts := strings.SplitN(f, ":", 2)
        fname := parts[0]
        ftype := "string"
        if len(parts) == 2 && parts[1] != "" {
            ftype = parts[1]
        }
        // Go field name (TitleCase)
        goName := strings.Title(fname)

        // Map types
        var goType string
        var sqlType string
        var notnull string
        switch strings.ToLower(ftype) {
        case "string", "text":
            goType = "string"
            sqlType = "TEXT"
            notnull = " NOT NULL"
        case "int", "integer":
            goType = "int"
            sqlType = "INTEGER"
        case "int64":
            goType = "int64"
            sqlType = "INTEGER"
        case "bool", "boolean":
            goType = "bool"
            sqlType = "BOOLEAN"
        case "float", "float64":
            goType = "float64"
            sqlType = "REAL"
        case "datetime", "time":
            goType = "time.Time"
            sqlType = "DATETIME"
            needTime = true
        default:
            // default to string
            goType = "string"
            sqlType = "TEXT"
        }

        // struct tag: bun and json
        tag := fmt.Sprintf("`bun:\"%s\" json:\"%s\"`", fname, fname)
        fieldsCodeLines = append(fieldsCodeLines, fmt.Sprintf("    %s %s %s", goName, goType, tag))

        // column SQL line (skip id/created/updated handled separately)
        colLine := fmt.Sprintf("    %s %s%s", fname, sqlType, notnull)
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
    // compute columns SQL for migration based on fields
    var columnsLines []string
    for _, f := range fields {
        parts := strings.SplitN(f, ":", 2)
        fname := parts[0]
        ftype := "string"
        if len(parts) == 2 && parts[1] != "" {
            ftype = parts[1]
        }
        var sqlType string
        var notnull string
        switch strings.ToLower(ftype) {
        case "string", "text":
            sqlType = "TEXT"
            notnull = " NOT NULL"
        case "int", "integer", "int64":
            sqlType = "INTEGER"
        case "bool", "boolean":
            sqlType = "BOOLEAN"
        case "float", "float64":
            sqlType = "REAL"
        case "datetime", "time":
            sqlType = "DATETIME"
        default:
            sqlType = "TEXT"
        }
        columnsLines = append(columnsLines, fmt.Sprintf("    %s %s%s", fname, sqlType, notnull))
    }
    cols := ""
    if len(columnsLines) > 0 {
        cols = ",\n" + strings.Join(columnsLines, ",\n")
    }

    // render migration templates
    if err := generateFile(migrationUpTmpl, map[string]string{"Timestamp": ts, "Table": table, "Columns": cols}, upPath, false); err != nil {
        return created, err
    }
    if err := generateFile(migrationDownTmpl, map[string]string{"Timestamp": ts, "Table": table}, downPath, false); err != nil {
        return created, err
    }
    created = append(created, upPath, downPath)

    // small delay to avoid duplicate timestamps when called rapidly
    time.Sleep(1 * time.Second)
    return created, nil
}
