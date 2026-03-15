package mcp

import (
	"context"
	"encoding/json"
	"testing"

	mcpgo "github.com/mark3labs/mcp-go/mcp"
	"github.com/mark3labs/mcp-go/client"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestNew(t *testing.T) {
	s := New("engram", "0.1.0-dev")
	require.NotNil(t, s)
	assert.Equal(t, "0.1.0-dev", s.version)
	assert.NotNil(t, s.mcpServer)
	assert.False(t, s.startTime.IsZero())
}

func TestRegisterTool(t *testing.T) {
	s := New("engram", "0.1.0-dev")
	tool := mcpgo.NewTool("test_tool",
		mcpgo.WithDescription("A test tool"),
	)
	// Should not panic
	s.RegisterTool(tool, func(ctx context.Context, req mcpgo.CallToolRequest) (*mcpgo.CallToolResult, error) {
		return mcpgo.NewToolResultText("ok"), nil
	})
}

func TestEngramStatusHandler(t *testing.T) {
	s := New("engram", "0.1.0-dev")
	RegisterBuiltinTools(s)

	result, err := s.handleStatus(context.Background(), mcpgo.CallToolRequest{})
	require.NoError(t, err)
	require.NotNil(t, result)

	require.Len(t, result.Content, 1)
	textContent, ok := mcpgo.AsTextContent(result.Content[0])
	require.True(t, ok)

	var response map[string]interface{}
	err = json.Unmarshal([]byte(textContent.Text), &response)
	require.NoError(t, err)

	assert.Equal(t, "0.1.0-dev", response["version"])
	assert.Equal(t, "healthy", response["status"])

	uptimeSeconds, ok := response["uptime_seconds"].(float64)
	assert.True(t, ok)
	assert.GreaterOrEqual(t, uptimeSeconds, 0.0)
}

func TestEngramStatusVersionMatchesNew(t *testing.T) {
	s := New("engram", "1.2.3")

	result, err := s.handleStatus(context.Background(), mcpgo.CallToolRequest{})
	require.NoError(t, err)

	textContent, ok := mcpgo.AsTextContent(result.Content[0])
	require.True(t, ok)

	var response map[string]interface{}
	err = json.Unmarshal([]byte(textContent.Text), &response)
	require.NoError(t, err)

	assert.Equal(t, "1.2.3", response["version"])
}

func TestEngramStatusHealthy(t *testing.T) {
	s := New("engram", "0.1.0-dev")

	result, err := s.handleStatus(context.Background(), mcpgo.CallToolRequest{})
	require.NoError(t, err)

	textContent, ok := mcpgo.AsTextContent(result.Content[0])
	require.True(t, ok)

	var response map[string]interface{}
	err = json.Unmarshal([]byte(textContent.Text), &response)
	require.NoError(t, err)

	assert.Equal(t, "healthy", response["status"])
}

func TestEngramStatusUptimeNonNegative(t *testing.T) {
	s := New("engram", "0.1.0-dev")

	result, err := s.handleStatus(context.Background(), mcpgo.CallToolRequest{})
	require.NoError(t, err)

	textContent, ok := mcpgo.AsTextContent(result.Content[0])
	require.True(t, ok)

	var response map[string]interface{}
	err = json.Unmarshal([]byte(textContent.Text), &response)
	require.NoError(t, err)

	uptime := response["uptime_seconds"].(float64)
	assert.GreaterOrEqual(t, uptime, 0.0)
}

func TestIntegrationInProcessClient(t *testing.T) {
	s := New("engram", "0.1.0-dev")
	RegisterBuiltinTools(s)

	mcpClient, err := client.NewInProcessClient(s.mcpServer)
	require.NoError(t, err)
	defer mcpClient.Close()

	ctx := context.Background()

	// Initialize the client
	_, err = mcpClient.Initialize(ctx, mcpgo.InitializeRequest{
		Params: mcpgo.InitializeParams{
			ProtocolVersion: mcpgo.LATEST_PROTOCOL_VERSION,
			ClientInfo: mcpgo.Implementation{
				Name:    "test-client",
				Version: "1.0.0",
			},
		},
	})
	require.NoError(t, err)

	// List tools — should contain engram_status
	tools, err := mcpClient.ListTools(ctx, mcpgo.ListToolsRequest{})
	require.NoError(t, err)
	require.Len(t, tools.Tools, 1)
	assert.Equal(t, "engram_status", tools.Tools[0].Name)

	// Call engram_status
	result, err := mcpClient.CallTool(ctx, mcpgo.CallToolRequest{
		Params: mcpgo.CallToolParams{
			Name: "engram_status",
		},
	})
	require.NoError(t, err)
	require.Len(t, result.Content, 1)

	textContent, ok := mcpgo.AsTextContent(result.Content[0])
	require.True(t, ok)

	var response map[string]interface{}
	err = json.Unmarshal([]byte(textContent.Text), &response)
	require.NoError(t, err)

	assert.Equal(t, "0.1.0-dev", response["version"])
	assert.Equal(t, "healthy", response["status"])
	assert.Contains(t, response, "uptime_seconds")
}
