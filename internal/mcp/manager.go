package mcp

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"sort"
	"strings"
	"sync"
	"time"

	"github.com/alliecatowo/alliecode/internal/types"
)

// ServerConfig describes one MCP server process.
type ServerConfig struct {
	Name      string
	Transport TransportType
	Command   string
	Args      []string
	URL       string
	Disabled  bool
}

// ServerConnectionState tracks high-level MCP server lifecycle state.
type ServerConnectionState string

const (
	ServerConnectionConnected ServerConnectionState = "connected"
	ServerConnectionFailed    ServerConnectionState = "failed"
	ServerConnectionNeedsAuth ServerConnectionState = "needs_auth"
	ServerConnectionPending   ServerConnectionState = "pending"
	ServerConnectionDisabled  ServerConnectionState = "disabled"
)

// RemoteTool describes one tool exposed by a specific MCP server.
type RemoteTool struct {
	ServerName string
	Def        types.ToolDef
}

// Resource describes a resource exposed by an MCP server.
type Resource struct {
	ServerName string
	URI        string
	Name       string
	MIMEType   string
}

// ResourceContent contains read data for one resource.
type ResourceContent struct {
	URI      string
	MIMEType string
	Text     string
}

// ServerStatus captures resource/auth status for one server.
type ServerStatus struct {
	ServerName           string
	TransportType        TransportType
	ConnectionState      ServerConnectionState
	Authenticated        bool
	AuthStatus           AuthStatus
	LastErrorCategory    ErrorCategory
	CachedResourceCount  int
	CachedContentEntries int
	LastError            string
	LastErrorAt          time.Time
	LastSuccessAt        time.Time
	LastAuthAttemptAt    time.Time
}

type managerClient interface {
	Initialize(ctx context.Context) error
	ListTools(ctx context.Context) ([]types.ToolDef, error)
	CallTool(ctx context.Context, name string, input json.RawMessage) (string, error)
	ListResources(ctx context.Context) ([]Resource, error)
	ReadResource(ctx context.Context, uri string) (ResourceContent, error)
	AuthenticateLocal(ctx context.Context, token string) error
	Close() error
}

// AuthStatus represents local MCP auth lifecycle state.
type AuthStatus string

const (
	AuthStatusUnknown         AuthStatus = "unknown"
	AuthStatusUnauthenticated AuthStatus = "unauthenticated"
	AuthStatusAuthenticating  AuthStatus = "authenticating"
	AuthStatusAuthenticated   AuthStatus = "authenticated"
	AuthStatusFailed          AuthStatus = "failed"
)

// Manager owns MCP server clients and aggregates remote tools.
type Manager struct {
	mu                sync.RWMutex
	clients           map[string]managerClient
	serverConfigs     map[string]ServerConfig
	tools             []RemoteTool
	resourceList      []Resource
	resourceContents  map[string]ResourceContent
	authenticated     map[string]bool
	authStatus        map[string]AuthStatus
	connectionState   map[string]ServerConnectionState
	transportType     map[string]TransportType
	lastError         map[string]string
	lastErrorCategory map[string]ErrorCategory
	lastErrorAt       map[string]time.Time
	lastSuccessAt     map[string]time.Time
	lastAuthAttempt   map[string]time.Time
	onInvalidate      func(serverName string)
}

func NewManager(ctx context.Context, configs []ServerConfig) (*Manager, error) {
	m := &Manager{
		clients:           make(map[string]managerClient),
		serverConfigs:     make(map[string]ServerConfig),
		resourceContents:  make(map[string]ResourceContent),
		authenticated:     make(map[string]bool),
		authStatus:        make(map[string]AuthStatus),
		connectionState:   make(map[string]ServerConnectionState),
		transportType:     make(map[string]TransportType),
		lastError:         make(map[string]string),
		lastErrorCategory: make(map[string]ErrorCategory),
		lastErrorAt:       make(map[string]time.Time),
		lastSuccessAt:     make(map[string]time.Time),
		lastAuthAttempt:   make(map[string]time.Time),
	}

	for _, cfg := range configs {
		if cfg.Name == "" {
			cfg.Name = cfg.Command
			if cfg.Name == "" {
				cfg.Name = cfg.URL
			}
		}
		cfg.Name = strings.TrimSpace(cfg.Name)
		if cfg.Name == "" {
			m.Close()
			return nil, fmt.Errorf("mcp manager: server name cannot be empty")
		}
		transportType := normalizeTransportType(cfg.Transport)
		cfg.Transport = transportType
		m.serverConfigs[cfg.Name] = cfg
		m.transportType[cfg.Name] = transportType
		if cfg.Disabled {
			m.authenticated[cfg.Name] = false
			m.authStatus[cfg.Name] = AuthStatusUnauthenticated
			m.connectionState[cfg.Name] = ServerConnectionDisabled
			continue
		}

		client, err := NewClientWithConfig(ClientConfig{
			Transport: transportType,
			Command:   cfg.Command,
			Args:      cfg.Args,
			URL:       cfg.URL,
		})
		if err != nil {
			m.Close()
			return nil, fmt.Errorf("mcp manager: start %s: %w", cfg.Name, err)
		}
		m.clients[cfg.Name] = client
		m.authenticated[cfg.Name] = false
		m.authStatus[cfg.Name] = AuthStatusUnauthenticated
		m.connectionState[cfg.Name] = ServerConnectionPending

		if err := client.Initialize(ctx); err != nil {
			m.connectionState[cfg.Name] = ServerConnectionFailed
			m.setServerErrorLocked(cfg.Name, err)
			m.Close()
			return nil, fmt.Errorf("mcp manager: initialize %s: %w", cfg.Name, err)
		}

		defs, err := client.ListTools(ctx)
		if err != nil {
			m.connectionState[cfg.Name] = ServerConnectionFailed
			m.setServerErrorLocked(cfg.Name, err)
			m.Close()
			return nil, fmt.Errorf("mcp manager: list tools %s: %w", cfg.Name, err)
		}
		m.connectionState[cfg.Name] = ServerConnectionConnected
		m.setServerSuccessLocked(cfg.Name)
		for _, def := range defs {
			m.tools = append(m.tools, RemoteTool{ServerName: cfg.Name, Def: def})
		}
	}

	return m, nil
}

// AddServerConfig registers one server config in manager state.
// This mutates manager metadata only and does not establish a live connection.
func (m *Manager) AddServerConfig(cfg ServerConfig) error {
	name := strings.TrimSpace(cfg.Name)
	if name == "" {
		name = strings.TrimSpace(cfg.Command)
	}
	if name == "" {
		name = strings.TrimSpace(cfg.URL)
	}
	if name == "" {
		return fmt.Errorf("mcp manager: server name cannot be empty")
	}

	cfg.Name = name
	cfg.Transport = normalizeTransportType(cfg.Transport)

	m.mu.Lock()
	defer m.mu.Unlock()

	if m.serverConfigs == nil {
		m.serverConfigs = make(map[string]ServerConfig)
	}
	if _, exists := m.serverConfigs[name]; exists {
		return fmt.Errorf("mcp manager: server %q already exists", name)
	}

	m.serverConfigs[name] = cfg
	if m.transportType == nil {
		m.transportType = make(map[string]TransportType)
	}
	m.transportType[name] = cfg.Transport
	if m.authenticated == nil {
		m.authenticated = make(map[string]bool)
	}
	m.authenticated[name] = false
	if m.authStatus == nil {
		m.authStatus = make(map[string]AuthStatus)
	}
	m.authStatus[name] = AuthStatusUnauthenticated
	if m.connectionState == nil {
		m.connectionState = make(map[string]ServerConnectionState)
	}
	if cfg.Disabled {
		m.connectionState[name] = ServerConnectionDisabled
	} else {
		m.connectionState[name] = ServerConnectionPending
	}

	return nil
}

// RemoveServerConfig removes one server config and related cached manager state.
func (m *Manager) RemoveServerConfig(serverName string) bool {
	serverName = strings.TrimSpace(serverName)
	if serverName == "" {
		return false
	}

	var client managerClient
	var hook func(string)

	m.mu.Lock()
	_, hadConfig := m.serverConfigs[serverName]
	if !hadConfig {
		m.mu.Unlock()
		return false
	}
	delete(m.serverConfigs, serverName)

	client = m.clients[serverName]
	delete(m.clients, serverName)
	delete(m.transportType, serverName)
	delete(m.authenticated, serverName)
	delete(m.authStatus, serverName)
	delete(m.connectionState, serverName)
	delete(m.lastError, serverName)
	delete(m.lastErrorCategory, serverName)
	delete(m.lastErrorAt, serverName)
	delete(m.lastSuccessAt, serverName)
	delete(m.lastAuthAttempt, serverName)

	filteredTools := make([]RemoteTool, 0, len(m.tools))
	for _, tool := range m.tools {
		if tool.ServerName != serverName {
			filteredTools = append(filteredTools, tool)
		}
	}
	m.tools = filteredTools

	hook = m.clearServerCacheLocked(serverName)
	m.mu.Unlock()

	if client != nil {
		_ = client.Close()
	}
	if hook != nil {
		hook(serverName)
	}

	return true
}

// ServerConfigs returns a deterministic copy of registered server configs.
func (m *Manager) ServerConfigs() []ServerConfig {
	m.mu.RLock()
	defer m.mu.RUnlock()

	names := make([]string, 0, len(m.serverConfigs))
	for name := range m.serverConfigs {
		names = append(names, name)
	}
	sort.Strings(names)

	out := make([]ServerConfig, 0, len(names))
	for _, name := range names {
		cfg := m.serverConfigs[name]
		cfg.Name = name
		cfg.Transport = normalizeTransportType(cfg.Transport)
		out = append(out, cfg)
	}

	return out
}

// ServerTransportType reports the configured transport type for one server.
func (m *Manager) ServerTransportType(serverName string) TransportType {
	m.mu.RLock()
	defer m.mu.RUnlock()
	return m.serverTransportTypeLocked(serverName)
}

// ServerStatus reports one server status when known.
func (m *Manager) ServerStatus(serverName string) (ServerStatus, bool) {
	m.mu.RLock()
	defer m.mu.RUnlock()

	_, hasClient := m.clients[serverName]
	_, hasState := m.connectionState[serverName]
	_, hasAuth := m.authStatus[serverName]
	_, hasTransport := m.transportType[serverName]
	if !hasClient && !hasState && !hasAuth && !hasTransport {
		return ServerStatus{}, false
	}

	status := ServerStatus{
		ServerName:        serverName,
		TransportType:     m.serverTransportTypeLocked(serverName),
		ConnectionState:   m.connectionState[serverName],
		Authenticated:     m.authenticated[serverName],
		AuthStatus:        m.authStatus[serverName],
		LastErrorCategory: m.lastErrorCategory[serverName],
		LastError:         m.lastError[serverName],
		LastErrorAt:       m.lastErrorAt[serverName],
		LastSuccessAt:     m.lastSuccessAt[serverName],
		LastAuthAttemptAt: m.lastAuthAttempt[serverName],
	}

	for _, item := range m.resourceList {
		if item.ServerName == serverName {
			status.CachedResourceCount++
		}
	}

	prefix := serverName + "\n"
	for key := range m.resourceContents {
		if strings.HasPrefix(key, prefix) {
			status.CachedContentEntries++
		}
	}

	normalizeServerStatusFields(&status)

	return status, true
}

func (m *Manager) Tools() []RemoteTool {
	out := make([]RemoteTool, len(m.tools))
	copy(out, m.tools)
	return out
}

func (m *Manager) ServerStatusSummary() StatusSummary {
	return summarizeServerStatuses(m.ServerStatuses())
}

func (m *Manager) ResourceSummary() ResourceSummary {
	m.mu.RLock()
	resources := make([]Resource, len(m.resourceList))
	copy(resources, m.resourceList)
	m.mu.RUnlock()
	return summarizeResources(resources)
}

func (m *Manager) Call(ctx context.Context, serverName, toolName string, input json.RawMessage) (string, error) {
	m.mu.RLock()
	if m.connectionState[serverName] == ServerConnectionDisabled {
		m.mu.RUnlock()
		return "", fmt.Errorf("mcp server %q is disabled", serverName)
	}
	m.mu.RUnlock()

	client, ok := m.clients[serverName]
	if !ok {
		return "", fmt.Errorf("unknown MCP server %q", serverName)
	}
	out, err := client.CallTool(ctx, toolName, input)
	m.mu.Lock()
	if err != nil {
		m.setConnectionStateLocked(serverName, ServerConnectionFailed)
		m.setServerErrorLocked(serverName, err)
	} else if m.connectionState[serverName] != ServerConnectionDisabled {
		m.setConnectionStateLocked(serverName, ServerConnectionConnected)
		m.setServerSuccessLocked(serverName)
	}
	m.mu.Unlock()
	return out, err
}

// SetAuthStateNeedsAuth marks a server as unauthenticated and needing auth.
func (m *Manager) SetAuthStateNeedsAuth(serverName string) error {
	serverName = strings.TrimSpace(serverName)
	m.mu.Lock()
	defer m.mu.Unlock()
	if !m.serverKnownLocked(serverName) {
		return fmt.Errorf("unknown MCP server %q", serverName)
	}
	m.setServerAuthStateLocked(serverName, false, AuthStatusUnauthenticated, ServerConnectionNeedsAuth)
	return nil
}

// SetAuthStateAuthenticating marks a server auth flow as in progress.
func (m *Manager) SetAuthStateAuthenticating(serverName string) error {
	serverName = strings.TrimSpace(serverName)
	m.mu.Lock()
	defer m.mu.Unlock()
	if !m.serverKnownLocked(serverName) {
		return fmt.Errorf("unknown MCP server %q", serverName)
	}
	m.setServerAuthStateLocked(serverName, false, AuthStatusAuthenticating, ServerConnectionPending)
	return nil
}

// SetAuthStateAuthenticated marks a server as authenticated and connected.
func (m *Manager) SetAuthStateAuthenticated(serverName string) error {
	serverName = strings.TrimSpace(serverName)
	m.mu.Lock()
	defer m.mu.Unlock()
	if !m.serverKnownLocked(serverName) {
		return fmt.Errorf("unknown MCP server %q", serverName)
	}
	m.setServerAuthStateLocked(serverName, true, AuthStatusAuthenticated, ServerConnectionConnected)
	return nil
}

// SetAuthStateFailed marks a server auth flow as failed.
func (m *Manager) SetAuthStateFailed(serverName string) error {
	serverName = strings.TrimSpace(serverName)
	m.mu.Lock()
	defer m.mu.Unlock()
	if !m.serverKnownLocked(serverName) {
		return fmt.Errorf("unknown MCP server %q", serverName)
	}
	m.setServerAuthStateLocked(serverName, false, AuthStatusFailed, ServerConnectionFailed)
	return nil
}

// ListResourcesForServer lists resources for one server and refreshes its cache.
func (m *Manager) ListResourcesForServer(ctx context.Context, serverName string) ([]Resource, error) {
	serverName = strings.TrimSpace(serverName)
	if serverName == "" {
		return nil, fmt.Errorf("mcp manager: server name is required")
	}

	client, err := m.resourceClientForOperation(serverName, "list resources")
	if err != nil {
		return nil, err
	}

	listed, err := client.ListResources(ctx)
	if err != nil {
		m.mu.Lock()
		m.setConnectionStateLocked(serverName, ServerConnectionFailed)
		m.setServerErrorLocked(serverName, err)
		m.mu.Unlock()
		return nil, fmt.Errorf("mcp manager: list resources %s: %w", serverName, err)
	}

	resources := normalizeResourceServerName(serverName, listed)

	sortResourcesByIdentity(resources)

	m.mu.Lock()
	if m.connectionState[serverName] != ServerConnectionDisabled {
		m.setConnectionStateLocked(serverName, ServerConnectionConnected)
	}
	m.setServerSuccessLocked(serverName)
	m.replaceServerResourcesLocked(serverName, resources)
	m.mu.Unlock()

	return resources, nil
}

func (m *Manager) ListResources(ctx context.Context) ([]Resource, error) {
	serverNames := m.resourceServerNames()
	resources := make([]Resource, 0)
	for _, serverName := range serverNames {
		listed, err := m.ListResourcesForServer(ctx, serverName)
		if err != nil {
			return nil, err
		}
		resources = append(resources, listed...)
	}

	sortResourcesByIdentity(resources)

	m.mu.Lock()
	m.resourceList = make([]Resource, len(resources))
	copy(m.resourceList, resources)
	m.mu.Unlock()

	return resources, nil
}

// CachedResources returns the latest successful resource listing.
func (m *Manager) CachedResources() []Resource {
	m.mu.RLock()
	defer m.mu.RUnlock()

	out := make([]Resource, len(m.resourceList))
	copy(out, m.resourceList)
	return out
}

func (m *Manager) ReadResource(ctx context.Context, serverName, uri string) (ResourceContent, error) {
	client, err := m.resourceClientForOperation(serverName, "read resource")
	if err != nil {
		return ResourceContent{}, err
	}

	content, err := client.ReadResource(ctx, uri)
	if err != nil {
		m.mu.Lock()
		m.setConnectionStateLocked(serverName, ServerConnectionFailed)
		m.setServerErrorLocked(serverName, err)
		m.mu.Unlock()
		return ResourceContent{}, fmt.Errorf("mcp manager: read resource %s: %w", serverName, err)
	}

	m.mu.Lock()
	if m.connectionState[serverName] != ServerConnectionDisabled {
		m.setConnectionStateLocked(serverName, ServerConnectionConnected)
	}
	m.setServerSuccessLocked(serverName)
	m.resourceContents[resourceCacheKey(serverName, uri)] = content
	m.mu.Unlock()

	return content, nil
}

// ReadResourceCached returns a cached resource when available.
func (m *Manager) ReadResourceCached(ctx context.Context, serverName, uri string) (ResourceContent, error) {
	key := resourceCacheKey(serverName, uri)

	m.mu.RLock()
	content, ok := m.resourceContents[key]
	m.mu.RUnlock()
	if ok {
		return content, nil
	}

	return m.ReadResource(ctx, serverName, uri)
}

func (m *Manager) AuthenticateLocal(ctx context.Context, serverName, token string) error {
	serverName = strings.TrimSpace(serverName)
	if serverName == "" {
		return fmt.Errorf("unknown MCP server %q", serverName)
	}

	m.mu.RLock()
	client, ok := m.clients[serverName]
	m.mu.RUnlock()
	if !ok {
		return fmt.Errorf("unknown MCP server %q", serverName)
	}
	if err := ctx.Err(); err != nil {
		_ = m.SetAuthStateFailed(serverName)
		return fmt.Errorf("mcp manager: authenticate %s: %w", serverName, err)
	}
	if strings.TrimSpace(token) == "" {
		_ = m.SetAuthStateNeedsAuth(serverName)
		return fmt.Errorf("mcp manager: authenticate %s: empty token", serverName)
	}

	_ = m.SetAuthStateAuthenticating(serverName)
	m.mu.Lock()
	if m.lastAuthAttempt == nil {
		m.lastAuthAttempt = make(map[string]time.Time)
	}
	m.lastAuthAttempt[serverName] = time.Now().UTC()
	m.mu.Unlock()

	if err := client.AuthenticateLocal(ctx, token); err != nil {
		var hook func(string)
		m.mu.Lock()
		if !errors.Is(err, context.Canceled) && !errors.Is(err, context.DeadlineExceeded) {
			hook = m.clearServerCacheLocked(serverName)
		}
		m.setServerErrorLocked(serverName, err)
		if isNeedsAuthError(err) {
			m.setServerAuthStateLocked(serverName, false, AuthStatusUnauthenticated, ServerConnectionNeedsAuth)
		} else {
			m.setServerAuthStateLocked(serverName, false, AuthStatusFailed, ServerConnectionFailed)
		}
		m.mu.Unlock()
		if hook != nil {
			hook(serverName)
		}
		return fmt.Errorf("mcp manager: authenticate %s: %w", serverName, err)
	}

	var hook func(string)
	m.mu.Lock()
	hook = m.clearServerCacheLocked(serverName)
	m.setServerAuthStateLocked(serverName, true, AuthStatusAuthenticated, ServerConnectionConnected)
	m.setServerSuccessLocked(serverName)
	m.mu.Unlock()
	if hook != nil {
		hook(serverName)
	}

	return nil
}

// Authenticated reports whether local auth has completed for a server.
func (m *Manager) Authenticated(serverName string) bool {
	m.mu.RLock()
	defer m.mu.RUnlock()
	return m.authenticated[serverName]
}

// AuthStatus reports the lifecycle auth status for one server.
func (m *Manager) AuthStatus(serverName string) AuthStatus {
	m.mu.RLock()
	defer m.mu.RUnlock()
	status, ok := m.authStatus[serverName]
	if !ok {
		return AuthStatusUnknown
	}
	return status
}

// ServerStatuses returns auth and cache status for all known servers.
func (m *Manager) ServerStatuses() []ServerStatus {
	m.mu.RLock()
	defer m.mu.RUnlock()

	statusByServer := make(map[string]*ServerStatus, len(m.clients))
	for serverName := range m.clients {
		statusByServer[serverName] = &ServerStatus{
			ServerName:        serverName,
			TransportType:     m.serverTransportTypeLocked(serverName),
			ConnectionState:   m.connectionState[serverName],
			Authenticated:     m.authenticated[serverName],
			AuthStatus:        m.authStatus[serverName],
			LastErrorCategory: m.lastErrorCategory[serverName],
			LastError:         m.lastError[serverName],
			LastErrorAt:       m.lastErrorAt[serverName],
			LastSuccessAt:     m.lastSuccessAt[serverName],
			LastAuthAttemptAt: m.lastAuthAttempt[serverName],
		}
	}
	for serverName, state := range m.connectionState {
		if _, ok := statusByServer[serverName]; ok {
			continue
		}
		statusByServer[serverName] = &ServerStatus{
			ServerName:        serverName,
			TransportType:     m.serverTransportTypeLocked(serverName),
			ConnectionState:   state,
			Authenticated:     m.authenticated[serverName],
			AuthStatus:        m.authStatus[serverName],
			LastErrorCategory: m.lastErrorCategory[serverName],
			LastError:         m.lastError[serverName],
			LastErrorAt:       m.lastErrorAt[serverName],
			LastSuccessAt:     m.lastSuccessAt[serverName],
			LastAuthAttemptAt: m.lastAuthAttempt[serverName],
		}
	}
	for serverName := range m.transportType {
		if _, ok := statusByServer[serverName]; ok {
			continue
		}
		statusByServer[serverName] = &ServerStatus{
			ServerName:        serverName,
			TransportType:     m.serverTransportTypeLocked(serverName),
			ConnectionState:   m.connectionState[serverName],
			Authenticated:     m.authenticated[serverName],
			AuthStatus:        m.authStatus[serverName],
			LastError:         m.lastError[serverName],
			LastErrorAt:       m.lastErrorAt[serverName],
			LastSuccessAt:     m.lastSuccessAt[serverName],
			LastAuthAttemptAt: m.lastAuthAttempt[serverName],
		}
	}

	for _, item := range m.resourceList {
		if st, ok := statusByServer[item.ServerName]; ok {
			st.CachedResourceCount++
		}
	}

	for key := range m.resourceContents {
		serverName, _, ok := strings.Cut(key, "\n")
		if !ok {
			continue
		}
		if st, exists := statusByServer[serverName]; exists {
			st.CachedContentEntries++
		}
	}

	out := make([]ServerStatus, 0, len(statusByServer))
	for _, st := range statusByServer {
		normalizeServerStatusFields(st)
		out = append(out, *st)
	}

	sortServerStatuses(out)

	return out
}

// SetCacheInvalidationHook sets a callback for cache invalidation events.
func (m *Manager) SetCacheInvalidationHook(hook func(serverName string)) {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.onInvalidate = hook
}

// InvalidateServerCache drops cached resources for one server.
func (m *Manager) InvalidateServerCache(serverName string) {
	var hook func(string)
	m.mu.Lock()
	hook = m.clearServerCacheLocked(serverName)
	m.authenticated[serverName] = false
	m.authStatus[serverName] = AuthStatusUnauthenticated
	if m.connectionState[serverName] != ServerConnectionDisabled {
		m.setConnectionStateLocked(serverName, ServerConnectionPending)
	}
	m.mu.Unlock()
	if hook != nil {
		hook(serverName)
	}
}

// InvalidateAllCaches drops cached resources for all servers.
func (m *Manager) InvalidateAllCaches() {
	m.mu.RLock()
	serverNames := make([]string, 0, len(m.clients))
	for serverName := range m.clients {
		serverNames = append(serverNames, serverName)
	}
	m.mu.RUnlock()
	sort.Strings(serverNames)

	for _, serverName := range serverNames {
		m.InvalidateServerCache(serverName)
	}
}

func resourceCacheKey(serverName, uri string) string {
	return serverName + "\n" + uri
}

func (m *Manager) serverKnownLocked(serverName string) bool {
	if serverName == "" {
		return false
	}
	if _, ok := m.clients[serverName]; ok {
		return true
	}
	if _, ok := m.serverConfigs[serverName]; ok {
		return true
	}
	if _, ok := m.connectionState[serverName]; ok {
		return true
	}
	if _, ok := m.authStatus[serverName]; ok {
		return true
	}
	if _, ok := m.transportType[serverName]; ok {
		return true
	}
	return false
}

func (m *Manager) setServerAuthStateLocked(serverName string, authenticated bool, authStatus AuthStatus, connectionState ServerConnectionState) {
	if m.authenticated == nil {
		m.authenticated = make(map[string]bool)
	}
	if m.authStatus == nil {
		m.authStatus = make(map[string]AuthStatus)
	}
	m.authenticated[serverName] = authenticated
	m.authStatus[serverName] = authStatus
	m.setConnectionStateLocked(serverName, connectionState)
}

func (m *Manager) resourceClientForOperation(serverName, operation string) (managerClient, error) {
	serverName = strings.TrimSpace(serverName)
	m.mu.RLock()
	if m.connectionState[serverName] == ServerConnectionDisabled {
		m.mu.RUnlock()
		return nil, fmt.Errorf("mcp server %q is disabled", serverName)
	}
	transport := m.serverTransportTypeLocked(serverName)
	client, ok := m.clients[serverName]
	m.mu.RUnlock()

	if !ok {
		return nil, fmt.Errorf("unknown MCP server %q", serverName)
	}
	if transport != TransportStdio {
		return nil, fmt.Errorf("mcp manager: %s %s: %w: %s", operation, serverName, ErrTransportUnsupported, transport)
	}

	return client, nil
}

func (m *Manager) replaceServerResourcesLocked(serverName string, resources []Resource) {
	filtered := make([]Resource, 0, len(m.resourceList)+len(resources))
	for _, item := range m.resourceList {
		if item.ServerName != serverName {
			filtered = append(filtered, item)
		}
	}
	filtered = append(filtered, resources...)
	m.resourceList = filtered
}

func (m *Manager) resourceServerNames() []string {
	m.mu.RLock()
	names := make([]string, 0, len(m.clients))
	for serverName := range m.clients {
		if m.connectionState[serverName] == ServerConnectionDisabled {
			continue
		}
		names = append(names, serverName)
	}
	m.mu.RUnlock()
	sort.Strings(names)
	return names
}

func (m *Manager) serverTransportTypeLocked(serverName string) TransportType {
	if t, ok := m.transportType[serverName]; ok {
		return normalizeTransportType(t)
	}
	return TransportStdio
}

func (m *Manager) clearServerCacheLocked(serverName string) func(string) {
	filtered := make([]Resource, 0, len(m.resourceList))
	for _, item := range m.resourceList {
		if item.ServerName != serverName {
			filtered = append(filtered, item)
		}
	}
	m.resourceList = filtered

	prefix := serverName + "\n"
	for key := range m.resourceContents {
		if strings.HasPrefix(key, prefix) {
			delete(m.resourceContents, key)
		}
	}

	return m.onInvalidate
}

func (m *Manager) setConnectionStateLocked(serverName string, state ServerConnectionState) {
	if m.connectionState == nil {
		m.connectionState = make(map[string]ServerConnectionState)
	}
	m.connectionState[serverName] = state
}

func (m *Manager) setServerErrorLocked(serverName string, err error) {
	if m.lastError == nil {
		m.lastError = make(map[string]string)
	}
	if m.lastErrorCategory == nil {
		m.lastErrorCategory = make(map[string]ErrorCategory)
	}
	if m.lastErrorAt == nil {
		m.lastErrorAt = make(map[string]time.Time)
	}
	if err == nil {
		delete(m.lastError, serverName)
		delete(m.lastErrorCategory, serverName)
		delete(m.lastErrorAt, serverName)
		return
	}
	classified := ClassifyError(err)
	m.lastError[serverName] = classified.Message
	m.lastErrorCategory[serverName] = classified.Category
	m.lastErrorAt[serverName] = time.Now().UTC()
}

func (m *Manager) setServerSuccessLocked(serverName string) {
	if m.lastSuccessAt == nil {
		m.lastSuccessAt = make(map[string]time.Time)
	}
	m.lastSuccessAt[serverName] = time.Now().UTC()
	m.setServerErrorLocked(serverName, nil)
}

// IsNeedsAuthError reports whether one error indicates auth is required.
func IsNeedsAuthError(err error) bool {
	if err == nil {
		return false
	}
	msg := strings.ToLower(err.Error())
	return strings.Contains(msg, "unauthorized") ||
		strings.Contains(msg, "needs auth") ||
		strings.Contains(msg, "needs-auth") ||
		strings.Contains(msg, "authentication") ||
		strings.Contains(msg, "401")
}

func isNeedsAuthError(err error) bool {
	return IsNeedsAuthError(err)
}

func inferServerConnectionState(authenticated bool, authStatus AuthStatus) ServerConnectionState {
	switch authStatus {
	case AuthStatusFailed:
		return ServerConnectionFailed
	case AuthStatusAuthenticating:
		return ServerConnectionPending
	case AuthStatusUnauthenticated:
		if authenticated {
			return ServerConnectionConnected
		}
		return ServerConnectionNeedsAuth
	case AuthStatusAuthenticated:
		if authenticated {
			return ServerConnectionConnected
		}
		return ServerConnectionPending
	default:
		if authenticated {
			return ServerConnectionConnected
		}
		return ServerConnectionPending
	}
}

func normalizeServerStatusFields(status *ServerStatus) {
	if status == nil {
		return
	}
	if status.AuthStatus == AuthStatusAuthenticated && !status.Authenticated {
		status.AuthStatus = AuthStatusUnauthenticated
	}
	if status.AuthStatus == AuthStatusUnauthenticated && status.Authenticated {
		status.AuthStatus = AuthStatusAuthenticated
	}
	if status.AuthStatus == "" || status.AuthStatus == AuthStatusUnknown {
		if status.Authenticated {
			status.AuthStatus = AuthStatusAuthenticated
		} else {
			status.AuthStatus = AuthStatusUnauthenticated
		}
	}
	if status.ConnectionState == "" {
		status.ConnectionState = inferServerConnectionState(status.Authenticated, status.AuthStatus)
		return
	}
	if status.ConnectionState == ServerConnectionDisabled {
		return
	}
	if status.AuthStatus == AuthStatusFailed {
		status.ConnectionState = ServerConnectionFailed
		return
	}
	if status.AuthStatus == AuthStatusAuthenticating {
		status.ConnectionState = ServerConnectionPending
		return
	}
	if status.AuthStatus == AuthStatusUnauthenticated {
		if status.ConnectionState != ServerConnectionPending {
			status.ConnectionState = ServerConnectionNeedsAuth
		}
		return
	}
	if status.AuthStatus == AuthStatusAuthenticated && status.ConnectionState == ServerConnectionNeedsAuth {
		status.ConnectionState = ServerConnectionConnected
	}
}

func (m *Manager) Close() error {
	var firstErr error
	for _, c := range m.clients {
		if err := c.Close(); err != nil && firstErr == nil {
			firstErr = err
		}
	}
	return firstErr
}
