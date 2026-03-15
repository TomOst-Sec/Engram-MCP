package conventions

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestToolDefinition(t *testing.T) {
	tool := &ConventionsTool{}
	def := tool.Definition()
	assert.Equal(t, "get_conventions", def.Name)
}
