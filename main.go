package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"os"
	"sort"
	"time"

	tea "github.com/charmbracelet/bubbletea"

	"github.com/stevemurr/fastmail-agent/config"
	"github.com/stevemurr/fastmail-agent/export"
	"github.com/stevemurr/fastmail-agent/jmap"
	"github.com/stevemurr/fastmail-agent/tui"
)

// ThreadInfo represents a thread in CLI query output
type ThreadInfo struct {
	ID         int      `json:"id"`
	Subject    string   `json:"subject"`
	From       string   `json:"from"`
	Date       string   `json:"date"`
	EmailCount int      `json:"email_count"`
	Preview    string   `json:"preview"`
	EmailIDs   []string `json:"email_ids"`
}

// QueryResult represents the full query response
type QueryResult struct {
	Query   string       `json:"query"`
	Count   int          `json:"count"`
	Threads []ThreadInfo `json:"threads"`
}

func main() {
	// Define CLI flags
	query := flag.String("q", "", "Search query - returns list of threads as JSON with IDs")
	threadID := flag.Int("t", 0, "Thread ID from query results - returns full thread content")
	outputJSON := flag.Bool("json", false, "Output thread content as JSON (only for -t, -q always outputs JSON)")

	flag.Usage = func() {
		fmt.Fprintf(os.Stderr, `fastmail-agent - Search and export Fastmail emails

USAGE:
  fastmail-agent                    Launch interactive TUI
  fastmail-agent -q "search terms"  Search and list threads (JSON output)
  fastmail-agent -t <id>            Fetch thread by ID (text output, LLM-optimized)
  fastmail-agent -t <id> -json      Fetch thread by ID (JSON output)

AGENT WORKFLOW:
  1. Search for threads:
     $ fastmail-agent -q "from:alice@example.com invoice"
     Returns JSON with thread IDs, subjects, dates, and previews

  2. Fetch specific thread content:
     $ fastmail-agent -t 3
     Returns the full email thread in LLM-optimized text format

FLAGS:
`)
		flag.PrintDefaults()
	}

	flag.Parse()

	// Load configuration
	cfg, err := config.Load()
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error loading config: %v\n", err)
		fmt.Fprintln(os.Stderr, "\nPlease set FASTMAIL_API_TOKEN environment variable")
		fmt.Fprintln(os.Stderr, "or create ~/.config/fastmail-tui/config.json with:")
		fmt.Fprintln(os.Stderr, `  {"api_token": "fmu1-xxxxx"}`)
		os.Exit(1)
	}

	// Create JMAP client and connect
	client := jmap.NewClient(cfg.APIToken)
	if err := client.Connect(); err != nil {
		fmt.Fprintf(os.Stderr, "Error connecting to Fastmail: %v\n", err)
		os.Exit(1)
	}

	// CLI mode: query for threads
	if *query != "" {
		runQuery(client, *query)
		return
	}

	// CLI mode: fetch specific thread
	if *threadID > 0 {
		runFetchThread(client, *threadID, *outputJSON)
		return
	}

	// Interactive TUI mode
	p := tea.NewProgram(
		tui.New(client),
		tea.WithAltScreen(),
	)

	if _, err := p.Run(); err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		os.Exit(1)
	}
}

// runQuery searches for emails and outputs grouped threads
func runQuery(client *jmap.Client, query string) {
	emails, err := client.SearchEmails(query, 50)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error searching: %v\n", err)
		os.Exit(1)
	}

	if len(emails) == 0 {
		result := QueryResult{
			Query:   query,
			Count:   0,
			Threads: []ThreadInfo{},
		}
		outputJSONResult(result)
		return
	}

	// Group emails by subject (same logic as TUI)
	threads := groupEmailsBySubject(emails)

	// Build result
	result := QueryResult{
		Query:   query,
		Count:   len(threads),
		Threads: threads,
	}

	outputJSONResult(result)
}

// runFetchThread fetches and outputs a specific thread by its query result ID
func runFetchThread(client *jmap.Client, threadID int, asJSON bool) {
	// We need to re-run a broad search and find the thread by index
	// This is a limitation - we could cache results, but for agent use
	// the typical workflow is: query -> pick thread -> fetch thread
	// all in quick succession

	// Read thread data from stdin if piped, otherwise error
	// For now, require the user to pass the email IDs directly via a state file
	// or re-query

	// Actually, let's store the last query result in a temp file
	stateFile := getStateFilePath()

	data, err := os.ReadFile(stateFile)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error: No previous query results found.\n")
		fmt.Fprintf(os.Stderr, "Run a query first with: fastmail-agent -q \"search terms\"\n")
		os.Exit(1)
	}

	var lastResult QueryResult
	if err := json.Unmarshal(data, &lastResult); err != nil {
		fmt.Fprintf(os.Stderr, "Error reading state: %v\n", err)
		os.Exit(1)
	}

	// Find the thread by ID (1-indexed for user friendliness)
	if threadID < 1 || threadID > len(lastResult.Threads) {
		fmt.Fprintf(os.Stderr, "Error: Thread ID %d not found. Valid range: 1-%d\n", threadID, len(lastResult.Threads))
		os.Exit(1)
	}

	thread := lastResult.Threads[threadID-1]

	// Fetch full email content
	emails, err := client.GetEmails(thread.EmailIDs)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error fetching emails: %v\n", err)
		os.Exit(1)
	}

	// Sort by date (oldest first for reading)
	sort.Slice(emails, func(i, j int) bool {
		ti, _ := time.Parse(time.RFC3339, emails[i].ReceivedAt)
		tj, _ := time.Parse(time.RFC3339, emails[j].ReceivedAt)
		return ti.Before(tj)
	})

	if asJSON {
		outputThreadJSON(emails, thread.Subject)
	} else {
		// Output in LLM-optimized text format
		opts := export.DefaultLLMOptions()
		fmt.Print(export.FormatThreadForLLM(emails, opts))
	}
}

// outputThreadJSON outputs thread content as JSON
func outputThreadJSON(emails []jmap.Email, subject string) {
	type EmailContent struct {
		From    string `json:"from"`
		To      string `json:"to"`
		CC      string `json:"cc,omitempty"`
		Date    string `json:"date"`
		Subject string `json:"subject"`
		Body    string `json:"body"`
	}

	type ThreadContent struct {
		Subject string         `json:"subject"`
		Count   int            `json:"count"`
		Emails  []EmailContent `json:"emails"`
	}

	result := ThreadContent{
		Subject: subject,
		Count:   len(emails),
		Emails:  make([]EmailContent, len(emails)),
	}

	opts := export.DefaultLLMOptions()
	for i, email := range emails {
		body := email.GetBodyText()
		// Clean body same as LLM export
		body = export.CleanBody(body, opts)

		cc := ""
		if len(email.CC) > 0 {
			cc = formatEmailsOnly(email.CC)
		}

		result.Emails[i] = EmailContent{
			From:    formatEmailsOnly(email.From),
			To:      formatEmailsOnly(email.To),
			CC:      cc,
			Date:    formatDate(email.ReceivedAt),
			Subject: email.Subject,
			Body:    body,
		}
	}

	enc := json.NewEncoder(os.Stdout)
	enc.SetIndent("", "  ")
	enc.Encode(result)
}

// groupEmailsBySubject groups emails by normalized subject
func groupEmailsBySubject(emails []jmap.Email) []ThreadInfo {
	threads := tui.GroupEmailsBySubject(emails)

	result := make([]ThreadInfo, len(threads))
	for i, t := range threads {
		emailIDs := make([]string, len(t.Emails))
		for j, e := range t.Emails {
			emailIDs[j] = e.ID
		}

		from := t.From
		if from == "" && len(t.Emails) > 0 && len(t.Emails[0].From) > 0 {
			from = t.Emails[0].From[0].Email
		}

		result[i] = ThreadInfo{
			ID:         i + 1, // 1-indexed for user friendliness
			Subject:    t.Subject,
			From:       from,
			Date:       t.Date.Format("2006-01-02 15:04"),
			EmailCount: t.EmailCount,
			Preview:    truncate(t.Preview, 100),
			EmailIDs:   emailIDs,
		}
	}

	return result
}

// outputJSONResult outputs the query result and saves state
func outputJSONResult(result QueryResult) {
	// Save state for subsequent -t calls
	stateFile := getStateFilePath()
	data, _ := json.Marshal(result)
	os.WriteFile(stateFile, data, 0600)

	// Output to stdout
	enc := json.NewEncoder(os.Stdout)
	enc.SetIndent("", "  ")
	enc.Encode(result)
}

// getStateFilePath returns the path to the state file
func getStateFilePath() string {
	cacheDir, err := os.UserCacheDir()
	if err != nil {
		cacheDir = os.TempDir()
	}
	stateDir := cacheDir + "/fastmail-agent"
	os.MkdirAll(stateDir, 0700)
	return stateDir + "/last_query.json"
}

// truncate truncates a string to max length
func truncate(s string, max int) string {
	if len(s) <= max {
		return s
	}
	return s[:max-3] + "..."
}

// formatEmailsOnly formats addresses as just email addresses
func formatEmailsOnly(addrs []jmap.EmailAddress) string {
	if len(addrs) == 0 {
		return ""
	}
	emails := make([]string, len(addrs))
	for i, addr := range addrs {
		emails[i] = addr.Email
	}
	result := ""
	for i, e := range emails {
		if i > 0 {
			result += ", "
		}
		result += e
	}
	return result
}

// formatDate formats a JMAP date string
func formatDate(dateStr string) string {
	t, err := time.Parse(time.RFC3339, dateStr)
	if err != nil {
		return dateStr
	}
	return t.Local().Format("2006-01-02 3:04 PM")
}
