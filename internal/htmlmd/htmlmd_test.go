package htmlmd

import (
	"strings"
	"testing"
)

func TestConvert_ConfluenceViewHTML(t *testing.T) {
	input := `
<div class="contentLayout2">
  <div class="toc-macro"></div>
  <h2>Overview</h2>
  <p>Status: <span class="status-macro aui-lozenge">LIVE</span></p>
  <p><strong>Bold</strong> <em>Italic</em> <del>Old</del> and <code>inline()</code></p>
  <p><a class="confluence-userlink" href="/wiki/display/~abc">Jane Doe</a> updated <span class="confluence-jim-macro" data-jira-key="CF-42"><a href="https://jira.example/browse/CF-42">CF-42</a></span> on <time datetime="2025-10-01">01 Oct 2025</time></p>
  <div class="table-wrap">
    <table>
      <thead><tr><th>Name</th><th>Value</th></tr></thead>
      <tbody><tr><td>A</td><td>1</td></tr></tbody>
    </table>
  </div>
  <p><a href="https://example.com/path?atlOrigin=abc">https://example.com/path?atlOrigin=abc</a></p>
  <p><img alt="Diagram" src="/wiki/download/attachments/1/diag.png"/></p>
</div>`

	out, err := Convert(input)
	if err != nil {
		t.Fatalf("Convert() error = %v", err)
	}

	mustContain(t, out, "## Overview")
	mustContain(t, out, "[LIVE]")
	mustContain(t, out, "**Bold**")
	mustContain(t, out, "_Italic_")
	mustContain(t, out, "~~Old~~")
	mustContain(t, out, "`inline()`")
	mustContain(t, out, "Jane Doe")
	mustContain(t, out, "CF-42")
	mustContain(t, out, "2025-10-01")
	mustContain(t, out, "| Name | Value |")
	mustContain(t, out, "https://example.com/path")

	mustNotContain(t, out, "atlOrigin=")
	mustNotContain(t, out, "[Jane Doe](")
	mustNotContain(t, out, "toc-macro")
	mustNotContain(t, out, "![Diagram]")
	mustNotContain(t, out, "<h2>")
}

func TestConvert_RemovesHighNoiseMacros(t *testing.T) {
	input := `
<div>
  <p>Important intro.</p>
  <div class="recently-updated conf-macro output-block">
    <ul><li>Very long noisy feed item 1</li><li>Very long noisy feed item 2</li></ul>
  </div>
  <div class="plugin-contributors conf-macro output-block">
    <ul><li>Alice</li><li>Bob</li></ul>
  </div>
  <p>Important outro.</p>
</div>`

	out, err := Convert(input)
	if err != nil {
		t.Fatalf("Convert() error = %v", err)
	}

	mustContain(t, out, "Important intro.")
	mustContain(t, out, "Important outro.")
	mustNotContain(t, out, "noise")
	mustNotContain(t, out, "contributors")

	if len(out) >= len(strings.TrimSpace(input)) {
		t.Fatalf("expected converted output to be smaller; in=%d out=%d", len(strings.TrimSpace(input)), len(out))
	}
}

func TestConvert_TableCellLineBreaksDoNotLeaveHTML(t *testing.T) {
	input := `
<table>
  <thead><tr><th>Field</th><th>Details</th></tr></thead>
  <tbody>
    <tr><td>Example</td><td>Line one<br>Line two<br/>Line three</td></tr>
  </tbody>
</table>`

	out, err := Convert(input)
	if err != nil {
		t.Fatalf("Convert() error = %v", err)
	}

	mustContain(t, out, "| Field | Details |")
	mustContain(t, out, "Line one / Line two / Line three")
	mustNotContain(t, out, "<br>")
}

func mustContain(t *testing.T, got, want string) {
	t.Helper()
	if !strings.Contains(got, want) {
		t.Fatalf("expected output to contain %q\noutput:\n%s", want, got)
	}
}

func mustNotContain(t *testing.T, got, want string) {
	t.Helper()
	if strings.Contains(got, want) {
		t.Fatalf("expected output to NOT contain %q\noutput:\n%s", want, got)
	}
}
