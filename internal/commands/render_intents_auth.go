package commands

import (
	"fmt"

	"github.com/alliecatowo/alliecode/internal/types"
)

func loginFlowIntents(provider, account string, ready bool) []types.RenderIntent {
	return []types.RenderIntent{
		summaryCardIntent("Login", "Authenticated provider session.",
			field("Provider", defaultDash(provider)),
			field("Account", defaultDash(account)),
			field("Provider ready", boolState(ready, "yes", "no")),
		),
		actionHintsIntent("Actions", hint("Select model", "/model "+provider+"/<model>"), hint("Provider status", "/provider status"), hint("Check runtime", "/status")),
	}
}

func loginStatusIntents(loggedIn bool, provider, account string, ready bool, loginCount, logoutCount int) []types.RenderIntent {
	return []types.RenderIntent{
		summaryCardIntent("Login status", "Current authentication state.",
			field("Logged in", boolState(loggedIn, "yes", "no")),
			field("Provider", defaultDash(provider)),
			field("Account", defaultDash(account)),
			field("Provider ready", boolState(ready, "yes", "no")),
			field("Login count", fmt.Sprintf("%d", loginCount)),
			field("Logout count", fmt.Sprintf("%d", logoutCount)),
		),
	}
}

func loginProviderIntents(provider string, ready bool, loginCount int) []types.RenderIntent {
	return []types.RenderIntent{
		summaryCardIntent("Login provider", "Provider authenticated.",
			field("Provider", defaultDash(provider)),
			field("Provider ready", boolState(ready, "yes", "no")),
			field("Login count", fmt.Sprintf("%d", loginCount)),
		),
		actionHintsIntent("Actions", hint("List models", "/model list "+provider), hint("Select model", "/model "+provider+"/<model>"), hint("Verify provider", "/provider status")),
	}
}

func loginAccountIntents(account, provider string, ready bool, loginCount int) []types.RenderIntent {
	return []types.RenderIntent{
		summaryCardIntent("Login account", "Authenticated account and provider.",
			field("Account", defaultDash(account)),
			field("Provider", defaultDash(provider)),
			field("Provider ready", boolState(ready, "yes", "no")),
			field("Login count", fmt.Sprintf("%d", loginCount)),
		),
		actionHintsIntent("Actions", hint("List models", "/model list "+provider), hint("Select model", "/model "+provider+"/<model>"), hint("Verify provider", "/provider status")),
	}
}

func logoutStatusIntents(loggedIn bool, provider, account string, logoutCount int) []types.RenderIntent {
	return []types.RenderIntent{
		summaryCardIntent("Logout status", "Current logout counters and auth context.",
			field("Logged in", boolState(loggedIn, "yes", "no")),
			field("Provider", defaultDash(provider)),
			field("Account", defaultDash(account)),
			field("Logout count", fmt.Sprintf("%d", logoutCount)),
		),
	}
}

func logoutResultIntents(wasLoggedIn, ready bool, logoutCount int) []types.RenderIntent {
	return []types.RenderIntent{
		summaryCardIntent("Logout", "Session authentication cleared.",
			field("Was logged in", boolState(wasLoggedIn, "yes", "no")),
			field("Provider ready", boolState(ready, "yes", "no")),
			field("Logout count", fmt.Sprintf("%d", logoutCount)),
		),
		actionHintsIntent("Actions", hint("Re-authenticate", "/login provider <name>"), hint("Verify provider", "/provider status"), hint("Repair model", "/model <provider>/<model>")),
	}
}
