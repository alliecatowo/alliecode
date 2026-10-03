package commands

import (
	"context"
	"testing"
)

func TestAliasSigninDispatchesLogin(t *testing.T) {
	r := DefaultRegistry()
	res, err := r.Dispatch(context.Background(), Context{State: &RuntimeState{}}, "/signin status")
	if err != nil {
		t.Fatalf("signin alias failed: %v", err)
	}
	if res.Message == "" || res.Message[:12] != "LOGIN_STATUS" {
		t.Fatalf("unexpected signin output: %q", res.Message)
	}
}

func TestAliasSignoutDispatchesLogout(t *testing.T) {
	r := DefaultRegistry()
	res, err := r.Dispatch(context.Background(), Context{State: &RuntimeState{}}, "/signout")
	if err != nil {
		t.Fatalf("signout alias failed: %v", err)
	}
	if res.Message == "" || res.Message[:13] != "LOGOUT_RESULT" {
		t.Fatalf("unexpected signout output: %q", res.Message)
	}
}

func TestAliasThemesDispatchesTheme(t *testing.T) {
	r := DefaultRegistry()
	res, err := r.Dispatch(context.Background(), Context{State: &RuntimeState{}}, "/themes status")
	if err != nil {
		t.Fatalf("themes alias failed: %v", err)
	}
	if res.Message == "" || res.Message[:12] != "THEME_STATUS" {
		t.Fatalf("unexpected themes output: %q", res.Message)
	}
}

func TestAliasEditorDispatchesIDE(t *testing.T) {
	r := DefaultRegistry()
	res, err := r.Dispatch(context.Background(), Context{State: &RuntimeState{}}, "/editor status")
	if err != nil {
		t.Fatalf("editor alias failed: %v", err)
	}
	if res.Message == "" || res.Message[:10] != "IDE_STATUS" {
		t.Fatalf("unexpected editor output: %q", res.Message)
	}
}

func TestAliasGithubAppDispatchesInstallGithubApp(t *testing.T) {
	r := DefaultRegistry()
	res, err := r.Dispatch(context.Background(), Context{State: &RuntimeState{}}, "/github-app status")
	if err != nil {
		t.Fatalf("github-app alias failed: %v", err)
	}
	if res.Message == "" || res.Message[:25] != "INSTALL_GITHUB_APP_STATUS" {
		t.Fatalf("unexpected github-app output: %q", res.Message)
	}
}

func TestAliasSlackAppDispatchesInstallSlackApp(t *testing.T) {
	r := DefaultRegistry()
	res, err := r.Dispatch(context.Background(), Context{State: &RuntimeState{}}, "/slack-app status")
	if err != nil {
		t.Fatalf("slack-app alias failed: %v", err)
	}
	if res.Message == "" || res.Message[:24] != "INSTALL_SLACK_APP_STATUS" {
		t.Fatalf("unexpected slack-app output: %q", res.Message)
	}
}
