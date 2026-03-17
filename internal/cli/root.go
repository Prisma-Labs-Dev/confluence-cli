package cli

import "time"

// CLI defines the agent-first command surface for the confluence binary.
type CLI struct {
	URL     string        `name:"url" help:"Confluence base URL" env:"CONFLUENCE_URL"`
	Email   string        `name:"email" help:"Atlassian account email" env:"CONFLUENCE_EMAIL"`
	Token   string        `name:"token" help:"Atlassian API token" env:"CONFLUENCE_API_TOKEN"`
	Format  string        `name:"format" help:"Output format: json or plain" enum:"json,plain" default:"json"`
	Timeout time.Duration `name:"timeout" help:"HTTP timeout" default:"30s"`

	Spaces  SpacesCmd  `cmd:"" help:"Space discovery commands"`
	Pages   PagesCmd   `cmd:"" help:"Page discovery commands"`
	Auth    AuthCmd    `cmd:"" help:"Credential management commands"`
	Version VersionCmd `cmd:"" name:"version" help:"Print CLI version"`
}

// SpacesCmd groups space commands.
type SpacesCmd struct {
	List SpacesListCmd `cmd:"" help:"List spaces with compact summaries"`
}

// PagesCmd groups page commands.
type PagesCmd struct {
	List   PagesListCmd   `cmd:"" help:"List pages in a space"`
	Get    PagesGetCmd    `cmd:"" help:"Get a page by ID"`
	Tree   PagesTreeCmd   `cmd:"" help:"Traverse a bounded page tree"`
	Search PagesSearchCmd `cmd:"" help:"Search pages with safe query inputs"`
}

// AuthCmd groups credential commands.
type AuthCmd struct {
	Login AuthLoginCmd `cmd:"" help:"Store credentials for later non-interactive use"`
}
