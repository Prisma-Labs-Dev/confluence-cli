package confluence_test

import (
	"encoding/json"
	"os"
	"path/filepath"
	"reflect"
	"sort"
	"strings"
	"testing"
)

type liveSchemaSnapshot struct {
	ItemType string   `json:"itemType"`
	Fields   []string `json:"fields"`
}

type livePageSnapshot struct {
	Limit         int  `json:"limit,omitempty"`
	HasNextCursor bool `json:"hasNextCursor,omitempty"`
}

type liveSpacesFirstSnapshot struct {
	HasID   bool   `json:"hasId"`
	HasKey  bool   `json:"hasKey"`
	HasName bool   `json:"hasName"`
	Type    string `json:"type,omitempty"`
	Status  string `json:"status,omitempty"`
}

type livePagesFirstSnapshot struct {
	HasID            bool   `json:"hasId"`
	HasTitle         bool   `json:"hasTitle"`
	HasSpaceID       bool   `json:"hasSpaceId"`
	Status           string `json:"status,omitempty"`
	HasParentID      bool   `json:"hasParentId"`
	HasVersionNumber bool   `json:"hasVersionNumber"`
}

type livePageDetailSnapshot struct {
	HasID        bool   `json:"hasId"`
	HasTitle     bool   `json:"hasTitle"`
	HasSpaceID   bool   `json:"hasSpaceId"`
	Status       string `json:"status,omitempty"`
	HasVersion   bool   `json:"hasVersion"`
	HasBody      bool   `json:"hasBody"`
	BodyFormat   string `json:"bodyFormat,omitempty"`
	BodyHasValue bool   `json:"bodyHasValue,omitempty"`
}

type liveTreeSnapshot struct {
	HasRootPageID   bool `json:"hasRootPageId"`
	Depth           int  `json:"depth"`
	LimitPerLevel   int  `json:"limitPerLevel"`
	HasMoreChildren bool `json:"hasMoreChildren"`
	ChildCount      int  `json:"childCount"`
	FirstChildHasID bool `json:"firstChildHasId,omitempty"`
}

type liveSearchFirstSnapshot struct {
	HasID      bool `json:"hasId"`
	HasTitle   bool `json:"hasTitle"`
	HasSpaceID bool `json:"hasSpaceId"`
	HasExcerpt bool `json:"hasExcerpt"`
	HasURL     bool `json:"hasUrl"`
}

type liveListSnapshot[T any] struct {
	Schema      liveSchemaSnapshot `json:"schema"`
	Page        livePageSnapshot   `json:"page"`
	ResultCount int                `json:"resultCount"`
	First       T                  `json:"first"`
}

type liveContractSnapshot struct {
	SpacesList liveListSnapshot[liveSpacesFirstSnapshot] `json:"spacesList"`
	PagesList  liveListSnapshot[livePagesFirstSnapshot]  `json:"pagesList"`
	PagesGet   struct {
		Schema liveSchemaSnapshot     `json:"schema"`
		Item   livePageDetailSnapshot `json:"item"`
	} `json:"pagesGet"`
	PagesGetView struct {
		Schema liveSchemaSnapshot     `json:"schema"`
		Item   livePageDetailSnapshot `json:"item"`
	} `json:"pagesGetView"`
	PagesTree struct {
		Schema liveSchemaSnapshot `json:"schema"`
		Item   liveTreeSnapshot   `json:"item"`
	} `json:"pagesTree"`
	PagesSearch liveListSnapshot[liveSearchFirstSnapshot] `json:"pagesSearch"`
}

func TestLiveAPIContractGolden_Integration(t *testing.T) {
	if os.Getenv("CONFLUENCE_LIVE_E2E") != "1" {
		t.Skip("set CONFLUENCE_LIVE_E2E=1 to run live API golden validation")
	}

	tmp := t.TempDir()
	binPath := buildBinary(t, tmp)
	snapshot := collectLiveContractSnapshot(t, binPath)

	goldenPath := filepath.Join("testdata", "golden", "live", "contract.json")
	if os.Getenv("CONFLUENCE_LIVE_E2E_UPDATE") == "1" {
		writePrettyJSONGolden(t, goldenPath, snapshot)
	}

	var want liveContractSnapshot
	readJSONGolden(t, goldenPath, &want)
	if !reflect.DeepEqual(snapshot, want) {
		got, err := json.MarshalIndent(snapshot, "", "  ")
		if err != nil {
			t.Fatalf("marshal live snapshot: %v", err)
		}
		expected, err := json.MarshalIndent(want, "", "  ")
		if err != nil {
			t.Fatalf("marshal live golden: %v", err)
		}
		t.Fatalf("live golden mismatch\n--- want ---\n%s\n--- got ---\n%s", string(expected), string(got))
	}
}

func collectLiveContractSnapshot(t *testing.T, binPath string) liveContractSnapshot {
	t.Helper()

	spacesStdout, spacesStderr, err := runBinary(binPath, []string{"spaces", "list", "--limit", "1"}, "")
	if err != nil {
		t.Fatalf("live spaces list failed: %v\nstdout=%s\nstderr=%s", err, spacesStdout, spacesStderr)
	}

	var spaces struct {
		Results []struct {
			ID     string `json:"id"`
			Key    string `json:"key"`
			Name   string `json:"name"`
			Type   string `json:"type"`
			Status string `json:"status"`
		} `json:"results"`
		Page struct {
			Limit      int    `json:"limit"`
			NextCursor string `json:"nextCursor"`
		} `json:"page"`
		Schema liveSchemaSnapshot `json:"schema"`
	}
	unmarshalLiveJSON(t, spacesStdout, &spaces)
	if len(spaces.Results) == 0 {
		t.Fatal("live spaces list returned no results")
	}

	spaceID := strings.TrimSpace(os.Getenv("CONFLUENCE_LIVE_SPACE_ID"))
	if spaceID == "" {
		spaceID = spaces.Results[0].ID
	}

	pagesStdout, pagesStderr, err := runBinary(binPath, []string{"pages", "list", "--space-id", spaceID, "--limit", "1", "--sort", "title"}, "")
	if err != nil {
		t.Fatalf("live pages list failed: %v\nstdout=%s\nstderr=%s", err, pagesStdout, pagesStderr)
	}

	var pages struct {
		Results []struct {
			ID            string `json:"id"`
			Title         string `json:"title"`
			SpaceID       string `json:"spaceId"`
			Status        string `json:"status"`
			ParentID      string `json:"parentId"`
			VersionNumber int    `json:"versionNumber"`
		} `json:"results"`
		Page struct {
			Limit      int    `json:"limit"`
			NextCursor string `json:"nextCursor"`
		} `json:"page"`
		Schema liveSchemaSnapshot `json:"schema"`
	}
	unmarshalLiveJSON(t, pagesStdout, &pages)
	if len(pages.Results) == 0 {
		t.Skip("live pages list returned no results for the selected space")
	}

	pageID := strings.TrimSpace(os.Getenv("CONFLUENCE_LIVE_PAGE_ID"))
	if pageID == "" {
		pageID = pages.Results[0].ID
	}

	pageGetStdout, pageGetStderr, err := runBinary(binPath, []string{"pages", "get", "--page-id", pageID}, "")
	if err != nil {
		t.Fatalf("live pages get failed: %v\nstdout=%s\nstderr=%s", err, pageGetStdout, pageGetStderr)
	}

	var pageGet struct {
		Item struct {
			ID      string `json:"id"`
			Title   string `json:"title"`
			SpaceID string `json:"spaceId"`
			Status  string `json:"status"`
			Version *struct {
				Number int `json:"number"`
			} `json:"version"`
			Body *struct {
				Format string `json:"format"`
				Value  string `json:"value"`
			} `json:"body"`
		} `json:"item"`
		Schema liveSchemaSnapshot `json:"schema"`
	}
	unmarshalLiveJSON(t, pageGetStdout, &pageGet)

	pageViewStdout, pageViewStderr, err := runBinary(binPath, []string{"pages", "get", "--page-id", pageID, "--body-format", "view"}, "")
	if err != nil {
		t.Fatalf("live pages get view failed: %v\nstdout=%s\nstderr=%s", err, pageViewStdout, pageViewStderr)
	}

	var pageView struct {
		Item struct {
			Body *struct {
				Format string `json:"format"`
				Value  string `json:"value"`
			} `json:"body"`
		} `json:"item"`
		Schema liveSchemaSnapshot `json:"schema"`
	}
	unmarshalLiveJSON(t, pageViewStdout, &pageView)

	treeStdout, treeStderr, err := runBinary(binPath, []string{"pages", "tree", "--page-id", pageID, "--depth", "1"}, "")
	if err != nil {
		t.Fatalf("live pages tree failed: %v\nstdout=%s\nstderr=%s", err, treeStdout, treeStderr)
	}

	var tree struct {
		Item struct {
			RootPageID      string `json:"rootPageId"`
			Depth           int    `json:"depth"`
			LimitPerLevel   int    `json:"limitPerLevel"`
			HasMoreChildren bool   `json:"hasMoreChildren"`
			Children        []struct {
				ID string `json:"id"`
			} `json:"children"`
		} `json:"item"`
		Schema liveSchemaSnapshot `json:"schema"`
	}
	unmarshalLiveJSON(t, treeStdout, &tree)

	query := strings.TrimSpace(os.Getenv("CONFLUENCE_LIVE_SEARCH_QUERY"))
	if query == "" {
		query = firstSearchToken(pages.Results[0].Title)
	}
	searchStdout, searchStderr, err := runBinary(binPath, []string{"pages", "search", "--query", query, "--limit", "1"}, "")
	if err != nil {
		t.Fatalf("live pages search failed: %v\nstdout=%s\nstderr=%s", err, searchStdout, searchStderr)
	}

	var search struct {
		Results []struct {
			ID      string `json:"id"`
			Title   string `json:"title"`
			SpaceID string `json:"spaceId"`
			Excerpt string `json:"excerpt"`
			URL     string `json:"url"`
		} `json:"results"`
		Page struct {
			Limit      int    `json:"limit"`
			NextCursor string `json:"nextCursor"`
		} `json:"page"`
		Schema liveSchemaSnapshot `json:"schema"`
	}
	unmarshalLiveJSON(t, searchStdout, &search)
	if len(search.Results) == 0 {
		t.Skip("live pages search returned no results for the selected query")
	}

	snapshot := liveContractSnapshot{
		SpacesList: liveListSnapshot[liveSpacesFirstSnapshot]{
			Schema:      sanitizeSchema(spaces.Schema),
			Page:        livePageSnapshot{Limit: spaces.Page.Limit, HasNextCursor: spaces.Page.NextCursor != ""},
			ResultCount: len(spaces.Results),
			First: liveSpacesFirstSnapshot{
				HasID:   spaces.Results[0].ID != "",
				HasKey:  spaces.Results[0].Key != "",
				HasName: spaces.Results[0].Name != "",
				Type:    spaces.Results[0].Type,
				Status:  spaces.Results[0].Status,
			},
		},
		PagesList: liveListSnapshot[livePagesFirstSnapshot]{
			Schema:      sanitizeSchema(pages.Schema),
			Page:        livePageSnapshot{Limit: pages.Page.Limit, HasNextCursor: pages.Page.NextCursor != ""},
			ResultCount: len(pages.Results),
			First: livePagesFirstSnapshot{
				HasID:            pages.Results[0].ID != "",
				HasTitle:         pages.Results[0].Title != "",
				HasSpaceID:       pages.Results[0].SpaceID != "",
				Status:           pages.Results[0].Status,
				HasParentID:      pages.Results[0].ParentID != "",
				HasVersionNumber: pages.Results[0].VersionNumber > 0,
			},
		},
		PagesSearch: liveListSnapshot[liveSearchFirstSnapshot]{
			Schema:      sanitizeSchema(search.Schema),
			Page:        livePageSnapshot{Limit: search.Page.Limit, HasNextCursor: search.Page.NextCursor != ""},
			ResultCount: len(search.Results),
			First: liveSearchFirstSnapshot{
				HasID:      search.Results[0].ID != "",
				HasTitle:   search.Results[0].Title != "",
				HasSpaceID: search.Results[0].SpaceID != "",
				HasExcerpt: search.Results[0].Excerpt != "",
				HasURL:     search.Results[0].URL != "",
			},
		},
	}

	snapshot.PagesGet.Schema = sanitizeSchema(pageGet.Schema)
	snapshot.PagesGet.Item = livePageDetailSnapshot{
		HasID:      pageGet.Item.ID != "",
		HasTitle:   pageGet.Item.Title != "",
		HasSpaceID: pageGet.Item.SpaceID != "",
		Status:     pageGet.Item.Status,
		HasVersion: pageGet.Item.Version != nil,
		HasBody:    pageGet.Item.Body != nil,
	}

	snapshot.PagesGetView.Schema = sanitizeSchema(pageView.Schema)
	snapshot.PagesGetView.Item = livePageDetailSnapshot{
		HasBody:      pageView.Item.Body != nil,
		BodyHasValue: pageView.Item.Body != nil && pageView.Item.Body.Value != "",
	}
	if pageView.Item.Body != nil {
		snapshot.PagesGetView.Item.BodyFormat = pageView.Item.Body.Format
	}

	snapshot.PagesTree.Schema = sanitizeSchema(tree.Schema)
	snapshot.PagesTree.Item = liveTreeSnapshot{
		HasRootPageID:   tree.Item.RootPageID != "",
		Depth:           tree.Item.Depth,
		LimitPerLevel:   tree.Item.LimitPerLevel,
		HasMoreChildren: tree.Item.HasMoreChildren,
		ChildCount:      len(tree.Item.Children),
		FirstChildHasID: len(tree.Item.Children) > 0 && tree.Item.Children[0].ID != "",
	}

	return snapshot
}

func sanitizeSchema(schema liveSchemaSnapshot) liveSchemaSnapshot {
	fields := append([]string(nil), schema.Fields...)
	sort.Strings(fields)
	return liveSchemaSnapshot{ItemType: schema.ItemType, Fields: fields}
}

func unmarshalLiveJSON(t *testing.T, raw string, out any) {
	t.Helper()
	if err := json.Unmarshal([]byte(raw), out); err != nil {
		t.Fatalf("parse live JSON: %v\njson=%s", err, raw)
	}
}

func firstSearchToken(title string) string {
	for _, part := range strings.Fields(title) {
		part = strings.Trim(part, ":,.-_/()[]{}")
		if len(part) >= 4 {
			return part
		}
	}
	return "page"
}
