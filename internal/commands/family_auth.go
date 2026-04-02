package commands

import (
	"fmt"
	"strings"
)

const loginUsage = "usage: /login [status|provider <name>|account <name>|<provider>]"

func executeLoginCommand(cmdCtx Context, inv Invocation) (Result, error) {
	if cmdCtx.State == nil {
		return Result{}, fmt.Errorf("missing command runtime state")
	}

	if len(inv.Args) == 0 {
		provider := defaultLoginProvider(cmdCtx.State.ProviderName)
		cmdCtx.State.ProviderName = provider
		cmdCtx.State.LoggedIn = true
		cmdCtx.State.ProviderReady = true
		cmdCtx.State.LoginCount++
		return Result{Handled: true, Message: fmt.Sprintf("LOGIN_FLOW\nprovider=%s\naccount=%s\nlogged_in=true\nprovider_ready=true\nlogin_count=%d\nnext=follow_provider_oauth_or_api_key_setup", normalizeToken(provider), normalizeToken(cmdCtx.State.AuthAccount), cmdCtx.State.LoginCount)}, nil
	}

	if len(inv.Args) == 1 && equalFoldTrimmed(inv.Args[0], "status") {
		return Result{Handled: true, Message: fmt.Sprintf("LOGIN_STATUS\nlogged_in=%t\nprovider=%s\naccount=%s\nprovider_ready=%t\nlogin_count=%d\nlogout_count=%d", cmdCtx.State.LoggedIn, normalizeToken(cmdCtx.State.ProviderName), normalizeToken(cmdCtx.State.AuthAccount), cmdCtx.State.ProviderReady, cmdCtx.State.LoginCount, cmdCtx.State.LogoutCount)}, nil
	}

	if len(inv.Args) == 2 && equalFoldTrimmed(inv.Args[0], "provider") {
		provider := strings.ToLower(strings.TrimSpace(inv.Args[1]))
		if provider == "" {
			return Result{}, fmt.Errorf("%s", loginUsage)
		}
		applyLoginProvider(cmdCtx.State, provider)
		return Result{Handled: true, Message: fmt.Sprintf("LOGIN_PROVIDER\nprovider=%s\nlogged_in=true\nprovider_ready=true\nlogin_count=%d", normalizeToken(provider), cmdCtx.State.LoginCount)}, nil
	}

	if len(inv.Args) >= 2 && equalFoldTrimmed(inv.Args[0], "account") {
		account := normalizeAccountName(strings.Join(inv.Args[1:], " "))
		if account == "" {
			return Result{}, fmt.Errorf("%s", loginUsage)
		}
		cmdCtx.State.AuthAccount = account
		provider := defaultLoginProvider(cmdCtx.State.ProviderName)
		cmdCtx.State.ProviderName = provider
		cmdCtx.State.LoggedIn = true
		cmdCtx.State.ProviderReady = true
		cmdCtx.State.LoginCount++
		return Result{Handled: true, Message: fmt.Sprintf("LOGIN_ACCOUNT\naccount=%s\nprovider=%s\nlogged_in=true\nprovider_ready=true\nlogin_count=%d", normalizeToken(account), normalizeToken(provider), cmdCtx.State.LoginCount)}, nil
	}

	if len(inv.Args) == 1 {
		provider := strings.ToLower(strings.TrimSpace(inv.Args[0]))
		if provider == "" {
			return Result{}, fmt.Errorf("%s", loginUsage)
		}
		applyLoginProvider(cmdCtx.State, provider)
		return Result{Handled: true, Message: fmt.Sprintf("LOGIN_PROVIDER\nprovider=%s\nlogged_in=true\nprovider_ready=true\nlogin_count=%d", normalizeToken(provider), cmdCtx.State.LoginCount)}, nil
	}

	return Result{}, fmt.Errorf("%s", loginUsage)
}

func applyLoginProvider(state *RuntimeState, provider string) {
	state.ProviderName = provider
	state.ProviderReady = true
	state.LoggedIn = true
	state.LoginCount++
}

func executeLogoutCommand(cmdCtx Context, inv Invocation) (Result, error) {
	if cmdCtx.State == nil {
		return Result{}, fmt.Errorf("missing command runtime state")
	}
	if len(inv.Args) == 1 && equalFoldTrimmed(inv.Args[0], "status") {
		return Result{Handled: true, Message: fmt.Sprintf("LOGOUT_STATUS\nlogged_in=%t\nprovider=%s\naccount=%s\nlogout_count=%d", cmdCtx.State.LoggedIn, normalizeToken(cmdCtx.State.ProviderName), normalizeToken(cmdCtx.State.AuthAccount), cmdCtx.State.LogoutCount)}, nil
	}
	if len(inv.Args) > 0 {
		return Result{}, fmt.Errorf("usage: /logout [status]")
	}

	wasLoggedIn := cmdCtx.State.LoggedIn
	cmdCtx.State.LoggedIn = false
	cmdCtx.State.ProviderReady = false
	cmdCtx.State.AuthAccount = ""
	cmdCtx.State.LogoutCount++
	return Result{Handled: true, Message: fmt.Sprintf("LOGOUT_RESULT\nwas_logged_in=%t\nlogged_in=false\nprovider_ready=false\nlogout_count=%d", wasLoggedIn, cmdCtx.State.LogoutCount)}, nil
}
