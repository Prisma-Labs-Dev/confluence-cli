package cli

import confluence "github.com/Prisma-Labs-Dev/confluence-cli"

type PagesTreeCmd struct {
	PageID        string `help:"Root page ID" required:""`
	Depth         int    `help:"Maximum traversal depth" default:"1"`
	LimitPerLevel int    `name:"limit-per-level" help:"Maximum children fetched per node" default:"10"`
}

func (cmd *PagesTreeCmd) Run(app *App) error {
	if err := validateRange("depth", cmd.Depth, 1, maxTreeDepth, helpHint("pages tree")); err != nil {
		return err
	}
	if err := validateRange("limit-per-level", cmd.LimitPerLevel, 1, maxTreeLimitPerLevel, helpHint("pages tree")); err != nil {
		return err
	}

	children, hasMoreChildren, err := buildPageTree(app.Client, cmd.PageID, cmd.Depth, cmd.LimitPerLevel)
	if err != nil {
		return err
	}

	tree := PageTree{
		RootPageID:      cmd.PageID,
		Depth:           cmd.Depth,
		LimitPerLevel:   cmd.LimitPerLevel,
		HasMoreChildren: hasMoreChildren,
		Children:        children,
	}
	if app.IsPlain() {
		renderTreePlain(app.Stdout, tree)
		return nil
	}
	return renderJSON(app.Stdout, itemEnvelope(tree, "page-tree", []string{"rootPageId", "depth", "limitPerLevel", "hasMoreChildren", "children"}))
}

func buildPageTree(client *confluence.Client, pageID string, depth, limitPerLevel int) ([]PageTreeNode, bool, error) {
	if depth <= 0 {
		return nil, false, nil
	}

	result, err := client.GetPageChildren(confluence.GetPageChildrenOptions{
		PageID: pageID,
		Limit:  limitPerLevel,
	})
	if err != nil {
		return nil, false, err
	}

	nodes := make([]PageTreeNode, len(result.Results))
	for i, child := range result.Results {
		nodes[i] = PageTreeNode{
			ID:      child.ID,
			Title:   child.Title,
			SpaceID: child.SpaceID,
			Status:  child.Status,
		}
		if depth > 1 {
			grandChildren, childHasMore, err := buildPageTree(client, child.ID, depth-1, limitPerLevel)
			if err != nil {
				return nil, false, err
			}
			nodes[i].Children = grandChildren
			nodes[i].HasMoreChildren = childHasMore
		}
	}

	return nodes, result.NextCursor != "", nil
}
