//go:build cgo && treesitter_c_parity

package cgoharness

import (
	"fmt"
	"strings"
	"testing"

	gotreesitter "github.com/odvcencio/gotreesitter"
	"github.com/odvcencio/gotreesitter/grammars"
	sitter "github.com/tree-sitter/go-tree-sitter"
)

// Keep the issue #1399 witnesses strict: a clean parse alone would miss a
// directive attached to the preceding expression or an incorrect field/span.
func TestSwiftDirectiveNewlinesMatchLockedC(t *testing.T) {
	cLang, err := COracleLanguage("swift")
	if err != nil {
		t.Fatal(err)
	}
	cParser := sitter.NewParser()
	t.Cleanup(cParser.Close)
	if err := cParser.SetLanguage(cLang); err != nil {
		t.Fatal(err)
	}
	lang := grammars.SwiftLanguage()
	cases := []struct{ name, source string }{
		{"warning", "let x = 1\n#warning(\"w\")\n"},
		{"if", "let x = 1\n#if compiler(>=5.9)\nlet y = 2\n#endif\n"},
		{"if-import", "#if FOO\nimport A\n#else\nimport B\n#endif\n"},
		{"preview", "struct A {}\n#Preview { Text(\"hi\") }\n"},
		{"expect", "func f() {\n  let x = 1\n  #expect(x == 1)\n}\n"},
		{"explicit-separator", "let x = 1;\n#warning(\"w\")\n"},
		{"line-comment", "let x = 1 // comment\n#warning(\"w\")\n"},
		{"block-comment", "let x = 1 /* comment */\n#warning(\"w\")\n"},
		{"intervening-comment", "let x = 1\n// comment\n#warning(\"w\")\n"},
		{"empty-type-body", "struct S {\n#if FOO\nvar x: Int\n#else\nvar y: Int\n#endif\n}\n"},
		{"type-member-separator", "struct S {\nvar x: Int\n#if FOO\nvar y: Int\n#else\nvar z: Int\n#endif\n}\n"},
		{"function-body", "func f() {\nlet x = 1\n#if FOO\nprint(x)\n#elseif BAR\nprint(2)\n#else\nprint(3)\n#endif\n}\n"},
		{"nested-directives", "#if FOO\n#if BAR\nimport A\n#else\nimport B\n#endif\n#else\nimport C\n#endif\n"},
		{"consecutive-directives", "#warning(\"a\")\n#warning(\"b\")\n"},
		{"raw-string", "let x = #\"#if FOO\"#\n#warning(\"w\")\n"},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			// Insert and remove a newline at the first statement boundary,
			// then change line endings and EOF trivia in the same edit session.
			blank := strings.Replace(tc.source, "\n", "\n\n", 1)
			steps := []string{tc.source, blank, tc.source, strings.ReplaceAll(tc.source, "\n", "\r\n"), tc.source, strings.TrimSuffix(tc.source, "\n"), tc.source}
			for _, candidate := range []bool{false, true} {
				t.Run(fmt.Sprintf("candidate=%t", candidate), func(t *testing.T) {
					parser := gotreesitter.NewParser(lang)
					parser.SetAdmissionCandidateRoute(candidate)
					var oldSource []byte
					var oldTree *gotreesitter.Tree
					defer func() { releaseGoTree(oldTree) }()
					for i, step := range steps {
						source := []byte(step)
						cTree := cParser.Parse(source, nil)
						if cTree == nil {
							t.Fatalf("step %d: C returned no tree", i)
						}
						func() {
							defer cTree.Close()
							fresh, err := parser.Parse(source)
							if err != nil {
								t.Fatalf("step %d fresh: %v", i, err)
							}
							defer fresh.Release()
							assertLockedCTreeExact(t, fmt.Sprintf("%s step %d fresh", tc.name, i), fresh, lang, cTree)
							if oldTree != nil {
								start, oldEnd, newEnd := swiftDirectiveChangedRange(oldSource, source)
								oldTree.Edit(gotreesitter.InputEdit{
									StartByte: uint32(start), OldEndByte: uint32(oldEnd), NewEndByte: uint32(newEnd),
									StartPoint: pointAtOffset(oldSource, start), OldEndPoint: pointAtOffset(oldSource, oldEnd), NewEndPoint: pointAtOffset(source, newEnd),
								})
							}
							tree, err := parser.ParseIncremental(source, oldTree)
							releaseGoTree(oldTree)
							oldTree = tree
							if err != nil {
								t.Fatalf("step %d incremental: %v", i, err)
							}
							assertLockedCTreeExact(t, fmt.Sprintf("%s step %d incremental", tc.name, i), tree, lang, cTree)
						}()
						oldSource = source
					}
				})
			}
		})
	}
}

func swiftDirectiveChangedRange(before, after []byte) (start, oldEnd, newEnd int) {
	for start < len(before) && start < len(after) && before[start] == after[start] {
		start++
	}
	oldEnd, newEnd = len(before), len(after)
	for oldEnd > start && newEnd > start && before[oldEnd-1] == after[newEnd-1] {
		oldEnd--
		newEnd--
	}
	return
}
