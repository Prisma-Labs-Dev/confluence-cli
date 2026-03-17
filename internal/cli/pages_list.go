package cli

import confluence "github.com/Prisma-Labs-Dev/confluence-cli"

type PagesListCmd struct {
	SpaceID string `help:"Space ID from spaces list output" required:""`
	Limit   int    `help:"Maximum number of results per page" default:"10"`
	Cursor  string `help:"Opaque cursor from the previous response"`
	Sort    string `help:"Sort order: title, created-date, or -modified-date"`
}

func (cmd *PagesListCmd) Run(app *App) error {
	if err := validateRange("limit", cmd.Limit, 1, maxListLimit, helpHint("pages list")); err != nil {
		return err
	}

	result, err := app.Client.ListPages(confluence.ListPagesOptions{
		SpaceID: cmd.SpaceID,
		Limit:   cmd.Limit,
		Cursor:  cmd.Cursor,
		Sort:    cmd.Sort,
	})
	if err != nil {
		return err
	}

	items := make([]PageSummary, len(result.Results))
	for i, page := range result.Results {
		items[i] = newPageSummary(page)
	}

	if app.IsPlain() {
		renderPagesPlain(app.Stdout, items, result.NextCursor)
		return nil
	}
	return renderJSON(app.Stdout, listEnvelope(items, cmd.Limit, result.NextCursor, "page-summary", []string{"id", "title", "spaceId", "status", "parentId", "versionNumber"}))
}
