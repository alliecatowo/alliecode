package config

import (
	"sort"
	"strings"
)

type DiagnosticQuery struct {
	Severity        DiagnosticSeverity
	SeverityAtLeast DiagnosticSeverity
	CodePrefix      string
	CodeContains    string
	MessageContains string
	RemedyContains  string
	RequireRemedy   bool
	IncludeRemedy   bool
	Limit           int
	SortBySeverity  bool
	SortByCode      bool
	CaseInsensitive bool
}

func QueryDiagnostics(diags []Diagnostic, query DiagnosticQuery) []Diagnostic {
	if len(diags) == 0 {
		return nil
	}

	items := make([]Diagnostic, 0, len(diags))
	for _, item := range diags {
		if !diagnosticMatches(item, query) {
			continue
		}
		if !query.IncludeRemedy {
			item.Remediation = ""
		}
		items = append(items, item)
	}

	if query.SortBySeverity || query.SortByCode {
		sortDiagnostics(items, query.SortBySeverity, query.SortByCode)
	}

	if query.Limit > 0 && len(items) > query.Limit {
		items = items[:query.Limit]
	}
	return items
}

func diagnosticMatches(item Diagnostic, query DiagnosticQuery) bool {
	if query.Severity != "" && item.Severity != query.Severity {
		return false
	}
	if query.SeverityAtLeast != "" && severityRank(item.Severity) < severityRank(query.SeverityAtLeast) {
		return false
	}
	if query.RequireRemedy && strings.TrimSpace(item.Remediation) == "" {
		return false
	}

	code := strings.TrimSpace(item.Code)
	msg := strings.TrimSpace(item.Message)
	prefix := strings.TrimSpace(query.CodePrefix)
	containsCode := strings.TrimSpace(query.CodeContains)
	containsMsg := strings.TrimSpace(query.MessageContains)
	containsRemedy := strings.TrimSpace(query.RemedyContains)

	if query.CaseInsensitive {
		code = strings.ToLower(code)
		msg = strings.ToLower(msg)
		prefix = strings.ToLower(prefix)
		containsCode = strings.ToLower(containsCode)
		containsMsg = strings.ToLower(containsMsg)
		containsRemedy = strings.ToLower(containsRemedy)
	}

	if prefix != "" && !strings.HasPrefix(code, prefix) {
		return false
	}
	if containsCode != "" && !strings.Contains(code, containsCode) {
		return false
	}
	if containsMsg != "" && !strings.Contains(msg, containsMsg) {
		return false
	}
	if containsRemedy != "" {
		remedy := strings.TrimSpace(item.Remediation)
		if query.CaseInsensitive {
			remedy = strings.ToLower(remedy)
		}
		if !strings.Contains(remedy, containsRemedy) {
			return false
		}
	}
	return true
}

func sortDiagnostics(items []Diagnostic, sortBySeverity, sortByCode bool) {
	sort.SliceStable(items, func(i, j int) bool {
		if sortBySeverity {
			leftRank := severityRank(items[i].Severity)
			rightRank := severityRank(items[j].Severity)
			if leftRank != rightRank {
				return leftRank > rightRank
			}
		}
		if sortByCode && items[i].Code != items[j].Code {
			return items[i].Code < items[j].Code
		}
		if items[i].Message != items[j].Message {
			return items[i].Message < items[j].Message
		}
		return items[i].Remediation < items[j].Remediation
	})
}
