package parsercorephase0

import "testing"

type acceptancePublicationTestValue struct{ releases int }

func (v *acceptancePublicationTestValue) Release() { v.releases++ }

func TestAcceptanceMaterializationTransfersOnlyExactSelection(t *testing.T) {
	for _, count := range []int{1, 5, AcceptanceMaterializationPayloadLimit} {
		compact := newTinyCore(t, 4)
		head, err := compact.Seed(1, 0)
		if err != nil {
			t.Fatal(err)
		}
		payloads := make([]SubtreeID, count)
		for i := range payloads {
			payloads[i] = SubtreeID(i + 1)
		}
		original := append([]SubtreeID(nil), payloads...)
		value := &acceptancePublicationTestValue{}
		var slot AcceptanceMaterialization[*acceptancePublicationTestValue]
		if !slot.Store(compact, head, payloads, 7, value) || slot.Store(compact, head, payloads, 7, value) {
			t.Fatal("slot did not enforce single ownership")
		}
		payloads[0]++
		for _, key := range []struct {
			compact  *Core
			head     Head
			payloads []SubtreeID
			context  uint64
		}{
			{newTinyCore(t, 4), head, original, 7},
			{compact, Head{Node: head.Node + 1}, original, 7},
			{compact, head, payloads, 7},
			{compact, head, original, 8},
		} {
			if got, ok := slot.Take(key.compact, key.head, key.payloads, key.context); ok || got != nil {
				t.Fatal("mismatched selection acquired the publication")
			}
		}
		got, ok := slot.Take(compact, head, original, 7)
		if !ok || got != value || value.releases != 0 || slot.Retained() != nil {
			t.Fatal("exact selection did not transfer ownership")
		}
		slot.Discard()
		if _, ok := slot.Take(compact, head, original, 7); ok || value.releases != 0 {
			t.Fatal("transferred publication remained owned by slot")
		}
		got.Release()
	}
}

func TestAcceptanceMaterializationDoesNotRetainWideSelections(t *testing.T) {
	compact := newTinyCore(t, 4)
	value := &acceptancePublicationTestValue{}
	var slot AcceptanceMaterialization[*acceptancePublicationTestValue]
	if slot.Store(compact, Head{Node: 1}, make([]SubtreeID, AcceptanceMaterializationPayloadLimit+1), 0, value) ||
		slot.Retained() != nil || value.releases != 0 {
		t.Fatal("wide selection transferred ownership or allocated retained storage")
	}
}

func TestAcceptanceMaterializationInvalidatesResetAndFrontier(t *testing.T) {
	for _, reset := range []bool{false, true} {
		compact := newTinyCore(t, 4)
		head, err := compact.Seed(1, 0)
		if err != nil {
			t.Fatal(err)
		}
		value := &acceptancePublicationTestValue{}
		var slot AcceptanceMaterialization[*acceptancePublicationTestValue]
		if !slot.Store(compact, head, []SubtreeID{1}, 0, value) {
			t.Fatal("store failed")
		}
		if reset {
			err = compact.Reset()
		} else {
			err = compact.BeginFrontier()
		}
		if err != nil {
			t.Fatal(err)
		}
		if got, ok := slot.Take(compact, head, []SubtreeID{1}, 0); ok || got != nil {
			t.Fatal("stale generation acquired a publication")
		}
		slot.Discard()
		slot.Discard()
		if value.releases != 1 || slot.compact != nil || slot.payloads != nil {
			t.Fatal("discard leaked or released twice")
		}
	}
}
