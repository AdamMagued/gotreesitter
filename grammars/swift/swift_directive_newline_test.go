package swift_test

import (
	"testing"

	"github.com/odvcencio/gotreesitter"
	"github.com/odvcencio/gotreesitter/grammars/swift"
)

// Issue #1399 regression witnesses contributed by AdamMagued in PR #1442.
func TestSwiftDirectiveNewlineRepros(t *testing.T) {
	lang := swift.Language()

	cases := []struct {
		name string
		src  string
	}{
		{"warning", "let x = 1\n#warning(\"w\")\n"},
		{"if", "let x = 1\n#if compiler(>=5.9)\nlet y = 2\n#endif\n"},
		{"if-import", "#if FOO\nimport A\n#else\nimport B\n#endif\n"},
		{"preview", "struct A {}\n#Preview { Text(\"hi\") }\n"},
		{"expect", "func f() {\n  let x = 1\n  #expect(x == 1)\n}\n"},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			for _, candidate := range []bool{false, true} {
				t.Run(candidateRouteName(candidate), func(t *testing.T) {
					p := gotreesitter.NewParser(lang)
					p.SetAdmissionCandidateRoute(candidate)
					tree, err := p.Parse([]byte(tc.src))
					if err != nil {
						t.Fatalf("parse error: %v", err)
					}
					defer tree.Release()
					root := tree.RootNode()
					if root.HasError() {
						t.Errorf("expected clean parse, got error: %s", root.SExpr(lang))
					}
				})
			}
		})
	}
}

func candidateRouteName(candidate bool) string {
	if candidate {
		return "candidate"
	}
	return "production"
}
