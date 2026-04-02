package integration_test

import (
	"os"
	"path/filepath"
	"reflect"
	"runtime"
	"testing"

	"github.com/alliecatowo/alliecode/internal/plugins"
)

func TestPluginPolicyFixtureMatrixNormalization(t *testing.T) {
	t.Parallel()

	cases := []string{
		"case01",
		"case02",
		"case03",
		"case04",
		"case05",
		"case06",
		"case07",
		"case08",
		"case09",
		"case10",
		"case11",
		"case12",
		"case13",
		"case14",
		"case15",
		"case16",
		"case17",
		"case18",
		"case19",
		"case20",
	}

	for _, name := range cases {
		name := name
		t.Run(name, func(t *testing.T) {
			input := readPolicyFixture(t, name+"_input.yaml")
			expected := readPolicyFixture(t, name+"_expected.yaml")

			gotState := loadPolicyFixtureState(t, input)
			expectedState := loadPolicyFixtureState(t, expected)

			if !reflect.DeepEqual(gotState, expectedState) {
				t.Fatalf("normalized policy mismatch for %s\n got: %#v\nwant: %#v", name, gotState, expectedState)
			}
		})
	}
}

func loadPolicyFixtureState(t *testing.T, fixture []byte) plugins.PolicyState {
	t.Helper()
	tmp := filepath.Join(t.TempDir(), "policy.yaml")
	if err := os.WriteFile(tmp, fixture, 0o644); err != nil {
		t.Fatalf("WriteFile() fixture error = %v", err)
	}
	store := plugins.NewPolicyStore(tmp)
	state, err := store.Load()
	if err != nil {
		t.Fatalf("Load() fixture error = %v", err)
	}
	if err := store.Save(state); err != nil {
		t.Fatalf("Save() fixture error = %v", err)
	}
	roundTripped, err := store.Load()
	if err != nil {
		t.Fatalf("Load(round trip) error = %v", err)
	}
	return roundTripped
}

func readPolicyFixture(t *testing.T, name string) []byte {
	t.Helper()
	_, filename, _, ok := runtime.Caller(0)
	if !ok {
		t.Fatalf("runtime.Caller(0) failed")
	}
	path := filepath.Join(filepath.Dir(filename), "testdata", "plugin_policy_cases", name)
	b, err := os.ReadFile(path)
	if err != nil {
		t.Fatalf("read fixture %s: %v", path, err)
	}
	return b
}
