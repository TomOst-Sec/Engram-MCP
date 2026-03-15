package architecture

import (
	"context"
	"encoding/json"
	"testing"

	mcpgo "github.com/mark3labs/mcp-go/mcp"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestHandlerReturnsFullProject(t *testing.T) {
	store := setupTestDB(t)

	// Set up data
	insertSymbol(t, store, "internal/auth/handler.go", "go", "HandleLogin", "function", 10, 30)
	insertSymbol(t, store, "internal/auth/middleware.go", "go", "AuthMiddleware", "function", 5, 20)
	insertSymbol(t, store, "internal/storage/store.go", "go", "Open", "function", 10, 50)

	tool := NewArchitectureTool(store, "/repo", "github.com/TomOst-Sec/colony-project")

	req := mcpgo.CallToolRequest{}
	req.Params.Arguments = map[string]interface{}{}

	result, err := tool.Handle(context.Background(), req)
	require.NoError(t, err)
	require.NotNil(t, result)

	// Parse the result
	var response map[string]interface{}
	err = json.Unmarshal([]byte(result.Content[0].(mcpgo.TextContent).Text), &response)
	require.NoError(t, err)

	assert.Equal(t, "/repo", response["project_root"])
	assert.NotNil(t, response["modules"])
	assert.NotNil(t, response["dependency_graph"])
	assert.NotNil(t, response["total_modules"])
	assert.NotNil(t, response["total_files"])

	modules := response["modules"].([]interface{})
	assert.Equal(t, 2, len(modules))
}

func TestHandlerSingleModule(t *testing.T) {
	store := setupTestDB(t)

	insertSymbol(t, store, "internal/auth/handler.go", "go", "HandleLogin", "function", 10, 30)
	insertSymbol(t, store, "internal/auth/middleware.go", "go", "AuthMiddleware", "function", 5, 20)
	insertSymbol(t, store, "internal/storage/store.go", "go", "Open", "function", 10, 50)

	tool := NewArchitectureTool(store, "/repo", "github.com/TomOst-Sec/colony-project")

	req := mcpgo.CallToolRequest{}
	req.Params.Arguments = map[string]interface{}{
		"module": "internal/auth",
	}

	result, err := tool.Handle(context.Background(), req)
	require.NoError(t, err)
	require.NotNil(t, result)

	var response map[string]interface{}
	err = json.Unmarshal([]byte(result.Content[0].(mcpgo.TextContent).Text), &response)
	require.NoError(t, err)

	module := response["module"].(map[string]interface{})
	assert.Equal(t, "internal/auth", module["name"])
}

func TestHandlerModuleNotFound(t *testing.T) {
	store := setupTestDB(t)

	insertSymbol(t, store, "internal/auth/handler.go", "go", "HandleLogin", "function", 10, 30)

	tool := NewArchitectureTool(store, "/repo", "")

	req := mcpgo.CallToolRequest{}
	req.Params.Arguments = map[string]interface{}{
		"module": "nonexistent/module",
	}

	result, err := tool.Handle(context.Background(), req)
	require.NoError(t, err)
	require.NotNil(t, result)
	assert.True(t, result.IsError)
}

func TestHandlerEmptyDatabase(t *testing.T) {
	store := setupTestDB(t)

	tool := NewArchitectureTool(store, "/repo", "")

	req := mcpgo.CallToolRequest{}
	req.Params.Arguments = map[string]interface{}{}

	result, err := tool.Handle(context.Background(), req)
	require.NoError(t, err)
	require.NotNil(t, result)

	var response map[string]interface{}
	err = json.Unmarshal([]byte(result.Content[0].(mcpgo.TextContent).Text), &response)
	require.NoError(t, err)

	assert.Equal(t, float64(0), response["total_modules"])
	modules := response["modules"].([]interface{})
	assert.Empty(t, modules)
}

func TestHandlerIncludeExports(t *testing.T) {
	store := setupTestDB(t)

	insertSymbol(t, store, "internal/auth/handler.go", "go", "HandleLogin", "function", 10, 30)
	insertSymbol(t, store, "internal/auth/handler.go", "go", "handlePrivate", "function", 35, 45)

	tool := NewArchitectureTool(store, "/repo", "")

	req := mcpgo.CallToolRequest{}
	req.Params.Arguments = map[string]interface{}{
		"include_exports": true,
	}

	result, err := tool.Handle(context.Background(), req)
	require.NoError(t, err)

	var response map[string]interface{}
	err = json.Unmarshal([]byte(result.Content[0].(mcpgo.TextContent).Text), &response)
	require.NoError(t, err)

	modules := response["modules"].([]interface{})
	require.Len(t, modules, 1)

	mod := modules[0].(map[string]interface{})
	exports := mod["exports"].([]interface{})
	assert.Contains(t, exports, "HandleLogin")
	// handlePrivate should NOT be in exports (lowercase in Go)
	assert.NotContains(t, exports, "handlePrivate")
}

func TestHandlerCustomDepth(t *testing.T) {
	store := setupTestDB(t)

	insertSymbol(t, store, "internal/auth/handler.go", "go", "HandleLogin", "function", 10, 30)
	insertSymbol(t, store, "internal/storage/store.go", "go", "Open", "function", 10, 50)
	insertSymbol(t, store, "cmd/engram/main.go", "go", "main", "function", 1, 10)

	tool := NewArchitectureTool(store, "/repo", "")

	req := mcpgo.CallToolRequest{}
	req.Params.Arguments = map[string]interface{}{
		"depth": float64(1),
	}

	result, err := tool.Handle(context.Background(), req)
	require.NoError(t, err)

	var response map[string]interface{}
	err = json.Unmarshal([]byte(result.Content[0].(mcpgo.TextContent).Text), &response)
	require.NoError(t, err)

	// At depth 1: "internal" and "cmd"
	assert.Equal(t, float64(2), response["total_modules"])
}

func TestToolDefinition(t *testing.T) {
	store := setupTestDB(t)
	tool := NewArchitectureTool(store, "/repo", "")
	def := tool.Definition()
	assert.Equal(t, "get_architecture", def.Name)
}
