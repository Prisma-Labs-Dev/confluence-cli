package cli

import confluence "github.com/Prisma-Labs-Dev/confluence-cli"

type PagesGetCmd struct {
	PageID     string `help:"Page ID from list/search output" required:""`
	BodyFormat string `name:"body-format" help:"Optional body format"`
}

func (cmd *PagesGetCmd) Run(app *App) error {
	if cmd.BodyFormat != "" && cmd.BodyFormat != "view" && cmd.BodyFormat != "storage" && cmd.BodyFormat != "atlas_doc_format" {
		return validationError("body-format must be one of: view, storage, atlas_doc_format", helpHint("pages get"))
	}

	page, err := app.Client.GetPage(confluence.GetPageOptions{
		PageID:     cmd.PageID,
		BodyFormat: cmd.BodyFormat,
	})
	if err != nil {
		return err
	}

	item := newPageDetail(page, cmd.BodyFormat)
	if app.IsPlain() {
		renderPagePlain(app.Stdout, item)
		return nil
	}

	fields := []string{"id", "title", "spaceId", "status", "parentId", "parentType", "authorId", "createdAt", "version"}
	if cmd.BodyFormat != "" {
		fields = append(fields, "body")
	}
	return renderJSON(app.Stdout, itemEnvelope(item, "page-detail", fields))
}
