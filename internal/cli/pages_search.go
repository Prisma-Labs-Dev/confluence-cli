package cli

import confluence "github.com/Prisma-Labs-Dev/confluence-cli"

type PagesSearchCmd struct {
	Query     string `help:"Search text to match in page content or titles" required:""`
	TitleOnly bool   `help:"Restrict matching to page titles"`
	SpaceID   string `help:"Optional space ID filter"`
	Limit     int    `help:"Maximum number of results per page" default:"10"`
	Cursor    string `help:"Opaque cursor from the previous response"`
}

func (cmd *PagesSearchCmd) Run(app *App) error {
	if err := validateRange("limit", cmd.Limit, 1, maxListLimit, helpHint("pages search")); err != nil {
		return err
	}

	result, err := app.Client.SearchPages(confluence.PageSearchOptions{
		Query:     cmd.Query,
		TitleOnly: cmd.TitleOnly,
		SpaceID:   cmd.SpaceID,
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
