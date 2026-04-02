package references

import (
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"sync"
)

// ParsedReference captures a raw @reference extracted from text.
type ParsedReference struct {
	Raw     string
	Path    string
	Line    int
	HasLine bool
	Start   int
	End     int
}

// ResolvedReference is a parsed reference with filesystem metadata.
type ResolvedReference struct {
	ParsedReference
	AbsolutePath  string
	CanonicalPath string
	Exists        bool
	ResourceType  string
}

// Resolver resolves @references relative to a base directory.
type Resolver struct {
	baseDir      string
	indexOnce    sync.Once
	indexedPaths []indexedPath
	indexErr     error
}

type indexedPath struct {
	Path  string
	IsDir bool
}

// NewResolver constructs a resolver for the given base directory.
func NewResolver(baseDir string) *Resolver {
	if strings.TrimSpace(baseDir) == "" {
		if cwd, err := os.Getwd(); err == nil {
			baseDir = cwd
		}
	}
	return &Resolver{baseDir: baseDir}
}

// ParseReferences parses @file and @path/to/file:line references from text.
func ParseReferences(input string) []ParsedReference {
	refs := make([]ParsedReference, 0)
	for i := 0; i < len(input); i++ {
		if input[i] != '@' {
			continue
		}
		if i > 0 && isWordLike(input[i-1]) {
			continue
		}

		j := i + 1
		for j < len(input) && isRefChar(input[j]) {
			j++
		}
		if j == i+1 {
			continue
		}

		token := input[i+1 : j]
		pathPart, line, hasLine := splitPathAndLine(token)
		if pathPart == "" {
			continue
		}

		refs = append(refs, ParsedReference{
			Raw:     input[i:j],
			Path:    pathPart,
			Line:    line,
			HasLine: hasLine,
			Start:   i,
			End:     j,
		})

		i = j - 1
	}

	return refs
}

// Resolve parses and resolves all references in input.
func (r *Resolver) Resolve(input string) []ResolvedReference {
	parsed := ParseReferences(input)
	resolved := make([]ResolvedReference, 0, len(parsed))
	for _, ref := range parsed {
		abs := ref.Path
		if strings.HasPrefix(abs, "~/") || abs == "~" {
			if home, err := os.UserHomeDir(); err == nil {
				if abs == "~" {
					abs = home
				} else {
					abs = filepath.Join(home, abs[2:])
				}
			}
		}
		if !filepath.IsAbs(abs) {
			abs = filepath.Join(r.baseDir, abs)
		}
		abs = filepath.Clean(abs)

		canonical := abs
		if eval, err := filepath.EvalSymlinks(abs); err == nil {
			canonical = eval
		}

		entry := ResolvedReference{
			ParsedReference: ref,
			AbsolutePath:    abs,
			CanonicalPath:   canonical,
			Exists:          false,
			ResourceType:    "missing",
		}

		if info, err := os.Stat(canonical); err == nil {
			entry.Exists = true
			if info.IsDir() {
				entry.ResourceType = "directory"
			} else {
				entry.ResourceType = "file"
			}
		}

		resolved = append(resolved, entry)
	}

	return resolved
}

// Suggestion captures a candidate path for @reference autocompletion.
type Suggestion struct {
	Path        string
	Score       int
	IsDir       bool
	Source      string
	Section     string
	MatchReason string
	Preview     string
}

// Suggest returns ranked @reference suggestions from workspace and recent files.
// Query can include an optional :line suffix scaffold (for example src/main.go:12).
func (r *Resolver) Suggest(query string, recent []string, limit int) []Suggestion {
	if limit <= 0 {
		limit = 8
	}
	r.ensureIndex()
	if len(r.indexedPaths) == 0 && len(recent) == 0 {
		return nil
	}

	pathQuery, lineSuffix := splitSuggestionQuery(query)
	norm := normalizePathToken(pathQuery)
	recentSet := make(map[string]struct{}, len(recent))
	for _, raw := range recent {
		rel, ok := r.toRelativePath(raw)
		if !ok {
			continue
		}
		recentSet[rel] = struct{}{}
	}

	scored := make([]Suggestion, 0, len(r.indexedPaths)+len(recentSet))
	seen := make(map[string]struct{}, len(r.indexedPaths)+len(recentSet))

	addCandidate := func(path string, isDir bool, source string) {
		if path == "" {
			return
		}
		if _, exists := seen[path]; exists {
			return
		}
		score, reason, ok := scorePathCandidate(path, norm)
		if !ok {
			return
		}
		if _, isRecent := recentSet[path]; isRecent {
			score += 320
			if reason == "" {
				reason = "recent"
			}
		}
		if source == "recent" {
			score += 80
		}
		if reason == "" {
			reason = "browse"
		}
		seen[path] = struct{}{}
		scored = append(scored, Suggestion{Path: path, Score: score, IsDir: isDir, Source: source, Section: suggestionSectionFor(source, reason, isDir), MatchReason: reason, Preview: buildSuggestionPreview(path, isDir)})
	}

	for _, item := range r.indexedPaths {
		source := "workspace"
		if _, ok := recentSet[item.Path]; ok {
			source = "workspace+recent"
		}
		addCandidate(item.Path, item.IsDir, source)
	}
	for rel := range recentSet {
		addCandidate(rel, false, "recent")
	}

	if len(scored) == 0 {
		return nil
	}

	sortSuggestions(scored)
	if len(scored) > limit {
		scored = scored[:limit]
	}

	if lineSuffix != "" {
		for i := range scored {
			scored[i].Path += lineSuffix
		}
	}
	return scored
}

func (r *Resolver) ensureIndex() {
	r.indexOnce.Do(func() {
		items := make([]indexedPath, 0, 512)
		err := filepath.WalkDir(r.baseDir, func(path string, d os.DirEntry, err error) error {
			if err != nil {
				return nil
			}
			rel, relErr := filepath.Rel(r.baseDir, path)
			if relErr != nil {
				return nil
			}
			rel = filepath.ToSlash(strings.TrimSpace(rel))
			if rel == "." || rel == "" {
				return nil
			}
			if shouldSkipPath(rel, d) {
				if d.IsDir() {
					return filepath.SkipDir
				}
				return nil
			}
			items = append(items, indexedPath{Path: rel, IsDir: d.IsDir()})
			return nil
		})
		r.indexedPaths = items
		r.indexErr = err
	})
}

func splitSuggestionQuery(query string) (string, string) {
	trimmed := strings.TrimSpace(query)
	if trimmed == "" {
		return "", ""
	}
	idx := strings.LastIndex(trimmed, ":")
	if idx <= 0 || idx == len(trimmed)-1 {
		return trimmed, ""
	}
	line := strings.TrimSpace(trimmed[idx+1:])
	if line == "" {
		return trimmed, ""
	}
	for _, r := range line {
		if r < '0' || r > '9' {
			return trimmed, ""
		}
	}
	return strings.TrimSpace(trimmed[:idx]), ":" + line
}

func normalizePathToken(s string) string {
	s = strings.TrimSpace(s)
	s = strings.TrimPrefix(s, "@")
	s = filepath.ToSlash(s)
	return strings.ToLower(strings.TrimSpace(s))
}

func (r *Resolver) toRelativePath(path string) (string, bool) {
	if strings.TrimSpace(path) == "" {
		return "", false
	}
	p := path
	if strings.HasPrefix(p, "~/") || p == "~" {
		if home, err := os.UserHomeDir(); err == nil {
			if p == "~" {
				p = home
			} else {
				p = filepath.Join(home, p[2:])
			}
		}
	}
	if !filepath.IsAbs(p) {
		p = filepath.Join(r.baseDir, p)
	}
	p = filepath.Clean(p)
	rel, err := filepath.Rel(r.baseDir, p)
	if err != nil {
		return "", false
	}
	rel = filepath.ToSlash(strings.TrimSpace(rel))
	if rel == "." || rel == "" || strings.HasPrefix(rel, "../") {
		return "", false
	}
	return rel, true
}

func scorePathCandidate(candidate, normQuery string) (int, string, bool) {
	normCandidate := normalizePathToken(candidate)
	if normCandidate == "" {
		return 0, "", false
	}
	if normQuery == "" {
		return 120 + pathDepthBoost(normCandidate), "browse", true
	}

	base := strings.ToLower(filepath.Base(normCandidate))
	score := 0
	matched := false
	reason := ""

	if normCandidate == normQuery {
		score += 1400
		matched = true
		reason = "exact"
	} else if strings.HasPrefix(normCandidate, normQuery) {
		score += 1020
		matched = true
		reason = "path-prefix"
	} else if strings.Contains(normCandidate, normQuery) {
		score += 620
		matched = true
		reason = "path-contains"
	}

	if strings.HasPrefix(base, normQuery) {
		score += 280
		matched = true
		if reason == "" {
			reason = "name-prefix"
		}
	} else if strings.Contains(base, normQuery) {
		score += 120
		matched = true
		if reason == "" {
			reason = "name-contains"
		}
	}

	tokens := strings.Fields(strings.NewReplacer("/", " ", "-", " ", "_", " ", ".", " ").Replace(normQuery))
	tokenHits := 0
	for _, token := range tokens {
		if token == "" {
			continue
		}
		if strings.Contains(normCandidate, token) {
			score += 90
			tokenHits++
			matched = true
			if reason == "" {
				reason = "token"
			}
		}
	}
	if len(tokens) > 1 && tokenHits == len(tokens) {
		score += 130
	}

	if !matched && isSubsequence(normCandidate, normQuery) {
		score += 220
		matched = true
		reason = "fuzzy"
	}

	if !matched {
		return 0, "", false
	}
	score += pathDepthBoost(normCandidate)
	if reason == "" {
		reason = "browse"
	}
	return score, reason, true
}

func buildSuggestionPreview(path string, isDir bool) string {
	base := filepath.Base(path)
	depth := strings.Count(path, "/") + 1
	if isDir {
		return "directory " + base + " depth:" + strconv.Itoa(depth)
	}
	if ext := strings.TrimSpace(filepath.Ext(base)); ext != "" {
		return ext + " file " + base + " depth:" + strconv.Itoa(depth)
	}
	return "file " + base + " depth:" + strconv.Itoa(depth)
}

func suggestionSectionFor(source, reason string, isDir bool) string {
	prefix := "Workspace"
	switch strings.ToLower(strings.TrimSpace(source)) {
	case "recent":
		prefix = "Recent"
	case "workspace+recent":
		prefix = "Recent + Workspace"
	}

	suffix := "Files"
	if isDir {
		suffix = "Folders"
	}

	normReason := strings.ToLower(strings.TrimSpace(reason))
	switch {
	case normReason == "exact" || strings.Contains(normReason, "exact"):
		return "Best Match"
	case strings.Contains(normReason, "prefix"):
		return prefix + " Prefix " + suffix
	case strings.Contains(normReason, "contains") || strings.Contains(normReason, "token"):
		return prefix + " Contains " + suffix
	case strings.Contains(normReason, "fuzzy"):
		return prefix + " Fuzzy " + suffix
	default:
		return prefix + " " + suffix
	}
}

func pathDepthBoost(path string) int {
	depth := strings.Count(path, "/")
	if depth <= 0 {
		return 40
	}
	if depth > 5 {
		depth = 5
	}
	return 40 - depth*5
}

func shouldSkipPath(rel string, d os.DirEntry) bool {
	name := strings.ToLower(strings.TrimSpace(filepath.Base(rel)))
	if name == "" {
		return false
	}
	if strings.HasPrefix(name, ".") {
		if d.IsDir() {
			return true
		}
		switch name {
		case ".env", ".env.local", ".env.development", ".env.production", ".env.test":
			return true
		}
	}
	if d.IsDir() {
		switch name {
		case "node_modules", "dist", "build", "coverage", "vendor", "tmp", "temp", "out", ".git":
			return true
		}
	}
	return false
}

func sortSuggestions(items []Suggestion) {
	if len(items) < 2 {
		return
	}
	for i := 0; i < len(items)-1; i++ {
		for j := i + 1; j < len(items); j++ {
			if items[j].Score > items[i].Score {
				items[i], items[j] = items[j], items[i]
				continue
			}
			if items[j].Score == items[i].Score && items[j].Path < items[i].Path {
				items[i], items[j] = items[j], items[i]
			}
		}
	}
}

func isSubsequence(text, query string) bool {
	if query == "" {
		return true
	}
	j := 0
	for i := 0; i < len(text) && j < len(query); i++ {
		if text[i] == query[j] {
			j++
		}
	}
	return j == len(query)
}

func splitPathAndLine(token string) (string, int, bool) {
	idx := strings.LastIndex(token, ":")
	if idx <= 0 || idx >= len(token)-1 {
		return token, 0, false
	}

	linePart := token[idx+1:]
	lineNum, err := strconv.Atoi(linePart)
	if err != nil || lineNum <= 0 {
		return token, 0, false
	}

	return token[:idx], lineNum, true
}

func isWordLike(b byte) bool {
	return (b >= 'a' && b <= 'z') ||
		(b >= 'A' && b <= 'Z') ||
		(b >= '0' && b <= '9') ||
		b == '_' || b == '.' || b == '-'
}

func isRefChar(b byte) bool {
	if (b >= 'a' && b <= 'z') || (b >= 'A' && b <= 'Z') || (b >= '0' && b <= '9') {
		return true
	}
	switch b {
	case '/', '\\', '.', '_', '-', '~', ':':
		return true
	default:
		return false
	}
}
