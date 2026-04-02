package config

import "strings"

type DiagnosticPageQuery struct {
	DiagnosticQuery
	Offset             int
	Limit              int
	Codes              []string
	CodePrefixAny      []string
	MessageContainsAny []string
	RemedyContainsAny  []string
	RequireRemedy      *bool
}

type DiagnosticPage struct {
	Items   []Diagnostic `json:"items,omitempty"`
	Total   int          `json:"total"`
	Offset  int          `json:"offset"`
	Limit   int          `json:"limit"`
	HasMore bool         `json:"has_more"`
}

type DiagnosticSummary struct {
	Total          int            `json:"total"`
	Errors         int            `json:"errors"`
	Warns          int            `json:"warns"`
	Infos          int            `json:"infos"`
	WithRemedy     int            `json:"with_remedy"`
	DistinctCodes  int            `json:"distinct_codes"`
	CodePrefixHits map[string]int `json:"code_prefix_hits,omitempty"`
}

func QueryDiagnosticsPage(diags []Diagnostic, query DiagnosticPageQuery) DiagnosticPage {
	base := query.DiagnosticQuery
	base.Limit = 0
	items := QueryDiagnostics(diags, base)
	items = filterDiagnosticsSurface(items, query)
	total := len(items)
	start, end := diagPagination(total, query.Offset, query.Limit)
	return DiagnosticPage{Items: items[start:end], Total: total, Offset: start, Limit: end - start, HasMore: end < total}
}

func SummarizeDiagnostics(diags []Diagnostic) DiagnosticSummary {
	out := DiagnosticSummary{CodePrefixHits: make(map[string]int)}
	codes := make(map[string]struct{})
	for _, item := range diags {
		out.Total++
		switch item.Severity {
		case DiagnosticError:
			out.Errors++
		case DiagnosticWarn:
			out.Warns++
		case DiagnosticInfo:
			out.Infos++
		}
		if strings.TrimSpace(item.Remediation) != "" {
			out.WithRemedy++
		}
		if code := strings.TrimSpace(item.Code); code != "" {
			codes[code] = struct{}{}
		}
		prefix := diagnosticPrefix(item.Code)
		if prefix != "" {
			out.CodePrefixHits[prefix]++
		}
	}
	out.DistinctCodes = len(codes)
	if len(out.CodePrefixHits) == 0 {
		out.CodePrefixHits = nil
	}
	return out
}

func filterDiagnosticsSurface(items []Diagnostic, query DiagnosticPageQuery) []Diagnostic {
	if len(items) == 0 {
		return nil
	}
	allowedCodes := make(map[string]struct{}, len(query.Codes))
	for _, code := range query.Codes {
		trimmed := strings.TrimSpace(code)
		if trimmed != "" {
			allowedCodes[trimmed] = struct{}{}
		}
	}
	containsAny := make([]string, 0, len(query.MessageContainsAny))
	codePrefixes := make([]string, 0, len(query.CodePrefixAny))
	remedyContainsAny := make([]string, 0, len(query.RemedyContainsAny))
	for _, value := range query.MessageContainsAny {
		trimmed := strings.ToLower(strings.TrimSpace(value))
		if trimmed != "" {
			containsAny = append(containsAny, trimmed)
		}
	}
	for _, value := range query.CodePrefixAny {
		trimmed := strings.ToLower(strings.TrimSpace(value))
		if trimmed != "" {
			codePrefixes = append(codePrefixes, trimmed)
		}
	}
	for _, value := range query.RemedyContainsAny {
		trimmed := strings.ToLower(strings.TrimSpace(value))
		if trimmed != "" {
			remedyContainsAny = append(remedyContainsAny, trimmed)
		}
	}
	out := make([]Diagnostic, 0, len(items))
	for _, item := range items {
		if query.RequireRemedy != nil {
			hasRemedy := strings.TrimSpace(item.Remediation) != ""
			if hasRemedy != *query.RequireRemedy {
				continue
			}
		}
		if len(allowedCodes) > 0 {
			if _, ok := allowedCodes[item.Code]; !ok {
				continue
			}
		}
		if len(codePrefixes) > 0 {
			code := strings.ToLower(strings.TrimSpace(item.Code))
			matched := false
			for _, prefix := range codePrefixes {
				if strings.HasPrefix(code, prefix) {
					matched = true
					break
				}
			}
			if !matched {
				continue
			}
		}
		if len(containsAny) > 0 {
			msg := strings.ToLower(strings.TrimSpace(item.Message))
			matched := false
			for _, contains := range containsAny {
				if strings.Contains(msg, contains) {
					matched = true
					break
				}
			}
			if !matched {
				continue
			}
		}
		if len(remedyContainsAny) > 0 {
			remedy := strings.ToLower(strings.TrimSpace(item.Remediation))
			matched := false
			for _, contains := range remedyContainsAny {
				if strings.Contains(remedy, contains) {
					matched = true
					break
				}
			}
			if !matched {
				continue
			}
		}
		out = append(out, item)
	}
	return out
}

func diagnosticPrefix(code string) string {
	code = strings.TrimSpace(code)
	if code == "" {
		return ""
	}
	idx := strings.Index(code, "_")
	if idx <= 0 {
		return code
	}
	return code[:idx]
}

func diagPagination(total, offset, limit int) (int, int) {
	if total <= 0 {
		return 0, 0
	}
	if offset < 0 {
		offset = 0
	}
	if offset > total {
		offset = total
	}
	end := total
	if limit > 0 && offset+limit < end {
		end = offset + limit
	}
	return offset, end
}
