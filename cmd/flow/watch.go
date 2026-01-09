package main

import (
	"context"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"sync"
	"time"

	fsnotify "github.com/fsnotify/fsnotify"
)

// WatchAndRun watches the given paths and runs the provided command (cmdArgs)
// as a child process. On file changes it restarts the child. It returns when
// the parent context is cancelled.
func WatchAndRun(ctx context.Context, watchPaths []string, cmdArgs []string) error {
	w, err := fsnotify.NewWatcher()
	if err != nil {
		return err
	}
	defer w.Close()

	addPaths := func(paths []string) error {
		for _, p := range paths {
			p = strings.TrimSpace(p)
			if p == "" {
				continue
			}
			// walk and add dirs
			_ = filepath.Walk(p, func(path string, info os.FileInfo, err error) error {
				if err != nil {
					return nil
				}
				if !info.IsDir() {
					return nil
				}
				// ignore .git, vendor, node_modules
			base := filepath.Base(path)
			if base == ".git" || base == "vendor" || base == "node_modules" {
				return filepath.SkipDir
			}
			_ = w.Add(path)
			return nil
		})
		}
		return nil
	}

	if err := addPaths(watchPaths); err != nil {
		return err
	}

	// child process management
	var mu sync.Mutex
	var child *exec.Cmd
	startChild := func() error {
		mu.Lock()
		defer mu.Unlock()
		if child != nil && child.Process != nil {
			// already running
			return nil
		}
		cmd := exec.CommandContext(ctx, "go", append([]string{"run", "./cmd/flow"}, cmdArgs...)...)
		cmd.Stdout = os.Stdout
		cmd.Stderr = os.Stderr
		cmd.Stdin = os.Stdin
		if err := cmd.Start(); err != nil {
			return err
		}
		child = cmd
		go func() {
			_ = cmd.Wait()
			mu.Lock()
			child = nil
			mu.Unlock()
		}()
		fmt.Printf("[watch] started child pid=%d\n", cmd.Process.Pid)
		return nil
	}
	stopChild := func() error {
		mu.Lock()
		defer mu.Unlock()
		if child == nil || child.Process == nil {
			return nil
		}
		_ = child.Process.Kill()
		child = nil
		return nil
	}

	// start initial child
	if err := startChild(); err != nil {
		return err
	}

	debounce := time.NewTimer(0)
	if !debounce.Stop() {
		<-debounce.C
	}
	trigger := false

	for {
		select {
		case <-ctx.Done():
			_ = stopChild()
			return nil
		case ev, ok := <-w.Events:
			if !ok {
				return nil
			}
			// only consider write/create/remove/rename
			if ev.Op&(fsnotify.Write|fsnotify.Create|fsnotify.Rename|fsnotify.Remove) == 0 {
				continue
			}
			// ignore editor temp files
			if strings.HasSuffix(ev.Name, "~") || strings.HasSuffix(ev.Name, ".swp") {
				continue
			}
			fmt.Printf("[watch] change detected: %s\n", ev.Name)
			trigger = true
			// reset debounce
			debounce.Reset(300 * time.Millisecond)
		case err, ok := <-w.Errors:
			if !ok {
				return nil
			}
			fmt.Fprintln(os.Stderr, "watch error:", err)
		case <-debounce.C:
			if trigger {
				trigger = false
				// restart child
				_ = stopChild()
				fmt.Println("[watch] rebuilding and restarting...")
				if err := startChild(); err != nil {
					fmt.Fprintln(os.Stderr, "failed to restart child:", err)
				}
			}
		}
	}
}
