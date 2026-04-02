package mcp

import (
	"context"
	"encoding/json"
	"errors"
	"testing"
	"time"

	"github.com/alliecatowo/alliecode/internal/types"
)

func TestManagerToolsCopy(t *testing.T) {
	m := &Manager{
		tools: []RemoteTool{{ServerName: "s", Def: types.ToolDef{Name: "x"}}},
	}
	got := m.Tools()
	if len(got) != 1 || got[0].Def.Name != "x" {
		t.Fatalf("unexpected tools: %+v", got)
	}
	got[0].Def.Name = "mutated"
	if m.tools[0].Def.Name != "x" {
		t.Fatal("Tools should return a copy")
	}
}

type fakeClient struct {
	resources []Resource
	content   ResourceContent
	callOut   string
	callErr   error
	listErr   error
	readErr   error
	authErr   error
	readCalls int
	authStart func()
	authWait  <-chan struct{}
}

func (f *fakeClient) Initialize(ctx context.Context) error {
	return nil
}

func (f *fakeClient) ListTools(ctx context.Context) ([]types.ToolDef, error) {
	return nil, nil
}

func (f *fakeClient) CallTool(ctx context.Context, name string, input json.RawMessage) (string, error) {
	if f.callErr != nil {
		return "", f.callErr
	}
	if f.callOut != "" {
		return f.callOut, nil
	}
	return "ok", nil
}

func (f *fakeClient) ListResources(ctx context.Context) ([]Resource, error) {
	if f.listErr != nil {
		return nil, f.listErr
	}
	out := make([]Resource, len(f.resources))
	copy(out, f.resources)
	return out, nil
}

func (f *fakeClient) ReadResource(ctx context.Context, uri string) (ResourceContent, error) {
	f.readCalls++
	if f.readErr != nil {
		return ResourceContent{}, f.readErr
	}
	if f.content.URI != uri {
		return ResourceContent{}, errors.New("not found")
	}
	return f.content, nil
}

func (f *fakeClient) AuthenticateLocal(ctx context.Context, token string) error {
	if f.authStart != nil {
		f.authStart()
	}
	if f.authWait != nil {
		select {
		case <-ctx.Done():
			return ctx.Err()
		case <-f.authWait:
		}
	}
	return f.authErr
}

func (f *fakeClient) Close() error {
	return nil
}

func TestManagerResourcesAndLocalAuthFlows(t *testing.T) {
	m := &Manager{clients: map[string]managerClient{
		"alpha": &fakeClient{resources: []Resource{{URI: "mem://one", Name: "one"}}},
		"beta":  &fakeClient{resources: []Resource{{ServerName: "beta", URI: "mem://two", Name: "two"}}, content: ResourceContent{URI: "mem://two", MIMEType: "text/plain", Text: "hello"}},
	}, resourceContents: map[string]ResourceContent{}, authenticated: map[string]bool{}, authStatus: map[string]AuthStatus{"alpha": AuthStatusUnauthenticated, "beta": AuthStatusUnauthenticated}}

	resources, err := m.ListResources(context.Background())
	if err != nil {
		t.Fatalf("ListResources error: %v", err)
	}
	if len(resources) != 2 {
		t.Fatalf("expected two resources, got %+v", resources)
	}
	if resources[0].ServerName != "alpha" || resources[0].URI != "mem://one" {
		t.Fatalf("unexpected first resource ordering: %+v", resources[0])
	}
	if resources[1].ServerName != "beta" || resources[1].URI != "mem://two" {
		t.Fatalf("unexpected second resource ordering: %+v", resources[1])
	}
	if resources[0].ServerName == "" || resources[1].ServerName == "" {
		t.Fatalf("expected server names to be populated: %+v", resources)
	}

	cached := m.CachedResources()
	if len(cached) != 2 {
		t.Fatalf("expected cached resources, got %+v", cached)
	}
	cached[0].Name = "mutated"
	if m.CachedResources()[0].Name == "mutated" {
		t.Fatal("CachedResources should return a copy")
	}

	content, err := m.ReadResource(context.Background(), "beta", "mem://two")
	if err != nil {
		t.Fatalf("ReadResource error: %v", err)
	}
	if content.Text != "hello" {
		t.Fatalf("unexpected content: %+v", content)
	}

	if err := m.AuthenticateLocal(context.Background(), "beta", "token-123"); err != nil {
		t.Fatalf("AuthenticateLocal error: %v", err)
	}
	if !m.Authenticated("beta") {
		t.Fatal("expected beta auth status to be true")
	}
	if got := m.AuthStatus("beta"); got != AuthStatusAuthenticated {
		t.Fatalf("AuthStatus(beta) = %q, want %q", got, AuthStatusAuthenticated)
	}

	statuses := m.ServerStatuses()
	if len(statuses) != 2 {
		t.Fatalf("expected two server statuses, got %+v", statuses)
	}
	if statuses[0].ServerName != "alpha" || statuses[0].AuthStatus != AuthStatusUnauthenticated || statuses[0].CachedResourceCount != 1 || statuses[0].CachedContentEntries != 0 {
		t.Fatalf("unexpected alpha status: %+v", statuses[0])
	}
	if statuses[1].ServerName != "beta" || statuses[1].AuthStatus != AuthStatusAuthenticated || statuses[1].CachedResourceCount != 0 || statuses[1].CachedContentEntries != 0 {
		t.Fatalf("unexpected beta status after auth cache invalidation: %+v", statuses[1])
	}
}

func TestManagerUnknownServerAndAuthErrors(t *testing.T) {
	m := &Manager{clients: map[string]managerClient{
		"alpha": &fakeClient{authErr: errors.New("denied")},
	}, resourceContents: map[string]ResourceContent{}, authenticated: map[string]bool{}, authStatus: map[string]AuthStatus{"alpha": AuthStatusUnauthenticated}}

	if _, err := m.ReadResource(context.Background(), "missing", "mem://x"); err == nil {
		t.Fatal("expected unknown server error for ReadResource")
	}

	if err := m.AuthenticateLocal(context.Background(), "alpha", "bad"); err == nil {
		t.Fatal("expected auth error")
	}
	if m.Authenticated("alpha") {
		t.Fatal("expected alpha auth status to remain false")
	}
	if got := m.AuthStatus("alpha"); got != AuthStatusFailed {
		t.Fatalf("AuthStatus(alpha) = %q, want %q", got, AuthStatusFailed)
	}
}

func TestManagerReadResourceCachedUsesCache(t *testing.T) {
	beta := &fakeClient{content: ResourceContent{URI: "mem://two", MIMEType: "text/plain", Text: "hello"}}
	m := &Manager{clients: map[string]managerClient{
		"beta": beta,
	}, resourceContents: map[string]ResourceContent{}, authenticated: map[string]bool{}, authStatus: map[string]AuthStatus{"beta": AuthStatusUnauthenticated}}

	first, err := m.ReadResourceCached(context.Background(), "beta", "mem://two")
	if err != nil {
		t.Fatalf("first ReadResourceCached error: %v", err)
	}
	second, err := m.ReadResourceCached(context.Background(), "beta", "mem://two")
	if err != nil {
		t.Fatalf("second ReadResourceCached error: %v", err)
	}
	if first.Text != "hello" || second.Text != "hello" {
		t.Fatalf("unexpected cached content: first=%+v second=%+v", first, second)
	}
	if beta.readCalls != 1 {
		t.Fatalf("expected one backend read, got %d", beta.readCalls)
	}
}

func TestManagerResourceFailures(t *testing.T) {
	m := &Manager{clients: map[string]managerClient{
		"alpha": &fakeClient{listErr: errors.New("list failed")},
		"beta":  &fakeClient{readErr: errors.New("read failed")},
	}, resourceContents: map[string]ResourceContent{}, authenticated: map[string]bool{}, authStatus: map[string]AuthStatus{"alpha": AuthStatusUnauthenticated, "beta": AuthStatusUnauthenticated}}

	if _, err := m.ListResources(context.Background()); err == nil {
		t.Fatal("expected list resources error")
	}

	if _, err := m.ReadResource(context.Background(), "beta", "mem://two"); err == nil {
		t.Fatal("expected read resource error")
	}
}

func TestManagerAuthStatusAuthenticatingTransition(t *testing.T) {
	started := make(chan struct{}, 1)
	continueAuth := make(chan struct{})
	beta := &fakeClient{authStart: func() { started <- struct{}{} }, authWait: continueAuth}
	m := &Manager{clients: map[string]managerClient{"beta": beta}, resourceContents: map[string]ResourceContent{}, authenticated: map[string]bool{}, authStatus: map[string]AuthStatus{"beta": AuthStatusUnauthenticated}}
	errCh := make(chan error, 1)

	go func() {
		errCh <- m.AuthenticateLocal(context.Background(), "beta", "tok")
	}()

	<-started
	if got := m.AuthStatus("beta"); got != AuthStatusAuthenticating {
		t.Fatalf("AuthStatus(beta) during auth = %q, want %q", got, AuthStatusAuthenticating)
	}
	close(continueAuth)
	if err := <-errCh; err != nil {
		t.Fatalf("AuthenticateLocal() error = %v", err)
	}
}

func TestManagerCacheInvalidationHooks(t *testing.T) {
	beta := &fakeClient{resources: []Resource{{ServerName: "beta", URI: "mem://two", Name: "two"}}, content: ResourceContent{URI: "mem://two", MIMEType: "text/plain", Text: "hello"}}
	m := &Manager{clients: map[string]managerClient{"beta": beta}, resourceContents: map[string]ResourceContent{}, authenticated: map[string]bool{}, authStatus: map[string]AuthStatus{"beta": AuthStatusUnauthenticated}}

	if _, err := m.ListResources(context.Background()); err != nil {
		t.Fatalf("ListResources() error = %v", err)
	}
	if _, err := m.ReadResourceCached(context.Background(), "beta", "mem://two"); err != nil {
		t.Fatalf("ReadResourceCached() first read error = %v", err)
	}
	if beta.readCalls != 1 {
		t.Fatalf("expected one read before invalidation, got %d", beta.readCalls)
	}

	invalidated := make([]string, 0, 4)
	m.SetCacheInvalidationHook(func(serverName string) {
		invalidated = append(invalidated, serverName)
	})

	if err := m.AuthenticateLocal(context.Background(), "beta", "tok"); err != nil {
		t.Fatalf("AuthenticateLocal() error = %v", err)
	}
	if len(invalidated) != 1 || invalidated[0] != "beta" {
		t.Fatalf("unexpected invalidation hook calls after auth success: %+v", invalidated)
	}

	if len(m.CachedResources()) != 0 {
		t.Fatalf("expected cached resource list to be invalidated")
	}
	if _, err := m.ReadResourceCached(context.Background(), "beta", "mem://two"); err != nil {
		t.Fatalf("ReadResourceCached() after invalidation error = %v", err)
	}
	if beta.readCalls != 2 {
		t.Fatalf("expected cache miss after invalidation, readCalls=%d", beta.readCalls)
	}

	m.InvalidateServerCache("beta")
	if len(invalidated) != 2 || invalidated[1] != "beta" {
		t.Fatalf("unexpected invalidation hook calls after manual invalidate: %+v", invalidated)
	}
	if m.Authenticated("beta") {
		t.Fatalf("expected authenticated false after manual invalidation")
	}
	if got := m.AuthStatus("beta"); got != AuthStatusUnauthenticated {
		t.Fatalf("AuthStatus(beta) after manual invalidation = %q, want %q", got, AuthStatusUnauthenticated)
	}
}

func TestManagerAuthenticateLocalCanceledDoesNotInvalidateCache(t *testing.T) {
	beta := &fakeClient{
		resources: []Resource{{ServerName: "beta", URI: "mem://two", Name: "two"}},
		content:   ResourceContent{URI: "mem://two", MIMEType: "text/plain", Text: "hello"},
	}
	m := &Manager{clients: map[string]managerClient{"beta": beta}, resourceContents: map[string]ResourceContent{}, authenticated: map[string]bool{}, authStatus: map[string]AuthStatus{"beta": AuthStatusUnauthenticated}}

	if _, err := m.ListResources(context.Background()); err != nil {
		t.Fatalf("ListResources() error = %v", err)
	}
	if _, err := m.ReadResourceCached(context.Background(), "beta", "mem://two"); err != nil {
		t.Fatalf("ReadResourceCached() warm cache error = %v", err)
	}

	invalidated := 0
	m.SetCacheInvalidationHook(func(serverName string) {
		if serverName != "beta" {
			t.Fatalf("unexpected invalidation server %q", serverName)
		}
		invalidated++
	})

	ctx, cancel := context.WithCancel(context.Background())
	cancel()
	err := m.AuthenticateLocal(ctx, "beta", "tok")
	if !errors.Is(err, context.Canceled) {
		t.Fatalf("AuthenticateLocal() error = %v, want context canceled", err)
	}
	if got := m.AuthStatus("beta"); got != AuthStatusFailed {
		t.Fatalf("AuthStatus(beta) = %q, want %q", got, AuthStatusFailed)
	}
	if invalidated != 0 {
		t.Fatalf("expected no invalidation hook calls on canceled auth, got %d", invalidated)
	}

	if len(m.CachedResources()) != 1 {
		t.Fatalf("expected cached resource list preserved on canceled auth")
	}
	if _, err := m.ReadResourceCached(context.Background(), "beta", "mem://two"); err != nil {
		t.Fatalf("ReadResourceCached() after canceled auth error = %v", err)
	}
	if beta.readCalls != 1 {
		t.Fatalf("expected cache hit after canceled auth, readCalls=%d", beta.readCalls)
	}
}

func TestManagerAuthenticateLocalFailureInvalidatesCache(t *testing.T) {
	beta := &fakeClient{
		resources: []Resource{{ServerName: "beta", URI: "mem://two", Name: "two"}},
		content:   ResourceContent{URI: "mem://two", MIMEType: "text/plain", Text: "hello"},
		authErr:   errors.New("denied"),
	}
	m := &Manager{clients: map[string]managerClient{"beta": beta}, resourceContents: map[string]ResourceContent{}, authenticated: map[string]bool{}, authStatus: map[string]AuthStatus{"beta": AuthStatusUnauthenticated}}

	if _, err := m.ListResources(context.Background()); err != nil {
		t.Fatalf("ListResources() error = %v", err)
	}
	if _, err := m.ReadResourceCached(context.Background(), "beta", "mem://two"); err != nil {
		t.Fatalf("ReadResourceCached() warm cache error = %v", err)
	}

	invalidated := 0
	m.SetCacheInvalidationHook(func(serverName string) {
		if serverName != "beta" {
			t.Fatalf("unexpected invalidation server %q", serverName)
		}
		invalidated++
	})

	err := m.AuthenticateLocal(context.Background(), "beta", "bad")
	if err == nil {
		t.Fatalf("expected auth error")
	}
	if got := m.AuthStatus("beta"); got != AuthStatusFailed {
		t.Fatalf("AuthStatus(beta) = %q, want %q", got, AuthStatusFailed)
	}
	if invalidated != 1 {
		t.Fatalf("expected one invalidation hook call on auth failure, got %d", invalidated)
	}
	if len(m.CachedResources()) != 0 {
		t.Fatalf("expected cached resource list invalidated on auth failure")
	}
}

func TestManagerAuthenticateLocalContextTimeoutSetsFailed(t *testing.T) {
	continueAuth := make(chan struct{})
	beta := &fakeClient{authWait: continueAuth}
	m := &Manager{clients: map[string]managerClient{"beta": beta}, resourceContents: map[string]ResourceContent{}, authenticated: map[string]bool{}, authStatus: map[string]AuthStatus{"beta": AuthStatusUnauthenticated}}

	ctx, cancel := context.WithTimeout(context.Background(), 20*time.Millisecond)
	defer cancel()
	err := m.AuthenticateLocal(ctx, "beta", "tok")
	if !errors.Is(err, context.DeadlineExceeded) {
		t.Fatalf("AuthenticateLocal() error = %v, want deadline exceeded", err)
	}
	if got := m.AuthStatus("beta"); got != AuthStatusFailed {
		t.Fatalf("AuthStatus(beta) = %q, want %q", got, AuthStatusFailed)
	}
}

func TestManagerServerStatusesDeterministicOrderingAndCounts(t *testing.T) {
	m := &Manager{clients: map[string]managerClient{
		"gamma": &fakeClient{},
		"alpha": &fakeClient{},
	}, resourceContents: map[string]ResourceContent{}, authenticated: map[string]bool{"gamma": true}, authStatus: map[string]AuthStatus{"gamma": AuthStatusAuthenticated}}
	m.resourceList = []Resource{
		{ServerName: "gamma", URI: "mem://g1"},
		{ServerName: "alpha", URI: "mem://a1"},
		{ServerName: "gamma", URI: "mem://g2"},
	}
	m.resourceContents[resourceCacheKey("gamma", "mem://g1")] = ResourceContent{URI: "mem://g1"}
	m.resourceContents[resourceCacheKey("gamma", "mem://g2")] = ResourceContent{URI: "mem://g2"}
	m.resourceContents[resourceCacheKey("alpha", "mem://a1")] = ResourceContent{URI: "mem://a1"}

	statuses := m.ServerStatuses()
	if len(statuses) != 2 {
		t.Fatalf("expected two statuses, got %+v", statuses)
	}
	if statuses[0].ServerName != "alpha" || statuses[0].AuthStatus != AuthStatusUnauthenticated || statuses[0].CachedResourceCount != 1 || statuses[0].CachedContentEntries != 1 {
		t.Fatalf("unexpected alpha status: %+v", statuses[0])
	}
	if statuses[1].ServerName != "gamma" || statuses[1].AuthStatus != AuthStatusAuthenticated || statuses[1].CachedResourceCount != 2 || statuses[1].CachedContentEntries != 2 {
		t.Fatalf("unexpected gamma status: %+v", statuses[1])
	}
}

func TestManagerServerStatusesRepairsAuthStatusMismatch(t *testing.T) {
	m := &Manager{clients: map[string]managerClient{
		"alpha": &fakeClient{},
		"beta":  &fakeClient{},
	}, resourceContents: map[string]ResourceContent{}, authenticated: map[string]bool{"alpha": true, "beta": false}, authStatus: map[string]AuthStatus{"alpha": AuthStatusUnauthenticated, "beta": AuthStatusAuthenticated}}

	statuses := m.ServerStatuses()
	if len(statuses) != 2 {
		t.Fatalf("expected two statuses, got %+v", statuses)
	}
	if statuses[0].ServerName != "alpha" || statuses[0].AuthStatus != AuthStatusAuthenticated || !statuses[0].Authenticated {
		t.Fatalf("unexpected alpha status repair: %+v", statuses[0])
	}
	if statuses[1].ServerName != "beta" || statuses[1].AuthStatus != AuthStatusUnauthenticated || statuses[1].Authenticated {
		t.Fatalf("unexpected beta status repair: %+v", statuses[1])
	}
}

func TestManagerInvalidateAllCachesDeterministicHookOrder(t *testing.T) {
	m := &Manager{clients: map[string]managerClient{
		"beta":  &fakeClient{},
		"alpha": &fakeClient{},
		"zeta":  &fakeClient{},
	}, resourceContents: map[string]ResourceContent{}, authenticated: map[string]bool{}, authStatus: map[string]AuthStatus{}}

	called := make([]string, 0, 3)
	m.SetCacheInvalidationHook(func(serverName string) {
		called = append(called, serverName)
	})

	m.InvalidateAllCaches()

	if len(called) != 3 {
		t.Fatalf("expected three hook calls, got %+v", called)
	}
	if called[0] != "alpha" || called[1] != "beta" || called[2] != "zeta" {
		t.Fatalf("unexpected hook ordering: %+v", called)
	}
}

func TestManagerServerStatusesConnectionStateTransitions(t *testing.T) {
	beta := &fakeClient{}
	m := &Manager{
		clients:          map[string]managerClient{"beta": beta},
		resourceContents: map[string]ResourceContent{},
		authenticated:    map[string]bool{"beta": false},
		authStatus:       map[string]AuthStatus{"beta": AuthStatusUnauthenticated},
		connectionState:  map[string]ServerConnectionState{"beta": ServerConnectionPending},
		transportType:    map[string]TransportType{"beta": TransportStdio},
	}

	statuses := m.ServerStatuses()
	if len(statuses) != 1 || statuses[0].ConnectionState != ServerConnectionPending || statuses[0].TransportType != TransportStdio {
		t.Fatalf("pending state mismatch: %+v", statuses)
	}

	beta.authErr = errors.New("401 unauthorized")
	if err := m.AuthenticateLocal(context.Background(), "beta", "tok"); err == nil {
		t.Fatalf("expected auth error")
	}
	statuses = m.ServerStatuses()
	if statuses[0].ConnectionState != ServerConnectionNeedsAuth {
		t.Fatalf("state after needs-auth error = %q, want %q", statuses[0].ConnectionState, ServerConnectionNeedsAuth)
	}

	beta.authErr = nil
	if err := m.AuthenticateLocal(context.Background(), "beta", "tok"); err != nil {
		t.Fatalf("AuthenticateLocal() success error = %v", err)
	}
	statuses = m.ServerStatuses()
	if statuses[0].ConnectionState != ServerConnectionConnected {
		t.Fatalf("state after auth success = %q, want %q", statuses[0].ConnectionState, ServerConnectionConnected)
	}

	beta.readErr = errors.New("backend down")
	if _, err := m.ReadResource(context.Background(), "beta", "mem://x"); err == nil {
		t.Fatalf("expected read failure")
	}
	statuses = m.ServerStatuses()
	if statuses[0].ConnectionState != ServerConnectionFailed {
		t.Fatalf("state after read failure = %q, want %q", statuses[0].ConnectionState, ServerConnectionFailed)
	}

	m.InvalidateServerCache("beta")
	statuses = m.ServerStatuses()
	if statuses[0].ConnectionState != ServerConnectionPending {
		t.Fatalf("state after invalidate = %q, want %q", statuses[0].ConnectionState, ServerConnectionPending)
	}
}

func TestManagerTransportSelectionForDisabledServers(t *testing.T) {
	m, err := NewManager(context.Background(), []ServerConfig{
		{Name: "default", Disabled: true},
		{Name: "sse", Transport: TransportSSE, Disabled: true},
		{Name: "ws", Transport: "ws", Disabled: true},
	})
	if err != nil {
		t.Fatalf("NewManager() error = %v", err)
	}

	if got := m.ServerTransportType("default"); got != TransportStdio {
		t.Fatalf("ServerTransportType(default) = %q, want %q", got, TransportStdio)
	}
	if got := m.ServerTransportType("sse"); got != TransportSSE {
		t.Fatalf("ServerTransportType(sse) = %q, want %q", got, TransportSSE)
	}
	if got := m.ServerTransportType("ws"); got != TransportWebSocket {
		t.Fatalf("ServerTransportType(ws) = %q, want %q", got, TransportWebSocket)
	}

	statuses := m.ServerStatuses()
	if len(statuses) != 3 {
		t.Fatalf("expected three statuses, got %+v", statuses)
	}
	if statuses[0].ServerName != "default" || statuses[0].TransportType != TransportStdio || statuses[0].ConnectionState != ServerConnectionDisabled {
		t.Fatalf("unexpected default status: %+v", statuses[0])
	}
	if statuses[1].ServerName != "sse" || statuses[1].TransportType != TransportSSE || statuses[1].ConnectionState != ServerConnectionDisabled {
		t.Fatalf("unexpected sse status: %+v", statuses[1])
	}
	if statuses[2].ServerName != "ws" || statuses[2].TransportType != TransportWebSocket || statuses[2].ConnectionState != ServerConnectionDisabled {
		t.Fatalf("unexpected ws status: %+v", statuses[2])
	}
}

func TestManagerServerStatusAPIIncludesTransport(t *testing.T) {
	m := &Manager{
		clients:          map[string]managerClient{"beta": &fakeClient{}},
		resourceContents: map[string]ResourceContent{},
		authenticated:    map[string]bool{"beta": true},
		authStatus:       map[string]AuthStatus{"beta": AuthStatusAuthenticated},
		connectionState:  map[string]ServerConnectionState{"beta": ServerConnectionConnected},
		transportType:    map[string]TransportType{"beta": TransportSSE},
	}

	status, ok := m.ServerStatus("beta")
	if !ok {
		t.Fatalf("expected ServerStatus(beta) to exist")
	}
	if status.TransportType != TransportSSE || status.ConnectionState != ServerConnectionConnected {
		t.Fatalf("unexpected status: %+v", status)
	}

	if _, ok := m.ServerStatus("missing"); ok {
		t.Fatalf("expected missing server status to report false")
	}
}

func TestManagerAuthStateTransitionHelpers(t *testing.T) {
	m := &Manager{
		clients:          map[string]managerClient{"beta": &fakeClient{}},
		resourceContents: map[string]ResourceContent{},
		authenticated:    map[string]bool{"beta": false},
		authStatus:       map[string]AuthStatus{"beta": AuthStatusUnauthenticated},
		connectionState:  map[string]ServerConnectionState{"beta": ServerConnectionPending},
		transportType:    map[string]TransportType{"beta": TransportStdio},
	}

	if err := m.SetAuthStateNeedsAuth("beta"); err != nil {
		t.Fatalf("SetAuthStateNeedsAuth() error = %v", err)
	}
	if st, _ := m.ServerStatus("beta"); st.AuthStatus != AuthStatusUnauthenticated || st.ConnectionState != ServerConnectionNeedsAuth {
		t.Fatalf("unexpected needs-auth state: %+v", st)
	}

	if err := m.SetAuthStateAuthenticating("beta"); err != nil {
		t.Fatalf("SetAuthStateAuthenticating() error = %v", err)
	}
	if st, _ := m.ServerStatus("beta"); st.AuthStatus != AuthStatusAuthenticating || st.ConnectionState != ServerConnectionPending {
		t.Fatalf("unexpected authenticating state: %+v", st)
	}

	if err := m.SetAuthStateAuthenticated("beta"); err != nil {
		t.Fatalf("SetAuthStateAuthenticated() error = %v", err)
	}
	if st, _ := m.ServerStatus("beta"); st.AuthStatus != AuthStatusAuthenticated || st.ConnectionState != ServerConnectionConnected || !st.Authenticated {
		t.Fatalf("unexpected authenticated state: %+v", st)
	}

	if err := m.SetAuthStateFailed("beta"); err != nil {
		t.Fatalf("SetAuthStateFailed() error = %v", err)
	}
	if st, _ := m.ServerStatus("beta"); st.AuthStatus != AuthStatusFailed || st.ConnectionState != ServerConnectionFailed || st.Authenticated {
		t.Fatalf("unexpected failed state: %+v", st)
	}

	if err := m.SetAuthStateAuthenticated("missing"); err == nil {
		t.Fatalf("expected unknown server error")
	}
}

func TestManagerListResourcesForServerAndUnsupportedTransport(t *testing.T) {
	m := &Manager{
		clients: map[string]managerClient{
			"stdio": &fakeClient{resources: []Resource{{URI: "mem://a", Name: "a"}}},
			"sse":   &fakeClient{resources: []Resource{{URI: "mem://b", Name: "b"}}},
		},
		resourceContents: map[string]ResourceContent{},
		authenticated:    map[string]bool{},
		authStatus:       map[string]AuthStatus{"stdio": AuthStatusUnauthenticated, "sse": AuthStatusUnauthenticated},
		connectionState:  map[string]ServerConnectionState{"stdio": ServerConnectionPending, "sse": ServerConnectionPending},
		transportType:    map[string]TransportType{"stdio": TransportStdio, "sse": TransportSSE},
	}

	listed, err := m.ListResourcesForServer(context.Background(), "stdio")
	if err != nil {
		t.Fatalf("ListResourcesForServer(stdio) error = %v", err)
	}
	if len(listed) != 1 || listed[0].ServerName != "stdio" {
		t.Fatalf("unexpected stdio resources: %+v", listed)
	}

	if _, err := m.ListResourcesForServer(context.Background(), "sse"); !errors.Is(err, ErrTransportUnsupported) {
		t.Fatalf("expected unsupported transport error, got %v", err)
	}
}

func TestManagerServerStatusesIncludesDisabledWithoutClient(t *testing.T) {
	m := &Manager{
		clients:          map[string]managerClient{},
		resourceContents: map[string]ResourceContent{},
		authenticated:    map[string]bool{"disabled-one": false},
		authStatus:       map[string]AuthStatus{"disabled-one": AuthStatusUnauthenticated},
		connectionState:  map[string]ServerConnectionState{"disabled-one": ServerConnectionDisabled},
	}

	statuses := m.ServerStatuses()
	if len(statuses) != 1 {
		t.Fatalf("expected one status, got %+v", statuses)
	}
	if statuses[0].ServerName != "disabled-one" || statuses[0].ConnectionState != ServerConnectionDisabled {
		t.Fatalf("unexpected disabled status: %+v", statuses[0])
	}
}

func TestManagerAddRemoveServerConfigAndDeterministicConfigListing(t *testing.T) {
	m := &Manager{
		clients:          map[string]managerClient{"beta": &fakeClient{}},
		serverConfigs:    map[string]ServerConfig{"beta": {Name: "beta", Transport: TransportStdio, Command: "npx", Args: []string{"beta-mcp"}}},
		tools:            []RemoteTool{{ServerName: "beta", Def: types.ToolDef{Name: "betaTool"}}},
		resourceList:     []Resource{{ServerName: "beta", URI: "mem://beta"}},
		resourceContents: map[string]ResourceContent{resourceCacheKey("beta", "mem://beta"): {URI: "mem://beta", Text: "hello"}},
		authenticated:    map[string]bool{"beta": true},
		authStatus:       map[string]AuthStatus{"beta": AuthStatusAuthenticated},
		connectionState:  map[string]ServerConnectionState{"beta": ServerConnectionConnected},
		transportType:    map[string]TransportType{"beta": TransportStdio},
	}

	if err := m.AddServerConfig(ServerConfig{Name: "alpha", Transport: "ws", URL: "ws://localhost:4242"}); err != nil {
		t.Fatalf("AddServerConfig(alpha) error = %v", err)
	}
	if err := m.AddServerConfig(ServerConfig{Name: "gamma", Disabled: true}); err != nil {
		t.Fatalf("AddServerConfig(gamma) error = %v", err)
	}

	if err := m.AddServerConfig(ServerConfig{Name: "alpha", Command: "dup"}); err == nil {
		t.Fatalf("expected duplicate add to fail")
	}

	configs := m.ServerConfigs()
	if len(configs) != 3 {
		t.Fatalf("expected three configs, got %+v", configs)
	}
	if configs[0].Name != "alpha" || configs[0].Transport != TransportWebSocket || configs[0].URL != "ws://localhost:4242" {
		t.Fatalf("unexpected alpha config: %+v", configs[0])
	}
	if configs[1].Name != "beta" || configs[1].Transport != TransportStdio {
		t.Fatalf("unexpected beta config: %+v", configs[1])
	}
	if configs[2].Name != "gamma" || configs[2].Transport != TransportStdio || !configs[2].Disabled {
		t.Fatalf("unexpected gamma config: %+v", configs[2])
	}

	if !m.RemoveServerConfig("beta") {
		t.Fatalf("expected remove beta to succeed")
	}
	if m.RemoveServerConfig("missing") {
		t.Fatalf("expected remove missing to return false")
	}

	statuses := m.ServerStatuses()
	if len(statuses) != 2 {
		t.Fatalf("expected two statuses after remove, got %+v", statuses)
	}
	if statuses[0].ServerName != "alpha" || statuses[0].ConnectionState != ServerConnectionPending {
		t.Fatalf("unexpected alpha status after remove: %+v", statuses[0])
	}
	if statuses[1].ServerName != "gamma" || statuses[1].ConnectionState != ServerConnectionDisabled {
		t.Fatalf("unexpected gamma status after remove: %+v", statuses[1])
	}
}

func TestManagerServerStatusIncludesErrorAndSuccessDiagnostics(t *testing.T) {
	beta := &fakeClient{callErr: errors.New("backend exploded")}
	m := &Manager{
		clients:          map[string]managerClient{"beta": beta},
		resourceContents: map[string]ResourceContent{},
		authenticated:    map[string]bool{"beta": true},
		authStatus:       map[string]AuthStatus{"beta": AuthStatusAuthenticated},
		connectionState:  map[string]ServerConnectionState{"beta": ServerConnectionConnected},
		transportType:    map[string]TransportType{"beta": TransportStdio},
	}

	if _, err := m.Call(context.Background(), "beta", "sum", json.RawMessage(`{"a":1}`)); err == nil {
		t.Fatalf("expected call failure")
	}
	status, ok := m.ServerStatus("beta")
	if !ok {
		t.Fatalf("expected status for beta")
	}
	if status.ConnectionState != ServerConnectionFailed {
		t.Fatalf("ConnectionState = %q, want %q", status.ConnectionState, ServerConnectionFailed)
	}
	if status.LastError != "backend exploded" || status.LastErrorAt.IsZero() {
		t.Fatalf("unexpected error diagnostics: %+v", status)
	}

	beta.callErr = nil
	beta.callOut = "done"
	out, err := m.Call(context.Background(), "beta", "sum", json.RawMessage(`{"a":1}`))
	if err != nil || out != "done" {
		t.Fatalf("expected call success, out=%q err=%v", out, err)
	}
	status, _ = m.ServerStatus("beta")
	if status.ConnectionState != ServerConnectionConnected {
		t.Fatalf("ConnectionState = %q, want %q", status.ConnectionState, ServerConnectionConnected)
	}
	if status.LastError != "" || status.LastSuccessAt.IsZero() {
		t.Fatalf("unexpected success diagnostics: %+v", status)
	}
}

func TestManagerAuthAttemptTimestampRecorded(t *testing.T) {
	beta := &fakeClient{}
	m := &Manager{
		clients:          map[string]managerClient{"beta": beta},
		resourceContents: map[string]ResourceContent{},
		authenticated:    map[string]bool{"beta": false},
		authStatus:       map[string]AuthStatus{"beta": AuthStatusUnauthenticated},
		connectionState:  map[string]ServerConnectionState{"beta": ServerConnectionPending},
		transportType:    map[string]TransportType{"beta": TransportStdio},
	}

	if err := m.AuthenticateLocal(context.Background(), "beta", "tok"); err != nil {
		t.Fatalf("AuthenticateLocal() error = %v", err)
	}
	status, ok := m.ServerStatus("beta")
	if !ok {
		t.Fatalf("expected status for beta")
	}
	if status.LastAuthAttemptAt.IsZero() {
		t.Fatalf("expected LastAuthAttemptAt to be populated")
	}
}
