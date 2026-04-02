package fixtures

import (
	"os"
	"strings"
	"testing"
)

func RequireEnv(t *testing.T, key string) string {
	t.Helper()
	v := strings.TrimSpace(os.Getenv(key))
	if v == "" {
		t.Skipf("set %s=1 to run this test", key)
	}
	return v
}
