package cli

import confluence "github.com/Prisma-Labs-Dev/confluence-cli"

type PagesSearchCmd struct {
	Query     string `help:"Search text to match in page content or titles"`
	CQL       string `help:"Raw CQL expression for advanced search"`
	TitleOnly bool   `help:"Restrict matching to page titles (query mode only)"`
	SpaceID   string `help:"Optional space ID filter (query mode only)"`
	SpaceKey  string `help:"Optional space key filter such as SC or TNLTA (query mode only)"`
	Limit     int    `help:"Maximum number of results per page" default:"10"`
	Cursor    string `help:"Opaque cursor from the previous response"`
}

func (cmd *PagesSearchCmd) Run(app *App) error {
	if err := validateRange("limit", cmd.Limit, 1, maxListLimit, helpHint("pages search")); err != nil {
		return err
	}
	if cmd.Query == "" && cmd.CQL == "" {
		return validationError("provide exactly one of --query or --cql", helpHint("pages search"))
	}
	if cmd.Query != "" && cmd.CQL != "" {
		return validationError("provide exactly one of --query or --cql", helpHint("pages search"))
	}
	if cmd.CQL != "" {
		if cmd.TitleOnly {
			return validationError("--title-only cannot be combined with --cql", helpHint("pages search"))
		}
		if cmd.SpaceID != "" || cmd.SpaceKey != "" {
			return validationError("--space-id and --space-key cannot be combined with --cql", helpHint("pages search"))
		}
	}
	if cmd.SpaceID != "" && cmd.SpaceKey != "" {
		return validationError("provide at most one of --space-id or --space-key", helpHint("pages search"))
	}

	result, err := app.Client.SearchPages(confluence.PageSearchOptions{
		Query:     cmd.Query,
		CQL:       cmd.CQL,
		TitleOnly: cmd.TitleOnly,
		SpaceID:   cmd.SpaceID,
		SpaceKey:  cmd.SpaceKey,
		Limit:     cmd.Limit,
		Cursor:    cmd.Cursor,
	})
	if err != nil {
		return err
	}

	items := make([]SearchSummary, len(result.Results))
	for i, resultItem := range result.Results {
		items[i] = newSearchSummary(resultItem)
	}

	if app.IsPlain() {
		renderSearchPlain(app.Stdout, items, result.NextCursor)
		return nil
	}
	return renderJSON(app.Stdout, listEnvelope(items, cmd.Limit, result.NextCursor, "page-search-result", []string{"id", "title", "spaceId", "excerpt", "url"}))
}
