package htmlmd

import (
	"fmt"
	"net/url"
	"strings"

	md "github.com/JohannesKaufmann/html-to-markdown"
	"github.com/JohannesKaufmann/html-to-markdown/plugin"
	"github.com/PuerkitoBio/goquery"
)

var converter = newConverter()

// Convert transforms Confluence view HTML into compact, readable Markdown.
func Convert(html string) (string, error) {
	if strings.TrimSpace(html) == "" {
		return "", nil
	}
	return converter.ConvertString(html)
}

func newConverter() *md.Converter {
	conv := md.NewConverter("", true, &md.Options{
		CodeBlockStyle:   "fenced",
		BulletListMarker: "-",
		EscapeMode:       "basic",
	})

	conv.Use(plugin.GitHubFlavored())
	conv.Before(preprocessConfluenceHTML)
	conv.AddRules(
		md.Rule{
			Filter: []string{"br"},
			Replacement: func(_ string, _ *goquery.Selection, _ *md.Options) *string {
				return md.String("\n")
			},
		},
		md.Rule{
			Filter: []string{"img"},
			Replacement: func(_ string, _ *goquery.Selection, _ *md.Options) *string {
				// Images add significant noise in plain mode and are rarely useful for LLM context.
				return md.String("")
			},
		},
		md.Rule{
			Filter: []string{"time"},
			Replacement: func(content string, selec *goquery.Selection, _ *md.Options) *string {
				if datetime := strings.TrimSpace(selec.AttrOr("datetime", "")); datetime != "" {
					return md.String(datetime)
				}
				return md.String(normalizeInlineText(content))
			},
		},
		md.Rule{
			Filter: []string{"span"},
			Replacement: func(content string, selec *goquery.Selection, _ *md.Options) *string {
				if selec.HasClass("status-macro") {
					status := normalizeInlineText(content)
					if status == "" {
						return md.String("")
					}
					return md.String("[" + status + "]")
				}

				if selec.HasClass("confluence-jim-macro") {
					if jiraKey := extractJiraKey(selec, content); jiraKey != "" {
						return md.String(jiraKey)
					}
					return md.String("")
				}

				return nil
			},
		},
		md.Rule{
			Filter: []string{"a"},
			Replacement: func(content string, selec *goquery.Selection, _ *md.Options) *string {
				text := normalizeInlineText(content)
				if selec.HasClass("confluence-userlink") {
					return md.String(text)
				}

				href := strings.TrimSpace(selec.AttrOr("href", ""))
				if href == "" || href == "#" {
					if text == "" {
						return md.String("")
					}
					return md.String(text)
				}

				if looksLikeURL(text) {
					text = cleanConfluenceURL(text)
				}
				if text != "" && text == href {
					return md.String(cleanConfluenceURL(href))
				}
				if looksLikeURL(text) && text != "" {
					return md.String(fmt.Sprintf("[%s](%s)", text, cleanConfluenceURL(href)))
				}
				return nil
			},
		},
	)

	return conv
}

func preprocessConfluenceHTML(doc *goquery.Selection) {
	for _, selector := range []string{
		".toc-macro",
		".recently-updated",
		".plugin-contributors",
		"colgroup",
		"col",
	} {
		doc.Find(selector).Each(func(_ int, s *goquery.Selection) {
			s.Remove()
		})
	}

	for _, selector := range []string{
		"div.contentLayout2",
		"div.columnLayout",
		"div.cell",
		"div.innerCell",
		"div.table-wrap",
		"div.code.panel",
		"div.codeContent.panelContent",
	} {
		doc.Find(selector).Each(func(_ int, s *goquery.Selection) {
			s.ReplaceWithSelection(s.Contents())
		})
	}

	doc.Find("p").Each(func(_ int, s *goquery.Selection) {
		normalized := strings.ReplaceAll(s.Text(), "\u00a0", "")
		if strings.TrimSpace(normalized) == "" && s.Children().Length() == 0 {
			s.Remove()
		}
	})

	doc.Find("a[href]").Each(func(_ int, s *goquery.Selection) {
		href := strings.TrimSpace(s.AttrOr("href", ""))
		if href == "" {
			return
		}
		s.SetAttr("href", cleanConfluenceURL(href))
	})
}

func cleanConfluenceURL(raw string) string {
	parsed, err := url.Parse(strings.TrimSpace(raw))
	if err != nil {
		return raw
	}

	query := parsed.Query()
	query.Del("atlOrigin")
	query.Del("focusedCommentId")
	query.Del("src")
	parsed.RawQuery = query.Encode()

	return parsed.String()
}

func extractJiraKey(selec *goquery.Selection, content string) string {
	if key := strings.TrimSpace(selec.AttrOr("data-jira-key", "")); key != "" {
		return key
	}

	if key := strings.TrimSpace(selec.Find("[data-jira-key]").First().AttrOr("data-jira-key", "")); key != "" {
		return key
	}

	text := normalizeInlineText(content)
	if text == "" {
		return ""
	}

	return text
}

func normalizeInlineText(content string) string {
	return strings.Join(strings.Fields(strings.TrimSpace(content)), " ")
}

func looksLikeURL(v string) bool {
	v = strings.TrimSpace(v)
	if v == "" || strings.Contains(v, " ") {
		return false
	}
	return strings.HasPrefix(v, "http://") || strings.HasPrefix(v, "https://") || strings.HasPrefix(v, "/")
}
