package main

import (
	"encoding/json"
	"fmt"
	"io"
	"os"
	"strings"

	"atheon/core"
)

func main() {
	args := os.Args[1:]

	jsonOutput := len(args) > 0 && args[0] == "--json"
	if jsonOutput {
		args = args[1:]
	}

	cats, args := parseCategories(args)
	core.SetActiveCategories(cats)

	if len(args) == 0 {
		printHelp()
		return
	}

	switch args[0] {
	case "update":
		fmt.Println("downloading patterns bundle...")
		if err := core.DownloadBundle(); err != nil {
			fmt.Fprintln(os.Stderr, "error:", err)
			os.Exit(1)
		}
		fmt.Println("patterns updated.")

	case "list":
		cmdList(args[1:])

	case "--help", "help", "-h":
		printHelp()

	case "--env":
		findings := core.ScanEnv()
		printFindings(findings, nil, jsonOutput)
		if len(findings) > 0 {
			os.Exit(1)
		}

	case "-", "--stdin":
		data, err := io.ReadAll(os.Stdin)
		if err != nil {
			fmt.Fprintln(os.Stderr, "error: reading stdin:", err)
			os.Exit(1)
		}
		findings := core.ScanString(string(data), "stdin")
		printFindings(findings, nil, jsonOutput)
		if len(findings) > 0 {
			os.Exit(1)
		}

	case "--file":
		if len(args) < 2 {
			fmt.Fprintln(os.Stderr, "error: --file requires a path")
			os.Exit(1)
		}
		findings, stats, err := core.ScanFile(args[1])
		if err != nil {
			fmt.Fprintln(os.Stderr, "error:", err)
			os.Exit(1)
		}
		printFindings(findings, stats, jsonOutput)
		if len(findings) > 0 {
			os.Exit(1)
		}

	default:
		path := args[0]
		info, err := os.Stat(path)
		if err != nil {
			fmt.Fprintln(os.Stderr, "error: path not found:", path)
			os.Exit(1)
		}
		var findings []core.Finding
		var stats *core.Stats
		if info.IsDir() {
			findings, stats, err = core.ScanDir(path)
		} else {
			findings, stats, err = core.ScanFile(path)
		}
		if err != nil {
			fmt.Fprintln(os.Stderr, "error:", err)
			os.Exit(1)
		}
		printFindings(findings, stats, jsonOutput)
		if len(findings) > 0 {
			os.Exit(1)
		}
	}
}

func parseCategories(args []string) ([]string, []string) {
	var cats []string
	var rest []string
	all := false
	for _, a := range args {
		switch {
		case strings.HasPrefix(a, "--categories="):
			val := strings.TrimPrefix(a, "--categories=")
			for _, c := range strings.Split(val, ",") {
				if c = strings.TrimSpace(c); c != "" {
					cats = append(cats, c)
				}
			}
		case a == "--all":
			all = true
		default:
			rest = append(rest, a)
		}
	}
	if all {
		return nil, rest
	}
	return cats, rest
}

func printFindings(findings []core.Finding, stats *core.Stats, jsonOutput bool) {
	if jsonOutput {
		printJSONFindings(findings)
		return
	}
	if len(findings) == 0 {
		fmt.Println("no findings.")
	} else {
		for _, f := range findings {
			loc := f.File
			if f.Line > 0 {
				loc = fmt.Sprintf("%s:%d", f.File, f.Line)
			}
			fmt.Printf("%s  %s\n", f.Pattern, loc)
			if f.Content != "" {
				fmt.Println(" ", redact(f.Content))
			}
		}
		fmt.Printf("\n%d finding(s)\n", len(findings))
	}
	if stats != nil && stats.Files > 0 {
		fmt.Printf("scanned %d file(s)  %s  %dms\n",
			stats.Files, formatBytes(stats.Bytes), stats.ElapsedMs)
	}
}

func printJSONFindings(findings []core.Finding) {
	items := make([]map[string]any, 0, len(findings))
	for _, f := range findings {
		items = append(items, map[string]any{"pattern": f.Pattern, "file": f.File, "line": f.Line, "match": f.Content})
	}
	if err := json.NewEncoder(os.Stdout).Encode(items); err != nil {
		fmt.Fprintln(os.Stderr, "error:", err)
	}
}

func cmdList(args []string) {
	if len(args) > 0 && args[0] == "categories" {
		for _, c := range core.Categories() {
			fmt.Println(c)
		}
		return
	}

	// Check if category filter is specified
	var categoryFilter string
	if len(args) > 0 && strings.HasPrefix(args[0], "--category=") {
		categoryFilter = strings.TrimPrefix(args[0], "--category=")
	}

	patterns := core.All()
	if categoryFilter != "" {
		filtered := make([]core.Pattern, 0)
		for _, p := range patterns {
			if p.Category() == categoryFilter {
				filtered = append(filtered, p)
			}
		}
		patterns = filtered
	}

	for _, p := range patterns {
		fmt.Printf("%s [%s]\n", p.Name(), p.Category())
	}
	fmt.Printf("\n%d pattern(s) loaded\n", len(patterns))
}

func printHelp() {
	fmt.Print(`atheon - pattern matching engine

usage:
  atheon <path>                      scan a directory
  atheon --file <path>               scan a single file
  atheon --env                       scan environment variables
  atheon --json <path>               print findings as JSON
  atheon --categories=<c1,c2> <path> scan specific categories
  atheon --all <path>                scan all categories
  atheon list                        list loaded patterns
  atheon list categories             list available categories
  atheon update                      download latest patterns bundle
  atheon --help                      show this message
`)
}

func redact(s string) string {
	if len(s) <= 8 {
		return "***"
	}
	return s[:4] + "****" + s[len(s)-4:]
}

func formatBytes(b int64) string {
	if b >= 1<<20 {
		return fmt.Sprintf("%.1f MB", float64(b)/(1<<20))
	}
	if b >= 1<<10 {
		return fmt.Sprintf("%.1f KB", float64(b)/(1<<10))
	}
	return fmt.Sprintf("%d B", b)
}
