package commands

import "testing"

func TestParseTableDriven(t *testing.T) {
	tests := []struct {
		name    string
		input   string
		want    Invocation
		wantErr error
	}{
		{name: "basic", input: "/status", want: Invocation{Raw: "/status", Name: "status", Args: nil, IsSlash: true}},
		{name: "trimmed", input: "  /help model  ", want: Invocation{Raw: "/help model", Name: "help", Args: []string{"model"}, IsSlash: true}},
		{name: "escaped", input: `/copy hi\ there`, want: Invocation{Raw: `/copy hi\ there`, Name: "copy", Args: []string{"hi there"}, IsSlash: true}},
		{name: "not slash", input: "hello", wantErr: ErrNotSlashCommand},
		{name: "empty", input: "/", wantErr: ErrEmptyCommand},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			got, err := Parse(tc.input)
			if tc.wantErr != nil {
				if err != tc.wantErr {
					t.Fatalf("expected error %v, got %v", tc.wantErr, err)
				}
				return
			}
			if err != nil {
				t.Fatalf("Parse() error = %v", err)
			}
			if got.Name != tc.want.Name || got.Raw != tc.want.Raw || got.IsSlash != tc.want.IsSlash {
				t.Fatalf("unexpected invocation: %#v", got)
			}
			if len(got.Args) != len(tc.want.Args) {
				t.Fatalf("unexpected args length: %#v", got.Args)
			}
			for i := range got.Args {
				if got.Args[i] != tc.want.Args[i] {
					t.Fatalf("unexpected arg %d: %q", i, got.Args[i])
				}
			}
		})
	}
}
