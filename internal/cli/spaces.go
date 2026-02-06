package cli

import confluence "github.com/Prisma-Labs-Dev/confluence-cli"

type SpacesCmd struct {
	List SpacesListCmd `cmd:"" help:"List spaces"`
}

type SpacesListCmd struct {
	Limit  int    `help:"Maximum number of results" default:"25"`
	Cursor string `help:"Pagination cursor"`
}

func (cmd *SpacesListCmd) Run(app *App) error {
	result, err := app.Client.ListSpaces(confluence.ListSpacesOptions{
		Limit:  cmd.Limit,
		Cursor: cmd.Cursor,
	})
	if err != nil {
		return err
	}

	if app.Plain {
		renderSpaces(app.Stdout, result)
		return nil
	}
	return renderJSON(app.Stdout, result)
}
