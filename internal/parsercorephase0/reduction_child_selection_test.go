package parsercorephase0

import (
	"errors"
	"testing"
)

func TestFreshSchedulerSessionScopesReductionPressure(t *testing.T) {
	for _, exit := range []string{"success", "error", "panic"} {
		t.Run(exit, func(t *testing.T) {
			c := newCSelectionCompareTestCore(t)
			c.reductionScratch.pressureSeen = true
			sentinel := errors.New("fresh selection failure")
			panicked := false
			func() {
				defer func() {
					if recovered := recover(); recovered != nil {
						panicked = true
						if exit != "panic" || recovered != sentinel {
							t.Fatalf("unexpected panic: %v", recovered)
						}
					}
				}()
				err := c.RunFreshSchedulerSession(func(SchedulerTransactionToken) error {
					if c.reductionScratch.pressureSeen {
						t.Fatal("fresh session inherited reduction pressure")
					}
					c.reductionScratch.pressureSeen = true
					switch exit {
					case "error":
						return sentinel
					case "panic":
						panic(sentinel)
					}
					return nil
				})
				if (exit == "error") != errors.Is(err, sentinel) {
					t.Fatalf("session exit: %v", err)
				}
			}()
			if panicked != (exit == "panic") {
				t.Fatal("fresh session did not preserve its panic contract")
			}
			if c.reductionScratch.pressureSeen {
				t.Fatal("fresh session retained reduction pressure")
			}
		})
	}
}

func reductionSelectionPair(t *testing.T, c *Core, leftLeaf, rightLeaf SubtreeID) (linkRecord, linkInput) {
	t.Helper()
	appendNode := func(symbol Symbol, children []SubtreeID) SubtreeID {
		id, err := c.appendSubtree(subtreeRecord{symbol: symbol}, children, nil, nil)
		if err != nil {
			t.Fatal(err)
		}
		return id
	}
	leftChild := appendNode(20, []SubtreeID{leftLeaf})
	rightChild := appendNode(10, []SubtreeID{rightLeaf})
	left := appendNode(50, []SubtreeID{leftChild})
	right := appendNode(50, []SubtreeID{rightChild})
	c.reductionScratch.selectCleanChildren = true
	c.reductionScratch.batchParents = []SubtreeID{left, right}
	return linkRecord{prev: 1, payload: left}, linkInput{prev: 1, payload: right}
}

func reductionSelectionLeaf(t *testing.T, c *Core, record subtreeRecord) SubtreeID {
	t.Helper()
	record.terminal = true
	record.external = true
	id, err := c.appendSubtree(record, nil, nil, nil)
	if err != nil {
		t.Fatal(err)
	}
	c.externalProvenance = append(c.externalProvenance, externalPayloadProvenance{payload: id})
	return id
}

func TestReconvergedReductionChildSelectionRequiresCompleteScannerHistory(t *testing.T) {
	for _, mixed := range []bool{false, true} {
		c := newCSelectionCompareTestCore(t)
		dfa, err := c.appendSubtree(subtreeRecord{terminal: true, symbol: 3, endByte: 1}, nil, nil, nil)
		if err != nil {
			t.Fatal(err)
		}
		leaf := dfa
		if mixed {
			external := reductionSelectionLeaf(t, c, subtreeRecord{symbol: 2, endByte: 1})
			leaf, err = c.appendSubtree(subtreeRecord{symbol: 40}, []SubtreeID{external, dfa}, nil, nil)
			if err != nil {
				t.Fatal(err)
			}
		}
		left, right := reductionSelectionPair(t, c, leaf, leaf)
		if _, selected, err := c.compareReconvergedReductionChildren(left, right); err != nil || selected {
			t.Fatalf("mixed=%t selected=%t err=%v: DFA history must retain alternatives", mixed, selected, err)
		}
		if len(c.popScratch.rev) != 0 || len(c.popScratch.external) != 0 {
			t.Fatal("unsupported history retained live cursors")
		}
	}
}

func TestReconvergedReductionChildSelectionUsesCOrder(t *testing.T) {
	c := newCSelectionCompareTestCore(t)
	leaf := reductionSelectionLeaf(t, c, subtreeRecord{symbol: 2, endByte: 1})
	left, right := reductionSelectionPair(t, c, leaf, leaf)
	if cmp, selected, err := c.compareReconvergedReductionChildren(left, right); err != nil || !selected || cmp != 1 {
		t.Fatalf("selection=(%d,%t,%v), want incoming raw-symbol winner", cmp, selected, err)
	}
	if cmp, selected, err := c.compareReconvergedReductionChildren(linkRecord{prev: 1, payload: right.payload}, linkInput{prev: 1, payload: left.payload}); err != nil || !selected || cmp != -1 {
		t.Fatalf("reverse selection=(%d,%t,%v), want incumbent winner", cmp, selected, err)
	}
	if len(c.popScratch.rev) != 0 || len(c.popScratch.external) != 0 {
		t.Fatal("selection retained a live leaf cursor")
	}
}

func TestReconvergedReductionChildSelectionScope(t *testing.T) {
	for _, name := range []string{"disabled", "other batch", "other predecessor", "other production", "other precedence", "other span", "missing", "error", "different terminals"} {
		t.Run(name, func(t *testing.T) {
			c := newCSelectionCompareTestCore(t)
			leaf := reductionSelectionLeaf(t, c, subtreeRecord{symbol: 2, endByte: 1})
			other := leaf
			if name == "different terminals" {
				other = reductionSelectionLeaf(t, c, subtreeRecord{symbol: 3, endByte: 1})
			}
			left, right := reductionSelectionPair(t, c, leaf, other)
			switch name {
			case "disabled":
				c.reductionScratch.selectCleanChildren = false
			case "other batch":
				c.reductionScratch.batchParents = c.reductionScratch.batchParents[1:]
			case "other predecessor":
				right.prev = 2
			case "other production":
				c.subtrees[right.payload-1].productionID = 1
			case "other precedence":
				c.subtrees[right.payload-1].dynamicPrecedence = 1
			case "other span":
				c.subtrees[right.payload-1].endByte = 2
			case "missing":
				c.subtrees[leaf-1].missing = true
			case "error":
				c.subtrees[leaf-1].symbol = ErrorRegionSymbol
			}
			if _, selected, err := c.compareReconvergedReductionChildren(left, right); err != nil || selected {
				t.Fatalf("unsupported history selected=%t err=%v", selected, err)
			}
			if len(c.popScratch.rev) != 0 || len(c.popScratch.external) != 0 {
				t.Fatal("decline retained a live leaf cursor")
			}
		})
	}
}

func TestReconvergedReductionChildSelectionScannerHistory(t *testing.T) {
	for _, name := range []string{"same", "different start", "different end", "missing provenance", "invalid checkpoint", "opaque"} {
		t.Run(name, func(t *testing.T) {
			c := newCSelectionCompareTestCore(t)
			start := mustInternCheckpoint(t, c, []byte{1})
			end := mustInternCheckpoint(t, c, []byte{2})
			a := reductionSelectionLeaf(t, c, subtreeRecord{symbol: 2, endByte: 1, external: true})
			b := reductionSelectionLeaf(t, c, subtreeRecord{symbol: 2, endByte: 1, external: true})
			c.externalProvenance = []externalPayloadProvenance{{payload: a, start: start, end: end}, {payload: b, start: start, end: end}}
			switch name {
			case "different start":
				c.externalProvenance[1].start = end
			case "different end":
				c.externalProvenance[1].end = start
			case "missing provenance":
				c.externalProvenance = c.externalProvenance[:1]
			case "invalid checkpoint":
				c.externalProvenance[1].end = end + 100
			case "opaque":
				c.subtrees[b-1].externalProvenanceState = subtreeExternalProvenanceReusedOpaque
			}
			left, right := reductionSelectionPair(t, c, a, b)
			_, selected, err := c.compareReconvergedReductionChildren(left, right)
			if selected != (name == "same") || (err != nil) != (name == "invalid checkpoint" || name == "opaque") {
				t.Fatalf("selection=%t err=%v", selected, err)
			}
		})
	}
}

func TestReconvergedReductionChildSelectionBoundsAndClearsScratch(t *testing.T) {
	c := newCSelectionCompareTestCore(t)
	leaf := reductionSelectionLeaf(t, c, subtreeRecord{symbol: 2})
	left, right := reductionSelectionPair(t, c, leaf, leaf)
	child := c.children[c.subtrees[left.payload-1].firstChild]
	c.children[c.subtrees[child-1].firstChild] = child
	if _, selected, err := c.compareReconvergedReductionChildren(left, right); err == nil || selected {
		t.Fatalf("corrupt child selected=%t err=%v", selected, err)
	}
	if len(c.popScratch.rev) != 0 || len(c.popScratch.external) != 0 {
		t.Fatal("failed selection retained live cursors")
	}
	c.reductionScratch.finish()
	if c.reductionScratch.selectCleanChildren {
		t.Fatal("completed reduction leaked selection authority")
	}
}

func TestReconvergedReductionChildSelectionPreservesSkippedPrefixProof(t *testing.T) {
	c := newCSelectionCompareTestCore(t)
	a := reductionSelectionLeaf(t, c, subtreeRecord{symbol: 2, startByte: 1, endByte: 2})
	b := reductionSelectionLeaf(t, c, subtreeRecord{symbol: 2, startByte: 1, endByte: 2})
	c.lexerSkippedPrefixes = []lexerSkippedPrefixProvenance{{payload: a, start: 0}}
	left, right := reductionSelectionPair(t, c, a, b)
	if _, selected, err := c.compareReconvergedReductionChildren(left, right); err != nil || selected {
		t.Fatalf("different skipped-prefix proof selected=%t err=%v", selected, err)
	}
}

func TestReconvergedReductionChildSelectionBoundsRepeatedDAGWork(t *testing.T) {
	c, err := New(&fakeTable{}, Limits{MaxSubtrees: 64, MaxChildren: 64})
	if err != nil {
		t.Fatal(err)
	}
	left := reductionSelectionLeaf(t, c, subtreeRecord{symbol: 2})
	right := reductionSelectionLeaf(t, c, subtreeRecord{symbol: 2})
	for level := 0; level < 6; level++ {
		left, err = c.appendSubtree(subtreeRecord{symbol: 20}, []SubtreeID{left, left}, nil, nil)
		if err != nil {
			t.Fatal(err)
		}
		right, err = c.appendSubtree(subtreeRecord{symbol: 10}, []SubtreeID{right, right}, nil, nil)
		if err != nil {
			t.Fatal(err)
		}
	}
	a, b := reductionSelectionPair(t, c, left, right)
	if _, selected, err := c.compareReconvergedReductionChildren(a, b); err != nil || selected {
		t.Fatalf("repeated DAG selection=%t err=%v, want conservative retention", selected, err)
	}
}
