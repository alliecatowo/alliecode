package mcp

import "sort"

type ResourceSummary struct {
	Total        int            `json:"total"`
	ByServer     map[string]int `json:"by_server,omitempty"`
	ByMIMEType   map[string]int `json:"by_mime_type,omitempty"`
	MissingName  int            `json:"missing_name"`
	MissingMIME  int            `json:"missing_mime_type"`
	UniqueServer int            `json:"unique_servers"`
}

func normalizeResourceServerName(serverName string, listed []Resource) []Resource {
	out := make([]Resource, 0, len(listed))
	for _, item := range listed {
		if item.ServerName == "" {
			item.ServerName = serverName
		}
		out = append(out, item)
	}
	return out
}

func sortResourcesByIdentity(resources []Resource) {
	sort.Slice(resources, func(i, j int) bool {
		if resources[i].ServerName != resources[j].ServerName {
			return resources[i].ServerName < resources[j].ServerName
		}
		if resources[i].Name != resources[j].Name {
			return resources[i].Name < resources[j].Name
		}
		if resources[i].URI != resources[j].URI {
			return resources[i].URI < resources[j].URI
		}
		return resources[i].MIMEType < resources[j].MIMEType
	})
}

func sortResourcesByServerURI(resources []Resource) {
	sort.Slice(resources, func(i, j int) bool {
		if resources[i].ServerName != resources[j].ServerName {
			return resources[i].ServerName < resources[j].ServerName
		}
		return resources[i].URI < resources[j].URI
	})
}

func summarizeResources(resources []Resource) ResourceSummary {
	out := ResourceSummary{
		ByServer:   map[string]int{},
		ByMIMEType: map[string]int{},
	}
	for _, resource := range resources {
		out.Total++
		if resource.ServerName != "" {
			out.ByServer[resource.ServerName]++
		}
		if resource.MIMEType == "" {
			out.MissingMIME++
		} else {
			out.ByMIMEType[resource.MIMEType]++
		}
		if resource.Name == "" {
			out.MissingName++
		}
	}
	out.UniqueServer = len(out.ByServer)
	if len(out.ByServer) == 0 {
		out.ByServer = nil
	}
	if len(out.ByMIMEType) == 0 {
		out.ByMIMEType = nil
	}
	return out
}
