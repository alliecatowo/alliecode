package tools

import (
	"context"
	"sort"
	"sync"

	"github.com/alliecatowo/alliecode/internal/mcp"
	"github.com/alliecatowo/alliecode/internal/types"
)

// Registry holds all registered tools and provides lookup capabilities.
type Registry struct {
	tools map[string]types.Tool
}

var (
	registeredToolDefsMu sync.RWMutex
	registeredToolDefs   []types.ToolDef
)

// NewRegistry creates an empty tool registry.
func NewRegistry() *Registry {
	return &Registry{
		tools: make(map[string]types.Tool),
	}
}

// Register adds a tool to the registry. If a tool with the same name
// already exists, it is replaced.
func (r *Registry) Register(t types.Tool) {
	r.tools[t.Name()] = t
	r.refreshRegisteredToolDefs()
}

// Get returns the tool with the given name, or false if not found.
func (r *Registry) Get(name string) (types.Tool, bool) {
	t, ok := r.tools[name]
	return t, ok
}

// All returns all registered tools as a slice.
func (r *Registry) All() []types.Tool {
	result := make([]types.Tool, 0, len(r.tools))
	for _, t := range r.tools {
		result = append(result, t)
	}
	return result
}

// ToolDefs returns serializable tool definitions for all registered tools,
// suitable for sending to providers.
func (r *Registry) ToolDefs() []types.ToolDef {
	return types.ToToolDefs(r.All())
}

func (r *Registry) refreshRegisteredToolDefs() {
	all := r.All()
	sort.Slice(all, func(i, j int) bool {
		return all[i].Name() < all[j].Name()
	})
	defs := types.ToToolDefs(all)

	registeredToolDefsMu.Lock()
	registeredToolDefs = defs
	registeredToolDefsMu.Unlock()
}

func registeredToolDefsSnapshot() []types.ToolDef {
	registeredToolDefsMu.RLock()
	defer registeredToolDefsMu.RUnlock()
	defs := make([]types.ToolDef, len(registeredToolDefs))
	copy(defs, registeredToolDefs)
	return defs
}

// DefaultRegistry returns a registry pre-populated with all built-in tools.
func DefaultRegistry() *Registry {
	r := NewRegistry()
	registerShellFamilyTools(r)
	registerFileFamilyTools(r)
	registerWebFamilyTools(r)
	registerTaskFamilyTools(r)
	registerTeamFamilyTools(r)
	registerOrchestrationTools(r)

	return r
}

// DefaultRegistryWithMCP returns the default registry plus MCP tools.
func DefaultRegistryWithMCP(ctx context.Context, configs []mcp.ServerConfig) (*Registry, MCPManager, error) {
	r := DefaultRegistry()
	if len(configs) == 0 {
		return r, nil, nil
	}

	mgr, err := NewMCPManager(ctx, configs)
	if err != nil {
		return nil, nil, err
	}

	registerMCPFamilyTools(r, mgr)

	return r, mgr, nil
}
