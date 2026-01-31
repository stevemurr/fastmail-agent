# fastmail-agent

A terminal UI and CLI tool for searching and exporting Fastmail emails. Designed for both interactive use and integration with AI agents/LLMs.

## Features

- **Interactive TUI**: Browse and search emails with a terminal interface
- **Agent-friendly CLI**: JSON output for programmatic access
- **LLM-optimized export**: Strips quotes and signatures for cleaner context
- **Attachment handling**: Download attachments to folders
- **Clipboard integration**: Copy threads directly to clipboard

## Installation

```bash
go install github.com/YOUR_USERNAME/fastmail-agent@latest
```

Or build from source:

```bash
git clone https://github.com/YOUR_USERNAME/fastmail-agent.git
cd fastmail-agent
go build -o fastmail-agent
```

## Configuration

Set your Fastmail API token via environment variable:

```bash
export FASTMAIL_API_TOKEN="fmu1-xxxxx"
```

Or create a config file at `~/.config/fastmail-tui/config.json`:

```json
{"api_token": "fmu1-xxxxx"}
```

To get an API token:
1. Go to Fastmail Settings > Privacy & Security > API Tokens
2. Create a new token with Mail access

## Usage

### Interactive TUI

```bash
fastmail-agent
```

- Type your search query and press Enter
- Use arrow keys to navigate threads
- Press Enter to view a thread
- Press `c` to copy thread to clipboard (LLM format)
- Press `a` to copy attachment info
- Press `f` to copy full thread with attachments
- Press `j` to export thread + attachments to folder
- Press `q` to go back/quit

### CLI Mode (for agents)

**Search for threads:**

```bash
fastmail-agent -q "from:alice@example.com invoice"
```

Returns JSON with thread IDs, subjects, dates, and previews.

**Fetch a specific thread:**

```bash
fastmail-agent -t 3        # LLM-optimized text format
fastmail-agent -t 3 -json  # JSON format
```

### Agent Workflow Example

```bash
# 1. Search for relevant emails
$ fastmail-agent -q "project update"

# 2. Pick a thread ID from the results and fetch full content
$ fastmail-agent -t 2
```

## Output Formats

**LLM-optimized text** (default for `-t`):
- Strips quoted replies to reduce redundancy
- Removes email signatures
- Clean, readable format

**JSON** (with `-json` flag):
- Structured data for programmatic use
- Includes all metadata

## License

MIT
