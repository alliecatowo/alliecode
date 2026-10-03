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
	syncProviderReadiness(cmdCtx.State)

	if len(inv.Args) == 0 {
		provider := defaultLoginProvider(RuntimeSelectionTruth(cmdCtx.State).ProviderName)
		if err := applyLoginProvider(cmdCtx.State, provider); err != nil {
			return Result{}, err
		}
		message := fmt.Sprintf("LOGIN_FLOW\nprovider=%s\naccount=%s\nlogged_in=true\nprovider_ready=%t\nlogin_count=%d\nstatus=authenticated\nquick_fix_model=/model_%s/<model>\nquick_fix_verify=/provider_status\nnext=run_/model_to_select_a_model_then_/status", normalizeToken(provider), normalizeToken(cmdCtx.State.AuthAccount), cmdCtx.State.ProviderReady, cmdCtx.State.LoginCount, normalizeToken(provider))
		return resultWithIntents(message, loginFlowIntents(provider, cmdCtx.State.AuthAccount, cmdCtx.State.ProviderReady)...), nil
	}

	if len(inv.Args) == 1 && equalFoldTrimmed(inv.Args[0], "status") {
		message := fmt.Sprintf("LOGIN_STATUS\nlogged_in=%t\nprovider=%s\naccount=%s\nprovider_ready=%t\nlogin_count=%d\nlogout_count=%d", cmdCtx.State.LoggedIn, normalizeToken(cmdCtx.State.ProviderName), normalizeToken(cmdCtx.State.AuthAccount), cmdCtx.State.ProviderReady, cmdCtx.State.LoginCount, cmdCtx.State.LogoutCount)
		return resultWithIntents(message, loginStatusIntents(cmdCtx.State.LoggedIn, cmdCtx.State.ProviderName, cmdCtx.State.AuthAccount, cmdCtx.State.ProviderReady, cmdCtx.State.LoginCount, cmdCtx.State.LogoutCount)...), nil
	}

	if len(inv.Args) == 2 && equalFoldTrimmed(inv.Args[0], "provider") {
		provider := strings.ToLower(strings.TrimSpace(inv.Args[1]))
		if provider == "" {
			return Result{}, fmt.Errorf("%s", loginUsage)
		}
		if !isSupportedProviderName(provider) {
			return Result{}, fmt.Errorf("%s", unknownProviderGuidance(provider))
		}
		if err := applyLoginProvider(cmdCtx.State, provider); err != nil {
			return Result{}, err
		}
		message := fmt.Sprintf("LOGIN_PROVIDER\nprovider=%s\nlogged_in=true\nprovider_ready=true\nlogin_count=%d\nstatus=authenticated\nquick_fix_model=/model_%s/<model>\nquick_fix_verify=/provider_status\nnext=run_/model_list_%s_or_/model_%s/<model>", normalizeToken(provider), cmdCtx.State.LoginCount, normalizeToken(provider), normalizeToken(provider), normalizeToken(provider))
		return resultWithIntents(message, loginProviderIntents(provider, true, cmdCtx.State.LoginCount)...), nil
	}

	if len(inv.Args) >= 2 && equalFoldTrimmed(inv.Args[0], "account") {
		account := normalizeAccountName(strings.Join(inv.Args[1:], " "))
		if account == "" {
			return Result{}, fmt.Errorf("%s", loginUsage)
		}
		cmdCtx.State.AuthAccount = account
		provider := defaultLoginProvider(RuntimeSelectionTruth(cmdCtx.State).ProviderName)
		if err := applyLoginProvider(cmdCtx.State, provider); err != nil {
			return Result{}, err
		}
		message := fmt.Sprintf("LOGIN_ACCOUNT\naccount=%s\nprovider=%s\nlogged_in=true\nprovider_ready=%t\nlogin_count=%d\nstatus=authenticated\nquick_fix_model=/model_%s/<model>\nquick_fix_verify=/provider_status\nnext=run_/model_list_%s_or_/model_%s/<model>", normalizeToken(account), normalizeToken(provider), cmdCtx.State.ProviderReady, cmdCtx.State.LoginCount, normalizeToken(provider), normalizeToken(provider), normalizeToken(provider))
		return resultWithIntents(message, loginAccountIntents(account, provider, cmdCtx.State.ProviderReady, cmdCtx.State.LoginCount)...), nil
	}

	if len(inv.Args) == 1 {
		provider := strings.ToLower(strings.TrimSpace(inv.Args[0]))
		if provider == "" {
			return Result{}, fmt.Errorf("%s", loginUsage)
		}
		if !isSupportedProviderName(provider) {
			return Result{}, fmt.Errorf("%s", unknownProviderGuidance(provider))
		}
		if err := applyLoginProvider(cmdCtx.State, provider); err != nil {
			return Result{}, err
		}
		message := fmt.Sprintf("LOGIN_PROVIDER\nprovider=%s\nlogged_in=true\nprovider_ready=true\nlogin_count=%d\nstatus=authenticated\nquick_fix_model=/model_%s/<model>\nquick_fix_verify=/provider_status\nnext=run_/model_list_%s_or_/model_%s/<model>", normalizeToken(provider), cmdCtx.State.LoginCount, normalizeToken(provider), normalizeToken(provider), normalizeToken(provider))
		return resultWithIntents(message, loginProviderIntents(provider, true, cmdCtx.State.LoginCount)...), nil
	}

	return Result{}, fmt.Errorf("%s", loginUsage)
}

func applyLoginProvider(state *RuntimeState, provider string) error {
	selection, err := ReconcileProviderSelection(provider, RuntimeSelectionTruth(state).ModelName)
	if err != nil {
		return err
	}
	if err := ApplyProviderModelSelection(state, selection); err != nil {
		return err
	}
	state.LoggedIn = true
	state.AuthProvider = provider
	syncProviderReadiness(state)
	state.LoginCount++
	return nil
}

func executeLogoutCommand(cmdCtx Context, inv Invocation) (Result, error) {
	if cmdCtx.State == nil {
		return Result{}, fmt.Errorf("missing command runtime state")
	}
	syncProviderReadiness(cmdCtx.State)
	if len(inv.Args) == 1 && equalFoldTrimmed(inv.Args[0], "status") {
		message := fmt.Sprintf("LOGOUT_STATUS\nlogged_in=%t\nprovider=%s\naccount=%s\nlogout_count=%d", cmdCtx.State.LoggedIn, normalizeToken(cmdCtx.State.ProviderName), normalizeToken(cmdCtx.State.AuthAccount), cmdCtx.State.LogoutCount)
		return resultWithIntents(message, logoutStatusIntents(cmdCtx.State.LoggedIn, cmdCtx.State.ProviderName, cmdCtx.State.AuthAccount, cmdCtx.State.LogoutCount)...), nil
	}
	if len(inv.Args) > 0 {
		return Result{}, fmt.Errorf("usage: /logout [status]")
	}

	wasLoggedIn := cmdCtx.State.LoggedIn
	cmdCtx.State.LoggedIn = false
	cmdCtx.State.AuthProvider = ""
	cmdCtx.State.AuthAccount = ""
	syncProviderReadiness(cmdCtx.State)
	cmdCtx.State.LogoutCount++
	message := fmt.Sprintf("LOGOUT_RESULT\nwas_logged_in=%t\nlogged_in=false\nprovider_ready=%t\nlogout_count=%d\nquick_fix_auth=%s\nnext=run_/login_provider_<name>_to_reauthenticate", wasLoggedIn, cmdCtx.State.ProviderReady, cmdCtx.State.LogoutCount, normalizeToken(providerRecoveryHint(cmdCtx.State.ProviderName)))
	return resultWithIntents(message, logoutResultIntents(wasLoggedIn, cmdCtx.State.ProviderReady, cmdCtx.State.LogoutCount)...), nil
}
