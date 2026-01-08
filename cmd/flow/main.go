// Command-line interface for the Flow framework.
//
// This file implements a small, user-facing CLI using cobra. It provides
// a `serve` command to run an App and a `version` command. The CLI is
// intentionally minimal but fully functional so it can be extended with
// generators and other developer tools later.
package main

import (
    "context"
    "fmt"
    "os"
    "os/signal"
    "syscall"

    "github.com/spf13/cobra"

    flowpkg "github.com/dministrator/flow/pkg/flow"
    routerpkg "github.com/dministrator/flow/internal/router"
    "net/http"
    "database/sql"

    gen "github.com/dministrator/flow/internal/generator"
    mig "github.com/dministrator/flow/internal/migrations"
)

const version = "0.1.0"

func main() {
    if err := rootCmd.Execute(); err != nil {
        fmt.Fprintln(os.Stderr, err)
        os.Exit(1)
    }
}

var rootCmd = &cobra.Command{
    Use:   "flow",
    Short: "Flow — an opinionated Go MVC web framework (CLI)",
    Long:  "Flow CLI: run, generate and manage Flow web applications.",
}

func init() {
    rootCmd.AddCommand(serveCmd)
    rootCmd.AddCommand(versionCmd)
    rootCmd.AddCommand(dbCmd)
    rootCmd.AddCommand(generateCmd)
}

var serveAddr string

var serveCmd = &cobra.Command{
    Use:   "serve",
    Short: "Start the development server",
    RunE: func(cmd *cobra.Command, args []string) error {
        app := flowpkg.New("flow", flowpkg.WithAddr(serveAddr))

        // small demo router: exposes a health endpoint and root index
        r := routerpkg.New()
        r.Get("/", func(w http.ResponseWriter, r *http.Request) {
            w.Header().Set("Content-Type", "text/plain; charset=utf-8")
            w.WriteHeader(200)
            _, _ = w.Write([]byte("Flow app running"))
        })
        r.Get("/health", func(w http.ResponseWriter, r *http.Request) {
            w.Header().Set("Content-Type", "application/json; charset=utf-8")
            w.WriteHeader(200)
            _, _ = w.Write([]byte("{\"status\":\"ok\"}"))
        })

        app.SetRouter(r)

        // start and block until signal
        if err := app.Start(); err != nil {
            return err
        }

        // Wait for shutdown signal
        sig := make(chan os.Signal, 1)
        signal.Notify(sig, syscall.SIGINT, syscall.SIGTERM)
        <-sig

        return app.Shutdown(context.Background())
    },
}

func init() {
    serveCmd.Flags().StringVar(&serveAddr, "addr", ":3000", "listen address for the server")
}

var versionCmd = &cobra.Command{
    Use:   "version",
    Short: "Print the CLI version",
    Run: func(cmd *cobra.Command, args []string) {
        fmt.Println("flow", version)
    },
}

var dbCmd = &cobra.Command{
    Use:   "db",
    Short: "Database tasks (migrate, rollback)",
}

var dbDir string
var dbDriver string
var dbDSN string

var dbMigrateCmd = &cobra.Command{
    Use:   "migrate",
    Short: "Apply all pending migrations in a directory",
    RunE: func(cmd *cobra.Command, args []string) error {
        if dbDriver == "" || dbDSN == "" {
            return fmt.Errorf("driver and dsn flags are required to run migrations")
        }
        db, err := sql.Open(dbDriver, dbDSN)
        if err != nil {
            return err
        }
        defer db.Close()
        runner := &mig.MigrationRunner{}
        return runner.ApplyAll(dbDir, db)
    },
}

var dbRollbackCmd = &cobra.Command{
    Use:   "rollback",
    Short: "Rollback the most recent migration",
    RunE: func(cmd *cobra.Command, args []string) error {
        if dbDriver == "" || dbDSN == "" {
            return fmt.Errorf("driver and dsn flags are required to rollback migrations")
        }
        db, err := sql.Open(dbDriver, dbDSN)
        if err != nil {
            return err
        }
        defer db.Close()
        runner := &mig.MigrationRunner{}
        return runner.RollbackLast(dbDir, db)
    },
}

func init() {
    dbCmd.AddCommand(dbMigrateCmd)
    dbCmd.AddCommand(dbRollbackCmd)
    dbCmd.PersistentFlags().StringVar(&dbDir, "dir", "db/migrate", "migrations directory")
    dbCmd.PersistentFlags().StringVar(&dbDriver, "driver", "", "database driver (eg. postgres, mysql)")
    dbCmd.PersistentFlags().StringVar(&dbDSN, "dsn", "", "database DSN")
}

var generateCmd = &cobra.Command{
    Use:   "generate",
    Short: "Code generators (controller, model, scaffold)",
}

var generateTarget string

var genControllerCmd = &cobra.Command{
    Use:   "controller [name]",
    Short: "Generate a controller",
    Args:  cobra.ExactArgs(1),
    RunE: func(cmd *cobra.Command, args []string) error {
        name := args[0]
        root := generateTarget
        if root == "" {
            var err error
            root, err = os.Getwd()
            if err != nil {
                return err
            }
        }
        dst, err := gen.GenerateController(root, name)
        if err != nil {
            return err
        }
        fmt.Println("created", dst)
        return nil
    },
}

var genModelCmd = &cobra.Command{
    Use:   "model [name] [fields...]",
    Short: "Generate a model (optionally with fields, e.g. title:string published_at:datetime)",
    Args:  cobra.MinimumNArgs(1),
    RunE: func(cmd *cobra.Command, args []string) error {
        name := args[0]
        fields := []string{}
        if len(args) > 1 {
            fields = args[1:]
        }
        root := generateTarget
        if root == "" {
            var err error
            root, err = os.Getwd()
            if err != nil {
                return err
            }
        }
        dst, err := gen.GenerateModel(root, name, fields...)
        if err != nil {
            return err
        }
        fmt.Println("created", dst)
        return nil
    },
}

var genScaffoldCmd = &cobra.Command{
    Use:   "scaffold [name] [fields...]",
    Short: "Generate scaffold (controller, model, views) optionally with fields",
    Args:  cobra.MinimumNArgs(1),
    RunE: func(cmd *cobra.Command, args []string) error {
        name := args[0]
        fields := []string{}
        if len(args) > 1 {
            fields = args[1:]
        }
        root := generateTarget
        if root == "" {
            var err error
            root, err = os.Getwd()
            if err != nil {
                return err
            }
        }
        created, err := gen.GenerateScaffold(root, name, fields...)
        if err != nil {
            return err
        }
        for _, c := range created {
            fmt.Println("created", c)
        }
        return nil
    },
}

func init() {
    generateCmd.AddCommand(genControllerCmd)
    generateCmd.AddCommand(genModelCmd)
    generateCmd.AddCommand(genScaffoldCmd)
    generateCmd.PersistentFlags().StringVar(&generateTarget, "target", "", "target project root (defaults to cwd)")
}
