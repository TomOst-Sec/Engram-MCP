package architecture

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/TomOst-Sec/colony-project/internal/storage"
	mcpgo "github.com/mark3labs/mcp-go/mcp"
)

// ArchitectureTool provides the get_architecture MCP tool.
type ArchitectureTool struct {
	store        *storage.Store
	repoRoot     string
	goModulePath string
}

// NewArchitectureTool creates a new architecture tool instance.
func NewArchitectureTool(store *storage.Store, repoRoot, goModulePath string) *ArchitectureTool {
	return &ArchitectureTool{
		store:        store,
		repoRoot:     repoRoot,
		goModulePath: goModulePath,
	}
}

// Definition returns the MCP tool definition with schema.
func (t *ArchitectureTool) Definition() mcpgo.Tool {
	return mcpgo.NewTool("get_architecture",
		mcpgo.WithDescription("Returns a structured map of the project's architecture: modules, their responsibilities, dependencies, and key exports."),
		mcpgo.WithString("module",
			mcpgo.Description("Focus on a specific module/directory path (returns detailed view of that module)"),
		),
		mcpgo.WithNumber("depth",
			mcpgo.Description("Directory depth for module detection (default: 2, e.g., 'internal/auth' is depth 2)"),
		),
		mcpgo.WithBoolean("include_exports",
			mcpgo.Description("Include exported symbols per module (default: false, set true for detailed view)"),
		),
	)
}

// Handle processes a get_architecture request.
func (t *ArchitectureTool) Handle(ctx context.Context, request mcpgo.CallToolRequest) (*mcpgo.CallToolResult, error) {
	// Parse parameters
	depth := 2
	if d, ok := request.GetArguments()["depth"]; ok {
		if df, ok := d.(float64); ok {
			depth = int(df)
		}
	}

	includeExports := false
	if ie, ok := request.GetArguments()["include_exports"]; ok {
		if b, ok := ie.(bool); ok {
			includeExports = b
		}
	}

	moduleName := ""
	if m, ok := request.GetArguments()["module"]; ok {
		if s, ok := m.(string); ok {
			moduleName = s
		}
	}

	// Detect modules
	modules, err := DetectModules(t.store, depth)
	if err != nil {
		return nil, fmt.Errorf("detecting modules: %w", err)
	}
	if modules == nil {
		modules = []Module{}
	}

	// Build dependency graph
	internalDeps, externalDeps, err := BuildDependencyGraph(t.store, modules, t.goModulePath)
	if err != nil {
		return nil, fmt.Errorf("building dependency graph: %w", err)
	}

	// Count imports per module for complexity
	importCounts := make(map[string]int)
	for mod, deps := range internalDeps {
		importCounts[mod] += len(deps)
	}
	for mod, deps := range externalDeps {
		importCounts[mod] += len(deps)
	}

	// Enrich modules with deps and complexity
	for i := range modules {
		modules[i].Dependencies = internalDeps[modules[i].Path]
		if modules[i].Dependencies == nil {
			modules[i].Dependencies = []string{}
		}
		modules[i].ExternalDependencies = externalDeps[modules[i].Path]
		if modules[i].ExternalDependencies == nil {
			modules[i].ExternalDependencies = []string{}
		}
		modules[i].ComplexityScore = ComputeComplexity(
			modules[i].Symbols,
			importCounts[modules[i].Path],
			modules[i].Files,
		)

		if includeExports {
			exports, err := GetModuleExports(t.store, modules[i].Path)
			if err != nil {
				return nil, fmt.Errorf("getting exports for %s: %w", modules[i].Path, err)
			}
			modules[i].ExportDetails = exports
			names := make([]string, len(exports))
			for j, e := range exports {
				names[j] = e.Name
			}
			modules[i].Exports = names
		}
	}

	// Single module detail view
	if moduleName != "" {
		return t.handleSingleModule(modules, moduleName, internalDeps)
	}

	// Ensure dep graph isn't nil
	if internalDeps == nil {
		internalDeps = make(map[string][]string)
	}

	// Full project view
	return t.handleFullProject(modules, internalDeps)
}

func (t *ArchitectureTool) handleFullProject(modules []Module, depGraph map[string][]string) (*mcpgo.CallToolResult, error) {
	totalFiles := 0
	for _, m := range modules {
		totalFiles += m.Files
	}

	response := map[string]interface{}{
		"project_root":     t.repoRoot,
		"total_modules":    len(modules),
		"total_files":      totalFiles,
		"modules":          modules,
		"dependency_graph": depGraph,
	}

	data, err := json.Marshal(response)
	if err != nil {
		return nil, fmt.Errorf("marshaling response: %w", err)
	}
	return mcpgo.NewToolResultText(string(data)), nil
}

func (t *ArchitectureTool) handleSingleModule(modules []Module, moduleName string, depGraph map[string][]string) (*mcpgo.CallToolResult, error) {
	var target *Module
	for i := range modules {
		if modules[i].Path == moduleName || modules[i].Name == moduleName {
			target = &modules[i]
			break
		}
	}

	if target == nil {
		return mcpgo.NewToolResultError(fmt.Sprintf("module %q not found", moduleName)), nil
	}

	// Ensure exports are populated for single module view
	if target.ExportDetails == nil {
		exports, err := GetModuleExports(t.store, target.Path)
		if err != nil {
			return nil, fmt.Errorf("getting exports: %w", err)
		}
		target.ExportDetails = exports
		names := make([]string, len(exports))
		for j, e := range exports {
			names[j] = e.Name
		}
		target.Exports = names
	}

	target.Dependents = FindDependents(target.Path, depGraph)

	response := map[string]interface{}{
		"module": target,
	}

	data, err := json.Marshal(response)
	if err != nil {
		return nil, fmt.Errorf("marshaling response: %w", err)
	}
	return mcpgo.NewToolResultText(string(data)), nil
}
