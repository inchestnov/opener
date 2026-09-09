package pathx

import "testing"

func TestExpand_Tilde(t *testing.T) {
	t.Setenv("HOME", "/home/tester")
	cases := map[string]string{
		"~":             "/home/tester",
		"~/code":        "/home/tester/code",
		"relative/path": "relative/path",
		"/absolute":     "/absolute",
		"~user/thing":   "~user/thing",
	}
	for in, want := range cases {
		if got := Expand(in); got != want {
			t.Errorf("Expand(%q) = %q, want %q", in, got, want)
		}
	}
}

func TestExpand_Env(t *testing.T) {
	t.Setenv("HOME", "/home/tester")
	t.Setenv("SRC_ROOT", "/home/tester/workspace")
	t.Setenv("TILDE_ROOT", "~/from-var")
	t.Setenv("EMPTY", "")

	cases := map[string]string{
		"$SRC_ROOT":             "/home/tester/workspace",
		"$SRC_ROOT/core-apps":   "/home/tester/workspace/core-apps",
		"${SRC_ROOT}/core-apps": "/home/tester/workspace/core-apps",
		"$TILDE_ROOT/x":         "/home/tester/from-var/x",
		"$EMPTY/x":              "/x",
		"https://example.com":   "https://example.com",
		"/no/vars/here":         "/no/vars/here",
	}
	for in, want := range cases {
		if got := Expand(in); got != want {
			t.Errorf("Expand(%q) = %q, want %q", in, got, want)
		}
	}
}

// An unset variable stays literal rather than collapsing to "", so a typo
// fails loudly as a nonexistent path instead of silently rewriting the
// value into something rooted at /.
func TestExpand_UnsetVarStaysLiteral(t *testing.T) {
	if got := Expand("$DEFINITELY_UNSET_VAR/core-apps"); got != "$DEFINITELY_UNSET_VAR/core-apps" {
		t.Errorf("Expand(unset) = %q, want it left verbatim", got)
	}
}
