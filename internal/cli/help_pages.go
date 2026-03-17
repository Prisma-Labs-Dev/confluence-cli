package cli

import "fmt"

func pagesHelp() string {
	return `Usage: confluence pages <command>

Page discovery commands.

Commands:
  list [flags]
    List pages in a space with bounded summaries.

  get [flags]
    Get page metadata by default; request body content explicitly.

  tree [flags]
    Traverse a bounded page tree with per-level limits.

  search [flags]
    Search pages with safe query inputs or raw CQL.

Run "confluence pages <command> --help" for the live contract.
`
}

func pagesListHelp() string {
	return fmt.Sprintf(`Usage: confluence pages list --space-id=STRING [flags]

List pages in a space with bounded summaries.

Output (json):
  {
    "results": [
      {"id":"...","title":"...","spaceId":"...","status":"...","parentId":"...","versionNumber":123}
    ],
    "page": {"limit": %d, "nextCursor": "..."},
    "schema": {"itemType":"page-summary","fields":["id","title","spaceId","status","parentId","versionNumber"]}
  }

Pagination:
  - Results are bounded by --limit.
  - Pass response.page.nextCursor back via --cursor.

Examples:
  confluence pages list --space-id 12345
  confluence pages list --space-id 12345 --limit 20
  confluence pages list --space-id 12345 --cursor abc123
  confluence --format plain pages list --space-id 12345

Flags:
  -h, --help               Show command help.
      --url=STRING         Confluence base URL ($CONFLUENCE_URL)
      --email=STRING       Atlassian account email ($CONFLUENCE_EMAIL)
      --token=STRING       Atlassian API token ($CONFLUENCE_API_TOKEN)
      --format=json        Output format: json or plain
      --timeout=30s        HTTP timeout
      --space-id=STRING    Space ID from spaces list output
      --limit=%d           Maximum number of results per page (%d-%d)
      --cursor=STRING      Opaque cursor from response.page.nextCursor
      --sort=STRING        Sort order: title, created-date, or -modified-date
`, defaultListLimit, defaultListLimit, 1, maxListLimit)
}

func pagesGetHelp() string {
	return `Usage: confluence pages get --page-id=STRING [flags]

Get a page by ID.

Default behavior:
  - Returns metadata only.
  - Body content is omitted unless --body-format is set explicitly.

Output (json):
  {
    "item": {
      "id":"...",
      "title":"...",
      "spaceId":"...",
      "status":"...",
      "version":{"number":123},
      "body":{"format":"view","value":"..."}
    },
    "schema": {"itemType":"page-detail","fields":["id","title","spaceId","status","parentId","parentType","authorId","createdAt","version","body"]}
  }

Examples:
  confluence pages get --page-id 67890
  confluence pages get --page-id 67890 --body-format view
  confluence --format plain pages get --page-id 67890 --body-format view
  confluence pages get --page-id 67890 --body-format storage

Flags:
  -h, --help                  Show command help.
      --url=STRING            Confluence base URL ($CONFLUENCE_URL)
      --email=STRING          Atlassian account email ($CONFLUENCE_EMAIL)
      --token=STRING          Atlassian API token ($CONFLUENCE_API_TOKEN)
      --format=json           Output format: json or plain
      --timeout=30s           HTTP timeout
      --page-id=STRING        Page ID from list/search output
      --body-format=STRING    Optional body format: view, storage, atlas_doc_format
`
}

func pagesTreeHelp() string {
	return fmt.Sprintf(`Usage: confluence pages tree --page-id=STRING [flags]

Traverse a bounded page tree.

Default behavior:
  - Depth defaults to %d.
  - Each node fetches at most %d children.
  - If a node has more children, hasMoreChildren is set instead of fetching the entire subtree.

Output (json):
  {
    "item": {
      "rootPageId":"...",
      "depth": %d,
      "limitPerLevel": %d,
      "hasMoreChildren": true,
      "children": [
        {"id":"...","title":"...","spaceId":"...","status":"...","hasMoreChildren":true,"children":[...]}
      ]
    },
    "schema": {"itemType":"page-tree","fields":["rootPageId","depth","limitPerLevel","hasMoreChildren","children"]}
  }

Examples:
  confluence pages tree --page-id 67890
  confluence pages tree --page-id 67890 --depth 2
  confluence pages tree --page-id 67890 --depth 2 --limit-per-level 5
  confluence --format plain pages tree --page-id 67890

Flags:
  -h, --help                    Show command help.
      --url=STRING              Confluence base URL ($CONFLUENCE_URL)
      --email=STRING            Atlassian account email ($CONFLUENCE_EMAIL)
      --token=STRING            Atlassian API token ($CONFLUENCE_API_TOKEN)
      --format=json             Output format: json or plain
      --timeout=30s             HTTP timeout
      --page-id=STRING          Root page ID
      --depth=%d                Maximum traversal depth (%d-%d)
      --limit-per-level=%d      Maximum children fetched per node (%d-%d)
`, defaultTreeDepth, defaultTreeLimitPerLevel, defaultTreeDepth, defaultTreeLimitPerLevel, defaultTreeDepth, 1, maxTreeDepth, defaultTreeLimitPerLevel, 1, maxTreeLimitPerLevel)
}

func pagesSearchHelp() string {
	return fmt.Sprintf(`Usage: confluence pages search (--query=STRING | --cql=STRING) [flags]

Search pages with safe query inputs or raw CQL.

Default behavior:
  - Query mode searches full text unless --title-only is set.
  - Query mode can be scoped with either --space-id or --space-key.
  - Raw CQL mode is for advanced research queries and cannot be combined with query-only filters.
  - Results are bounded by --limit.
  - Internally this uses Atlassian's Confluence Cloud REST v1 search endpoint because REST v2 does not yet provide equivalent search capability.

Output (json):
  {
    "results": [
      {"id":"...","title":"...","spaceId":"...","excerpt":"...","url":"..."}
    ],
    "page": {"limit": %d, "nextCursor": "..."},
    "schema": {"itemType":"page-search-result","fields":["id","title","spaceId","excerpt","url"]}
  }

Pagination:
  - Pass response.page.nextCursor back via --cursor.
  - Use query mode space filters or targeted CQL to keep results compact.

Examples:
  confluence pages search --query "deployment"
  confluence pages search --query "meeting notes" --title-only
  confluence pages search --query "runbook" --space-key TNLTA
  confluence pages search --cql 'space = "SC" AND title ~ "slotting"'
  confluence --format plain pages search --cql 'space = "TNLTA" AND text ~ "OutSystems"'

Flags:
  -h, --help               Show command help.
      --url=STRING         Confluence base URL ($CONFLUENCE_URL)
      --email=STRING       Atlassian account email ($CONFLUENCE_EMAIL)
      --token=STRING       Atlassian API token ($CONFLUENCE_API_TOKEN)
      --format=json        Output format: json or plain
      --timeout=30s        HTTP timeout
      --query=STRING       Search text to match in page content or titles
      --cql=STRING         Raw CQL expression for advanced search
      --title-only         Restrict matching to page titles (query mode only)
      --space-id=STRING    Optional space ID filter (query mode only)
      --space-key=STRING   Optional space key filter such as SC or TNLTA (query mode only)
      --limit=%d           Maximum number of results per page (%d-%d)
      --cursor=STRING      Opaque cursor from response.page.nextCursor
`, defaultListLimit, defaultListLimit, 1, maxListLimit)
}
