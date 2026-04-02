// Package mcp implements a Model Context Protocol client.
// MCP enables AllieCode to connect to external tool servers.
package mcp

import (
	"context"
	"encoding/json"
	"fmt"
	"sync"
	"sync/atomic"

	"github.com/alliecatowo/alliecode/internal/types"
)

// Client connects to an MCP server over stdio using JSON-RPC 2.0.
type Client struct {
	transportType TransportType
	transport     rpcTransport
	mu            sync.Mutex
	nextID        atomic.Int64
	pending       map[int64]chan jsonRPCResponse
	tools         []types.ToolDef
}

// ClientConfig configures MCP client transport selection.
type ClientConfig struct {
	Transport TransportType
	Command   string
	Args      []string
	URL       string
}

// jsonRPCRequest is a JSON-RPC 2.0 request.
type jsonRPCRequest struct {
	JSONRPC string      `json:"jsonrpc"`
	ID      int64       `json:"id"`
	Method  string      `json:"method"`
	Params  interface{} `json:"params,omitempty"`
}

// jsonRPCResponse is a JSON-RPC 2.0 response.
type jsonRPCResponse struct {
	JSONRPC string          `json:"jsonrpc"`
	ID      int64           `json:"id"`
	Result  json.RawMessage `json:"result,omitempty"`
	Error   *jsonRPCError   `json:"error,omitempty"`
}

type jsonRPCError struct {
	Code    int    `json:"code"`
	Message string `json:"message"`
}

// NewClient launches an MCP server as a subprocess and connects via stdio.
func NewClient(serverCmd string, args []string) (*Client, error) {
	return NewClientWithConfig(ClientConfig{Transport: TransportStdio, Command: serverCmd, Args: args})

}

// NewClientWithConfig creates an MCP client using the requested transport.
func NewClientWithConfig(cfg ClientConfig) (*Client, error) {
	transportType := normalizeTransportType(cfg.Transport)

	transport, err := newTransport(transportType, cfg)
	if err != nil {
		return nil, err
	}

	client := &Client{
		transportType: transportType,
		transport:     transport,
		pending:       make(map[int64]chan jsonRPCResponse),
	}

	// Start reading responses in background
	go client.readLoop()

	return client, nil
}

func newTransport(transportType TransportType, cfg ClientConfig) (rpcTransport, error) {
	switch transportType {
	case TransportStdio:
		if cfg.Command == "" {
			return nil, fmt.Errorf("mcp: stdio transport requires command")
		}
		return newStdioTransport(cfg.Command, cfg.Args)
	case TransportSSE, TransportWebSocket:
		return &unsupportedTransport{transportType: transportType}, nil
	default:
		return nil, fmt.Errorf("mcp: unknown transport %q", transportType)
	}
}

// Initialize performs the MCP initialization handshake.
func (c *Client) Initialize(ctx context.Context) error {
	params := map[string]interface{}{
		"protocolVersion": "2024-11-05",
		"capabilities":    map[string]interface{}{},
		"clientInfo": map[string]interface{}{
			"name":    "alliecode",
			"version": "0.1.0",
		},
	}

	_, err := c.call(ctx, "initialize", params)
	if err != nil {
		return fmt.Errorf("mcp: initialization failed: %w", err)
	}

	// Send initialized notification
	c.notify("notifications/initialized", nil)
	return nil
}

// ListTools queries the MCP server for available tools.
func (c *Client) ListTools(ctx context.Context) ([]types.ToolDef, error) {
	result, err := c.call(ctx, "tools/list", nil)
	if err != nil {
		return nil, fmt.Errorf("mcp: listing tools: %w", err)
	}

	var response struct {
		Tools []struct {
			Name        string          `json:"name"`
			Description string          `json:"description"`
			InputSchema json.RawMessage `json:"inputSchema"`
		} `json:"tools"`
	}

	if err := json.Unmarshal(result, &response); err != nil {
		return nil, fmt.Errorf("mcp: parsing tools response: %w", err)
	}

	tools := make([]types.ToolDef, len(response.Tools))
	for i, t := range response.Tools {
		var schema types.ToolSchema
		if t.InputSchema != nil {
			json.Unmarshal(t.InputSchema, &schema)
		}
		tools[i] = types.ToolDef{
			Name:        t.Name,
			Description: t.Description,
			InputSchema: schema,
		}
	}

	c.tools = tools
	return tools, nil
}

// CallTool invokes a tool on the MCP server.
func (c *Client) CallTool(ctx context.Context, name string, input json.RawMessage) (string, error) {
	params := map[string]interface{}{
		"name":      name,
		"arguments": json.RawMessage(input),
	}

	result, err := c.call(ctx, "tools/call", params)
	if err != nil {
		return "", fmt.Errorf("mcp: calling tool %q: %w", name, err)
	}

	var response struct {
		Content []struct {
			Type string `json:"type"`
			Text string `json:"text"`
		} `json:"content"`
		IsError bool `json:"isError"`
	}

	if err := json.Unmarshal(result, &response); err != nil {
		return "", fmt.Errorf("mcp: parsing tool result: %w", err)
	}

	var text string
	for _, c := range response.Content {
		if c.Type == "text" {
			text += c.Text
		}
	}

	if response.IsError {
		return "", fmt.Errorf("mcp tool error: %s", text)
	}

	return text, nil
}

// ListResources queries the MCP server for available resources.
func (c *Client) ListResources(ctx context.Context) ([]Resource, error) {
	result, err := c.call(ctx, "resources/list", nil)
	if err != nil {
		return nil, fmt.Errorf("mcp: listing resources: %w", err)
	}

	var response struct {
		Resources []struct {
			URI      string `json:"uri"`
			Name     string `json:"name"`
			MIMEType string `json:"mimeType"`
		} `json:"resources"`
	}

	if err := json.Unmarshal(result, &response); err != nil {
		return nil, fmt.Errorf("mcp: parsing resources response: %w", err)
	}

	resources := make([]Resource, len(response.Resources))
	for i, r := range response.Resources {
		resources[i] = Resource{
			URI:      r.URI,
			Name:     r.Name,
			MIMEType: r.MIMEType,
		}
	}

	return resources, nil
}

// ReadResource reads one resource URI from the MCP server.
func (c *Client) ReadResource(ctx context.Context, uri string) (ResourceContent, error) {
	params := map[string]interface{}{"uri": uri}
	result, err := c.call(ctx, "resources/read", params)
	if err != nil {
		return ResourceContent{}, fmt.Errorf("mcp: reading resource %q: %w", uri, err)
	}

	var response struct {
		Contents []struct {
			URI      string `json:"uri"`
			MIMEType string `json:"mimeType"`
			Text     string `json:"text"`
		} `json:"contents"`
	}

	if err := json.Unmarshal(result, &response); err != nil {
		return ResourceContent{}, fmt.Errorf("mcp: parsing resource response: %w", err)
	}
	if len(response.Contents) == 0 {
		return ResourceContent{}, fmt.Errorf("mcp: resource %q returned no content", uri)
	}

	first := response.Contents[0]
	return ResourceContent{URI: first.URI, MIMEType: first.MIMEType, Text: first.Text}, nil
}

// AuthenticateLocal performs a local auth flow with a token.
func (c *Client) AuthenticateLocal(ctx context.Context, token string) error {
	params := map[string]interface{}{"token": token}
	_, err := c.call(ctx, "auth/local", params)
	if err != nil {
		return fmt.Errorf("mcp: local auth failed: %w", err)
	}
	return nil
}

// Close shuts down the MCP server connection.
func (c *Client) Close() error {
	if c.transport == nil {
		return nil
	}
	return c.transport.Close()
}

// call sends a JSON-RPC request and waits for the response.
func (c *Client) call(ctx context.Context, method string, params interface{}) (json.RawMessage, error) {
	id := c.nextID.Add(1)

	req := jsonRPCRequest{
		JSONRPC: "2.0",
		ID:      id,
		Method:  method,
		Params:  params,
	}

	ch := make(chan jsonRPCResponse, 1)
	c.mu.Lock()
	c.pending[id] = ch
	c.mu.Unlock()

	defer func() {
		c.mu.Lock()
		delete(c.pending, id)
		c.mu.Unlock()
	}()

	data, err := json.Marshal(req)
	if err != nil {
		return nil, err
	}

	c.mu.Lock()
	err = c.transport.WriteMessage(append(data, '\n'))
	c.mu.Unlock()
	if err != nil {
		return nil, fmt.Errorf("writing request: %w", err)
	}

	select {
	case resp := <-ch:
		if resp.Error != nil {
			return nil, fmt.Errorf("json-rpc error %d: %s", resp.Error.Code, resp.Error.Message)
		}
		return resp.Result, nil
	case <-ctx.Done():
		return nil, ctx.Err()
	}
}

// notify sends a JSON-RPC notification (no response expected).
func (c *Client) notify(method string, params interface{}) {
	req := jsonRPCRequest{
		JSONRPC: "2.0",
		Method:  method,
		Params:  params,
	}
	data, _ := json.Marshal(req)
	c.mu.Lock()
	_ = c.transport.WriteMessage(append(data, '\n'))
	c.mu.Unlock()
}

// readLoop continuously reads JSON-RPC responses from the server.
func (c *Client) readLoop() {
	for {
		line, err := c.transport.ReadMessage()
		if err != nil {
			return // server closed
		}

		var resp jsonRPCResponse
		if err := json.Unmarshal(line, &resp); err != nil {
			continue
		}

		c.mu.Lock()
		ch, ok := c.pending[resp.ID]
		c.mu.Unlock()

		if ok {
			ch <- resp
		}
	}
}
