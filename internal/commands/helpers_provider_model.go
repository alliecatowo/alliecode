package commands

import (
	"fmt"
	"strings"
)

func providerList() []string {
	return []string{"anthropic", "gemini", "ollama", "openai"}
}

func providerListMessage() string {
	known := providerList()
	lines := []string{"PROVIDER_LIST", fmt.Sprintf("count=%d", len(known))}
	for i, name := range known {
		lines = append(lines, fmt.Sprintf("provider.%d=%s", i+1, name))
	}
	lines = append(lines, "next=use_/provider_set_<name>_to_switch")
	return strings.Join(lines, "\n")
}
