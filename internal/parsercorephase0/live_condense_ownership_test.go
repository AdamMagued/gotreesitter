//go:build gts_parsercorephase0

package parsercorephase0

import (
	"errors"
	"testing"
)

func TestLiveCondenseIndependentOwnerKeepsHistorySeparate(t *testing.T) {
	for _, independent := range []bool{false, true} {
		for _, rollback := range []bool{false, true} {
			compact, left, right, _, _ := newStage2PhysicalHeadPair(t)
			key := compact.boundaryKey(3, 1)
			if err := compact.writeBoundary(key, right.Node); err != nil {
				t.Fatal(err)
			}
			if err := compact.recordHeadOwner(left, 101); err != nil {
				t.Fatal(err)
			}
			if err := compact.recordHeadOwner(right, 202); err != nil {
				t.Fatal(err)
			}
			leftNode, _ := compact.node(left.Node)
			links, err := compact.publishedNodeLinksInto(nil, *leftNode)
			if err != nil || len(links) != 1 {
				t.Fatalf("left links=%v err=%v", links, err)
			}
			beforeNodes := len(compact.nodes)
			sentinel := errors.New("rollback witness")
			err = compact.ApplySchedulerAtomic(func(owner SchedulerTransactionToken) error {
				return compact.runLiveCondenseCandidates([]CondenseCandidate{{Head: right, Independent: independent}}, func() error {
					out, err := compact.condenseWithOutcomeAtomic(key, linkInput{prev: links[0].prev, payload: links[0].payload})
					if err != nil {
						return err
					}
					wantPaths := uint64(2)
					if independent {
						wantPaths = 1
						if out.change != condenseNew || out.historicalBoundarySplit || out.historicalNode != 0 {
							t.Fatalf("live independent owner became dead history: %+v", out)
						}
					}
					stats, err := compact.Stats(out.head)
					if err != nil || stats.CurrentExactPaths != wantPaths {
						t.Fatalf("independent=%t paths=%d want=%d err=%v", independent, stats.CurrentExactPaths, wantPaths, err)
					}
					if err := compact.RecordHeadOwnerOwned(owner, out.head, 101); err != nil {
						return err
					}
					if err := compact.RecordHeadOwnerOwned(owner, right, 202); err != nil {
						return err
					}
					if rollback {
						return sentinel
					}
					return nil
				})
			})
			if rollback {
				if !errors.Is(err, sentinel) || len(compact.nodes) != beforeNodes {
					t.Fatalf("rollback nodes=%d want=%d err=%v", len(compact.nodes), beforeNodes, err)
				}
				if canonical, ok := compact.CanonicalBoundary(3, 1, false, 0); !ok || canonical != right {
					t.Fatalf("rollback canonical=%+v/%t want=%+v", canonical, ok, right)
				}
			} else if err != nil {
				t.Fatal(err)
			}
		}
	}
}

func TestCanonicalVersionRemapRequiresMatchingLiveOwner(t *testing.T) {
	type version struct {
		head    Head
		request uint32
	}
	current, canonical := Head{Node: 94}, Head{Node: 95}
	for _, request := range []uint32{1, 2} {
		versions := []version{{current, 1}, {canonical, request}}
		matches := func(v version, head Head) bool { return v.head == head && v.request == 1 }
		if got := CanonicalVersionRemap(current, canonical, true, versions, matches); got != (request == 1) {
			t.Fatalf("owned remap with request=%d: %t", request, got)
		}
		if !CanonicalVersionRemap(current, canonical, false, versions, matches) {
			t.Fatal("shared lexer remap changed")
		}
	}
	if CanonicalVersionRemap(current, canonical, true, []version(nil), func(version, Head) bool { return true }) {
		t.Fatal("an absent version authenticated a remap")
	}
}

func TestExternalLexerOwnsAllTerminals(t *testing.T) {
	for _, test := range []struct {
		tokens  uint32
		symbols []uint16
		want    bool
	}{
		{4, []uint16{1, 2, 3}, true},
		{4, []uint16{3, 1, 2}, true},
		{4, []uint16{1, 1, 3}, false},
		{4, []uint16{0, 2, 3}, false},
		{4, []uint16{1, 2, 4}, false},
		{4, []uint16{1, 2}, false},
		{1, nil, false},
	} {
		if got := ExternalLexerOwnsAllTerminals(test.tokens, test.symbols); got != test.want {
			t.Fatalf("tokens=%d symbols=%v got=%t want=%t", test.tokens, test.symbols, got, test.want)
		}
	}
}

func TestExternalLexerOwnershipProbeScope(t *testing.T) {
	if ExternalLexerOwnershipProbeEnabled() {
		t.Fatal("early external lexer ownership is enabled by default")
	}
	restore := SetExternalLexerOwnershipProbeForTest(true)
	t.Cleanup(restore)
	if !ExternalLexerOwnershipProbeEnabled() {
		t.Fatal("probe did not enable ownership")
	}
	inner := SetExternalLexerOwnershipProbeForTest(false)
	if ExternalLexerOwnershipProbeEnabled() {
		t.Fatal("inner probe did not disable ownership")
	}
	inner()
	if !ExternalLexerOwnershipProbeEnabled() {
		t.Fatal("inner restore lost the outer probe")
	}
	restore()
	if ExternalLexerOwnershipProbeEnabled() {
		t.Fatal("probe leaked into ordinary parsing")
	}
}
