//go:build cgo && treesitter_c_parity && gts_parsercorephase0 && !gts_no_parsercorephase0

package cgoharness

import (
	"fmt"
	"testing"

	gts "github.com/odvcencio/gotreesitter"
	"github.com/odvcencio/gotreesitter/grammars"
)

// Appending an unrelated typed function must preserve the earlier generic
// call. C elects the call both before and after packed histories reconverge.
func TestTSXGenericCallAppendIssue1443LockedC(t *testing.T) {
	base := []byte("export default function Page(p: inventory.DetailsPageProps) {\n    const Form = form<inventory.InventoryItem>(p.Data)\n}\n")
	suffix := []byte("function Editor({ form: Form }: { form: Form<inventory.InventoryItem> }) {\n}\n")
	extended := append(append([]byte(nil), base...), suffix...)
	entry := grammars.DetectLanguageByName("tsx")
	if entry == nil {
		t.Fatal("missing tsx")
	}
	lang := entry.Language()
	cl, err := ParityCLanguage("tsx")
	if err != nil {
		t.Fatal(err)
	}
	compare := func(t *testing.T, source []byte, tree *gts.Tree) {
		t.Helper()
		ct := compactT3ParseC(t, cl, source)
		defer ct.Close()
		if tree == nil || tree.RootNode() == nil {
			t.Fatal("nil tree")
		}
		if diffs := compactT3StructuralDivergences(tree.RootNode(), lang, ct.RootNode()); len(diffs) != 0 {
			for _, d := range diffs[:min(8, len(diffs))] {
				t.Error(compactT3FormatDivergence(d))
			}
			t.Log(tree.RootNode().SExpr(lang))
			t.Fatalf("differences=%d", len(diffs))
		}
	}
	for _, candidate := range []bool{false, true} {
		t.Run(fmt.Sprintf("candidate=%t", candidate), func(t *testing.T) {
			for _, test := range []struct {
				name   string
				source []byte
			}{{"base", base}, {"extended", extended}} {
				t.Run("fresh/"+test.name, func(t *testing.T) {
					p := gts.NewParser(lang)
					p.SetAdmissionCandidateRoute(candidate)
					tree, err := p.Parse(test.source)
					if err != nil {
						t.Fatal(err)
					}
					defer tree.Release()
					compare(t, test.source, tree)
				})
			}
			t.Run("append", func(t *testing.T) {
				p := gts.NewParser(lang)
				p.SetAdmissionCandidateRoute(candidate)
				old, err := p.Parse(base)
				if err != nil {
					t.Fatal(err)
				}
				defer old.Release()
				end := old.RootNode().EndPoint()
				newEnd := end
				for _, b := range suffix {
					if b == '\n' {
						newEnd.Row++
						newEnd.Column = 0
					} else {
						newEnd.Column++
					}
				}
				old.Edit(gts.InputEdit{StartByte: uint32(len(base)), OldEndByte: uint32(len(base)), NewEndByte: uint32(len(extended)), StartPoint: end, OldEndPoint: end, NewEndPoint: newEnd})
				tree, err := p.ParseIncremental(extended, old)
				if err != nil {
					t.Fatal(err)
				}
				if tree != old {
					defer tree.Release()
				}
				compare(t, extended, tree)
			})
		})
	}
}
