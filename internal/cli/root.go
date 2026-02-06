package cli

import "time"

type CLI struct {
	URL     string        `help:"Confluence base URL" env:"CONFLUENCE_URL"`
	Email   string        `help:"Atlassian account email" env:"CONFLUENCE_EMAIL"`
	Token   string        `help:"Atlassian API token" env:"CONFLUENCE_API_TOKEN"`
	Plain   bool          `help:"Plain text output instead of JSON" default:"false"`
	Color   bool          `help:"Enable color in plain output" default:"false"`
	Timeout time.Duration `help:"HTTP timeout" default:"30s"`

	Spaces SpacesCmd  `cmd:"" help:"Manage spaces"`
	Pages  PagesCmd   `cmd:"" help:"Manage pages"`
	Ver    VersionCmd `cmd:"" name:"version" help:"Print version"`
}
