// go-plm — Local Git-native PLM/PDM system for engineering documentation.
//
// A single binary that serves a complete PLM application:
//   - Embedded React frontend (Vite + Tailwind + TypeScript)
//   - JSON-RPC 2.0 API for all PLM operations
//   - SQLite FTS5 search index
//   - Git-native version history
//   - FSM-driven object lifecycles
//
// Usage:
//   plm init <name>         Create a new PLM project
//   plm [path]               Open a project (starts HTTP server + opens browser)
//   plm stats [path]         Show project statistics
//   plm rebuild [path]       Rebuild search index from files
package main

import (
	"context"
	"fmt"
	"os"
	"os/signal"
	"path/filepath"
	"strconv"

	"github.com/Homiakus/go-plm/internal/app/service"
)

func main() {
	if len(os.Args) < 2 {
		// Default: open current directory
		runServer(".")
		return
	}

	cmd := os.Args[1]

	switch cmd {
	case "init":
		runInit()
	case "stats":
		runStats()
	case "rebuild":
		runRebuild()
	case "help", "-h", "--help":
		printHelp()
	default:
		// Treat as project path
		path := cmd
		if _, err := os.Stat(filepath.Join(path, "project.md")); os.IsNotExist(err) {
			fmt.Fprintf(os.Stderr, "ERROR: No PLM project found at %q\n", path)
			fmt.Println("Run 'plm init <name>' to create a new project.")
			os.Exit(1)
		}
		runServer(path)
	}
}

func runServer(path string) {
	if _, err := os.Stat(filepath.Join(path, "project.md")); os.IsNotExist(err) {
		fmt.Fprintf(os.Stderr, "ERROR: No PLM project found at %q\n", path)
		fmt.Println("Run 'plm init <name>' to create a new project, or 'plm help' for usage.")
		os.Exit(1)
	}

	// Port: PLM_PORT env var or default 8470
	port := 8470
	if p := os.Getenv("PLM_PORT"); p != "" {
		if n, err := strconv.Atoi(p); err == nil && n > 0 && n < 65536 {
			port = n
		}
	}

	srv, err := NewServer(path)
	if err != nil {
		fmt.Fprintf(os.Stderr, "ERROR: %v\n", err)
		os.Exit(1)
	}
	defer srv.Shutdown()

	if err := srv.Start(port); err != nil {
		fmt.Fprintf(os.Stderr, "ERROR: %v\n", err)
		os.Exit(1)
	}

	// Graceful shutdown on SIGINT/SIGTERM (Ctrl+C)
	sigCh := make(chan os.Signal, 1)
	signal.Notify(sigCh, os.Interrupt)
	<-sigCh
	fmt.Println("\nShutting down...")
	srv.Shutdown()
}

func runInit() {
	if len(os.Args) < 3 {
		fmt.Println("Usage: plm init <project-name> [project-title]")
		fmt.Println()
		fmt.Println("Creates a new PLM project directory with:")
		fmt.Println("  project.md          — project configuration (YAML frontmatter)")
		fmt.Println("  config/lifecycle.md — lifecycle state machine definition")
		fmt.Println("  objects/            — engineering objects directory")
		fmt.Println("  .plm/index.db       — SQLite search index (rebuildable)")
		fmt.Println()
		fmt.Println("Examples:")
		fmt.Println("  plm init my-project")
		fmt.Println("  plm init a320 'Airbus A320 Retrofit'")
		os.Exit(1)
	}

	name := os.Args[2]
	title := name
	if len(os.Args) > 3 {
		title = os.Args[3]
	}

	cwd, _ := os.Getwd()
	root := filepath.Join(cwd, name)

	app, err := service.InitProject(root, name, title)
	if err != nil {
		fmt.Fprintf(os.Stderr, "ERROR: %v\n", err)
		os.Exit(1)
	}
	defer app.Close()

	fmt.Println()
	fmt.Printf("✅ Project created: %s\n", root)
	fmt.Printf("   Code:  %s\n", app.Config.Project.Code)
	fmt.Printf("   Title: %s\n", app.Config.Project.Title)
	fmt.Println()
	fmt.Println("Next:")
	fmt.Printf("  plm %s    # Open in browser\n", name)
	fmt.Println()
}

func runStats() {
	root := "."
	if len(os.Args) > 2 {
		root = os.Args[2]
	}

	app, err := service.Open(root)
	if err != nil {
		fmt.Fprintf(os.Stderr, "ERROR: %v\n", err)
		os.Exit(1)
	}
	defer app.Close()

	ctx := context.Background()
	objs, rels, arts, _ := app.Index.Stats(ctx)
	objects, _ := app.Qry.ListObjects(ctx)

	fmt.Println()
	fmt.Printf("═══ %s ═══\n", app.Config.Project.Title)
	fmt.Printf("Path: %s\n", root)
	fmt.Println()
	fmt.Printf("  Objects:   %d\n", objs)
	fmt.Printf("  Relations: %d\n", rels)
	fmt.Printf("  Artifacts: %d\n", arts)
	fmt.Println()

	if len(objects) > 0 {
		fmt.Println("  By class:")
		byClass := map[string]int{}
		for _, o := range objects {
			byClass[o.Class]++
		}
		for class, count := range byClass {
			fmt.Printf("    %-8s %d\n", class+":", count)
		}
	}
	fmt.Println()
}

func runRebuild() {
	root := "."
	if len(os.Args) > 2 {
		root = os.Args[2]
	}

	app, err := service.Open(root)
	if err != nil {
		fmt.Fprintf(os.Stderr, "ERROR: %v\n", err)
		os.Exit(1)
	}
	defer app.Close()

	if err := app.RebuildIndex(context.Background()); err != nil {
		fmt.Fprintf(os.Stderr, "ERROR: %v\n", err)
		os.Exit(1)
	}

	fmt.Println()
	fmt.Println("✅ Index rebuilt successfully.")
	fmt.Println()
}

func printHelp() {
	fmt.Print(`go-plm — Local Git-native PLM/PDM System

No server. No database. No cloud.
Just files. Just Git. Just Markdown.

USAGE:
  plm [path]               Open a project (starts server + opens browser)
  plm init <name> [title]  Create a new PLM project
  plm stats [path]         Show project statistics
  plm rebuild [path]       Rebuild search index from files
  plm help                 Show this help

EXAMPLES:
  plm init my-project
  plm .                    # Open current directory
  plm ~/projects/a320      # Open specific project
  plm stats .              # Show stats for current project

ARCHITECTURE:
  Terminal:  go-plm binary (Go 1.26, embedded frontend)
  UI:        React 19 + TypeScript + Vite + Tailwind
  API:       JSON-RPC 2.0 over HTTP (localhost:8470)
  Storage:   Markdown files + YAML frontmatter + SQLite FTS5
  VCS:       Git-native (go-git v5)
`)
}
