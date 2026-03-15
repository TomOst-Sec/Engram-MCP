package parser

import (
	"crypto/sha256"
	"fmt"
	"regexp"
	"strings"
)

// ZigParser extracts symbols from Zig source files using regex-based fallback.
// Tree-sitter Zig grammar is not available as a Go binding, so this parser
// uses regular expressions for basic symbol extraction.
type ZigParser struct{}

// NewZigParser creates a new Zig parser.
func NewZigParser() *ZigParser { return &ZigParser{} }

func (p *ZigParser) Language() string     { return "zig" }
func (p *ZigParser) Extensions() []string { return []string{".zig"} }

var (
	zigFuncRe    = regexp.MustCompile(`(?m)^(pub\s+)?fn\s+(\w+)\(`)
	zigStructRe  = regexp.MustCompile(`(?m)^(pub\s+)?const\s+(\w+)\s*=\s*struct\s*\{`)
	zigEnumRe    = regexp.MustCompile(`(?m)^(pub\s+)?const\s+(\w+)\s*=\s*enum\s*\{`)
	zigUnionRe   = regexp.MustCompile(`(?m)^(pub\s+)?const\s+(\w+)\s*=\s*union\(`)
	zigErrorRe   = regexp.MustCompile(`(?m)^(pub\s+)?const\s+(\w+)\s*=\s*error\s*\{`)
	zigImportRe  = regexp.MustCompile(`(?m)^(pub\s+)?const\s+(\w+)\s*=\s*@import\("([^"]+)"\)`)
	zigTestRe    = regexp.MustCompile(`(?m)^test\s+"([^"]+)"\s*\{`)
	zigConstRe   = regexp.MustCompile(`(?m)^(pub\s+)?const\s+(\w+)\s*:\s*\w+\s*=`)
	zigDocRe     = regexp.MustCompile(`(?m)^///\s*(.*)$`)
)

func (p *ZigParser) Parse(filePath string, source []byte) ([]Symbol, error) {
	lines := strings.Split(string(source), "\n")
	var symbols []Symbol

	// Track doc comments
	var docLines []string

	for lineIdx, line := range lines {
		lineNum := lineIdx + 1

		// Accumulate doc comments
		if match := zigDocRe.FindStringSubmatch(line); match != nil {
			docLines = append(docLines, match[1])
			continue
		}

		doc := strings.Join(docLines, "\n")

		// Check imports first
		if match := zigImportRe.FindStringSubmatch(line); match != nil {
			symbols = append(symbols, Symbol{
				Name:      match[2],
				Type:      "import",
				Language:  "zig",
				Signature: strings.TrimSpace(line),
				StartLine: lineNum,
				EndLine:   lineNum,
				FilePath:  filePath,
			})
			docLines = nil
			continue
		}

		// Check test blocks
		if match := zigTestRe.FindStringSubmatch(line); match != nil {
			endLine := p.findBlockEnd(lines, lineIdx)
			symbols = append(symbols, Symbol{
				Name:      match[1],
				Type:      "test",
				Language:  "zig",
				Signature: strings.TrimSpace(line),
				StartLine: lineNum,
				EndLine:   endLine,
				FilePath:  filePath,
				BodyHash:  hashZigBlock(lines, lineIdx, endLine),
			})
			docLines = nil
			continue
		}

		// Check structs
		if match := zigStructRe.FindStringSubmatch(line); match != nil {
			endLine := p.findBlockEnd(lines, lineIdx)
			symbols = append(symbols, Symbol{
				Name:      match[2],
				Type:      "type",
				Language:  "zig",
				Signature: strings.TrimSpace(strings.TrimSuffix(line, "{")),
				Docstring: doc,
				StartLine: lineNum,
				EndLine:   endLine,
				FilePath:  filePath,
				BodyHash:  hashZigBlock(lines, lineIdx, endLine),
			})
			docLines = nil
			continue
		}

		// Check enums
		if match := zigEnumRe.FindStringSubmatch(line); match != nil {
			endLine := p.findBlockEnd(lines, lineIdx)
			symbols = append(symbols, Symbol{
				Name:      match[2],
				Type:      "enum",
				Language:  "zig",
				Signature: strings.TrimSpace(strings.TrimSuffix(line, "{")),
				Docstring: doc,
				StartLine: lineNum,
				EndLine:   endLine,
				FilePath:  filePath,
				BodyHash:  hashZigBlock(lines, lineIdx, endLine),
			})
			docLines = nil
			continue
		}

		// Check unions
		if match := zigUnionRe.FindStringSubmatch(line); match != nil {
			endLine := p.findBlockEnd(lines, lineIdx)
			symbols = append(symbols, Symbol{
				Name:      match[2],
				Type:      "type",
				Language:  "zig",
				Signature: strings.TrimSpace(line),
				Docstring: doc,
				StartLine: lineNum,
				EndLine:   endLine,
				FilePath:  filePath,
				BodyHash:  hashZigBlock(lines, lineIdx, endLine),
			})
			docLines = nil
			continue
		}

		// Check errors
		if match := zigErrorRe.FindStringSubmatch(line); match != nil {
			endLine := p.findBlockEnd(lines, lineIdx)
			symbols = append(symbols, Symbol{
				Name:      match[2],
				Type:      "type",
				Language:  "zig",
				Signature: strings.TrimSpace(line),
				Docstring: doc,
				StartLine: lineNum,
				EndLine:   endLine,
				FilePath:  filePath,
				BodyHash:  hashZigBlock(lines, lineIdx, endLine),
			})
			docLines = nil
			continue
		}

		// Check functions
		if match := zigFuncRe.FindStringSubmatch(line); match != nil {
			endLine := p.findBlockEnd(lines, lineIdx)
			sig := strings.TrimSpace(line)
			// Remove body
			if idx := strings.Index(sig, " {"); idx > 0 {
				sig = sig[:idx]
			}

			symbols = append(symbols, Symbol{
				Name:      match[2],
				Type:      "function",
				Language:  "zig",
				Signature: sig,
				Docstring: doc,
				StartLine: lineNum,
				EndLine:   endLine,
				FilePath:  filePath,
				BodyHash:  hashZigBlock(lines, lineIdx, endLine),
			})
			docLines = nil
			continue
		}

		// Check const declarations (non-type)
		if match := zigConstRe.FindStringSubmatch(line); match != nil {
			symbols = append(symbols, Symbol{
				Name:      match[2],
				Type:      "type",
				Language:  "zig",
				Signature: strings.TrimSpace(strings.TrimSuffix(line, ";")),
				Docstring: doc,
				StartLine: lineNum,
				EndLine:   lineNum,
				FilePath:  filePath,
			})
			docLines = nil
			continue
		}

		// Reset doc comments if line doesn't match anything
		docLines = nil
	}

	return symbols, nil
}

// findBlockEnd finds the matching closing brace for a block starting at lineIdx.
func (p *ZigParser) findBlockEnd(lines []string, startIdx int) int {
	depth := 0
	for i := startIdx; i < len(lines); i++ {
		depth += strings.Count(lines[i], "{") - strings.Count(lines[i], "}")
		if depth <= 0 {
			return i + 1 // 1-based
		}
	}
	return len(lines)
}

func hashZigBlock(lines []string, startIdx, endLine int) string {
	end := endLine - 1 // convert to 0-based
	if end > len(lines) {
		end = len(lines)
	}
	block := strings.Join(lines[startIdx:end], "\n")
	h := sha256.Sum256([]byte(block))
	return fmt.Sprintf("%x", h)
}
