package cli

import (
	"fmt"

	confluence "github.com/Prisma-Labs-Dev/confluence-cli"
)

type PagesCmd struct {
	List   PagesListCmd   `cmd:"" help:"List pages in a space"`
	Get    PagesGetCmd    `cmd:"" help:"Get a page by ID"`
	Tree   PagesTreeCmd   `cmd:"" help:"Show page tree (children)"`
	Search PagesSearchCmd `cmd:"" help:"Search pages with CQL"`
}

type PagesListCmd struct {
	SpaceID string `help:"Space ID" required:""`
	Limit   int    `help:"Maximum number of results" default:"25"`
	Cursor  string `help:"Pagination cursor"`
	Sort    string `help:"Sort order (title, created-date, -modified-date)"`
}

func (cmd *PagesListCmd) Run(app *App) error {
	result, err := app.Client.ListPages(confluence.ListPagesOptions{
		SpaceID: cmd.SpaceID,
		Limit:   cmd.Limit,
		Cursor:  cmd.Cursor,
		Sort:    cmd.Sort,
	})
	if err != nil {
		return err
	}

	if app.Plain {
		renderPages(app.Stdout, result)
		return nil
	}
	return renderJSON(app.Stdout, result)
}

type PagesGetCmd struct {
	PageID     string `help:"Page ID" required:""`
	BodyFormat string `help:"Body format (view, storage, atlas_doc_format)" name:"body-format"`
}

func (cmd *PagesGetCmd) Run(app *App) error {
	page, err := app.Client.GetPage(confluence.GetPageOptions{
		PageID:     cmd.PageID,
		BodyFormat: cmd.BodyFormat,
	})
	if err != nil {
		return err
	}

	if app.Plain {
		renderPage(app.Stdout, page)
		return nil
	}
	return renderJSON(app.Stdout, page)
}

type PagesTreeCmd struct {
	PageID string `help:"Root page ID" required:""`
	Depth  int    `help:"Maximum depth to traverse" default:"1"`
}

func (cmd *PagesTreeCmd) Run(app *App) error {
	if cmd.Depth < 1 {
		cmd.Depth = 1
	}

	type treeNode struct {
		confluence.Page
		Children []treeNode `json:"children,omitempty"`
	}

	var buildTree func(pageID string, depth int) ([]treeNode, error)
	buildTree = func(pageID string, depth int) ([]treeNode, error) {
		if depth <= 0 {
			return nil, nil
		}

		var allChildren []confluence.Page
		cursor := ""
		for {
			result, err := app.Client.GetPageChildren(confluence.GetPageChildrenOptions{
				PageID: pageID,
				Limit:  25,
				Cursor: cursor,
			})
			if err != nil {
				return nil, err
			}
			allChildren = append(allChildren, result.Results...)
			if result.NextCursor == "" {
				break
			}
			cursor = result.NextCursor
		}

		nodes := make([]treeNode, len(allChildren))
		for i, child := range allChildren {
			nodes[i] = treeNode{Page: child}
			if depth > 1 {
				children, err := buildTree(child.ID, depth-1)
				if err != nil {
					return nil, err
				}
				nodes[i].Children = children
			}
		}
		return nodes, nil
	}

	tree, err := buildTree(cmd.PageID, cmd.Depth)
	if err != nil {
		return err
	}

	result := struct {
		PageID   string     `json:"pageId"`
		Children []treeNode `json:"children"`
	}{
		PageID:   cmd.PageID,
		Children: tree,
	}

	if app.Plain {
		fmt.Fprintf(app.Stdout, "Page %s\n", cmd.PageID)
		var printNodes func(nodes []treeNode, prefix string)
		printNodes = func(nodes []treeNode, prefix string) {
			for i, n := range nodes {
				connector := "├── "
				childPrefix := prefix + "│   "
				if i == len(nodes)-1 {
					connector = "└── "
					childPrefix = prefix + "    "
				}
				fmt.Fprintf(app.Stdout, "%s%s%s (id:%s)\n", prefix, connector, n.Title, n.ID)
				if len(n.Children) > 0 {
					printNodes(n.Children, childPrefix)
				}
			}
		}
		printNodes(tree, "")
		return nil
	}
	return renderJSON(app.Stdout, result)
}

type PagesSearchCmd struct {
	CQL     string `help:"CQL query" required:""`
	SpaceID string `help:"Filter by space ID"`
	Limit   int    `help:"Maximum number of results" default:"25"`
	Cursor  string `help:"Pagination cursor"`
}

func (cmd *PagesSearchCmd) Run(app *App) error {
	cql := cmd.CQL
	if cmd.SpaceID != "" {
		cql = fmt.Sprintf("%s AND space.id=%s", cql, cmd.SpaceID)
	}

	result, err := app.Client.Search(confluence.SearchOptions{
		CQL:    cql,
		Limit:  cmd.Limit,
		Cursor: cmd.Cursor,
	})
	if err != nil {
		return err
	}

	if app.Plain {
		renderSearchResults(app.Stdout, result)
		return nil
	}
	return renderJSON(app.Stdout, result)
}
