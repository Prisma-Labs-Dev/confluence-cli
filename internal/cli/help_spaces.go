package cli

import "fmt"

func spacesHelp() string {
	return `Usage: confluence spaces <command>

Space discovery commands.

Commands:
  list [flags]
    List spaces with compact summaries and cursor pagination.

Run "confluence spaces list --help" for the live contract.
`
}

func spacesListHelp() string {
	return fmt.Sprintf(`Usage: confluence spaces list [flags]

List spaces with compact summaries.

Output (json):
  {
    "results": [
      {"id":"...","key":"...","name":"...","type":"...","status":"..."}
    ],
    "page": {"limit": %d, "nextCursor": "..."},
    "schema": {"itemType":"space-summary","fields":["id","key","name","type","status"]}
  }

Pagination:
  - Results are bounded by --limit.
  - Pass response.page.nextCursor back via --cursor.

Examples:
  confluence spaces list
  confluence spaces list --limit 25
  confluence spaces list --cursor abc123
  confluence --format plain spaces list

Flags:
  -h, --help              Show command help.
      --url=STRING        Confluence base URL ($CONFLUENCE_URL)
      --email=STRING      Atlassian account email ($CONFLUENCE_EMAIL)
      --token=STRING      Atlassian API token ($CONFLUENCE_API_TOKEN)
      --format=json       Output format: json or plain
      --timeout=30s       HTTP timeout
      --limit=%d          Maximum number of results per page (%d-%d)
      --cursor=STRING     Opaque cursor from response.page.nextCursor
`, defaultListLimit, defaultListLimit, 1, maxListLimit)
}
