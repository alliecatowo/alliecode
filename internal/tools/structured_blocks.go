package tools

import "strings"

type structuredField struct {
	Key   string
	Value string
}

func renderStructuredBlock(tag string, fields []structuredField) string {
	var b strings.Builder
	b.WriteString("[")
	b.WriteString(tag)
	b.WriteString("]\n")
	for _, field := range fields {
		b.WriteString(field.Key)
		b.WriteString(": ")
		b.WriteString(field.Value)
		b.WriteString("\n")
	}
	b.WriteString("[/")
	b.WriteString(tag)
	b.WriteString("]")
	return b.String()
}
