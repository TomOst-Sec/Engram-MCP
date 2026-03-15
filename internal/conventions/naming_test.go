package conventions

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestIsSnakeCase(t *testing.T) {
	assert.True(t, IsSnakeCase("my_function"))
	assert.True(t, IsSnakeCase("get_user_by_id"))
	assert.True(t, IsSnakeCase("handle"))
	assert.False(t, IsSnakeCase("myFunction"))
	assert.False(t, IsSnakeCase("MyFunction"))
	assert.False(t, IsSnakeCase(""))
	assert.False(t, IsSnakeCase("_private"))
}

func TestIsCamelCase(t *testing.T) {
	assert.True(t, IsCamelCase("myFunction"))
	assert.True(t, IsCamelCase("getUserById"))
	assert.False(t, IsCamelCase("my_function"))
	assert.False(t, IsCamelCase("MyFunction"))
	assert.False(t, IsCamelCase("handle")) // pure lowercase is snake_case
}

func TestIsPascalCase(t *testing.T) {
	assert.True(t, IsPascalCase("MyFunction"))
	assert.True(t, IsPascalCase("GetUserById"))
	assert.True(t, IsPascalCase("Handler"))
	assert.False(t, IsPascalCase("myFunction"))
	assert.False(t, IsPascalCase("my_function"))
}

func TestDetectNamingStyleSnakeCase(t *testing.T) {
	names := []string{"get_user", "set_name", "handle_request", "process_data", "validate_input",
		"check_auth", "parse_config", "load_data", "save_file", "myFunction"}

	style, conf := DetectNamingStyle(names)
	assert.Equal(t, "snake_case", style)
	assert.Greater(t, conf, 0.8)
}

func TestDetectNamingStyleCamelCase(t *testing.T) {
	names := []string{"getUser", "setName", "handleRequest", "processData", "validateInput",
		"checkAuth", "parseConfig", "loadData", "saveFile", "my_func"}

	style, conf := DetectNamingStyle(names)
	assert.Equal(t, "camelCase", style)
	assert.Greater(t, conf, 0.8)
}

func TestDetectNamingStyleLowConfidence(t *testing.T) {
	// 50/50 split — should not report a convention
	names := []string{"get_user", "set_name", "handle_request", "processData", "validateInput", "checkAuth"}

	style, conf := DetectNamingStyle(names)
	assert.Equal(t, "", style)
	assert.Equal(t, float64(0), conf)
}

func TestDetectNamingStyleEmpty(t *testing.T) {
	style, conf := DetectNamingStyle(nil)
	assert.Equal(t, "", style)
	assert.Equal(t, float64(0), conf)
}
