package cli

import (
	"encoding/json"
	"fmt"
	"io"
	"text/tabwriter"

	confluence "github.com/Prisma-Labs-Dev/confluence-cli"
)

func renderJSON(w io.Writer, v any) error {
	enc := json.NewEncoder(w)
	enc.SetIndent("", "  ")
	return enc.Encode(v)
}

func renderSpaces(w io.Writer, result *confluence.ListResult[confluence.Space]) {
	tw := tabwriter.NewWriter(w, 0, 0, 2, ' ', 0)
	fmt.Fprintln(tw, "ID\tKEY\tNAME\tTYPE\tSTATUS")
	for _, s := range result.Results {
		fmt.Fprintf(tw, "%s\t%s\t%s\t%s\t%s\n", s.ID, s.Key, s.Name, s.Type, s.Status)
	}
	tw.Flush()
	if result.NextCursor != "" {
		fmt.Fprintf(w, "\nNext cursor: %s\n", result.NextCursor)
	}
}

func renderPages(w io.Writer, result *confluence.ListResult[confluence.Page]) {
	tw := tabwriter.NewWriter(w, 0, 0, 2, ' ', 0)
	fmt.Fprintln(tw, "ID\tTITLE\tSTATUS\tSPACE ID")
	for _, p := range result.Results {
		fmt.Fprintf(tw, "%s\t%s\t%s\t%s\n", p.ID, p.Title, p.Status, p.SpaceID)
	}
	tw.Flush()
	if result.NextCursor != "" {
		fmt.Fprintf(w, "\nNext cursor: %s\n", result.NextCursor)
	}
}

func renderPage(w io.Writer, page *confluence.Page) {
	fmt.Fprintf(w, "ID:      %s\n", page.ID)
	fmt.Fprintf(w, "Title:   %s\n", page.Title)
	fmt.Fprintf(w, "Space:   %s\n", page.SpaceID)
	fmt.Fprintf(w, "Status:  %s\n", page.Status)
	if page.Version != nil {
		fmt.Fprintf(w, "Version: %d\n", page.Version.Number)
	}
	if page.Body != nil {
		if page.Body.View != nil {
			fmt.Fprintf(w, "\n--- Body (view) ---\n%s\n", page.Body.View.Value)
		}
		if page.Body.Storage != nil {
			fmt.Fprintf(w, "\n--- Body (storage) ---\n%s\n", page.Body.Storage.Value)
		}
		if page.Body.AtlasDocFormat != nil {
			fmt.Fprintf(w, "\n--- Body (atlas_doc_format) ---\n%s\n", page.Body.AtlasDocFormat.Value)
		}
	}
}

func renderSearchResults(w io.Writer, result *confluence.ListResult[confluence.SearchResult]) {
	tw := tabwriter.NewWriter(w, 0, 0, 2, ' ', 0)
	fmt.Fprintln(tw, "ID\tTITLE\tTYPE\tSPACE ID")
	for _, r := range result.Results {
		fmt.Fprintf(tw, "%s\t%s\t%s\t%s\n", r.ID, r.Title, r.Type, r.SpaceID)
	}
	tw.Flush()
	if result.NextCursor != "" {
		fmt.Fprintf(w, "\nNext cursor: %s\n", result.NextCursor)
	}
}

