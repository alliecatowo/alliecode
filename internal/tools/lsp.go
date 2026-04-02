package tools

import (
	"bufio"
	"context"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"regexp"
	"sort"
	"strings"

	"github.com/alliecatowo/alliecode/internal/types"
)

type LSPTool struct{}

type lspInput struct {
	Operation string `json:"operation"`
	FilePath  string `json:"file_path"`
	Line      int    `json:"line"`
	Character int    `json:"character"`
	Query     string `json:"query,omitempty"`
}

type lspLocation struct {
	FilePath string `json:"file_path"`
	Line     int    `json:"line"`
	Column   int    `json:"column"`
	Preview  string `json:"preview"`
}

type lspOutput struct {
	Success     bool          `json:"success"`
	Operation   string        `json:"operation"`
	FilePath    string        `json:"file_path"`
	Symbol      string        `json:"symbol,omitempty"`
	Result      string        `json:"result"`
	ResultCount int           `json:"result_count"`
	Locations   []lspLocation `json:"locations,omitempty"`
	ErrorText   string        `json:"error,omitempty"`
}

func (t *LSPTool) Name() string { return "lsp" }

func (t *LSPTool) Description() string {
	return "Provider-agnostic code intelligence fallback (definitions, references, symbols, hover)."
}

func (t *LSPTool) InputSchema() types.ToolSchema {
	return types.ToolSchema{
		Type: "object",
		Properties: map[string]types.PropertySchema{
			"operation": {Type: "string", Description: "goToDefinition, findReferences, hover, documentSymbol, workspaceSymbol, goToImplementation, prepareCallHierarchy, incomingCalls, outgoingCalls", Enum: []string{"goToDefinition", "findReferences", "hover", "documentSymbol", "workspaceSymbol", "goToImplementation", "prepareCallHierarchy", "incomingCalls", "outgoingCalls"}},
			"file_path": {Type: "string", Description: "File path for context."},
			"line":      {Type: "integer", Description: "1-based line number."},
			"character": {Type: "integer", Description: "1-based character position."},
			"query":     {Type: "string", Description: "Optional symbol query for workspaceSymbol."},
		},
		Required: []string{"operation", "file_path", "line", "character"},
	}
}

func (t *LSPTool) Execute(_ context.Context, input types.ToolInput, toolCtx types.ToolContext) (types.ToolResult, error) {
	var in lspInput
	if err := json.Unmarshal(input, &in); err != nil {
		return lspResult(lspOutput{Success: false, Result: "", ErrorText: fmt.Sprintf("invalid input: %v", err)}, true)
	}
	op := strings.TrimSpace(in.Operation)
	if !isSupportedLSPOperation(op) {
		return lspResult(lspOutput{Success: false, Operation: op, FilePath: in.FilePath, Result: "", ErrorText: "unsupported operation"}, true)
	}
	if in.Line <= 0 || in.Character <= 0 {
		return lspResult(lspOutput{Success: false, Operation: op, FilePath: in.FilePath, Result: "", ErrorText: "line and character must be >= 1"}, true)
	}

	targetPath := in.FilePath
	if !filepath.IsAbs(targetPath) {
		targetPath = filepath.Join(toolCtx.WorkingDir, in.FilePath)
	}
	lines, err := readLines(targetPath)
	if err != nil {
		return lspResult(lspOutput{Success: false, Operation: op, FilePath: in.FilePath, Result: "", ErrorText: err.Error()}, true)
	}

	symbol := strings.TrimSpace(in.Query)
	if symbol == "" {
		symbol = extractSymbol(lines, in.Line, in.Character)
	}

	switch op {
	case "hover":
		if in.Line > len(lines) {
			return lspResult(lspOutput{Success: false, Operation: op, FilePath: in.FilePath, Result: "", ErrorText: "line out of range"}, true)
		}
		preview := strings.TrimSpace(lines[in.Line-1])
		result := fmt.Sprintf("Symbol: %s\nLine %d: %s", symbol, in.Line, preview)
		return lspResult(lspOutput{Success: true, Operation: op, FilePath: in.FilePath, Symbol: symbol, Result: result, ResultCount: 1, Locations: []lspLocation{{FilePath: in.FilePath, Line: in.Line, Column: in.Character, Preview: preview}}}, false)
	case "documentSymbol":
		locations := documentSymbols(lines, in.FilePath)
		return lspResult(lspOutput{Success: true, Operation: op, FilePath: in.FilePath, Result: fmt.Sprintf("Found %d symbols", len(locations)), ResultCount: len(locations), Locations: locations}, false)
	case "workspaceSymbol":
		if symbol == "" {
			return lspResult(lspOutput{Success: false, Operation: op, FilePath: in.FilePath, Result: "", ErrorText: "query or resolvable symbol is required"}, true)
		}
		locations := searchWorkspace(toolCtx.WorkingDir, symbol, true)
		return lspResult(lspOutput{Success: true, Operation: op, FilePath: in.FilePath, Symbol: symbol, Result: fmt.Sprintf("Found %d workspace symbol matches", len(locations)), ResultCount: len(locations), Locations: locations}, false)
	case "goToDefinition", "goToImplementation", "prepareCallHierarchy":
		if symbol == "" {
			return lspResult(lspOutput{Success: false, Operation: op, FilePath: in.FilePath, Result: "", ErrorText: "could not resolve symbol at requested position"}, true)
		}
		locations := searchDefinitions(toolCtx.WorkingDir, symbol)
		if len(locations) == 0 {
			return lspResult(lspOutput{Success: true, Operation: op, FilePath: in.FilePath, Symbol: symbol, Result: "No definition found", ResultCount: 0}, false)
		}
		return lspResult(lspOutput{Success: true, Operation: op, FilePath: in.FilePath, Symbol: symbol, Result: fmt.Sprintf("Found %d candidate definitions", len(locations)), ResultCount: len(locations), Locations: locations}, false)
	case "findReferences", "incomingCalls", "outgoingCalls":
		if symbol == "" {
			return lspResult(lspOutput{Success: false, Operation: op, FilePath: in.FilePath, Result: "", ErrorText: "could not resolve symbol at requested position"}, true)
		}
		locations := searchWorkspace(toolCtx.WorkingDir, symbol, false)
		return lspResult(lspOutput{Success: true, Operation: op, FilePath: in.FilePath, Symbol: symbol, Result: fmt.Sprintf("Found %d references", len(locations)), ResultCount: len(locations), Locations: locations}, false)
	default:
		return lspResult(lspOutput{Success: false, Operation: op, FilePath: in.FilePath, Result: "", ErrorText: "unsupported operation"}, true)
	}
}

func isSupportedLSPOperation(op string) bool {
	switch op {
	case "goToDefinition", "findReferences", "hover", "documentSymbol", "workspaceSymbol", "goToImplementation", "prepareCallHierarchy", "incomingCalls", "outgoingCalls":
		return true
	default:
		return false
	}
}

func readLines(path string) ([]string, error) {
	f, err := os.Open(path)
	if err != nil {
		return nil, fmt.Errorf("cannot open file: %w", err)
	}
	defer f.Close()
	lines := []string{}
	scanner := bufio.NewScanner(f)
	for scanner.Scan() {
		lines = append(lines, scanner.Text())
	}
	if err := scanner.Err(); err != nil {
		return nil, fmt.Errorf("cannot read file: %w", err)
	}
	return lines, nil
}

var symbolPattern = regexp.MustCompile(`[A-Za-z_][A-Za-z0-9_]*`)

func extractSymbol(lines []string, line, character int) string {
	if line <= 0 || line > len(lines) {
		return ""
	}
	text := lines[line-1]
	if text == "" {
		return ""
	}
	idx := character - 1
	if idx < 0 {
		idx = 0
	}
	if idx >= len(text) {
		idx = len(text) - 1
	}
	matches := symbolPattern.FindAllStringIndex(text, -1)
	for _, m := range matches {
		if idx >= m[0] && idx < m[1] {
			return text[m[0]:m[1]]
		}
	}
	if len(matches) > 0 {
		m := matches[0]
		return text[m[0]:m[1]]
	}
	return ""
}

var declarationPattern = regexp.MustCompile(`\b(func|type|class|interface|struct|const|var)\s+([A-Za-z_][A-Za-z0-9_]*)`)

func documentSymbols(lines []string, filePath string) []lspLocation {
	out := []lspLocation{}
	for idx, line := range lines {
		if declarationPattern.MatchString(line) {
			out = append(out, lspLocation{FilePath: filePath, Line: idx + 1, Column: 1, Preview: strings.TrimSpace(line)})
		}
	}
	return out
}

func searchDefinitions(root, symbol string) []lspLocation {
	if symbol == "" {
		return nil
	}
	re := regexp.MustCompile(fmt.Sprintf(`\b(func|type|class|interface|struct|const|var)\s+%s\b`, regexp.QuoteMeta(symbol)))
	return searchWithPattern(root, re)
}

func searchWorkspace(root, symbol string, onlyDefinitions bool) []lspLocation {
	if symbol == "" {
		return nil
	}
	if onlyDefinitions {
		return searchDefinitions(root, symbol)
	}
	re := regexp.MustCompile(fmt.Sprintf(`\b%s\b`, regexp.QuoteMeta(symbol)))
	return searchWithPattern(root, re)
}

func searchWithPattern(root string, re *regexp.Regexp) []lspLocation {
	locations := []lspLocation{}
	allowedExt := map[string]struct{}{".go": {}, ".ts": {}, ".tsx": {}, ".js": {}, ".jsx": {}, ".py": {}, ".java": {}, ".rs": {}, ".c": {}, ".cpp": {}, ".h": {}}
	_ = filepath.WalkDir(root, func(path string, d os.DirEntry, err error) error {
		if err != nil {
			return nil
		}
		if d.IsDir() {
			name := d.Name()
			if name == ".git" || name == "node_modules" || strings.HasPrefix(name, ".alliecode-worktrees") {
				return filepath.SkipDir
			}
			return nil
		}
		if len(locations) >= 200 {
			return filepath.SkipDir
		}
		if _, ok := allowedExt[strings.ToLower(filepath.Ext(path))]; !ok {
			return nil
		}
		lines, readErr := readLines(path)
		if readErr != nil {
			return nil
		}
		for i, line := range lines {
			loc := re.FindStringIndex(line)
			if loc == nil {
				continue
			}
			locations = append(locations, lspLocation{FilePath: path, Line: i + 1, Column: loc[0] + 1, Preview: strings.TrimSpace(line)})
			if len(locations) >= 200 {
				break
			}
		}
		return nil
	})
	sort.Slice(locations, func(i, j int) bool {
		if locations[i].FilePath != locations[j].FilePath {
			return locations[i].FilePath < locations[j].FilePath
		}
		if locations[i].Line != locations[j].Line {
			return locations[i].Line < locations[j].Line
		}
		return locations[i].Column < locations[j].Column
	})
	return locations
}

func lspResult(out lspOutput, isError bool) (types.ToolResult, error) {
	b, err := json.Marshal(out)
	if err != nil {
		return types.ToolResult{Content: fmt.Sprintf("serialization error: %v", err), IsError: true}, nil
	}
	return types.ToolResult{Content: string(b), IsError: isError}, nil
}

func (t *LSPTool) IsReadOnly(input types.ToolInput) bool { return true }

func (t *LSPTool) IsDestructive(input types.ToolInput) bool { return false }

func (t *LSPTool) IsConcurrencySafe(input types.ToolInput) bool { return true }

func (t *LSPTool) CheckPermissions(input types.ToolInput, toolCtx types.ToolContext) types.ToolPermission {
	return types.PermissionAllowed
}
