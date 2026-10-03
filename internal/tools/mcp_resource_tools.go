package tools

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"sort"
	"strings"

	"github.com/alliecatowo/alliecode/internal/mcp"
	"github.com/alliecatowo/alliecode/internal/types"
)

type mcpContractMetadata struct {
	Family          string `json:"family"`
	SchemaVersion   string `json:"schema_version"`
	StateBacked     bool   `json:"state_backed"`
	HasStatus       bool   `json:"has_status"`
	PermissionAware bool   `json:"permission_aware"`
}

func defaultMCPContractMetadata() *mcpContractMetadata {
	return &mcpContractMetadata{
		Family:          "mcp",
		SchemaVersion:   "v2",
		StateBacked:     true,
		HasStatus:       true,
		PermissionAware: true,
	}
}

type MCPResourceListTool struct {
	manager MCPManager
}

type mcpResourceListInput struct {
	ServerName string `json:"server_name,omitempty"`
}

type mcpResourceListOutput struct {
	Resources      []mcp.Resource       `json:"resources"`
	Total          int                  `json:"total"`
	ServerStatuses []mcp.ServerStatus   `json:"server_statuses,omitempty"`
	Summary        any                  `json:"summary,omitempty"`
	Contract       *mcpContractMetadata `json:"contract,omitempty"`
}

type mcpToolError struct {
	Code       string               `json:"code"`
	ErrorCode  string               `json:"error_code,omitempty"`
	Message    string               `json:"message"`
	Category   mcp.ErrorCategory    `json:"category,omitempty"`
	Hint       string               `json:"hint,omitempty"`
	Retryable  bool                 `json:"retryable,omitempty"`
	Operation  string               `json:"operation,omitempty"`
	ServerName string               `json:"server_name,omitempty"`
	ToolName   string               `json:"tool_name,omitempty"`
	Transport  mcp.TransportType    `json:"transport,omitempty"`
	Status     *mcp.ServerStatus    `json:"status,omitempty"`
	Contract   *mcpContractMetadata `json:"contract,omitempty"`
}

type mcpAuthStatusOutput struct {
	ServerName      string                    `json:"server_name,omitempty"`
	Authenticated   bool                      `json:"authenticated"`
	AuthStatus      mcp.AuthStatus            `json:"auth_status"`
	ConnectionState mcp.ServerConnectionState `json:"connection_state"`
	Status          *mcp.ServerStatus         `json:"status,omitempty"`
	Contract        *mcpContractMetadata      `json:"contract,omitempty"`
}

func deterministicMCPToolError(err mcpToolError) types.ToolResult {
	b, marshalErr := json.Marshal(err)
	if marshalErr != nil {
		return types.ToolResult{Content: `{"code":"MCP_SERIALIZATION_ERROR","message":"failed to serialize error"}`, IsError: true}
	}
	return types.ToolResult{Content: string(b), IsError: true}
}

func newMCPToolError(code string, err error) mcpToolError {
	classified := mcp.ClassifyError(err)
	return mcpToolError{Code: code, ErrorCode: classified.Code, Message: classified.Message, Category: classified.Category, Hint: classified.Hint, Retryable: classified.Retryable, Contract: defaultMCPContractMetadata()}
}

type mcpResourceSummary struct {
	ByServer       map[string]int `json:"by_server,omitempty"`
	ByMIMEType     map[string]int `json:"by_mime_type,omitempty"`
	UniqueServers  int            `json:"unique_servers"`
	MissingName    int            `json:"missing_name"`
	MissingMIME    int            `json:"missing_mime_type"`
	StatusOverview map[string]int `json:"status_overview,omitempty"`
}

func summarizeMCPResources(resources []mcp.Resource, statuses []mcp.ServerStatus) mcpResourceSummary {
	out := mcpResourceSummary{ByServer: map[string]int{}, ByMIMEType: map[string]int{}, StatusOverview: map[string]int{}}
	for _, resource := range resources {
		if resource.ServerName != "" {
			out.ByServer[resource.ServerName]++
		}
		if strings.TrimSpace(resource.MIMEType) != "" {
			out.ByMIMEType[resource.MIMEType]++
		} else {
			out.MissingMIME++
		}
		if strings.TrimSpace(resource.Name) == "" {
			out.MissingName++
		}
	}
	for _, status := range statuses {
		out.StatusOverview[string(status.ConnectionState)]++
	}
	out.UniqueServers = len(out.ByServer)
	if len(out.ByServer) == 0 {
		out.ByServer = nil
	}
	if len(out.ByMIMEType) == 0 {
		out.ByMIMEType = nil
	}
	if len(out.StatusOverview) == 0 {
		out.StatusOverview = nil
	}
	return out
}

func (t *MCPResourceListTool) Name() string { return "mcp_resource_list" }

func (t *MCPResourceListTool) Description() string {
	return "Lists resources exposed by configured MCP servers."
}

func (t *MCPResourceListTool) InputSchema() types.ToolSchema {
	return types.ToolSchema{
		Type: "object",
		Properties: map[string]types.PropertySchema{
			"server_name": {Type: "string", Description: "Optional MCP server name filter."},
		},
	}
}

func (t *MCPResourceListTool) Execute(ctx context.Context, input types.ToolInput, toolCtx types.ToolContext) (types.ToolResult, error) {
	_ = toolCtx
	mgr := t.manager
	if mgr == nil {
		return deterministicMCPToolError(mcpToolError{Code: "MCP_MANAGER_UNAVAILABLE", Message: "mcp manager is not configured", Contract: defaultMCPContractMetadata()}), nil
	}

	var in mcpResourceListInput
	if len(input) > 0 {
		if err := json.Unmarshal(input, &in); err != nil {
			return deterministicMCPToolError(mcpToolError{Code: "MCP_INVALID_INPUT", Message: fmt.Sprintf("invalid input: %v", err), Contract: defaultMCPContractMetadata()}), nil
		}
	}

	serverFilter := strings.TrimSpace(in.ServerName)
	resources := make([]mcp.Resource, 0)
	if serverFilter != "" {
		listed, err := mgr.ListResourcesForServer(ctx, serverFilter)
		if err != nil {
			if errors.Is(err, mcp.ErrTransportUnsupported) {
				status, _ := mgr.ServerStatus(serverFilter)
				return deterministicMCPToolError(mcpToolError{Code: "MCP_TRANSPORT_UNSUPPORTED", Message: fmt.Sprintf("resource listing unsupported for server %q", serverFilter), ServerName: serverFilter, Transport: status.TransportType, Status: &status, Operation: "resource_list"}), nil
			}
			status, ok := mgr.ServerStatus(serverFilter)
			if ok {
				outErr := newMCPToolError("MCP_RESOURCE_LIST_FAILED", err)
				outErr.Operation = "resource_list"
				outErr.ServerName = serverFilter
				outErr.Transport = status.TransportType
				outErr.Status = &status
				return deterministicMCPToolError(outErr), nil
			}
			outErr := newMCPToolError("MCP_RESOURCE_LIST_FAILED", err)
			outErr.Operation = "resource_list"
			outErr.ServerName = serverFilter
			return deterministicMCPToolError(outErr), nil
		}
		resources = append(resources, listed...)
	} else {
		listed, err := mgr.ListResources(ctx)
		if err != nil {
			if errors.Is(err, mcp.ErrTransportUnsupported) {
				return deterministicMCPToolError(mcpToolError{Code: "MCP_TRANSPORT_UNSUPPORTED", Message: "resource listing unsupported for at least one MCP server transport", Operation: "resource_list", Contract: defaultMCPContractMetadata()}), nil
			}
			outErr := newMCPToolError("MCP_RESOURCE_LIST_FAILED", err)
			outErr.Operation = "resource_list"
			return deterministicMCPToolError(outErr), nil
		}
		resources = listed
	}

	sort.Slice(resources, func(i, j int) bool {
		if resources[i].ServerName != resources[j].ServerName {
			return resources[i].ServerName < resources[j].ServerName
		}
		return resources[i].URI < resources[j].URI
	})

	serverStatuses := mgr.ServerStatuses()
	if len(serverStatuses) > 0 {
		sort.Slice(serverStatuses, func(i, j int) bool {
			return serverStatuses[i].ServerName < serverStatuses[j].ServerName
		})
	}

	summary := summarizeMCPResources(resources, serverStatuses)
	b, err := json.Marshal(mcpResourceListOutput{Resources: resources, Total: len(resources), ServerStatuses: serverStatuses, Summary: summary, Contract: defaultMCPContractMetadata()})
	if err != nil {
		return deterministicMCPToolError(mcpToolError{Code: "MCP_SERIALIZATION_ERROR", Message: fmt.Sprintf("serialization error: %v", err)}), nil
	}

	return types.ToolResult{Content: string(b)}, nil
}

func (t *MCPResourceListTool) IsReadOnly(input types.ToolInput) bool { return true }

func (t *MCPResourceListTool) IsDestructive(input types.ToolInput) bool { return false }

func (t *MCPResourceListTool) IsConcurrencySafe(input types.ToolInput) bool { return true }

func (t *MCPResourceListTool) CheckPermissions(input types.ToolInput, toolCtx types.ToolContext) types.ToolPermission {
	return evaluateToolPermissionByFamily(toolFamilyMCP, "mcp_resource_list", input, toolCtx, types.PermissionAllowed)
}

type MCPResourceReadTool struct {
	manager MCPManager
}

type mcpResourceReadInput struct {
	ServerName string `json:"server_name"`
	URI        string `json:"uri"`
}

type mcpResourceReadOutput struct {
	ServerName      string                    `json:"server_name"`
	Transport       mcp.TransportType         `json:"transport"`
	Authenticated   bool                      `json:"authenticated"`
	AuthStatus      mcp.AuthStatus            `json:"auth_status"`
	ConnectionState mcp.ServerConnectionState `json:"connection_state"`
	Content         mcp.ResourceContent       `json:"content"`
	Status          mcp.ServerStatus          `json:"status"`
	Contract        *mcpContractMetadata      `json:"contract,omitempty"`
}

type mcpToolInvokeInput struct {
	ServerName string          `json:"server_name"`
	ToolName   string          `json:"tool_name"`
	Arguments  json.RawMessage `json:"arguments,omitempty"`
}

type mcpToolInvokeOutput struct {
	ServerName      string                    `json:"server_name"`
	ToolName        string                    `json:"tool_name"`
	Transport       mcp.TransportType         `json:"transport"`
	Authenticated   bool                      `json:"authenticated"`
	AuthStatus      mcp.AuthStatus            `json:"auth_status"`
	ConnectionState mcp.ServerConnectionState `json:"connection_state"`
	Result          string                    `json:"result"`
	Status          mcp.ServerStatus          `json:"status"`
	Contract        *mcpContractMetadata      `json:"contract,omitempty"`
}

func (t *MCPResourceReadTool) Name() string { return "mcp_resource_read" }

func (t *MCPResourceReadTool) Description() string {
	return "Reads resource content from an MCP server by URI."
}

func (t *MCPResourceReadTool) InputSchema() types.ToolSchema {
	return types.ToolSchema{
		Type: "object",
		Properties: map[string]types.PropertySchema{
			"server_name": {Type: "string", Description: "MCP server name exposing the resource."},
			"uri":         {Type: "string", Description: "Resource URI."},
		},
		Required: []string{"server_name", "uri"},
	}
}

func (t *MCPResourceReadTool) Execute(ctx context.Context, input types.ToolInput, toolCtx types.ToolContext) (types.ToolResult, error) {
	_ = toolCtx
	mgr := t.manager
	if mgr == nil {
		return deterministicMCPToolError(mcpToolError{Code: "MCP_MANAGER_UNAVAILABLE", Message: "mcp manager is not configured", Contract: defaultMCPContractMetadata()}), nil
	}

	var in mcpResourceReadInput
	if err := json.Unmarshal(input, &in); err != nil {
		return deterministicMCPToolError(mcpToolError{Code: "MCP_INVALID_INPUT", Message: fmt.Sprintf("invalid input: %v", err), Contract: defaultMCPContractMetadata()}), nil
	}
	if strings.TrimSpace(in.ServerName) == "" {
		return deterministicMCPToolError(mcpToolError{Code: "MCP_SERVER_REQUIRED", Message: "server_name is required"}), nil
	}
	if strings.TrimSpace(in.URI) == "" {
		return deterministicMCPToolError(mcpToolError{Code: "MCP_URI_REQUIRED", Message: "uri is required", ServerName: in.ServerName}), nil
	}

	status, ok := mgr.ServerStatus(in.ServerName)
	if !ok {
		return deterministicMCPToolError(mcpToolError{Code: "MCP_SERVER_UNKNOWN", Message: fmt.Sprintf("unknown MCP server %q", in.ServerName), ServerName: in.ServerName}), nil
	}
	if status.TransportType != mcp.TransportStdio {
		return deterministicMCPToolError(mcpToolError{Code: "MCP_TRANSPORT_UNSUPPORTED", Message: fmt.Sprintf("resource read unsupported for server %q", in.ServerName), ServerName: in.ServerName, Transport: status.TransportType, Status: &status}), nil
	}
	if status.AuthStatus == mcp.AuthStatusUnauthenticated || status.ConnectionState == mcp.ServerConnectionNeedsAuth {
		return deterministicMCPToolError(mcpToolError{Code: "MCP_AUTH_REQUIRED", Message: fmt.Sprintf("MCP server %q requires authentication before reading resources", in.ServerName), ServerName: in.ServerName, Transport: status.TransportType, Status: &status, Operation: "resource_read", Hint: "Authenticate with mcp_auth_local before reading resources"}), nil
	}

	content, err := mgr.ReadResourceCached(ctx, in.ServerName, in.URI)
	if err != nil {
		if errors.Is(err, mcp.ErrTransportUnsupported) {
			return deterministicMCPToolError(mcpToolError{Code: "MCP_TRANSPORT_UNSUPPORTED", Message: fmt.Sprintf("resource read unsupported for server %q", in.ServerName), ServerName: in.ServerName, Transport: status.TransportType, Status: &status, Operation: "resource_read"}), nil
		}
		outErr := newMCPToolError("MCP_RESOURCE_READ_FAILED", err)
		outErr.Operation = "resource_read"
		outErr.ServerName = in.ServerName
		outErr.Transport = status.TransportType
		outErr.Status = &status
		return deterministicMCPToolError(outErr), nil
	}

	b, err := json.Marshal(mcpResourceReadOutput{
		Contract:        defaultMCPContractMetadata(),
		ServerName:      in.ServerName,
		Transport:       status.TransportType,
		Authenticated:   mgr.Authenticated(in.ServerName),
		AuthStatus:      status.AuthStatus,
		ConnectionState: status.ConnectionState,
		Content:         content,
		Status:          status,
	})
	if err != nil {
		return deterministicMCPToolError(mcpToolError{Code: "MCP_SERIALIZATION_ERROR", Message: fmt.Sprintf("serialization error: %v", err), ServerName: in.ServerName, Transport: status.TransportType, Status: &status}), nil
	}

	return types.ToolResult{Content: string(b)}, nil
}

func (t *MCPResourceReadTool) IsReadOnly(input types.ToolInput) bool { return true }

func (t *MCPResourceReadTool) IsDestructive(input types.ToolInput) bool { return false }

func (t *MCPResourceReadTool) IsConcurrencySafe(input types.ToolInput) bool { return true }

func (t *MCPResourceReadTool) CheckPermissions(input types.ToolInput, toolCtx types.ToolContext) types.ToolPermission {
	return evaluateToolPermissionByFamily(toolFamilyMCP, "mcp_resource_read", input, toolCtx, types.PermissionAllowed)
}

type MCPAuthLocalTool struct {
	manager MCPManager
}

type mcpAuthLocalInput struct {
	ServerName string `json:"server_name"`
	Token      string `json:"token"`
}

type mcpAuthLocalOutput struct {
	ServerName    string               `json:"server_name"`
	Authenticated bool                 `json:"authenticated"`
	Contract      *mcpContractMetadata `json:"contract,omitempty"`
}

func (t *MCPAuthLocalTool) Name() string { return "mcp_auth_local" }

func (t *MCPAuthLocalTool) Description() string {
	return "Authenticates to a local MCP server using a token."
}

func (t *MCPAuthLocalTool) InputSchema() types.ToolSchema {
	return types.ToolSchema{
		Type: "object",
		Properties: map[string]types.PropertySchema{
			"server_name": {Type: "string", Description: "MCP server name to authenticate."},
			"token":       {Type: "string", Description: "Authentication token for auth/local."},
		},
		Required: []string{"server_name", "token"},
	}
}

func (t *MCPAuthLocalTool) Execute(ctx context.Context, input types.ToolInput, toolCtx types.ToolContext) (types.ToolResult, error) {
	_ = toolCtx
	mgr := t.manager
	if mgr == nil {
		return deterministicMCPToolError(mcpToolError{Code: "MCP_MANAGER_UNAVAILABLE", Message: "mcp manager is not configured", Contract: defaultMCPContractMetadata()}), nil
	}

	var in mcpAuthLocalInput
	if err := json.Unmarshal(input, &in); err != nil {
		return deterministicMCPToolError(mcpToolError{Code: "MCP_INVALID_INPUT", Message: fmt.Sprintf("invalid input: %v", err), Contract: defaultMCPContractMetadata()}), nil
	}
	if strings.TrimSpace(in.ServerName) == "" {
		return deterministicMCPToolError(mcpToolError{Code: "MCP_SERVER_REQUIRED", Message: "server_name is required"}), nil
	}
	if strings.TrimSpace(in.Token) == "" {
		return deterministicMCPToolError(mcpToolError{Code: "MCP_TOKEN_REQUIRED", Message: "token is required", ServerName: in.ServerName}), nil
	}

	if err := mgr.AuthenticateLocal(ctx, in.ServerName, in.Token); err != nil {
		status, _ := mgr.ServerStatus(in.ServerName)
		outErr := newMCPToolError("MCP_AUTH_FAILED", err)
		outErr.Operation = "auth_local"
		outErr.ServerName = in.ServerName
		outErr.Transport = status.TransportType
		outErr.Status = &status
		return deterministicMCPToolError(outErr), nil
	}

	b, err := json.Marshal(mcpAuthLocalOutput{ServerName: in.ServerName, Authenticated: mgr.Authenticated(in.ServerName), Contract: defaultMCPContractMetadata()})
	if err != nil {
		return deterministicMCPToolError(mcpToolError{Code: "MCP_SERIALIZATION_ERROR", Message: fmt.Sprintf("serialization error: %v", err), ServerName: in.ServerName}), nil
	}

	return types.ToolResult{Content: string(b)}, nil
}

func (t *MCPAuthLocalTool) IsReadOnly(input types.ToolInput) bool { return false }

func (t *MCPAuthLocalTool) IsDestructive(input types.ToolInput) bool { return false }

func (t *MCPAuthLocalTool) IsConcurrencySafe(input types.ToolInput) bool { return true }

func (t *MCPAuthLocalTool) CheckPermissions(input types.ToolInput, toolCtx types.ToolContext) types.ToolPermission {
	return evaluateToolPermissionByFamily(toolFamilyMCP, "mcp_auth_local", input, toolCtx, types.PermissionAllowed)
}

type MCPAuthStatusTool struct {
	manager MCPManager
}

type mcpAuthStatusInput struct {
	ServerName string `json:"server_name,omitempty"`
}

func (t *MCPAuthStatusTool) Name() string { return "mcp_auth_status" }

func (t *MCPAuthStatusTool) Description() string {
	return "Reports authentication status for one MCP server."
}

func (t *MCPAuthStatusTool) InputSchema() types.ToolSchema {
	return types.ToolSchema{
		Type: "object",
		Properties: map[string]types.PropertySchema{
			"server_name": {Type: "string", Description: "MCP server name to inspect."},
		},
		Required: []string{"server_name"},
	}
}

func (t *MCPAuthStatusTool) Execute(ctx context.Context, input types.ToolInput, toolCtx types.ToolContext) (types.ToolResult, error) {
	_ = ctx
	_ = toolCtx
	mgr := t.manager
	if mgr == nil {
		return deterministicMCPToolError(mcpToolError{Code: "MCP_MANAGER_UNAVAILABLE", Message: "mcp manager is not configured", Contract: defaultMCPContractMetadata()}), nil
	}

	var in mcpAuthStatusInput
	if err := json.Unmarshal(input, &in); err != nil {
		return deterministicMCPToolError(mcpToolError{Code: "MCP_INVALID_INPUT", Message: fmt.Sprintf("invalid input: %v", err), Contract: defaultMCPContractMetadata()}), nil
	}
	serverName := strings.TrimSpace(in.ServerName)
	if serverName == "" {
		return deterministicMCPToolError(mcpToolError{Code: "MCP_SERVER_REQUIRED", Message: "server_name is required"}), nil
	}

	status, ok := mgr.ServerStatus(serverName)
	if !ok {
		return deterministicMCPToolError(mcpToolError{Code: "MCP_SERVER_UNKNOWN", Message: fmt.Sprintf("unknown MCP server %q", serverName), ServerName: serverName}), nil
	}

	b, err := json.Marshal(mcpAuthStatusOutput{
		Contract:        defaultMCPContractMetadata(),
		ServerName:      serverName,
		Authenticated:   status.Authenticated,
		AuthStatus:      status.AuthStatus,
		ConnectionState: status.ConnectionState,
		Status:          &status,
	})
	if err != nil {
		return deterministicMCPToolError(mcpToolError{Code: "MCP_SERIALIZATION_ERROR", Message: fmt.Sprintf("serialization error: %v", err), ServerName: serverName, Transport: status.TransportType, Status: &status}), nil
	}

	return types.ToolResult{Content: string(b)}, nil
}

func (t *MCPAuthStatusTool) IsReadOnly(input types.ToolInput) bool { return true }

func (t *MCPAuthStatusTool) IsDestructive(input types.ToolInput) bool { return false }

func (t *MCPAuthStatusTool) IsConcurrencySafe(input types.ToolInput) bool { return true }

func (t *MCPAuthStatusTool) CheckPermissions(input types.ToolInput, toolCtx types.ToolContext) types.ToolPermission {
	return evaluateToolPermissionByFamily(toolFamilyMCP, "mcp_auth_status", input, toolCtx, types.PermissionAllowed)
}

type MCPToolInvokeTool struct {
	manager MCPManager
}

func (t *MCPToolInvokeTool) Name() string { return "mcp_tool_invoke" }

func (t *MCPToolInvokeTool) Description() string {
	return "Invokes a tool on a specific MCP server with explicit auth and status diagnostics."
}

func (t *MCPToolInvokeTool) InputSchema() types.ToolSchema {
	return types.ToolSchema{
		Type: "object",
		Properties: map[string]types.PropertySchema{
			"server_name": {Type: "string", Description: "MCP server name that hosts the tool."},
			"tool_name":   {Type: "string", Description: "Exact MCP tool name exposed by the server."},
			"arguments":   {Type: "object", Description: "JSON object passed to MCP tools/call arguments."},
		},
		Required: []string{"server_name", "tool_name"},
	}
}

func (t *MCPToolInvokeTool) Execute(ctx context.Context, input types.ToolInput, toolCtx types.ToolContext) (types.ToolResult, error) {
	_ = toolCtx
	mgr := t.manager
	if mgr == nil {
		return deterministicMCPToolError(mcpToolError{Code: "MCP_MANAGER_UNAVAILABLE", Message: "mcp manager is not configured", Contract: defaultMCPContractMetadata()}), nil
	}

	var in mcpToolInvokeInput
	if err := json.Unmarshal(input, &in); err != nil {
		return deterministicMCPToolError(mcpToolError{Code: "MCP_INVALID_INPUT", Message: fmt.Sprintf("invalid input: %v", err), Contract: defaultMCPContractMetadata()}), nil
	}
	in.ServerName = strings.TrimSpace(in.ServerName)
	in.ToolName = strings.TrimSpace(in.ToolName)
	if in.ServerName == "" {
		return deterministicMCPToolError(mcpToolError{Code: "MCP_SERVER_REQUIRED", Message: "server_name is required"}), nil
	}
	if in.ToolName == "" {
		return deterministicMCPToolError(mcpToolError{Code: "MCP_TOOL_REQUIRED", Message: "tool_name is required", ServerName: in.ServerName}), nil
	}

	status, ok := mgr.ServerStatus(in.ServerName)
	if !ok {
		return deterministicMCPToolError(mcpToolError{Code: "MCP_SERVER_UNKNOWN", Message: fmt.Sprintf("unknown MCP server %q", in.ServerName), ServerName: in.ServerName}), nil
	}
	if status.ConnectionState == mcp.ServerConnectionDisabled {
		return deterministicMCPToolError(mcpToolError{Code: "MCP_SERVER_DISABLED", Message: fmt.Sprintf("mcp server %q is disabled", in.ServerName), ServerName: in.ServerName, ToolName: in.ToolName, Transport: status.TransportType, Status: &status}), nil
	}

	args := json.RawMessage(`{}`)
	if len(in.Arguments) > 0 {
		var argsObj map[string]any
		if err := json.Unmarshal(in.Arguments, &argsObj); err != nil {
			return deterministicMCPToolError(mcpToolError{Code: "MCP_INVALID_ARGUMENTS", Message: fmt.Sprintf("arguments must be a JSON object: %v", err), ServerName: in.ServerName, ToolName: in.ToolName, Transport: status.TransportType, Status: &status}), nil
		}
		args = in.Arguments
	}

	result, err := mgr.Call(ctx, in.ServerName, in.ToolName, args)
	if err != nil {
		if mcp.IsNeedsAuthError(err) {
			_ = mgr.SetAuthStateNeedsAuth(in.ServerName)
			refreshed, _ := mgr.ServerStatus(in.ServerName)
			return deterministicMCPToolError(mcpToolError{Code: "MCP_AUTH_REQUIRED", Message: fmt.Sprintf("MCP server %q requires authentication before invoking tool %q", in.ServerName, in.ToolName), ServerName: in.ServerName, ToolName: in.ToolName, Transport: refreshed.TransportType, Status: &refreshed, Operation: "tool_invoke", Hint: "Authenticate with mcp_auth_local and retry"}), nil
		}
		_ = mgr.SetAuthStateFailed(in.ServerName)
		refreshed, _ := mgr.ServerStatus(in.ServerName)
		outErr := newMCPToolError("MCP_TOOL_INVOKE_FAILED", err)
		outErr.Operation = "tool_invoke"
		outErr.ServerName = in.ServerName
		outErr.ToolName = in.ToolName
		outErr.Transport = refreshed.TransportType
		outErr.Status = &refreshed
		return deterministicMCPToolError(outErr), nil
	}

	_ = mgr.SetAuthStateAuthenticated(in.ServerName)
	refreshed, ok := mgr.ServerStatus(in.ServerName)
	if !ok {
		refreshed = status
	}

	b, err := json.Marshal(mcpToolInvokeOutput{
		Contract:        defaultMCPContractMetadata(),
		ServerName:      in.ServerName,
		ToolName:        in.ToolName,
		Transport:       refreshed.TransportType,
		Authenticated:   mgr.Authenticated(in.ServerName),
		AuthStatus:      refreshed.AuthStatus,
		ConnectionState: refreshed.ConnectionState,
		Result:          result,
		Status:          refreshed,
	})
	if err != nil {
		return deterministicMCPToolError(mcpToolError{Code: "MCP_SERIALIZATION_ERROR", Message: fmt.Sprintf("serialization error: %v", err), ServerName: in.ServerName, ToolName: in.ToolName, Transport: refreshed.TransportType, Status: &refreshed}), nil
	}

	return types.ToolResult{Content: string(b)}, nil
}

func (t *MCPToolInvokeTool) IsReadOnly(input types.ToolInput) bool { return false }

func (t *MCPToolInvokeTool) IsDestructive(input types.ToolInput) bool { return false }

func (t *MCPToolInvokeTool) IsConcurrencySafe(input types.ToolInput) bool { return true }

func (t *MCPToolInvokeTool) CheckPermissions(input types.ToolInput, toolCtx types.ToolContext) types.ToolPermission {
	return evaluateToolPermissionByFamily(toolFamilyMCP, "mcp_tool_invoke", input, toolCtx, types.PermissionAllowed)
}
