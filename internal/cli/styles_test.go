package cli

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestStylesAreDefined(t *testing.T) {
	assert.NotEmpty(t, string(Primary))
	assert.NotEmpty(t, string(Secondary))
	assert.NotEmpty(t, string(Success))
	assert.NotEmpty(t, string(Warning))
	assert.NotEmpty(t, string(Error))
	assert.NotEmpty(t, string(Muted))
}

func TestTitleRenders(t *testing.T) {
	result := Title.Render("Test Title")
	assert.Contains(t, result, "Test Title")
}

func TestSubtitleRenders(t *testing.T) {
	result := Subtitle.Render("Section")
	assert.Contains(t, result, "Section")
}

func TestFilePathRenders(t *testing.T) {
	result := FilePath.Render("internal/auth/handler.go")
	assert.Contains(t, result, "internal/auth/handler.go")
}

func TestSymbolNameRenders(t *testing.T) {
	result := SymbolName.Render("HandleLogin")
	assert.Contains(t, result, "HandleLogin")
}

func TestSuccessTextRenders(t *testing.T) {
	result := SuccessText.Render("OK")
	assert.Contains(t, result, "OK")
}

func TestWarningTextRenders(t *testing.T) {
	result := WarningText.Render("Warning")
	assert.Contains(t, result, "Warning")
}
