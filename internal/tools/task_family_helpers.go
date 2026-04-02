package tools

import "strings"

func hasTaskUpdateFields(in taskUpdateInput) bool {
	return strings.TrimSpace(in.Status) != "" ||
		strings.TrimSpace(in.Result) != "" ||
		strings.TrimSpace(in.Error) != "" ||
		strings.TrimSpace(in.Subject) != "" ||
		strings.TrimSpace(in.Description) != "" ||
		strings.TrimSpace(in.ActiveForm) != "" ||
		strings.TrimSpace(in.Owner) != "" ||
		in.Metadata != nil
}
