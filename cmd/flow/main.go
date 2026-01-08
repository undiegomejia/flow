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
