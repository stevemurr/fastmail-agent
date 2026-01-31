# CLAUDE.md

This file provides guidance to Claude Code (claude.ai/code) when working with this codebase.

## Project Overview

fastmail-agent is a Go application that provides both a TUI (Terminal User Interface) and CLI for searching and exporting Fastmail emails via the JMAP protocol. It's designed for interactive use and integration with AI agents.

## Build & Run Commands

```bash
# Build the binary
go build -o fastmail-agent

# Run in TUI mode
./fastmail-agent

# Run in CLI mode
./fastmail-agent -q "search query"   # Search and list threads (JSON)
./fastmail-agent -t 3                # Fetch thread by ID (text)
./fastmail-agent -t 3 -json          # Fetch thread by ID (JSON)
```

## Architecture

```
.
├── main.go           # Entry point, CLI flag handling, query/thread commands
├── config/
│   └── config.go     # Configuration loading (env var or config file)
├── jmap/
│   ├── client.go     # JMAP client for Fastmail API
│   ├── email.go      # Email search and fetch operations
│   └── types.go      # JMAP data types (Email, Session, etc.)
├── tui/
│   ├── app.go        # Main TUI model and update logic
│   ├── keys.go       # Keybinding definitions
│   ├── search.go     # Search input component
│   ├── styles.go     # Lipgloss styling
│   ├── threadlist.go # Thread list component
│   └── threadview.go # Thread detail view component
└── export/
    ├── export.go     # Export formatting and clipboard
    ├── html.go       # HTML to text conversion
    └── quotes.go     # Quote/signature stripping
```

## Key Design Decisions

- **Email grouping**: Emails are grouped by normalized subject (stripping Re:/Fwd: prefixes) rather than JMAP thread IDs for more intuitive conversation threading
- **State persistence**: CLI mode stores last query results in a cache file to enable `fastmail-agent -t <id>` without re-querying
- **LLM optimization**: Default export strips quoted content and signatures to reduce token usage

## Dependencies

- `github.com/charmbracelet/bubbletea` - TUI framework
- `github.com/charmbracelet/bubbles` - TUI components
- `github.com/charmbracelet/lipgloss` - Styling
- `github.com/atotto/clipboard` - Clipboard access
- `golang.org/x/net` - HTML parsing
