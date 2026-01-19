package utils

import (
	"strings"
	"testing"
)

func TestRenderMermaidBlocks_PlainMode(t *testing.T) {
	input := "# Hello\n\n```mermaid\ngraph LR\nA --> B\n```\n\nMore text"
	result := RenderMermaidBlocks(input, "plain")
	if result != input {
		t.Errorf("plain mode should return input unchanged\ngot: %s\nwant: %s", result, input)
	}
}

func TestRenderMermaidBlocks_AsciiMode_SimpleGraph(t *testing.T) {
	input := "# Hello\n\n```mermaid\ngraph LR\nA --> B\n```\n\nMore text"
	result := RenderMermaidBlocks(input, "ascii")

	// Should not contain the original mermaid block
	if strings.Contains(result, "```mermaid") {
		t.Error("ascii mode should replace mermaid blocks")
	}
	// Should still have surrounding content
	if !strings.Contains(result, "# Hello") {
		t.Error("should preserve content before mermaid block")
	}
	if !strings.Contains(result, "More text") {
		t.Error("should preserve content after mermaid block")
	}
	// Should contain box-drawing characters (the rendered diagram)
	if !strings.Contains(result, "─") && !strings.Contains(result, "-") {
		t.Error("should contain rendered diagram with box characters")
	}
}

func TestRenderMermaidBlocks_MultipleMermaidBlocks(t *testing.T) {
	input := "```mermaid\ngraph LR\nA --> B\n```\n\nText\n\n```mermaid\ngraph TD\nC --> D\n```"
	result := RenderMermaidBlocks(input, "ascii")

	// Count occurrences of mermaid - should be zero
	if strings.Contains(result, "```mermaid") {
		t.Error("all mermaid blocks should be replaced")
	}
}

func TestRenderMermaidBlocks_NonMermaidCodeBlocks(t *testing.T) {
	input := "```go\nfunc main() {}\n```\n\n```mermaid\ngraph LR\nA --> B\n```"
	result := RenderMermaidBlocks(input, "ascii")

	// Go block should remain unchanged
	if !strings.Contains(result, "```go") {
		t.Error("non-mermaid code blocks should remain unchanged")
	}
	if !strings.Contains(result, "func main()") {
		t.Error("non-mermaid code block content should remain unchanged")
	}
}

func TestRenderMermaidBlocks_TildeFence(t *testing.T) {
	input := "~~~mermaid\ngraph LR\nA --> B\n~~~"
	result := RenderMermaidBlocks(input, "ascii")

	if strings.Contains(result, "~~~mermaid") {
		t.Error("tilde-fenced mermaid blocks should be replaced")
	}
}

func TestRenderMermaidBlocks_MermaidWithExtraInfo(t *testing.T) {
	// Some markdown processors allow extra info after language
	input := "```mermaid some-extra-info\ngraph LR\nA --> B\n```"
	result := RenderMermaidBlocks(input, "ascii")

	if strings.Contains(result, "```mermaid") {
		t.Error("mermaid blocks with extra info should be replaced")
	}
}

func TestRenderMermaidBlocks_InvalidMermaid_Fallback(t *testing.T) {
	// Invalid mermaid syntax should fall back to original with visible error
	input := "```mermaid\nthis is not valid mermaid syntax @@##$$\n```"
	result := RenderMermaidBlocks(input, "ascii")

	// Should have visible error message (not HTML comment)
	if !strings.Contains(result, "mermaid render error:") {
		t.Error("invalid mermaid should include visible error message")
	}
	// Should NOT use HTML comment format
	if strings.Contains(result, "<!--") {
		t.Error("error should not use HTML comment format")
	}
	// Should keep original block content
	if !strings.Contains(result, "```mermaid") || !strings.Contains(result, "this is not valid") {
		t.Error("invalid mermaid should fall back to original block")
	}
}

func TestRenderMermaidBlocks_SequenceDiagram(t *testing.T) {
	input := "```mermaid\nsequenceDiagram\nAlice->>Bob: Hello\n```"
	result := RenderMermaidBlocks(input, "ascii")

	if strings.Contains(result, "```mermaid") {
		t.Error("sequence diagrams should be rendered")
	}
	// Should contain participant names
	if !strings.Contains(result, "Alice") || !strings.Contains(result, "Bob") {
		t.Error("rendered sequence diagram should contain participant names")
	}
}

func TestRenderMermaidBlocks_EmptyContent(t *testing.T) {
	result := RenderMermaidBlocks("", "ascii")
	if result != "" {
		t.Error("empty input should return empty output")
	}
}

func TestRenderMermaidBlocks_NoMermaidBlocks(t *testing.T) {
	input := "# Just markdown\n\nNo mermaid here."
	result := RenderMermaidBlocks(input, "ascii")
	if result != input {
		t.Error("content without mermaid blocks should be unchanged")
	}
}

func TestRenderMermaidBlocks_CRLF_NoMermaid(t *testing.T) {
	// CRLF input with no mermaid blocks should return unchanged (including CRLF)
	input := "# Just markdown\r\n\r\nNo mermaid here.\r\n"
	result := RenderMermaidBlocks(input, "ascii")
	if result != input {
		t.Errorf("CRLF content without mermaid should be unchanged\ngot: %q\nwant: %q", result, input)
	}
}

func TestRenderMermaidBlocks_NestedFence_ShouldNotRender(t *testing.T) {
	// Mermaid block inside another code block (e.g., markdown example) should NOT be rendered
	input := "````markdown\n```mermaid\ngraph LR\nA --> B\n```\n````"
	result := RenderMermaidBlocks(input, "ascii")

	// The inner mermaid block should remain unchanged
	if !strings.Contains(result, "```mermaid") {
		t.Error("mermaid block nested inside another fence should NOT be rendered")
	}
	if !strings.Contains(result, "graph LR") {
		t.Error("nested mermaid content should remain unchanged")
	}
}

func TestRenderMermaidBlocks_IndentedInList(t *testing.T) {
	// Mermaid block indented inside a list item
	input := "- Item\n  ```mermaid\n  graph LR\n  A --> B\n  ```"
	result := RenderMermaidBlocks(input, "ascii")

	// Should render but preserve indentation
	if strings.Contains(result, "```mermaid") {
		t.Error("indented mermaid block should be rendered")
	}
	// Result should have indented lines (list structure preserved)
	lines := strings.Split(result, "\n")
	foundIndented := false
	for _, line := range lines {
		if strings.HasPrefix(line, "  ") && strings.Contains(line, "─") {
			foundIndented = true
			break
		}
	}
	if !foundIndented {
		t.Error("rendered diagram should preserve list indentation")
	}
}

func TestRenderMermaidBlocks_CRLF(t *testing.T) {
	// Windows-style line endings
	input := "```mermaid\r\ngraph LR\r\nA --> B\r\n```\r\n"
	result := RenderMermaidBlocks(input, "ascii")

	if strings.Contains(result, "```mermaid") {
		t.Error("CRLF input should be handled correctly")
	}
}

func TestRenderMermaidBlocks_LongerClosingFence(t *testing.T) {
	// Closing fence can be longer than opening fence
	input := "```mermaid\ngraph LR\nA --> B\n`````"
	result := RenderMermaidBlocks(input, "ascii")

	if strings.Contains(result, "```mermaid") {
		t.Error("longer closing fence should be valid")
	}
}

func TestRenderMermaidBlocks_NoClosingFence(t *testing.T) {
	// Unclosed fence should be left unchanged
	input := "```mermaid\ngraph LR\nA --> B"
	result := RenderMermaidBlocks(input, "ascii")

	if result != input {
		t.Error("unclosed fence should be left unchanged")
	}
}

func TestRenderMermaidBlocks_CaseInsensitive(t *testing.T) {
	input := "```MERMAID\ngraph LR\nA --> B\n```"
	result := RenderMermaidBlocks(input, "ascii")

	if strings.Contains(result, "```MERMAID") {
		t.Error("MERMAID (uppercase) should be recognized")
	}
}
