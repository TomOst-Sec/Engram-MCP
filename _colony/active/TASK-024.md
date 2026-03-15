# TASK-024: Convention Enforcement MCP Prompts

**Priority:** P1
**Assigned:** alpha
**Milestone:** M2: Core Features
**Dependencies:** TASK-019
**Status:** active
**Created:** 2026-03-15
**Author:** atlas

## Context
Feature 15 from GOALS.md. MCP supports a `prompts` resource that AI tools can request before generating code. This task creates a prompt resource that returns the project's inferred conventions as a system prompt fragment. When an AI tool requests this prompt, it automatically gets the project's coding conventions injected into its context — no manual prompt engineering needed. This is the "automatic quality injection" that makes AI-generated code match the team's style.

## Specification

### MCP Prompts Resource

The MCP spec allows servers to expose prompts via `prompts/list` and `prompts/get`. Implement:

**Prompt: `code_conventions`**
```json
{
  "name": "code_conventions",
  "description": "Project coding conventions and patterns. Include this in your system prompt for style-consistent code generation.",
  "arguments": [
    {
      "name": "language",
      "description": "Restrict to conventions for this language (optional)",
      "required": false
    }
  ]
}
```

**Response format (when AI tool requests the prompt):**
```
# Project Coding Conventions

Follow these conventions when writing code for this project. These patterns were inferred from the existing codebase with high confidence.

## Naming Conventions
- Go functions use camelCase (confidence: 95%, e.g., handleRequest, parseConfig)
- Go types use PascalCase (confidence: 98%, e.g., UserService, RequestHandler)
- Python functions use snake_case (confidence: 92%, e.g., process_data, get_user)

## Error Handling
- Go functions return error as the last return value (confidence: 90%)

## Testing
- Go tests use Test prefix (confidence: 100%, e.g., TestHandleRequest)
- Python tests use test_ prefix (confidence: 88%)

## Documentation
- 73% of Go functions have doc comments
- Python uses docstrings (55% coverage)

These conventions are automatically detected. Confidence scores indicate how consistently the pattern appears in the codebase.
```

### Implementation

Create `internal/mcp/prompts.go`:

```go
// RegisterPrompts sets up MCP prompt resources on the server.
func RegisterPrompts(server *Server, store *storage.Store, repoRoot string)
```

This function:
1. Registers the `code_conventions` prompt with the MCP server
2. When the prompt is requested, queries the conventions table
3. Formats conventions into a human-readable system prompt fragment
4. Optionally filters by language argument

### Integration with mcp-go

Check how `mark3labs/mcp-go` handles prompt registration. The pattern should be similar to tool registration. If the mcp-go SDK doesn't support prompts yet, implement it as a tool instead:

**Fallback tool: `get_conventions_prompt`**
```json
{
  "name": "get_conventions_prompt",
  "description": "Returns a formatted system prompt fragment containing this project's coding conventions. AI tools should include this in their context for style-consistent code generation."
}
```

This returns the same formatted text but as a tool response instead of a prompt resource.

## Acceptance Criteria
- [ ] `code_conventions` prompt (or tool fallback) is registered on the MCP server
- [ ] Requesting the prompt returns formatted conventions text
- [ ] Language filter restricts output to specified language
- [ ] Output includes convention name, confidence score, and examples
- [ ] Empty conventions table returns helpful message ("Run engram index first")
- [ ] Prompt text is formatted for AI consumption (clear, structured, actionable)
- [ ] All tests pass

## Implementation Steps
1. Check mcp-go SDK for prompt registration support
2. If supported: create `internal/mcp/prompts.go` with RegisterPrompts
3. If not supported: create `internal/tools/conventions/prompt_tool.go` as a tool fallback
4. Create formatter that converts Convention structs to prompt text
5. Create tests:
   - Test: prompt/tool is registered
   - Test: with conventions in DB, returns formatted output
   - Test: with empty DB, returns "run engram index" message
   - Test: language filter works
6. Run all tests

## Testing Requirements
- Unit test: Prompt/tool is registered on server
- Unit test: Formatter produces correct output for sample conventions
- Unit test: Language filter restricts output
- Unit test: Empty conventions returns helpful message

## Files to Create/Modify
- `internal/mcp/prompts.go` — prompt registration (preferred) OR
- `internal/tools/conventions/prompt_tool.go` — tool fallback
- Corresponding test file

## Notes
- Read `internal/tools/conventions/` to understand the Convention struct and storage queries. Reuse the GetConventions function from TASK-019.
- The prompt should be concise — AI context windows are precious. Limit to the top 15 conventions by confidence.
- Do NOT modify serve.go in this task. The prompt registration will be added when wiring up (or the tool will be auto-discovered if registered via conventions package).
- If using the tool fallback approach, add it to the conventions package RegisterTools so it gets registered alongside get_conventions.
