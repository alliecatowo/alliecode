package tools

import (
	"bufio"
	"context"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strconv"
	"strings"
	"time"

	"github.com/alliecatowo/alliecode/internal/types"
)

// FileReadTool reads file contents with line numbers.
type FileReadTool struct{}

type fileReadInput struct {
	FilePath string `json:"file_path"`
	Offset   int    `json:"offset"`
	Limit    int    `json:"limit"`
}

const maxReadLineLength = 2000

var blockedReadDevicePaths = map[string]struct{}{
	"/dev/zero":    {},
	"/dev/random":  {},
	"/dev/urandom": {},
	"/dev/full":    {},
	"/dev/stdin":   {},
	"/dev/tty":     {},
	"/dev/console": {},
	"/dev/stdout":  {},
	"/dev/stderr":  {},
	"/dev/fd/0":    {},
	"/dev/fd/1":    {},
	"/dev/fd/2":    {},
}

func (t *FileReadTool) Name() string { return "Read" }

func (t *FileReadTool) Description() string {
	return "Reads a file from the local filesystem, returning contents with line numbers."
}

func (t *FileReadTool) InputSchema() types.ToolSchema {
	return types.ToolSchema{
		Type: "object",
		Properties: map[string]types.PropertySchema{
			"file_path": {
				Type:        "string",
				Description: "The absolute path to the file to read.",
			},
			"offset": {
				Type:        "integer",
				Description: "Line number to start reading from (1-based). Defaults to 1.",
			},
			"limit": {
				Type:        "integer",
				Description: "Maximum number of lines to read. Defaults to 2000.",
			},
		},
		Required: []string{"file_path"},
	}
}

func (t *FileReadTool) Execute(ctx context.Context, input types.ToolInput, toolCtx types.ToolContext) (types.ToolResult, error) {
	started := time.Now()
	var in fileReadInput
	if err := json.Unmarshal(input, &in); err != nil {
		return types.ToolResult{Content: fmt.Sprintf("Invalid input: %v", err), IsError: true}, nil
	}

	if in.FilePath == "" {
		return types.ToolResult{Content: "file_path is required", IsError: true}, nil
	}

	path := in.FilePath
	if !filepath.IsAbs(path) {
		path = filepath.Join(toolCtx.WorkingDir, path)
	}
	path = filepath.Clean(path)

	if isBlockedDevicePath(path) {
		return types.ToolResult{Content: "Reading from this device path is blocked for safety", IsError: true}, nil
	}

	stat, err := os.Stat(path)
	if err != nil {
		return types.ToolResult{Content: fmt.Sprintf("Error opening file: %v", err), IsError: true}, nil
	}

	if stat.IsDir() {
		return t.readDirectory(path, stat.ModTime(), stat.Size(), started, in)
	}

	limit := in.Limit
	if limit <= 0 {
		limit = 2000
	}
	offset := in.Offset
	if offset <= 0 {
		offset = 1
	}

	f, err := os.Open(path)
	if err != nil {
		return types.ToolResult{Content: fmt.Sprintf("Error opening file: %v", err), IsError: true}, nil
	}
	defer f.Close()

	stat, err = f.Stat()
	if err != nil {
		return types.ToolResult{Content: fmt.Sprintf("Error reading file metadata: %v", err), IsError: true}, nil
	}

	if meta, ok := readAttachmentMetadata(path, stat.ModTime(), stat.Size()); ok {
		return types.ToolResult{Content: meta}, nil
	}

	scanner := bufio.NewScanner(f)
	// Increase scanner buffer for long lines.
	scanner.Buffer(make([]byte, 0, 1024*1024), 1024*1024)

	var lines []string
	lineNum := 0
	hitLimit := false
	for scanner.Scan() {
		lineNum++
		if lineNum < offset {
			continue
		}
		if len(lines) >= limit {
			hitLimit = true
			break
		}
		lines = append(lines, fmt.Sprintf("%d: %s", lineNum, truncateReadLine(scanner.Text())))
	}

	if err := scanner.Err(); err != nil {
		return types.ToolResult{Content: fmt.Sprintf("Error reading file: %v", err), IsError: true}, nil
	}

	if len(lines) == 0 {
		meta := renderReadMetadata(path, stat.ModTime(), stat.Size(), offset > 1 || hitLimit, offset, limit)
		block := renderStructuredBlock("file_read_result", []structuredField{
			{Key: "path", Value: path},
			{Key: "offset", Value: strconv.Itoa(offset)},
			{Key: "limit", Value: strconv.Itoa(limit)},
			{Key: "returned_lines", Value: "0"},
			{Key: "hit_limit", Value: strconv.FormatBool(hitLimit)},
			{Key: "duration_ms", Value: strconv.FormatInt(time.Since(started).Milliseconds(), 10)},
		})
		return types.ToolResult{Content: "(empty file or no content in the requested range)\n\n" + meta + "\n\n" + block}, nil
	}

	meta := renderReadMetadata(path, stat.ModTime(), stat.Size(), offset > 1 || hitLimit, offset, limit)
	block := renderStructuredBlock("file_read_result", []structuredField{
		{Key: "path", Value: path},
		{Key: "offset", Value: strconv.Itoa(offset)},
		{Key: "limit", Value: strconv.Itoa(limit)},
		{Key: "returned_lines", Value: strconv.Itoa(len(lines))},
		{Key: "hit_limit", Value: strconv.FormatBool(hitLimit)},
		{Key: "duration_ms", Value: strconv.FormatInt(time.Since(started).Milliseconds(), 10)},
	})
	return types.ToolResult{Content: strings.Join(lines, "\n") + "\n\n" + meta + "\n\n" + block}, nil
}

func (t *FileReadTool) readDirectory(path string, mtime time.Time, sizeBytes int64, started time.Time, in fileReadInput) (types.ToolResult, error) {
	entries, err := os.ReadDir(path)
	if err != nil {
		return types.ToolResult{Content: fmt.Sprintf("Error opening file: %v", err), IsError: true}, nil
	}

	limit := in.Limit
	if limit <= 0 {
		limit = 2000
	}
	offset := in.Offset
	if offset <= 0 {
		offset = 1
	}

	names := make([]string, 0, len(entries))
	for _, entry := range entries {
		name := entry.Name()
		if entry.IsDir() {
			name += "/"
		}
		names = append(names, name)
	}
	sort.Strings(names)

	start := offset - 1
	if start < 0 {
		start = 0
	}
	if start > len(names) {
		start = len(names)
	}
	end := start + limit
	hitLimit := false
	if end < len(names) {
		hitLimit = true
	} else {
		end = len(names)
	}
	contentLines := names[start:end]

	meta := renderReadMetadata(path, mtime, sizeBytes, offset > 1 || hitLimit, offset, limit)
	block := renderStructuredBlock("file_read_result", []structuredField{
		{Key: "path", Value: path},
		{Key: "offset", Value: strconv.Itoa(offset)},
		{Key: "limit", Value: strconv.Itoa(limit)},
		{Key: "returned_lines", Value: strconv.Itoa(len(contentLines))},
		{Key: "hit_limit", Value: strconv.FormatBool(hitLimit)},
		{Key: "is_directory", Value: "true"},
		{Key: "duration_ms", Value: strconv.FormatInt(time.Since(started).Milliseconds(), 10)},
	})
	if len(contentLines) == 0 {
		return types.ToolResult{Content: "(empty directory or no entries in the requested range)\n\n" + meta + "\n\n" + block}, nil
	}
	return types.ToolResult{Content: strings.Join(contentLines, "\n") + "\n\n" + meta + "\n\n" + block}, nil
}

func truncateReadLine(line string) string {
	if len(line) <= maxReadLineLength {
		return line
	}
	return line[:maxReadLineLength]
}

func isBlockedDevicePath(path string) bool {
	if _, ok := blockedReadDevicePaths[path]; ok {
		return true
	}
	if strings.HasPrefix(path, "/proc/") && (strings.HasSuffix(path, "/fd/0") || strings.HasSuffix(path, "/fd/1") || strings.HasSuffix(path, "/fd/2")) {
		return true
	}
	return false
}

func readAttachmentMetadata(path string, mtime time.Time, sizeBytes int64) (string, bool) {
	ext := strings.ToLower(filepath.Ext(path))
	mime, ok := map[string]string{
		".png":  "image/png",
		".jpg":  "image/jpeg",
		".jpeg": "image/jpeg",
		".gif":  "image/gif",
		".webp": "image/webp",
		".bmp":  "image/bmp",
		".svg":  "image/svg+xml",
		".pdf":  "application/pdf",
	}[ext]
	if !ok {
		return "", false
	}

	mediaKind := "image"
	if ext == ".pdf" {
		mediaKind = "pdf"
	}

	return renderReadAttachmentMetadata(path, mediaKind, mime, mtime, sizeBytes), true
}

func (t *FileReadTool) IsReadOnly(input types.ToolInput) bool {
	return true
}

func (t *FileReadTool) IsDestructive(input types.ToolInput) bool {
	return false
}

func (t *FileReadTool) IsConcurrencySafe(input types.ToolInput) bool {
	return true
}

func (t *FileReadTool) CheckPermissions(input types.ToolInput, toolCtx types.ToolContext) types.ToolPermission {
	return evaluateToolPermissionByFamily(toolFamilyFile, "read", input, toolCtx, types.PermissionAllowed)
}
