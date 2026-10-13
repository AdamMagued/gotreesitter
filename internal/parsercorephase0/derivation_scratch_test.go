package parsercorephase0

import (
	"errors"
	"reflect"
	"testing"
)

func TestDerivationScratchWarmEnumerationAllocatesNothing(t *testing.T) {
	for _, ambiguous := range []bool{false, true} {
		c, head := newSoleDerivationBranches()
		if !ambiguous {
			head = Head{Node: 2}
		}
		want, err := c.Derivations(head)
		if err != nil {
			t.Fatal(err)
		}
		var scratch DerivationScratch
		got, err := c.DerivationsWithScratch(head, &scratch, nil)
		if err != nil || !reflect.DeepEqual(got, want) {
			t.Fatalf("got=%+v err=%v want=%+v", got, err, want)
		}
		if allocations := testing.AllocsPerRun(100, func() {
			paths, err := c.DerivationsWithScratch(head, &scratch, nil)
			if err != nil || len(paths) != len(want) {
				panic("enumeration failed")
			}
			scratch.Reset()
		}); allocations != 0 {
			t.Fatalf("ambiguous=%v warm allocations=%v", ambiguous, allocations)
		}
		// The copying API owns its result across reuse and reset of scratch.
		if len(want[0].Payloads) == 0 || want[0].Payloads[0] != 11 {
			t.Fatal("scratch reuse changed an independently owned derivation")
		}
	}
}

func TestDerivationScratchResetsAfterCancellationAndRejectsReentry(t *testing.T) {
	c, head := newSoleDerivationBranches()
	var scratch DerivationScratch
	stop := errors.New("stop enumeration")
	polls := 0
	if paths, err := c.DerivationsWithScratch(head, &scratch, func() error {
		polls++
		if polls == 2 {
			return stop
		}
		return nil
	}); !errors.Is(err, stop) || paths != nil {
		t.Fatalf("paths=%v error=%v", paths, err)
	}
	if scratch.inUse {
		t.Fatal("cancelled enumeration retained its lease")
	}
	for _, color := range scratch.colors {
		if color != 0 {
			t.Fatal("cancelled enumeration retained graph state")
		}
	}
	for _, f := range scratch.frames {
		if len(f.paths) != 0 || len(f.links) != 0 || len(f.linkIDs) != 0 {
			t.Fatal("cancelled enumeration retained results")
		}
	}
	if _, err := c.DerivationsWithScratch(head, &scratch, func() error {
		if _, err := c.DerivationsWithScratch(head, &scratch, nil); err == nil {
			t.Fatal("reentrant enumeration was accepted")
		}
		return nil
	}); err != nil {
		t.Fatal("failed enumeration poisoned scratch:", err)
	}
}

func TestDerivationScratchRetentionDropsCompleteStorage(t *testing.T) {
	var scratch DerivationScratch
	f := scratch.frame(0)
	f.payloads = [][]SubtreeID{make([]SubtreeID, 0, 256*1024/4)}
	if scratch.RetainedBytes() <= 256*1024 {
		t.Fatal("footprint omitted frame headers or inactive payload capacity")
	}
	scratch.Reset()
	if scratch.RetainedBytes() != 0 || len(scratch.frames) != 0 {
		t.Fatal("oversized storage survived reset")
	}
}
