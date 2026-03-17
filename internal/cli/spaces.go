package cli

import confluence "github.com/Prisma-Labs-Dev/confluence-cli"

type SpacesListCmd struct {
	Limit  int    `help:"Maximum number of results per page" default:"10"`
	Cursor string `help:"Opaque cursor from the previous response"`
}

func (cmd *SpacesListCmd) Run(app *App) error {
	if err := validateRange("limit", cmd.Limit, 1, maxListLimit, helpHint("spaces list")); err != nil {
		return err
	}

	result, err := app.Client.ListSpaces(confluence.ListSpacesOptions{Limit: cmd.Limit, Cursor: cmd.Cursor})
	if err != nil {
		return err
	}

	items := make([]SpaceSummary, len(result.Results))
	for i, space := range result.Results {
		items[i] = newSpaceSummary(space)
	}

	if app.IsPlain() {
		renderSpacesPlain(app.Stdout, items, result.NextCursor)
		return nil
	}
	return renderJSON(app.Stdout, listEnvelope(items, cmd.Limit, result.NextCursor, "space-summary", []string{"id", "key", "name", "type", "status"}))
}
