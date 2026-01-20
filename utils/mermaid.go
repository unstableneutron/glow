package utils

import (
	"strings"

	mermaidcmd "github.com/AlexanderGrooff/mermaid-ascii/cmd"
)

// RenderMermaidBlocks processes markdown content and renders mermaid code blocks.
// Mode "raw" returns content unchanged; "ascii" and "unicode" render diagrams.
func RenderMermaidBlocks(content string, mode string, maxWidth int) string {
	if content == "" {
		return content
	}

	mode = strings.ToLower(mode)
	if mode == "raw" {
		return content
	}

	useAscii := mode == "ascii"
	return processMermaidBlocks(content, maxWidth, useAscii)
}

// fencedBlock represents a parsed fenced code block.
type fencedBlock struct {
	startLine    int    // line index where block starts
	endLine      int    // line index where block ends (inclusive)
	fenceChar    rune   // '`' or '~'
	fenceLen     int    // length of fence (>= 3)
	indentPrefix string // leading whitespace (up to 3 spaces)
	infoString   string // language/info after fence
	content      string // content inside the block
}

// processMermaidBlocks uses a line-scanner to find and replace mermaid blocks.
// This correctly handles nested fences, indentation, and CRLF line endings.
func processMermaidBlocks(content string, maxWidth int, useAscii bool) string {
	// First check if we have any mermaid blocks before normalizing
	// Normalize CRLF to LF for consistent processing
	normalized := strings.ReplaceAll(content, "\r\n", "\n")

	lines := strings.Split(normalized, "\n")
	blocks := findMermaidBlocks(lines)

	// If no mermaid blocks found, return original content unchanged
	if len(blocks) == 0 {
		return content
	}

	// Process blocks in reverse order to preserve line indices
	for i := len(blocks) - 1; i >= 0; i-- {
		block := blocks[i]
		rendered := renderMermaidBlock(block, maxWidth, useAscii)
		lines = replaceLines(lines, block.startLine, block.endLine, rendered)
	}

	return strings.Join(lines, "\n")
}

// findMermaidBlocks scans lines and returns all top-level mermaid fenced blocks.
// Blocks nested inside other fenced blocks are ignored.
func findMermaidBlocks(lines []string) []fencedBlock {
	var blocks []fencedBlock
	var currentBlock *fencedBlock
	inFence := false
	var fenceChar rune
	var fenceLen int

	for i, line := range lines {
		// Check if this line is a fence
		indent, char, length, info := parseFenceLine(line)

		if !inFence {
			// Not currently in a fence - check for opening fence
			if length >= 3 {
				inFence = true
				fenceChar = char
				fenceLen = length

				// Check if this is a mermaid block (case-insensitive)
				infoToken := strings.Fields(info)
				if len(infoToken) > 0 && strings.EqualFold(infoToken[0], "mermaid") {
					currentBlock = &fencedBlock{
						startLine:    i,
						fenceChar:    char,
						fenceLen:     length,
						indentPrefix: indent,
						infoString:   info,
					}
				}
			}
		} else {
			// Currently in a fence - check for closing fence
			// Closing fence must use same char and length >= opening length
			if char == fenceChar && length >= fenceLen && strings.TrimSpace(info) == "" {
				if currentBlock != nil {
					// End of a mermaid block
					currentBlock.endLine = i
					// Extract content (lines between start and end)
					var contentLines []string
					for j := currentBlock.startLine + 1; j < i; j++ {
						// Remove the indent prefix from content lines
						contentLine := lines[j]
						if strings.HasPrefix(contentLine, currentBlock.indentPrefix) {
							contentLine = contentLine[len(currentBlock.indentPrefix):]
						}
						contentLines = append(contentLines, contentLine)
					}
					currentBlock.content = strings.Join(contentLines, "\n")
					blocks = append(blocks, *currentBlock)
					currentBlock = nil
				}
				inFence = false
				fenceChar = 0
				fenceLen = 0
			}
		}
	}

	return blocks
}

// parseFenceLine checks if a line is a fence line.
// Returns: indent prefix, fence char, fence length, info string.
// If not a fence line, returns length=0.
func parseFenceLine(line string) (indent string, char rune, length int, info string) {
	// Count leading spaces (up to 3 allowed for fenced code blocks)
	spaces := 0
	for _, c := range line {
		if c == ' ' && spaces < 3 {
			spaces++
		} else {
			break
		}
	}
	indent = line[:spaces]
	rest := line[spaces:]

	if len(rest) < 3 {
		return indent, 0, 0, ""
	}

	// Check for fence character
	firstChar := rune(rest[0])
	if firstChar != '`' && firstChar != '~' {
		return indent, 0, 0, ""
	}

	// Count consecutive fence characters
	fenceCount := 0
	for _, c := range rest {
		if c == firstChar {
			fenceCount++
		} else {
			break
		}
	}

	if fenceCount < 3 {
		return indent, 0, 0, ""
	}

	// Info string is everything after the fence chars
	info = strings.TrimSpace(rest[fenceCount:])

	// Backtick fences cannot have backticks in info string
	if firstChar == '`' && strings.Contains(info, "`") {
		return indent, 0, 0, ""
	}

	return indent, firstChar, fenceCount, info
}

// renderMermaidBlock renders a mermaid block to ASCII and returns replacement lines.
func renderMermaidBlock(block fencedBlock, maxWidth int, useAscii bool) []string {
	availableWidth := maxWidth
	if availableWidth > 0 {
		availableWidth -= len(block.indentPrefix)
		const codeBlockMargin = 4
		if availableWidth > codeBlockMargin {
			availableWidth -= codeBlockMargin
		} else {
			availableWidth = 0
		}
	}
	options := []mermaidcmd.RenderOption{mermaidcmd.WithMaxWidth(availableWidth)}
	if useAscii {
		options = append(options, mermaidcmd.WithAscii())
	}
	rendered, err := mermaidcmd.RenderDiagramWithOptions(block.content, options...)
	if err != nil {
		// On error, show visible error message and keep original block
		var result []string
		result = append(result, block.indentPrefix+"```")
		result = append(result, block.indentPrefix+"mermaid render error: "+err.Error())
		result = append(result, block.indentPrefix+"```")
		result = append(result, block.indentPrefix+strings.Repeat(string(block.fenceChar), block.fenceLen)+block.infoString)
		for _, line := range strings.Split(block.content, "\n") {
			result = append(result, block.indentPrefix+line)
		}
		result = append(result, block.indentPrefix+strings.Repeat(string(block.fenceChar), block.fenceLen))
		return result
	}

	// Wrap rendered output in a plain code block, preserving indentation
	rendered = strings.TrimRight(rendered, "\n\r\t ")
	var result []string
	result = append(result, block.indentPrefix+"```")
	for _, line := range strings.Split(rendered, "\n") {
		result = append(result, block.indentPrefix+line)
	}
	result = append(result, block.indentPrefix+"```")
	return result
}

// replaceLines replaces lines[start:end+1] with newLines.
func replaceLines(lines []string, start, end int, newLines []string) []string {
	result := make([]string, 0, len(lines)-end+start-1+len(newLines))
	result = append(result, lines[:start]...)
	result = append(result, newLines...)
	result = append(result, lines[end+1:]...)
	return result
}
