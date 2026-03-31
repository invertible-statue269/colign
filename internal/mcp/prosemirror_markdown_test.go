package mcp

import (
	"encoding/json"
	"strings"
	"testing"
)

func TestProseMirrorJSONToMarkdown(t *testing.T) {
	input := `{
		"type":"doc",
		"content":[
			{"type":"heading","attrs":{"level":2},"content":[{"type":"text","text":"Design"}]},
			{"type":"paragraph","content":[
				{"type":"text","text":"Use "},
				{"type":"text","text":"Client","marks":[{"type":"code"}]}
			]},
			{"type":"codeBlock","attrs":{"language":"go"},"content":[{"type":"text","text":"fmt.Println(1)"}]}
		]
	}`

	got, err := proseMirrorJSONToMarkdown(input)
	if err != nil {
		t.Fatalf("proseMirrorJSONToMarkdown returned error: %v", err)
	}

	for _, want := range []string{
		"## Design",
		"Use `Client`",
		"```go",
		"fmt.Println(1)",
	} {
		if !strings.Contains(got, want) {
			t.Fatalf("expected markdown to contain %q, got %q", want, got)
		}
	}
}

func TestConvertProposalToMarkdown(t *testing.T) {
	input := `{"problem":"<p>사용자 의도를 파악하기 어렵다</p>","scope":"<ul><li><p>AI Assistant Side Panel 추가</p></li><li><p>대화형 Proposal 생성</p></li></ul>","outOfScope":"<ul><li><p>모바일 반응형 레이아웃</p></li></ul>"}`

	got, err := convertProposalToMarkdown(input)
	if err != nil {
		t.Fatalf("convertProposalToMarkdown returned error: %v", err)
	}

	var result map[string]any
	if err := json.Unmarshal([]byte(got), &result); err != nil {
		t.Fatalf("result is not valid JSON: %v", err)
	}

	// HTML tags should be removed
	for _, key := range []string{"problem", "scope", "outOfScope"} {
		val, ok := result[key].(string)
		if !ok {
			t.Fatalf("expected %q to be a string", key)
		}
		if strings.Contains(val, "<p>") || strings.Contains(val, "<ul>") || strings.Contains(val, "<li>") {
			t.Fatalf("%q still contains HTML tags: %q", key, val)
		}
	}

	// Content should be preserved as markdown
	problem := result["problem"].(string)
	if !strings.Contains(problem, "사용자 의도를 파악하기 어렵다") {
		t.Fatalf("problem text lost: %q", problem)
	}

	scope := result["scope"].(string)
	if !strings.Contains(scope, "AI Assistant Side Panel") {
		t.Fatalf("scope text lost: %q", scope)
	}
	if !strings.Contains(scope, "- ") {
		t.Fatalf("scope should contain markdown list markers: %q", scope)
	}
}

func TestConvertProposalToMarkdownPreservesAngleBrackets(t *testing.T) {
	input := `{"problem":"latency < 100ms is required","scope":"use <https://example.com> for reference"}`

	got, err := convertProposalToMarkdown(input)
	if err != nil {
		t.Fatalf("convertProposalToMarkdown returned error: %v", err)
	}

	var result map[string]any
	if err := json.Unmarshal([]byte(got), &result); err != nil {
		t.Fatalf("result is not valid JSON: %v", err)
	}

	// Plain text with < should NOT be treated as HTML
	if result["problem"] != "latency < 100ms is required" {
		t.Fatalf("angle bracket text was corrupted: %q", result["problem"])
	}
	if result["scope"] != "use <https://example.com> for reference" {
		t.Fatalf("angle bracket text was corrupted: %q", result["scope"])
	}
}

func TestConvertProposalToMarkdownPlainText(t *testing.T) {
	input := `{"problem":"plain text problem","scope":"plain text scope"}`

	got, err := convertProposalToMarkdown(input)
	if err != nil {
		t.Fatalf("convertProposalToMarkdown returned error: %v", err)
	}

	var result map[string]any
	if err := json.Unmarshal([]byte(got), &result); err != nil {
		t.Fatalf("result is not valid JSON: %v", err)
	}

	if result["problem"] != "plain text problem" {
		t.Fatalf("plain text should be unchanged, got %q", result["problem"])
	}
}

func TestConvertProposalFieldsToMarkdown(t *testing.T) {
	proposal := map[string]any{
		"problem":    "<p>문제 설명</p>",
		"scope":      "<ul><li><p>항목 1</p></li></ul>",
		"outOfScope": "이건 plain text",
	}

	convertProposalFieldsToMarkdown(proposal)

	if strings.Contains(proposal["problem"].(string), "<p>") {
		t.Fatalf("problem still has HTML: %q", proposal["problem"])
	}
	if strings.Contains(proposal["scope"].(string), "<ul>") {
		t.Fatalf("scope still has HTML: %q", proposal["scope"])
	}
	// plain text without < should remain unchanged
	if proposal["outOfScope"] != "이건 plain text" {
		t.Fatalf("outOfScope should be unchanged: %q", proposal["outOfScope"])
	}
}

func TestExportDocumentToMarkdownHTMLFallback(t *testing.T) {
	got, err := exportDocumentToMarkdown("<h2>Design</h2><p>Hello <code>world()</code></p>")
	if err != nil {
		t.Fatalf("exportDocumentToMarkdown returned error: %v", err)
	}

	for _, want := range []string{"## Design", "Hello `world()`"} {
		if !strings.Contains(got, want) {
			t.Fatalf("expected markdown to contain %q, got %q", want, got)
		}
	}
}
