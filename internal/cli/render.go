package cli

import (
	"encoding/json"
	"fmt"
	"io"
	"strings"
	"text/tabwriter"

	"github.com/Prisma-Labs-Dev/confluence-cli/internal/htmlmd"
)

func renderJSON(w io.Writer, v any) error {
	encoder := json.NewEncoder(w)
	encoder.SetIndent("", "  ")
	return encoder.Encode(v)
}

func renderSpacesPlain(w io.Writer, results []SpaceSummary, nextCursor string) {
	tw := tabwriter.NewWriter(w, 0, 0, 2, ' ', 0)
	discardWrite(fmt.Fprintln(tw, "ID\tKEY\tNAME\tTYPE\tSTATUS"))
	for _, result := range results {
		discardWrite(fmt.Fprintf(tw, "%s\t%s\t%s\t%s\t%s\n", result.ID, result.Key, result.Name, result.Type, result.Status))
	}
	_ = tw.Flush()
	if nextCursor != "" {
		discardWrite(fmt.Fprintf(w, "\nNext cursor: %s\n", nextCursor))
	}
}

func renderPagesPlain(w io.Writer, results []PageSummary, nextCursor string) {
	tw := tabwriter.NewWriter(w, 0, 0, 2, ' ', 0)
	discardWrite(fmt.Fprintln(tw, "ID\tTITLE\tSTATUS\tSPACE ID\tVERSION"))
	for _, result := range results {
		discardWrite(fmt.Fprintf(tw, "%s\t%s\t%s\t%s\t%d\n", result.ID, result.Title, result.Status, result.SpaceID, result.VersionNumber))
	}
	_ = tw.Flush()
	if nextCursor != "" {
		discardWrite(fmt.Fprintf(w, "\nNext cursor: %s\n", nextCursor))
	}
}

func renderPagePlain(w io.Writer, page PageDetail) {
	discardWrite(fmt.Fprintf(w, "ID: %s\n", page.ID))
	discardWrite(fmt.Fprintf(w, "Title: %s\n", page.Title))
	discardWrite(fmt.Fprintf(w, "Space ID: %s\n", page.SpaceID))
	discardWrite(fmt.Fprintf(w, "Status: %s\n", page.Status))
	if page.ParentID != "" {
		discardWrite(fmt.Fprintf(w, "Parent ID: %s\n", page.ParentID))
	}
	if page.Version != nil {
		discardWrite(fmt.Fprintf(w, "Version: %d\n", page.Version.Number))
	}
	if page.Body == nil {
		return
	}

	discardWrite(fmt.Fprintf(w, "\nBody (%s):\n", page.Body.Format))
	switch page.Body.Format {
	case "view":
		markdown, err := htmlmd.Convert(page.Body.Value)
		if err != nil {
			discardWrite(fmt.Fprintln(w, page.Body.Value))
			return
		}
		discardWrite(fmt.Fprintln(w, strings.TrimSpace(markdown)))
	default:
		discardWrite(fmt.Fprintln(w, page.Body.Value))
	}
}

func renderSearchPlain(w io.Writer, results []SearchSummary, nextCursor string) {
	tw := tabwriter.NewWriter(w, 0, 0, 2, ' ', 0)
	discardWrite(fmt.Fprintln(tw, "ID\tTITLE\tSPACE ID"))
	for _, result := range results {
		discardWrite(fmt.Fprintf(tw, "%s\t%s\t%s\n", result.ID, result.Title, result.SpaceID))
	}
	_ = tw.Flush()
	if nextCursor != "" {
		discardWrite(fmt.Fprintf(w, "\nNext cursor: %s\n", nextCursor))
	}
}

func renderTreePlain(w io.Writer, tree PageTree) {
	discardWrite(fmt.Fprintf(w, "Root page: %s\n", tree.RootPageID))
	discardWrite(fmt.Fprintf(w, "Depth: %d\n", tree.Depth))
	discardWrite(fmt.Fprintf(w, "Limit per level: %d\n", tree.LimitPerLevel))
	if tree.HasMoreChildren {
		discardWrite(fmt.Fprintln(w, "More children available: yes"))
	}
	if len(tree.Children) == 0 {
		discardWrite(fmt.Fprintln(w, "\n(no children)"))
		return
	}
	discardWrite(fmt.Fprintln(w))
	renderTreeNodes(w, tree.Children, "")
}

func renderTreeNodes(w io.Writer, nodes []PageTreeNode, prefix string) {
	for i, node := range nodes {
		connector := "|- "
		nextPrefix := prefix + "|  "
		if i == len(nodes)-1 {
			connector = "`- "
			nextPrefix = prefix + "   "
		}

		suffix := ""
		if node.HasMoreChildren {
			suffix = " [more]"
		}
		discardWrite(fmt.Fprintf(w, "%s%s%s (id:%s)%s\n", prefix, connector, node.Title, node.ID, suffix))
		if len(node.Children) > 0 {
			renderTreeNodes(w, node.Children, nextPrefix)
		}
	}
}
